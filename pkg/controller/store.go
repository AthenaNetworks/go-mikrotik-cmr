package controller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// FileStore persists the device registry to a JSON file. It is safe for
// concurrent use and writes atomically (temp file + rename). It is a reasonable
// default; a larger deployment can supply its own Store (e.g. a database).
type FileStore struct {
	path string
	mu   sync.Mutex
}

// NewFileStore returns a FileStore backed by the file at path. The file is
// created on first Save; a missing file Loads as an empty registry.
func NewFileStore(path string) *FileStore { return &FileStore{path: path} }

// Load reads the device set from the file. A missing file is not an error.
func (f *FileStore) Load() ([]Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	var devs []Device
	if err := json.Unmarshal(b, &devs); err != nil {
		return nil, err
	}
	return devs, nil
}

// Save writes the device set to the file atomically.
func (f *FileStore) Save(devices []Device) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(f.path)
	tmp, err := os.CreateTemp(dir, ".cmr-registry-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, f.path)
}
