package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	resolverP "github.com/twopow/srd/resolver"
)

type InspectResponse struct {
	Host          string `json:"host"`
	Destination   string `json:"destination,omitempty"`
	Code          int    `json:"code,omitempty"`
	PreserveRoute bool   `json:"preserve_route,omitempty"`
	RefererPolicy string `json:"referer_policy,omitempty"`
	NotFound      bool   `json:"not_found,omitempty"`
	Loop          bool   `json:"loop,omitempty"`
	Error         string `json:"error,omitempty"`
}

func HandleInspect(ctx context.Context, w http.ResponseWriter, r *http.Request, resolver resolverP.ResolverProvider) error {
	host := r.URL.Query().Get("host")
	if host == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return json.NewEncoder(w).Encode(InspectResponse{Error: "missing required query parameter: host"})
	}

	rr, err := resolver.Resolve(ctx, host)

	resp := InspectResponse{
		Host:     host,
		NotFound: rr.NotFound,
	}

	if err != nil {
		if errors.Is(err, resolverP.ErrLoop) {
			resp.Loop = true
		} else {
			resp.Error = err.Error()
		}
	}

	if !rr.NotFound {
		resp.Destination = rr.To
		resp.Code = rr.Code
		resp.PreserveRoute = rr.PreserveRoute
		resp.RefererPolicy = rr.RefererPolicy.String()
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(resp)
}
