package pkg

import (
	"encoding/json"
	"os"
)

// Lockfile represents the rfh.lock file format
type Lockfile struct {
	Registry string                    `json:"registry"`
	Packages map[string]LockfileEntry `json:"packages"`
}

// LockfileEntry represents a single package in the lockfile
type LockfileEntry struct {
	Version     string   `json:"version"`
	SHA256      string   `json:"sha256"`
	Targets     []string `json:"targets"`
	InstallPath string   `json:"install_path"`
	Registry    string   `json:"registry,omitempty"`
}

// LoadLockfile reads and parses a lockfile
func LoadLockfile(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Return empty lockfile if doesn't exist
		return &Lockfile{
			Packages: make(map[string]LockfileEntry),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var lockfile Lockfile
	if err := json.Unmarshal(data, &lockfile); err != nil {
		return nil, err
	}

	if lockfile.Packages == nil {
		lockfile.Packages = make(map[string]LockfileEntry)
	}

	return &lockfile, nil
}

// SaveLockfile writes the lockfile to disk
func (l *Lockfile) SaveLockfile(path string) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// AddPackage adds or updates a package in the lockfile
func (l *Lockfile) AddPackage(name string, entry LockfileEntry) {
	if l.Packages == nil {
		l.Packages = make(map[string]LockfileEntry)
	}
	l.Packages[name] = entry
}

// RemovePackage removes a package from the lockfile
func (l *Lockfile) RemovePackage(name string) {
	if l.Packages != nil {
		delete(l.Packages, name)
	}
}

// HasPackage checks if a package exists in the lockfile
func (l *Lockfile) HasPackage(name string) bool {
	if l.Packages == nil {
		return false
	}
	_, exists := l.Packages[name]
	return exists
}

// GetPackage gets a package from the lockfile
func (l *Lockfile) GetPackage(name string) (LockfileEntry, bool) {
	if l.Packages == nil {
		return LockfileEntry{}, false
	}
	entry, exists := l.Packages[name]
	return entry, exists
}