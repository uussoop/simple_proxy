package cache

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

var keyLocks = make(map[string]*sync.Mutex)
var keyLocksMutex sync.Mutex

// Cache is a wrapper for go-cache
type Cache struct {
	*cache.Cache
}

func GetCache() *Cache {
	return &Cache{
		Cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

// Set sets a value in the cache
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	c.Cache.Set(key, value, expiration)
}

// Get gets a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	return c.Cache.Get(key)
}

// Delete deletes a value from the cache
func (c *Cache) Delete(key string) {
	c.Cache.Delete(key)
}

// Clear clears the cache
func (c *Cache) Clear() {
	c.Cache.Flush()
}

// GetOrSet gets a value from the cache or sets it if it doesn't exist
func (c *Cache) GetOrSet(key string, value interface{}, expiration time.Duration) interface{} {
	if v, ok := c.Get(key); ok {
		return v
	}

	c.Set(key, value, expiration)

	return value
}

func GetLockForKey(key string) *sync.Mutex {
	keyLocksMutex.Lock()
	defer keyLocksMutex.Unlock()

	if lock, ok := keyLocks[key]; ok {
		return lock
	}

	lock := &sync.Mutex{}
	keyLocks[key] = lock
	return lock
}
