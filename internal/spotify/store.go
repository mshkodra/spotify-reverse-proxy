package spotify

import (
	"sync"

	"golang.org/x/oauth2"
)

type TokenStore struct {
	sync.RWMutex
	m map[string]*oauth2.Token
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		m: make(map[string]*oauth2.Token),
	}
}

func (s *TokenStore) Get(userID string) (*oauth2.Token, bool) {
	s.RLock()
	defer s.RUnlock()
	t, ok := s.m[userID]
	return t, ok
}

func (s *TokenStore) Set(userID string, token *oauth2.Token) {
	s.Lock()
	defer s.Unlock()
	s.m[userID] = token
}
