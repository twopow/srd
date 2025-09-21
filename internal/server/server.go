package server

import (
	"fmt"
	"net/http"

	"srd/internal/log"
	"srd/internal/resolver"
)

type ServerConfig struct {
	Host        string            `help:"Host for the HTTP server." default:"localhost"`
	Port        int               `help:"Port for the HTTP server." default:"8080"`
	CaddyHelper CaddyHelperConfig `help:"Caddy helper server configuration." embed:"" prefix:"caddyhelper."`
}

type CaddyHelperConfig struct {
	Enabled bool   `help:"Enable Caddy helper server." default:"false"`
	Host    string `help:"Host for the Caddy helper server." default:"localhost"`
	Port    int    `help:"Port for the Caddy helper server." default:"8081"`
}

func Start(cfg ServerConfig) error {
	// Start both servers concurrently
	go func() {
		if err := startServer(cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to start main server")
		}
	}()

	if cfg.CaddyHelper.Enabled {
		go func() {
			if err := startCaddyHelper(cfg.CaddyHelper); err != nil {
				log.Fatal().Err(err).Msg("failed to start caddy helper server")
			}
		}()
	}

	// Keep the main goroutine alive
	select {}
}

// StartServer starts an HTTP server on the specified host and port
func startServer(cfg ServerConfig) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Info().With("addr", addr).Msg("booting server")

	http.HandleFunc("/", ResolveHandler(resolver.DefaultResolver))

	return http.ListenAndServe(addr, nil)
}

// StartCaddyHelper starts a Caddy helper server on the specified host and port
// CadddyHelper is used to check if a cert should be issued for a given hostname
// ref: https://caddyserver.com/docs/caddyfile/options#on-demand-tls
func startCaddyHelper(cfg CaddyHelperConfig) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Info().With("addr", addr).Msg("booting caddy helper")

	http.HandleFunc("/ask", CaddyHelperHandler(resolver.DefaultResolver))

	return http.ListenAndServe(addr, nil)
}
