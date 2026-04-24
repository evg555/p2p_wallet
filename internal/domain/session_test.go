package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsSessionsSame(t *testing.T) {
	s := NewSession(UserID(1), time.Hour)
	expected := string(s.SessionID())

	assert.True(t, s.Equal(expected))
}

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
