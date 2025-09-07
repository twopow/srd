package server

import (
	"net/http"
	"net/url"
	"strings"

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

		// ensure destination url has a scheme
		// so url.Parse can parse the url appropriately
		if !strings.Contains(value.To, "://") {
			value.To = "http://" + value.To
		}

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
