package domain

import (
	"testing"

	"p2p_wallet/internal/errs"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Run("passwords are similar", func(t *testing.T) {
		login := "test"
		password := "123456"
		name := "alex"
		lastName := "joe"

		user, err := NewUser(login, password, name, lastName)
		assert.Nil(t, err)
		assert.False(t, password == user.Password)
		assert.True(t, user.CheckPassword(password))
	})

	t.Run("login is incorrect", func(t *testing.T) {
		login := ""
		password := "123456"
		name := "alex"
		lastName := "joe"

		_, err := NewUser(login, password, name, lastName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrEmptyField)
	})

	t.Run("password is incorrect", func(t *testing.T) {
		login := "test"
		password := ""
		name := "alex"
		lastName := "joe"

		_, err := NewUser(login, password, name, lastName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrEmptyField)
	})

	t.Run("first name is incorrect", func(t *testing.T) {
		login := "test"
		password := "123456"
		name := ""
		lastName := "joe"

		_, err := NewUser(login, password, name, lastName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrEmptyField)

		arr := make([]byte, 0, 100)
		for i := 0; i < 100; i++ {
			arr = append(arr, byte('a'))
		}

		name = string(arr)

		_, err = NewUser(login, password, name, lastName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrFieldTooMuch)
	})

	t.Run("last name is incorrect", func(t *testing.T) {
		login := "test"
		password := "123456"
		name := "alex"
		lastName := ""

		_, err := NewUser(login, password, name, lastName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrEmptyField)

		arr := make([]byte, 0, 100)
		for i := 0; i < 100; i++ {
			arr = append(arr, byte(i))
		}

		lastName = string(arr)

		_, err = NewUser(login, password, name, lastName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrFieldTooMuch)
	})
}
