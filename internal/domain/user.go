package domain

import (
	"fmt"
	"time"

	"p2p_wallet/internal/errs"

	"golang.org/x/crypto/bcrypt"
)

var maxFieldLength = 50

type User struct {
	ID        int64
	Login     string
	Password  string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt *time.Time
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

	hash, err := encodePassword(password)
	if err != nil {
		return nil, fmt.Errorf("user: encode password: %w", err)
	}
	user.Password = hash

	return user, nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func encodePassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
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
