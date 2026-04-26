package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// storedKey is defined in config_store.go (shared with TOML serialisation).

type apiKeyStore struct {
	mu   sync.RWMutex
	keys []storedKey
}

func generateKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "voc_" + hex.EncodeToString(b), nil
}

func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func (s *apiKeyStore) hasKeys() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.keys) > 0
}

func (s *apiKeyStore) validate(raw string) (id string, ok bool) {
	h := hashKey(raw)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, k := range s.keys {
		if k.Hash == h {
			return k.ID, true
		}
	}
	return "", false
}

func (s *apiKeyStore) list() []storedKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]storedKey, len(s.keys))
	copy(cp, s.keys)
	return cp
}

func (s *apiKeyStore) create(name string) (storedKey, string, error) {
	raw, err := generateKey()
	if err != nil {
		return storedKey{}, "", err
	}
	id, err := generateID()
	if err != nil {
		return storedKey{}, "", err
	}
	entry := storedKey{
		ID:        id,
		Name:      name,
		Prefix:    raw[:12],
		Hash:      hashKey(raw),
		CreatedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.keys = append(s.keys, entry)
	s.mu.Unlock()
	if err := s.save(); err != nil {
		// Non-fatal: key is in memory, just failed to persist.
		_ = err
	}
	return entry, raw, nil
}

func (s *apiKeyStore) delete(id string) (bool, error) {
	s.mu.Lock()
	found := false
	filtered := s.keys[:0]
	for _, k := range s.keys {
		if k.ID == id {
			found = true
		} else {
			filtered = append(filtered, k)
		}
	}
	s.keys = filtered
	s.mu.Unlock()
	if !found {
		return false, nil
	}
	return true, s.save()
}

func (s *apiKeyStore) touchLastUsed(id string) {
	now := time.Now().UTC()
	s.mu.Lock()
	for i := range s.keys {
		if s.keys[i].ID == id {
			s.keys[i].LastUsedAt = &now
			break
		}
	}
	s.mu.Unlock()
	_ = s.save()
}

func (s *apiKeyStore) save() error {
	s.mu.RLock()
	keys := make([]storedKey, len(s.keys))
	copy(keys, s.keys)
	s.mu.RUnlock()

	fileMu.Lock()
	defer fileMu.Unlock()
	vc := readVocalizeConfigUnlocked()
	vc.APIKeys = keys
	return writeVocalizeConfigUnlocked(vc)
}

func loadAPIKeyStore() *apiKeyStore {
	fileMu.Lock()
	vc := readVocalizeConfigUnlocked()
	fileMu.Unlock()
	return &apiKeyStore{keys: vc.APIKeys}
}
