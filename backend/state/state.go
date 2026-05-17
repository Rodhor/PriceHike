package state

import (
	"PriceHike/backend/domain"
	"sync"
)

type Store struct {
	mu      sync.RWMutex
	dataset *domain.Dataset
	path    string
}

func New() *Store {
	return &Store{}
}

func (s *Store) Set(ds domain.Dataset, filepath string) {

	clone := ds.Clone()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dataset = &clone
	s.path = filepath
}

func (s *Store) Get() (domain.Dataset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.dataset == nil {
		return domain.Dataset{}, domain.ErrNoDatasetLoaded
	}

	return s.dataset.Clone(), nil
}

func (s *Store) Path() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.path
}

func (s *Store) IsLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.dataset != nil
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dataset = nil
	s.path = ""
}
