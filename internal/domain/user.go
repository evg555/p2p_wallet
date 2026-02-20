package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
)

type User struct {
	ID        int64      `json:"id"`
	Login     string     `json:"login"`
	Password  string     `json:"password"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (u *User) NewUserID() {
	u.ID = rand.Int63()
}

func EncodePassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func CheckPassword(password, encoded string) bool {
	return EncodePassword(password) == encoded
}
