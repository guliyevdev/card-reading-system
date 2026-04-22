package state

import (
	"sync"
	"time"
)

type Card struct {
	UID       string    `json:"uid"`
	ATR       string    `json:"atr"`
	Reader    string    `json:"reader,omitempty"`
	Present   bool      `json:"present,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type Store struct {
	mu   sync.RWMutex
	card Card
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Snapshot() Card {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.card
}

func (s *Store) Set(card Card) {
	s.mu.Lock()
	defer s.mu.Unlock()
	card.UpdatedAt = time.Now().UTC()
	s.card = card
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.card = Card{
		UpdatedAt: time.Now().UTC(),
	}
}
