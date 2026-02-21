package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/twopow/srd/handlers"
	"github.com/twopow/srd/resolver"
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

var log *slog.Logger

func Start(cfg ServerConfig, rp resolver.ResolverProvider, logger *slog.Logger) error {
	log = logger

	// Start both servers concurrently
	go func() {
		if err := startServer(cfg, rp); err != nil {
			log.Error("failed to start main server", "error", err)
			os.Exit(1)
		}
	}()

	if cfg.CaddyHelper.Enabled {
		go func() {
			if err := startCaddyHelper(cfg.CaddyHelper, rp); err != nil {
				log.Error("failed to start caddy helper server", "error", err)
				os.Exit(1)
			}
		}()
	}

	// Keep the main goroutine alive
	select {}
}

// StartServer starts an HTTP server on the specified host and port
func startServer(cfg ServerConfig, rp resolver.ResolverProvider) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Info("booting server", "addr", addr)

	http.HandleFunc("/", handlers.ResolveHandler(rp))

	return http.ListenAndServe(addr, nil)
}

// StartCaddyHelper starts a Caddy helper server on the specified host and port
// CadddyHelper is used to check if a cert should be issued for a given hostname
// ref: https://caddyserver.com/docs/caddyfile/options#on-demand-tls
func startCaddyHelper(cfg CaddyHelperConfig, rp resolver.ResolverProvider) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Info("booting caddy helper", "addr", addr)

	http.HandleFunc("/ask", handlers.CaddyHelperHandler(rp))

	return http.ListenAndServe(addr, nil)
}
