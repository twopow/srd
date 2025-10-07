package cache

type MockCache struct {
	items map[string]interface{}
}

// Mock creates a new MockCache instance
func Mock() CacheProvider {
	return &MockCache{
		items: make(map[string]interface{}),
	}
}

func (c *MockCache) Get(key string) (interface{}, bool) {
	return c.items[key], true
}

func (c *MockCache) Set(key string, value interface{}) {
	c.items[key] = value
}

func (c *MockCache) Cleanup() {
	c.items = make(map[string]interface{})
}
