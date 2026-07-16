/*
Suppression Service - local do-not-send list.

Addresses on the suppression list are blocked during campaign preflight. The
list is normalized with the same policy as duplicate detection (see
backend/email) and persisted atomically as JSON in the user's config directory.

This is a local suppression list only. No tracking pixels or unsubscribe links
are added to messages; unsubscribe handling is a manual/administrative process.
*/
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"MailMergeApp/backend/email"
)

// SuppressionService manages a local suppression list.
type SuppressionService struct {
	mu       sync.RWMutex
	set      map[string]string // normalized key -> original address as entered
	filePath string
}

// NewSuppressionService loads (or initializes) the suppression list.
func NewSuppressionService() (*SuppressionService, error) {
	appData, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}
	dir := filepath.Join(appData, "MailMergeGo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	s := &SuppressionService{
		set:      map[string]string{},
		filePath: filepath.Join(dir, "suppression.json"),
	}
	if err := s.load(); err != nil {
		fmt.Printf("Warning: failed to load suppression list: %v\n", err)
	}
	return s, nil
}

func (s *SuppressionService) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("corrupted suppression list: %w", err)
	}
	for _, e := range list {
		s.set[email.NormalizeKey(e)] = e
	}
	return nil
}

// save writes the list atomically (temp file + rename).
func (s *SuppressionService) save() error {
	list := make([]string, 0, len(s.set))
	for _, orig := range s.set {
		list = append(list, orig)
	}
	sort.Strings(list)
	return atomicWriteJSON(s.filePath, list)
}

// Add inserts an address (idempotent). Returns error for invalid addresses.
func (s *SuppressionService) Add(addr string) error {
	if !email.IsValid(addr) {
		return fmt.Errorf("invalid email address: %s", addr)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set[email.NormalizeKey(addr)] = addr
	return s.save()
}

// Remove deletes an address (idempotent).
func (s *SuppressionService) Remove(addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.set, email.NormalizeKey(addr))
	return s.save()
}

// Contains reports whether an address is suppressed.
func (s *SuppressionService) Contains(addr string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.set[email.NormalizeKey(addr)]
	return ok
}

// List returns all suppressed addresses (sorted copy).
func (s *SuppressionService) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.set))
	for _, orig := range s.set {
		out = append(out, orig)
	}
	sort.Strings(out)
	return out
}

// Set returns a copy of the normalized suppression set for preflight use.
func (s *SuppressionService) Set() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]bool, len(s.set))
	for k := range s.set {
		out[k] = true
	}
	return out
}

// Import adds many addresses, returning the count added and any invalid entries.
func (s *SuppressionService) Import(addrs []string) (added int, invalid []string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range addrs {
		if !email.IsValid(a) {
			invalid = append(invalid, a)
			continue
		}
		key := email.NormalizeKey(a)
		if _, exists := s.set[key]; !exists {
			added++
		}
		s.set[key] = a
	}
	return added, invalid, s.save()
}
