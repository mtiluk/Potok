package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrNotFound = errors.New("config: not initialised, run `potok init`")

const maxVaultNameLen = 64
const invalidNameChars = `/\:*?"<>|` + " "

type Config struct {
	ServerURL string  `json:"server_url"`
	Vaults    []Vault `json:"vaults"`
}

type Vault struct {
	Name         string     `json:"name"`
	Path         string     `json:"path"`
	RemoteID     string     `json:"remote_id,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}

type Store interface {
	Load() (*Config, error)
	Save(cfg *Config) error
	Path() string
}

type FileStore struct {
	Dir string
}

var _ Store = (*FileStore)(nil)

func Dir() (string, error) {
	if dir := os.Getenv("POTOK_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "potok"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: locate home directory: %w", err)
	}
	return filepath.Join(home, ".potok"), nil
}

func (s *FileStore) Path() string {
	dir := s.Dir
	if dir == "" {
		resolved, err := Dir()
		if err != nil {
			return "config.json"
		}
		dir = resolved
	}
	return filepath.Join(dir, "config.json")
}

func (s *FileStore) Load() (*Config, error) {
	path := s.Path()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return &cfg, nil
}

func (s *FileStore) Save(cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: encode: %w", err)
	}
	data = append(data, '\n')

	path := s.Path()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("config: create %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".config-*.json")
	if err != nil {
		return fmt.Errorf("config: create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {

		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(0o600); err != nil {
		return fmt.Errorf("config: set permissions on temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("config: write temp file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("config: sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("config: close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("config: replace %s: %w", path, err)
	}
	return nil
}

func (c *Config) Validate() error {
	if c.ServerURL != "" {
		u, err := url.Parse(c.ServerURL)
		if err != nil {
			return fmt.Errorf("config: invalid server_url %q: %w", c.ServerURL, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("config: server_url must use http or https, got %q", u.Scheme)
		}
		if u.Host == "" {
			return fmt.Errorf("config: server_url %q is missing a host", c.ServerURL)
		}
	}

	seen := make(map[string]struct{}, len(c.Vaults))
	for _, v := range c.Vaults {
		if err := ValidateVaultName(v.Name); err != nil {
			return err
		}
		if _, dup := seen[v.Name]; dup {
			return fmt.Errorf("config: vault %q is registered twice", v.Name)
		}
		seen[v.Name] = struct{}{}

		if !filepath.IsAbs(v.Path) {
			return fmt.Errorf("config: vault %q needs an absolute path, got %q", v.Name, v.Path)
		}
	}
	return nil
}

func ValidateVaultName(name string) error {
	switch {
	case name == "":
		return errors.New("config: vault name is empty")
	case len(name) > maxVaultNameLen:
		return fmt.Errorf("config: vault name %q is longer than %d characters", name, maxVaultNameLen)
	case strings.ContainsAny(name, invalidNameChars):
		return fmt.Errorf("config: vault name %q contains a space or one of %s", name, invalidNameChars)
	case strings.HasPrefix(name, "-"):
		return fmt.Errorf("config: vault name %q may not start with a dash", name)
	case strings.HasPrefix(name, "."):
		return fmt.Errorf("config: vault name %q may not start with a dot", name)
	}
	return nil
}

func (c *Config) Vault(name string) (*Vault, bool) {
	for i := range c.Vaults {
		if c.Vaults[i].Name == name {
			return &c.Vaults[i], true
		}
	}
	return nil, false
}

func (c *Config) AddVault(v Vault) error {
	if err := ValidateVaultName(v.Name); err != nil {
		return err
	}
	if _, exists := c.Vault(v.Name); exists {
		return fmt.Errorf("config: vault %q is already registered", v.Name)
	}
	if !filepath.IsAbs(v.Path) {
		return fmt.Errorf("config: vault %q needs an absolute path, got %q", v.Name, v.Path)
	}
	c.Vaults = append(c.Vaults, v)
	return nil
}

func (c *Config) RemoveVault(name string) bool {
	for i := range c.Vaults {
		if c.Vaults[i].Name == name {
			c.Vaults = append(c.Vaults[:i], c.Vaults[i+1:]...)
			return true
		}
	}
	return false
}
