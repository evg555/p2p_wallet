package infra

import (
	"sync"
)

type cache struct {
	mu   sync.RWMutex
	data map[int64]any
}

func NewCache() *cache {
	return &cache{
		data: make(map[int64]any),
	}
}

func (c *cache) Get(id int64) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[id]
	if !ok {
		return nil, nil
	}

	return v, nil
}

func (c *cache) GetAll() ([]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data := make([]any, 0, len(c.data))
	for _, v := range c.data {
		data = append(data, v)
	}

	return data, nil
}

func (c *cache) Set(id int64, v any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[id] = v
}
