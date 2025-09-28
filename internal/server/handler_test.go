package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"srd/internal/cache"
	"srd/internal/resolver"
)

type TestData struct {
	Hostname        string
	Path            string
	ExpectedBody    string
	ExpectedStatus  int
	ExpectedTo      string
	ExpectedHeaders map[string]string
}

func doResolverTest(t *testing.T, test TestData) {
	req, err := http.NewRequest("GET", test.Path, nil)
	if err != nil {
		t.Fatal(err)
	}

	resolver.Init(resolver.ResolverConfig{}, cache.Mock())

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

	for key, value := range test.ExpectedHeaders {
		actual := rr.Header().Get(key)
		if actual != value {
			t.Fatalf("handler returned unexpected header: got %v want %v",
				actual, value)
		}
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
		Hostname:       "success-url.test",
		Path:           "/otherpath?otherquery=string",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "https://to.test/path?query=string",
	})
}
func TestResolveHandler_Success_Query(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success.test",
		Path:           "/?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "http://to.test",
	})
}

func TestResolveHandler_Success_Path_Query(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success.test",
		Path:           "/path?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "http://to.test",
	})
}

func TestResolveHandler_Success_Preserve_Path_Query(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success-preserve-path.test",
		Path:           "/path?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "https://to.test/path?key=value",
	})
}

func TestResolveHandler_Success_Preserve_Path_Query_No_Scheme(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success-preserve-path-no-scheme.test",
		Path:           "/path?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "http://to.test/path?key=value",
	})
}

func TestResolveHandler_NotFound(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "not-found.test",
		Path:           "/",
		ExpectedBody:   "Not found",
		ExpectedStatus: http.StatusNotFound,
	})
}

func TestResolveHandler_InvalidToUrl(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "invalid-to-url.test",
		Path:           "/",
		ExpectedBody:   "Not found",
		ExpectedStatus: http.StatusNotFound,
	})
}

func TestResolveHandler_Error(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "error.test",
		Path:           "/",
		ExpectedBody:   "internal server error",
		ExpectedStatus: http.StatusInternalServerError,
	})
}

func TestResolveHandler_NoHostBaseRedirect(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "127.0.0.1:8080",
		Path:           "/",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "https://github.com/twopow/srd",
	})
}

func TestResolveHandler_RefererPolicy_DefaultIsHost(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success-referer-policy-default.test",
		Path:           "/route?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "https://to.test/path?query=string",
		ExpectedHeaders: map[string]string{
			"Referer": "success-referer-policy-default.test",
		},
	})
}

func TestResolveHandler_RefererPolicy_None(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success-referer-policy-none.test",
		Path:           "/route?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "https://to.test/path?query=string",
		ExpectedHeaders: map[string]string{
			"Referer": "",
		},
	})
}

func TestResolveHandler_RefererPolicy_Host(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success-referer-policy-host.test",
		Path:           "/route?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "https://to.test/path?query=string",
		ExpectedHeaders: map[string]string{
			"Referer": "success-referer-policy-host.test",
		},
	})
}

func TestResolveHandler_RefererPolicy_Full(t *testing.T) {
	doResolverTest(t, TestData{
		Hostname:       "success-referer-policy-full.test",
		Path:           "/route?key=value",
		ExpectedStatus: http.StatusFound,
		ExpectedTo:     "https://to.test/path?query=string",
		ExpectedHeaders: map[string]string{
			"Referer": "success-referer-policy-full.test/route?key=value",
		},
	})
}
