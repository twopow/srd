package cache

import (
	"srd/internal/config"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cfg := config.CacheConfig{
		TTL:             time.Second * 5,
		CleanupInterval: time.Second * 10,
	}

	cache := New(cfg)

	tests := []struct {
		name     string
		key      string
		value    interface{}
		wantGet  interface{}
		wantFind bool
	}{
		{
			name:     "string value",
			key:      "test-key",
			value:    "test-value",
			wantGet:  "test-value",
			wantFind: true,
		},
		{
			name:     "integer value",
			key:      "number",
			value:    42,
			wantGet:  42,
			wantFind: true,
		},
		{
			name:     "struct value",
			key:      "struct",
			value:    struct{ Name string }{Name: "test"},
			wantGet:  struct{ Name string }{Name: "test"},
			wantFind: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.Set(tt.key, tt.value)
			got, found := cache.Get(tt.key)

			if found != tt.wantFind {
				t.Errorf("Cache.Get() found = %v, want %v", found, tt.wantFind)
			}
			if got != tt.wantGet {
				t.Errorf("Cache.Get() = %v, want %v", got, tt.wantGet)
			}
		})
	}
}

func TestCache_GetExpiredCleanup(t *testing.T) {
	cfg := config.CacheConfig{
		TTL:             time.Millisecond * 100, // Very short TTL for testing
		CleanupInterval: time.Second * 10,
	}

	cache := New(cfg)

	// Set a value
	cache.Set("test-key", "test-value")

	// Wait for the TTL to expire
	time.Sleep(time.Millisecond * 150)

	// Try to get the expired value
	_, found := cache.Get("test-key")
	if found {
		t.Error("Cache.Get() found expired value, want not found")
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	cfg := config.CacheConfig{
		TTL:             time.Second * 5,
		CleanupInterval: time.Second * 10,
	}

	cache := New(cfg)

	_, found := cache.Get("non-existent-key")
	if found {
		t.Error("Cache.Get() found non-existent value, want not found")
	}
}
