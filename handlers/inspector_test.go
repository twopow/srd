package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twopow/srd/resolver"
)

func doInspectTest(t *testing.T, query string, check func(t *testing.T, code int, resp InspectResponse)) {
	t.Helper()

	req, err := http.NewRequest("GET", "/inspect?"+query, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = HandleInspect(context.Background(), rr, req, resolver.Mock())
	if err != nil {
		t.Fatalf("HandleInspect returned error: %v", err)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}

	var resp InspectResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	check(t, rr.Code, resp)
}

func TestInspect_MissingHost(t *testing.T) {
	doInspectTest(t, "", func(t *testing.T, code int, resp InspectResponse) {
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", code)
		}
		if resp.Error == "" {
			t.Fatal("expected error message")
		}
	})
}

func TestInspect_Success(t *testing.T) {
	doInspectTest(t, "host=success.test", func(t *testing.T, code int, resp InspectResponse) {
		if code != http.StatusOK {
			t.Fatalf("expected 200, got %d", code)
		}
		if resp.Host != "success.test" {
			t.Fatalf("expected host success.test, got %s", resp.Host)
		}
		if resp.Destination != "to.test" {
			t.Fatalf("expected destination to.test, got %s", resp.Destination)
		}
		if resp.Code != http.StatusFound {
			t.Fatalf("expected code 302, got %d", resp.Code)
		}
		if resp.NotFound {
			t.Fatal("expected not_found to be false")
		}
		if resp.Loop {
			t.Fatal("expected loop to be false")
		}
	})
}

func TestInspect_PreserveRoute(t *testing.T) {
	doInspectTest(t, "host=success-preserve-path.test", func(t *testing.T, code int, resp InspectResponse) {
		if code != http.StatusOK {
			t.Fatalf("expected 200, got %d", code)
		}
		if !resp.PreserveRoute {
			t.Fatal("expected preserve_route to be true")
		}
	})
}

func TestInspect_RefererPolicy(t *testing.T) {
	doInspectTest(t, "host=success-referer-policy-full.test", func(t *testing.T, code int, resp InspectResponse) {
		if code != http.StatusOK {
			t.Fatalf("expected 200, got %d", code)
		}
		if resp.RefererPolicy != "full" {
			t.Fatalf("expected referer_policy full, got %s", resp.RefererPolicy)
		}
	})
}

func TestInspect_NotFound(t *testing.T) {
	doInspectTest(t, "host=not-found.test", func(t *testing.T, code int, resp InspectResponse) {
		if code != http.StatusOK {
			t.Fatalf("expected 200, got %d", code)
		}
		if !resp.NotFound {
			t.Fatal("expected not_found to be true")
		}
		if resp.Destination != "" {
			t.Fatalf("expected empty destination, got %s", resp.Destination)
		}
	})
}

func TestInspect_Loop(t *testing.T) {
	doInspectTest(t, "host=loop.test", func(t *testing.T, code int, resp InspectResponse) {
		if code != http.StatusOK {
			t.Fatalf("expected 200, got %d", code)
		}
		if !resp.Loop {
			t.Fatal("expected loop to be true")
		}
		if resp.Error != "" {
			t.Fatalf("loop should not set error, got %s", resp.Error)
		}
	})
}

func TestInspect_ResolveError(t *testing.T) {
	doInspectTest(t, "host=error.test", func(t *testing.T, code int, resp InspectResponse) {
		if code != http.StatusOK {
			t.Fatalf("expected 200, got %d", code)
		}
		if resp.Error == "" {
			t.Fatal("expected error message")
		}
		if resp.Loop {
			t.Fatal("expected loop to be false")
		}
	})
}

