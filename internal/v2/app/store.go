package app

import (
	"sync"
	"time"

	"github.com/AnTengye/lol-shield/internal/v2/domain"
)

type Store struct {
	mu    sync.RWMutex
	state domain.StateSnapshot
}

func NewStore() *Store {
	return &Store{
		state: domain.NewInitialSnapshot(),
	}
}

func (s *Store) Snapshot() domain.StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Store) Set(next domain.StateSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next.UpdatedAt = time.Now()
	s.state = next
}

func (s *Store) Update(fn func(current domain.StateSnapshot) domain.StateSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := fn(s.state)
	next.UpdatedAt = time.Now()
	s.state = next
}
