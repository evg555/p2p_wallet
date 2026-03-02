package repository

import (
	"errors"
	"time"

	"p2p_wallet/internal/domain"
	"p2p_wallet/internal/infra"
)

var errWrongType = errors.New("cache value is not string")

type Cache interface {
	Get(id int64) (any, error)
	Set(id int64, v any, ttl time.Duration)
	Delete(id int64)
}
type sessionRepo struct {
	cache Cache
}

func NewSessionRepo() *sessionRepo {
	return &sessionRepo{
		cache: infra.NewCache(),
	}
}

func (s *sessionRepo) Get(id int64) (domain.SessionID, error) {
	val, err := s.cache.Get(id)
	if err != nil {
		return "", err
	}

	valStr, ok := val.(domain.SessionID)
	if !ok {
		return "", errWrongType
	}

	return valStr, nil
}

func (s *sessionRepo) Set(id int64, v domain.SessionID, ttl time.Duration) {
	s.cache.Set(id, v, ttl)
}

func (s *sessionRepo) Delete(id int64) {
	s.cache.Delete(id)
}
