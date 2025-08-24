package cache

import (
	"sync"
	"time"

	"srd/internal/config"
	"srd/internal/log"
)

type CacheProvider interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Cleanup()
}

type item struct {
	value      interface{}
	expiration time.Time
}

type Cache struct {
	items  map[string]item
	mu     sync.RWMutex
	config config.CacheConfig
}

// New creates a new Cache instance
func New(cfg config.CacheConfig) CacheProvider {
	c := &Cache{
		items:  make(map[string]item),
		config: cfg,
	}

	// Start cleanup goroutine
	go c.cleanupTimer()

	return c
}

// Get retrieves a value from the cache by key
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.expiration) {
		return nil, false
	}

	// reset the expiration time
	item.expiration = time.Now().Add(c.config.TTL)
	c.items[key] = item

	return item.value, true
}

// Set stores a value in the cache with the specified key
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		value:      value,
		expiration: time.Now().Add(c.config.TTL),
	}
}

// cleanup periodically removes expired items from the cache
func (c *Cache) cleanupTimer() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.Cleanup()
	}
}

// Cleanup removes expired items from the cache
func (c *Cache) Cleanup() {
	deleted := 0

	c.mu.Lock()
	for key, item := range c.items {
		if time.Now().After(item.expiration) {
			delete(c.items, key)
			deleted++
		}
	}
	c.mu.Unlock()

	if deleted > 0 {
		log.Info().Int("deleted", deleted).Msg("cache cleanup")
	}
}
