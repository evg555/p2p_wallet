package domain

import (
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
