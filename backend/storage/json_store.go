package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
)

var validRecordID = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// JSONCampaignRepository stores each campaign run as a JSON file in a directory.
type JSONCampaignRepository struct {
	mu  sync.Mutex
	dir string
}

// NewJSONCampaignRepository creates a repository rooted at dir (created if
// needed).
func NewJSONCampaignRepository(dir string) (*JSONCampaignRepository, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create campaign store: %w", err)
	}
	return &JSONCampaignRepository{dir: dir}, nil
}

func (r *JSONCampaignRepository) path(id string) (string, error) {
	if !validRecordID.MatchString(id) {
		return "", fmt.Errorf("invalid campaign record id")
	}
	return filepath.Join(r.dir, id+".json"), nil
}

// Save writes a record atomically.
func (r *JSONCampaignRepository) Save(rec CampaignRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, err := r.path(rec.ID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// List returns all records, newest first. Corrupted files are skipped.
func (r *JSONCampaignRepository) List() ([]CampaignRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, err
	}
	var out []CampaignRecord
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(r.dir, e.Name()))
		if err != nil {
			continue
		}
		var rec CampaignRecord
		if json.Unmarshal(data, &rec) != nil {
			continue
		}
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out, nil
}

// Get returns a single record by ID.
func (r *JSONCampaignRepository) Get(id string) (*CampaignRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, err := r.path(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var rec CampaignRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("corrupted campaign record: %w", err)
	}
	return &rec, nil
}

// Delete removes a record.
func (r *JSONCampaignRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, err := r.path(id)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Compile-time check.
var _ CampaignRepository = (*JSONCampaignRepository)(nil)
