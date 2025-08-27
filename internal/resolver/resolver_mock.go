package resolver

import "fmt"

type MockResolver struct{}

var MockData = map[string]RR{
	"success": {
		Hostname: "success.test",
		To:       "to.test",
		NotFound: false,
	},
	"success-url": {
		Hostname: "success-url.test",
		To:       "https://to.test/path?query=string",
		NotFound: false,
	},
	"invalid-to-url": {
		Hostname: "invalid-to-url.test",
		NotFound: true,
	},
	"not-found": {
		Hostname: "not-found.test",
		NotFound: true,
	},
	"ip-port": {
		Hostname: "127.0.0.1:8080",
		To:       "https://github.com/twopow/srd",
		NotFound: false,
	},
}

var MockErrorHost = "error.test"

func Mock() ResolverProvider {
	return &MockResolver{}
}

func (r *MockResolver) Resolve(hostname string) (RR, error) {
	if hostname == MockErrorHost {
		return RR{}, fmt.Errorf("error")
	}

	for _, rr := range MockData {
		if rr.Hostname == hostname {
			return rr, nil
		}
	}

	return RR{}, fmt.Errorf("not found")
}
