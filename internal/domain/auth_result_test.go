package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSessionEmpty(t *testing.T) {
	s := SessionID("")
	assert.True(t, s.IsEmpty())
}

func TestIsSessionsSame(t *testing.T) {
	s := SessionID("test")
	assert.True(t, s.Equal("test"))
}
