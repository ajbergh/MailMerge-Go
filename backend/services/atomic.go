package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// atomicWrite writes data to path durably: it writes to a temp file in the same
// directory, flushes it, then atomically renames it over the destination. The
// destination is only replaced once the new content is safely on disk, so a
// crash mid-write can never leave a half-written file.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup if we bail out before the rename.
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to flush temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

// atomicWriteJSON marshals v as indented JSON and writes it atomically.
func atomicWriteJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return atomicWrite(path, data, 0o644)
}
