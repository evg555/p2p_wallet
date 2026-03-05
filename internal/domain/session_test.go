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
