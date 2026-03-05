package repository

import (
	"errors"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/infra"
)

var errWrongType = errors.New("cache value is not string")

type Cache interface {
	Get(key string) (any, error)
	Set(key string, v any, ttl time.Duration)
	Delete(key string)
}

type sessionRepo struct {
	cache Cache
}

func NewSessionRepo() *sessionRepo {
	return &sessionRepo{
		cache: infra.NewCache(),
	}
}

func (s *sessionRepo) Get(key domain.SessionID) (*domain.Session, error) {
	val, err := s.cache.Get(string(key))
	if err != nil {
		return nil, err
	}

	valStr, ok := val.(*domain.Session)
	if !ok {
		return nil, errWrongType
	}

	return valStr, nil
}

func (s *sessionRepo) Set(key domain.SessionID, v *domain.Session, ttl time.Duration) {
	s.cache.Set(string(key), v, ttl)
}

func (s *sessionRepo) Delete(key domain.SessionID) {
	s.cache.Delete(string(key))
}
