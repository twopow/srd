package server

import (
	"errors"
	"net/http"
	"net/url"

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

		to, err := url.Parse(value.To)
		if err != nil {
			l.Err(err).Msg("failed to parse to url")

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if to.Scheme == "" {
			to.Scheme = "http"
		}

		if value.PreserveRoute {
			to.Path = r.URL.Path
			to.RawQuery = r.URL.RawQuery
		}

		if value.Code == 0 {
			value.Code = http.StatusFound
		}

		// note, the full destination url is not logged
		// because it may contain sensitive information

		http.Redirect(w, r, to.String(), value.Code)
	}
}
