package handlers

import (
	"context"
	"net/http"

	"github.com/twopow/srd/internal/log"
	"github.com/twopow/srd/internal/util"
	"github.com/twopow/srd/resolver"
)

func CaddyHelperHandler(resv resolver.ResolverProvider) http.HandlerFunc {
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

		ctx := context.Background()
		value, err := resv.Resolve(ctx, domain)
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
