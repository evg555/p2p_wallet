package infra

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheWithoutTTLDoesNotExpire(t *testing.T) {
	c := NewCache()
	c.Set("test", "value", 0)

	time.Sleep(20 * time.Millisecond)

	got, err := c.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, "value", got)
}

func TestCacheDelete(t *testing.T) {
	c := NewCache()
	c.Set("test", "value", 10*time.Millisecond)

	val, err := c.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, val, "value")

	c.Delete("test")
	val, err = c.Get("test")
	assert.NoError(t, err)
	assert.Nil(t, val)
}
