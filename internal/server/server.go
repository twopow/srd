package server

import (
	"fmt"
	"net/http"

	"srd/internal/log"
	"srd/internal/resolver"
)

// Response represents the JSON response structure
type Response struct {
	To string `json:"to"`
}

// StartServer starts an HTTP server on the specified host and port
func StartServer(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Info().With("addr", addr).Msg("booting server")

	http.HandleFunc("/", ResolveHandler(resolver.DefaultResolver))

	return http.ListenAndServe(addr, nil)
}
