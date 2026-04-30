package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsSessionsExpired(t *testing.T) {
	now := time.Now()

	s := Session{
		ExpiresAt: now.Add(-time.Hour),
	}
	assert.True(t, s.IsExpired(now))

	s = Session{
		ExpiresAt: now.Add(time.Hour),
	}
	assert.False(t, s.IsExpired(now))
}
