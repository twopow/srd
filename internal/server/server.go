package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"srd/internal/resolver"
	"srd/internal/util"
)

// Response represents the JSON response structure
type Response struct {
	To string `json:"to"`
}

// StartServer starts an HTTP server on the specified host and port
func StartServer(host string, port int) error {
	// Define a handler function for all routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

		// Create response object
		response := Response{
			To: value.To,
		}

		log.Info().
			Str("request", rid).
			Str("from", r.Host).
			Str("to", response.To).
			Msg("handled request")

		// Encode response as JSON
		json.NewEncoder(w).Encode(response)
	})

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Info().Str("addr", addr).Msg("booting server")

	return http.ListenAndServe(addr, nil)
}
