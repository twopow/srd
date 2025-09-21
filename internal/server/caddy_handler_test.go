package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"srd/internal/cache"
	"srd/internal/resolver"
)

type CaddyHandlerTestData struct {
	Path           string
	ExpectedStatus int
	ExpectedBody   string
}

func doCaddyHandlerTest(t *testing.T, test CaddyHandlerTestData) {
	req, err := http.NewRequest("GET", test.Path, nil)
	if err != nil {
		t.Fatal(err)
	}

	resolver.Init(resolver.ResolverConfig{}, cache.Mock())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CaddyHelperHandler(resolver.Mock()))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != test.ExpectedStatus {
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, test.ExpectedStatus)
	}

	if test.ExpectedBody != "" {
		actual := strings.TrimSpace(rr.Body.String())
		if actual != test.ExpectedBody {
			t.Fatalf("handler returned unexpected body: got %v want %v",
				actual, test.ExpectedBody)
		}
	}
}

func TestCaddyHandler_Success(t *testing.T) {
	doCaddyHandlerTest(t, CaddyHandlerTestData{
		Path:           "/ask?domain=success.test",
		ExpectedStatus: http.StatusOK,
		ExpectedBody:   "ok",
	})
}

func TestCaddyHandler_NotFound(t *testing.T) {
	doCaddyHandlerTest(t, CaddyHandlerTestData{
		Path:           "/ask?domain=not-found.test",
		ExpectedStatus: http.StatusBadRequest,
		ExpectedBody:   "rejected",
	})
}

func TestCaddyHandler_Ip(t *testing.T) {
	doCaddyHandlerTest(t, CaddyHandlerTestData{
		Path:           "/ask?domain=127.0.0.1",
		ExpectedStatus: http.StatusBadRequest,
		ExpectedBody:   "ip address not allowed",
	})
}

func TestCaddyHandler_IpPort(t *testing.T) {
	doCaddyHandlerTest(t, CaddyHandlerTestData{
		Path:           "/ask?domain=127.0.0.1:8888",
		ExpectedStatus: http.StatusBadRequest,
		ExpectedBody:   "ip address not allowed",
	})
}
