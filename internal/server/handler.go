package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"srd/internal/log"
	resolverP "srd/internal/resolver"
	"srd/internal/util"
)

// Define a handler function for all routes
func ResolveHandler(resolver resolverP.ResolverProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := util.UUID7().String()
		w.Header().Set("x-request-id", rid)

		value, err := resolver.Resolve(r.Host)
		if err != nil {
			if errors.Is(err, resolverP.ErrLoop) {
				http.Error(w, "loop detected", http.StatusBadRequest)
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
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	referer := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.RequestURI())
	w.Header().Set("Referer", referer)
}
