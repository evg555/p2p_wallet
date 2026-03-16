package cache

import (
	"sync"
	"time"
)

type List interface {
	Len() int
	Back() any
	PushFront(el any) any
	Remove(el any)
	MoveToFront(el any)
}

type lruCache struct {
	mu       sync.RWMutex
	capacity int
	queue    List
	data     map[string]*entry
}

type entry struct {
	key       string
	value     any
	expiresAt time.Time
}

func NewCache(size int) *lruCache {
	return &lruCache{
		capacity: size,
		queue:    NewList(),
		data:     make(map[string]*entry),
	}
}

func (c *lruCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.data[key]; ok {
		if item.isExpired() {
			c.queue.Remove(item)
			delete(c.data, key)
			return nil, false
		}

		c.queue.MoveToFront(item)
		return item.value, true
	}

	return nil, false
}

func (c *lruCache) Set(key string, v any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.data[key]; ok {
		item.value = v
		c.queue.MoveToFront(item)
		return true
	}

	if c.queue.Len() == c.capacity {
		back := c.queue.Back()
		if backItem, ok := back.(*entry); ok {
			c.queue.Remove(backItem)
			delete(c.data, backItem.key)
		}
	}

	e := &entry{key: key, value: v}
	c.queue.PushFront(e)
	c.data[key] = e

	return false
}

func (c *lruCache) SetWithTTL(key string, v any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := &entry{key: key, value: v}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}

	if item, ok := c.data[key]; ok {
		item.value = v
		item.expiresAt = e.expiresAt
		c.queue.MoveToFront(item)
		return
	}

	if c.queue.Len() == c.capacity {
		back := c.queue.Back()
		if backItem, ok := back.(*entry); ok {
			c.queue.Remove(backItem)
			delete(c.data, backItem.key)
		}
	}

	c.queue.PushFront(e)
	c.data[key] = e
}

func (c *lruCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.data[key]
	if !ok {
		return
	}

	c.queue.Remove(item)
	delete(c.data, key)
}

func (l *lruCache) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.data = make(map[string]*entry, l.capacity)
	l.queue = NewList()
}

func (e *entry) isExpired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}
