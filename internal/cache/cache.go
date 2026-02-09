package cache

import (
	"sync"
	"time"

	"github.com/twopow/srd/internal/log"
)

type CacheConfig struct {
	// TTL is the cache TTL
	TTL time.Duration

	// CleanupInterval is how often to cleanup the cache
	CleanupInterval time.Duration
}

var DefaultCacheConfig = CacheConfig{
	TTL:             time.Second * 300, // 5 minutes
	CleanupInterval: time.Second * 900, // 15 minutes
}

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
	config CacheConfig
}

// New creates a new Cache instance
func New(cfg CacheConfig) (CacheProvider, error) {
	c := &Cache{
		items:  make(map[string]item),
		config: cfg,
	}

	// Start cleanup goroutine
	go c.cleanupTimer()

	return c, nil
}

// Get retrieves a value from the cache by key
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// bail early if the item has already expired
	if time.Now().After(item.expiration) {
		return nil, false
	}

	// upgrade to a write lock only when we need to bump the ttl
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists = c.items[key]
	if !exists {
		return nil, false
	}

	// item might have been updated or expired while waiting for the write lock
	if time.Now().After(item.expiration) {
		return nil, false
	}

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
		log.Info().With("deleted", deleted).Msg("cache cleanup")
	}
}
