package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodePassword(t *testing.T) {
	t.Run("encode and check password", func(t *testing.T) {
		password := "123456"

		hash := EncodePassword(password)
		assert.True(t, CheckPassword(password, hash))
	})
}
