package infra

import (
	"sync"
	"time"
)

type cache struct {
	mu   sync.RWMutex
	data map[string]entry
}

type entry struct {
	value     any
	expiresAt time.Time
}

func NewCache() *cache {
	return &cache{
		data: make(map[string]entry),
	}
}

func (c *cache) Get(key string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.data[key]
	if !ok {
		return nil, nil
	}

	if v.isExpired() {
		delete(c.data, key)
		return nil, nil
	}

	return v.value, nil
}

func (c *cache) Set(key string, v any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := entry{value: v}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}

	c.data[key] = e
}

func (c *cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

func (e entry) isExpired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}
