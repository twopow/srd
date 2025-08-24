package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"srd/internal/cache"
	"srd/internal/config"
	"srd/internal/resolver"
)

type TestData struct {
	Hostname       string
	Path           string
	ExpectedBody   string
	ExpectedStatus int
	ExpectedTo     string
}

func doResolverTest(t *testing.T, test TestData) {
	req, err := http.NewRequest("GET", test.Path, nil)
	if err != nil {
		t.Fatal(err)
	}

	resolver.Init(config.ResolverConfig{}, cache.Mock())

	req.Host = test.Hostname
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ResolveHandler(resolver.Mock()))
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

	if test.ExpectedTo != "" {
		actual := rr.Header().Get("Location")
		if actual != test.ExpectedTo {
			t.Fatalf("handler returned unexpected location: got %v want %v",
				actual, test.ExpectedTo)
		}
	}

	if rr.Header().Get("x-request-id") == "" {
		t.Fatalf("handler returned no x-request-id")
	}
}

func TestResolveHandler_Success(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success.test",
		Path:           "/",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "http://to.test",
	})
}

func TestResolveHandler_Success_Path(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success.test",
		Path:           "/path",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "http://to.test/path",
	})
}
func TestResolveHandler_Success_Query(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success.test",
		Path:           "/?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "http://to.test?key=value",
	})
}

func TestResolveHandler_Success_Path_Query(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success.test",
		Path:           "/path?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "http://to.test/path?key=value",
	})
}

func TestResolveHandler_NotFound(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "not-found.test",
		Path:           "/",
		ExpectedBody:   `Not found`,
		ExpectedStatus: http.StatusNotFound,
	})
}

func TestResolveHandler_Error(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "error.test",
		Path:           "/",
		ExpectedBody:   `error`,
		ExpectedStatus: http.StatusInternalServerError,
	})
}
