package server

import (
	"net/http"
	"net/url"

	"srd/internal/log"
	"srd/internal/resolver"
	"srd/internal/util"
)

// Define a handler function for all routes
func ResolveHandler(resolver resolver.ResolverProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := util.UUID7().String()
		w.Header().Set("x-request-id", rid)

		value, err := resolver.Resolve(r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		l := log.Info().
			Str("request", rid).
			Str("from", r.Host).
			Str("to", value.To)

		if value.NotFound {
			l.Msg("not found")
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		l.Msg("redirecting")

		// parse value.To to get scheme
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

		http.Redirect(w, r, to.String(), value.Code)
	}
}
