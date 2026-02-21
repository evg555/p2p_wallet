package domain

import (
	"github.com/google/uuid"
)

var SessionKey = "session_id"

type CtxKey string

type SessionID string

type AuthResult struct {
	UserID       int64     `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	UserLastName string    `json:"user_last_name"`
	SessionID    SessionID `json:"session_id"`
}

func NewSessionID() SessionID {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return ""
	}

	return SessionID(newUUID.String())
}

func (s SessionID) IsEmpty() bool {
	return len(s) == 0
}

func (s SessionID) Equal(sessionID string) bool {
	return string(s) == sessionID
}
