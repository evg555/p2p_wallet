package infra

import (
	"sync"
	"time"
)

type cache struct {
	mu   sync.RWMutex
	data map[int64]entry
}

type entry struct {
	value     any
	expiresAt time.Time
}

func NewCache() *cache {
	return &cache{
		data: make(map[int64]entry),
	}
}

func (c *cache) Get(id int64) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.data[id]
	if !ok {
		return nil, nil
	}

	if v.isExpired() {
		delete(c.data, id)
		return nil, nil
	}

	return v.value, nil
}

func (c *cache) GetAll() ([]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := make([]any, 0, len(c.data))
	for id, v := range c.data {
		if v.isExpired() {
			delete(c.data, id)
			continue
		}

		data = append(data, v.value)
	}

	return data, nil
}

func (c *cache) Set(id int64, v any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := entry{value: v}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}

	c.data[id] = e
}

func (e entry) isExpired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}
