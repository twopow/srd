package server

import (
	"net/http"

	"srd/internal/log"
	resolverP "srd/internal/resolver"
	"srd/internal/util"
)

func CaddyHelperHandler(resolver resolverP.ResolverProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		if domain == "" {
			log.Error().Msg("caddy domain check: domain is required")
			http.Error(w, "domain is required", http.StatusBadRequest)
			return
		}

		l := log.With("domain", domain)

		if util.IsIp(domain) {
			l.Debug().Msg("caddy domain check: ip address not allowed")
			http.Error(w, "ip address not allowed", http.StatusBadRequest)
			return
		}

		value, err := resolver.Resolve(domain)
		if err == nil && !value.NotFound {
			l.Debug().Msg("caddy domain check: ok")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
			return
		}

		l.Debug().Msg("caddy domain check: rejected")
		http.Error(w, "rejected", http.StatusBadRequest)
	}
}
