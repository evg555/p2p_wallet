package repository

import (
	"errors"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/infra/cache"
)

var cacheSize = 256

var (
	errWrongType = errors.New("cache value is not string")
	errNotFound  = errors.New("cache value not found")
)

type Cache interface {
	Get(key string) (any, bool)
	SetWithTTL(key string, v any, ttl time.Duration)
	Delete(key string)
}

type sessionRepo struct {
	cache Cache
}

func NewSessionRepo() *sessionRepo {
	return &sessionRepo{
		cache: cache.NewCache(cacheSize),
	}
}

func (s *sessionRepo) Get(key domain.SessionID) (*domain.Session, error) {
	val, ok := s.cache.Get(string(key))
	if !ok {
		return nil, errNotFound
	}

	valStr, ok := val.(*domain.Session)
	if !ok {
		return nil, errWrongType
	}

	return valStr, nil
}

func (s *sessionRepo) Set(key domain.SessionID, v *domain.Session, ttl time.Duration) {
	s.cache.SetWithTTL(string(key), v, ttl)
}

func (s *sessionRepo) Delete(key domain.SessionID) {
	s.cache.Delete(string(key))
}
