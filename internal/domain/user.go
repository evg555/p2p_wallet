package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"p2p_wallet/internal/errs"
)

var maxFieldLength = 50

type User struct {
	ID        int64      `json:"id"`
	Login     string     `json:"login"`
	Password  string     `json:"password"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func NewUser(login, password, firstName, lastName string) (*User, error) {
	if err := validate(login, password, firstName, lastName); err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	user := &User{
		Login:     login,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: time.Now(),
	}

	user.ID = rand.Int63()
	user.Password = encodePassword(password)

	return user, nil
}

func (u *User) CheckPassword(password string) bool {
	return u.Password == encodePassword(password)
}

func encodePassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func validate(login, password, firstName, lastName string) error {
	err := validateLogin(login)
	if err != nil {
		return err
	}

	err = validatePassword(password)
	if err != nil {
		return err
	}

	err = validateFirstName(firstName)
	if err != nil {
		return err
	}

	err = validateLastName(lastName)
	if err != nil {
		return err
	}

	return nil
}

func validateLogin(login string) error {
	if login == "" {
		return fmt.Errorf("invalid login: %w", errs.ErrEmptyField)
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("invalid password: %w", errs.ErrEmptyField)
	}
	return nil
}

func validateFirstName(firstName string) error {
	if firstName == "" {
		return fmt.Errorf("invalid first name: %w", errs.ErrEmptyField)
	}

	if len(firstName) > maxFieldLength {
		return fmt.Errorf("invalid first name: %w", errs.ErrFieldTooMuch)
	}

	return nil
}

func validateLastName(lastName string) error {
	if lastName == "" {
		return fmt.Errorf("invalid last name: %w", errs.ErrEmptyField)
	}

	if len(lastName) > maxFieldLength {
		return fmt.Errorf("invalid last name: %w", errs.ErrFieldTooMuch)
	}

	return nil
}
