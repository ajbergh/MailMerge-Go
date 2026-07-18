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

// Create writes a new record and fails if the ID already exists.
func (r *JSONCampaignRepository) Create(rec CampaignRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, err := r.path(rec.ID)
	if err != nil {
		return err
	}
	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("campaign record %q already exists", rec.ID)
	} else if !os.IsNotExist(err) {
		return err
	}
	return r.writeAtomic(p, rec)
}

// Update atomically replaces an existing record.
func (r *JSONCampaignRepository) Update(rec CampaignRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, err := r.path(rec.ID)
	if err != nil {
		return err
	}
	if _, err := os.Stat(p); err != nil {
		return err
	}
	return r.writeAtomic(p, rec)
}

// Save is an upsert-compatible persistence method retained for existing callers.
func (r *JSONCampaignRepository) Save(rec CampaignRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, err := r.path(rec.ID)
	if err != nil {
		return err
	}
	return r.writeAtomic(p, rec)
}

func (r *JSONCampaignRepository) writeAtomic(path string, rec CampaignRecord) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(r.dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// List returns all records, newest first. Corrupted files are isolated and
// skipped so one bad record never prevents the rest of campaign history loading.
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
	sort.Slice(out, func(i, j int) bool {
		li := out[i].StartedAt
		lj := out[j].StartedAt
		if li.IsZero() {
			li = out[i].CreatedAt
		}
		if lj.IsZero() {
			lj = out[j].CreatedAt
		}
		return li.After(lj)
	})
	return out, nil
}

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

var _ CampaignRepository = (*JSONCampaignRepository)(nil)
