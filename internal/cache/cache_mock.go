package cache

type MockCache struct {
}

// Mock creates a new MockCache instance
func Mock() CacheProvider {
	return &MockCache{}
}

func (c *MockCache) Get(key string) (interface{}, bool) {
	return nil, false
}

func (c *MockCache) Set(key string, value interface{}) {
}

func (c *MockCache) Cleanup() {
}
