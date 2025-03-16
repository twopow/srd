package server

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"srd/internal/resolver"
	"srd/internal/util"
)

// Define a handler function for all routes
func ResolveHandler(resolver resolver.ResolverProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value, err := resolver.Resolve(r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if value.NotFound {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		rid := util.UUID7().String()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-request-id", rid)

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
