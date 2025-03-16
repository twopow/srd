package resolver

import "fmt"

type MockResolver struct{}

var MockData = map[string]RR{
	"success": {
		Hostname: "success.test",
		To:       "to.test",
		NotFound: false,
	},
	"not-found": {
		Hostname: "not-found.test",
		NotFound: true,
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
