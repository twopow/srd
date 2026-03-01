package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/twopow/srd/internal/util"
	resolverP "github.com/twopow/srd/resolver"
)

// Define a handler function for all routes
func ResolveHandler(resolver resolverP.ResolverProvider) http.HandlerFunc {
	log := resolver.Logger()

	return func(w http.ResponseWriter, r *http.Request) {
		timeout := time.Duration(2 * time.Second)
		ctx, cancel := context.WithTimeoutCause(r.Context(), timeout, errors.New("timeout"))
		defer cancel()

		rid := util.UUID7().String()
		w.Header().Set("x-request-id", rid)

		ctx = context.WithValue(ctx, resolverP.ResolverContextKey("request_id"), rid)

		if isInspectRequest(r, resolver.Config()) {
			if err := HandleInspect(ctx, w, r, resolver); err != nil {
				handleResolveError(w, r, resolver, fmt.Errorf("failed to inspect: %w", err))
			}
			return
		}

		if r.Host == "" || util.IsIp(r.Host) {
			handleResolveError(w, r, resolver, resolverP.ErrHostIsIp)
			return
		}

		value, err := resolver.Resolve(ctx, r.Host)
		if err != nil {
			handleResolveError(w, r, resolver, err)
			return
		}

		l := log.With("request", rid, "from", r.Host, "to", value.To)

		if value.NotFound {
			l.Info("not found")
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		l.Info("redirecting")

		to, err := constructTo(r, value)
		if err != nil {
			l.Error("failed to construct to", "error", err)
			handleResolveError(w, r, resolver, err)
			return
		}

		if value.Code == 0 {
			value.Code = http.StatusFound
		}

		setReferer(w, r, value)

		// note, the full urls are not logged
		// as they may contain sensitive information

		dest := to.String()
		http.Redirect(w, r, dest, value.Code)
	}
}

func handleResolveError(w http.ResponseWriter, r *http.Request, resolver resolverP.ResolverProvider, err error) {
	log := resolver.Logger().With("hostname", r.Host)

	if errors.Is(err, resolverP.ErrLoop) {
		toolboxHost := resolver.Config().ToolboxHost
		msg := "loop detected"

		if toolboxHost != "" {
			msg = fmt.Sprintf(
				"loop detected.\n\nInspect with toolbox: https://%s/inspect?r=%s",
				toolboxHost,
				url.QueryEscape(r.Host),
			)
		}

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if errors.Is(err, resolverP.ErrHostIsIp) {
		http.Redirect(w, r, resolver.Config().NoHostBaseRedirect, http.StatusFound)
		return
	}

	log.Error("resolve error", "error", err)

	// Context timeouts/cancellations are distinguishable:
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		http.Error(w, "timeout", http.StatusInternalServerError)
		return
	}

	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func isInspectRequest(r *http.Request, cfg *resolverP.ResolverConfig) bool {
	if r.URL.Path != "/inspect" {
		return false
	}

	if r.Host == "" || util.IsIp(r.Host) {
		return true
	}

	return cfg.InHost != "" && r.Host == cfg.InHost
}

func constructTo(r *http.Request, value resolverP.RR) (*url.URL, error) {
	// url.Parse expects a scheme
	if !strings.Contains(value.To, "://") {
		value.To = "http://" + value.To
	}

	to, err := url.Parse(value.To)
	if err != nil {
		return nil, fmt.Errorf("failed to parse to url: %w", err)
	}

	if value.PreserveRoute {
		to.Path = r.URL.Path
		to.RawQuery = r.URL.RawQuery
	}

	return to, nil
}

func setReferer(w http.ResponseWriter, r *http.Request, value resolverP.RR) {
	// none
	if value.RefererPolicy == resolverP.RefererPolicyNone {
		return
	}

	// host
	if value.RefererPolicy == resolverP.RefererPolicyHost {
		w.Header().Set("Referer", r.Host)
		return
	}

	// full
	referer := fmt.Sprintf("%s%s", r.Host, r.URL.RequestURI())
	w.Header().Set("Referer", referer)
}
