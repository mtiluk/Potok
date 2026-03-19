// internal/sync/manifest.go
package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ManifestEntry struct {
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

type Manifest struct {
	path    string
	Entries map[string]ManifestEntry `json:"entries"`
}

func manifestDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".potok", "manifests")
	return dir, os.MkdirAll(dir, 0700)
}

func LoadManifest(vaultName string) (*Manifest, error) {
	dir, err := manifestDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, vaultName+".json")

	m := &Manifest{
		path:    path,
		Entries: make(map[string]ManifestEntry),
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return m, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&m.Entries); err != nil {
		return m, nil
	}

	return m, nil
}

func (m *Manifest) Save() error {
	f, err := os.Create(m.path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(m.Entries)
}

func (m *Manifest) Set(relPath string, entry ManifestEntry) {
	m.Entries[relPath] = entry
}

func (m *Manifest) Remove(relPath string) {
	delete(m.Entries, relPath)
}

func (m *Manifest) Get(relPath string) (ManifestEntry, bool) {
	e, ok := m.Entries[relPath]
	return e, ok
}
