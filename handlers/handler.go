package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/twopow/srd/internal/log"
	"github.com/twopow/srd/internal/util"
	resolverP "github.com/twopow/srd/resolver"
)

// Define a handler function for all routes
func ResolveHandler(resolver resolverP.ResolverProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeout := time.Duration(2 * time.Second)
		ctx, cancel := context.WithTimeoutCause(r.Context(), timeout, errors.New("timeout"))
		defer cancel()

		rid := util.UUID7().String()
		w.Header().Set("x-request-id", rid)

		ctx = context.WithValue(ctx, resolverP.ResolverContextKey("request_id"), rid)

		value, err := resolver.Resolve(ctx, r.Host)
		if err != nil {
			if errors.Is(err, resolverP.ErrLoop) {
				http.Error(w, "loop detected", http.StatusBadRequest)
				return
			}

			// Context timeouts/cancellations are distinguishable:
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				http.Error(w, "timeout", http.StatusInternalServerError)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		l := log.Info().
			WithMap(map[string]any{
				"request": rid,
				"from":    r.Host,
				"to":      value.To,
			})

		if value.NotFound {
			l.Msg("not found")
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		l.Msg("redirecting")

		to, err := constructTo(r, value)
		if err != nil {
			l.Err(err).Msg("failed to construct to")
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
