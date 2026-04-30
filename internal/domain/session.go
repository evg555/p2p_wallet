package domain

import (
	"time"

	"github.com/google/uuid"
)

type SessionID string
type Session struct {
	ID        SessionID
	UserID    UserID
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewSession(userID UserID, ttl time.Duration) *Session {
	now := time.Now()

	return &Session{
		ID:        NewSessionID(),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (s *Session) SessionID() SessionID {
	return s.ID
}

func NewSessionID() SessionID {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return ""
	}

	return SessionID(newUUID.String())
}

func (s *Session) IsExpired(t time.Time) bool {
	return s.ExpiresAt.Before(t)
}
