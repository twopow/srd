package server

import (
	"net/http"

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

		if value.NotFound {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		log.Info().
			Str("request", rid).
			Str("from", r.Host).
			Str("to", value.To).
			Msg("handled request")

		scheme := r.URL.Scheme
		if scheme == "" {
			scheme = "http"
		}

		to := scheme + "://" + value.To

		if r.URL.Path != "/" {
			to = to + r.URL.Path
		}

		if r.URL.RawQuery != "" {
			to = to + "?" + r.URL.RawQuery
		}

		http.Redirect(w, r, to, http.StatusFound)
	}
}
