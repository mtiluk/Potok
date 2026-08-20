package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func tempConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("POTOK_CONFIG_DIR", dir)
	return dir
}

func TestLoadReturnsNotFoundWhenMissing(t *testing.T) {
	tempConfig(t)

	if _, err := Load(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load() = %v, want ErrNotFound", err)
	}
}

func TestSaveThenLoad(t *testing.T) {
	dir := tempConfig(t)
	want := &Config{
		ServerURL: "https://potok.example.com",
		Vaults:    []Vault{{Name: "notes", Path: filepath.Join(dir, "notes")}},
	}

	if err := Save(want); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if got.ServerURL != want.ServerURL {
		t.Errorf("ServerURL = %q, want %q", got.ServerURL, want.ServerURL)
	}
	if len(got.Vaults) != 1 || got.Vaults[0].Name != "notes" {
		t.Errorf("Vaults = %+v, want one vault called notes", got.Vaults)
	}
	if got.Vaults[0].LastSyncedAt != nil {
		t.Error("LastSyncedAt should be nil for a vault that has never synced")
	}
}

func TestSaveOverwritesAndLeavesNoTempFiles(t *testing.T) {
	dir := tempConfig(t)

	if err := Save(&Config{ServerURL: "https://one.example.com"}); err != nil {
		t.Fatalf("Save() = %v", err)
	}
	if err := Save(&Config{ServerURL: "https://two.example.com"}); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if got.ServerURL != "https://two.example.com" {
		t.Errorf("ServerURL = %q, want the second save to win", got.ServerURL)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() = %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d files, want config.json only", len(entries))
	}
}

func TestSaveUsesRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file modes are not meaningful on Windows")
	}
	tempConfig(t)
	if err := Save(&Config{ServerURL: "https://potok.example.com"}); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	path, _ := Path()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() = %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config.json mode = %o, want 600", perm)
	}
}

func TestSaveRejectsInvalidConfig(t *testing.T) {
	tempConfig(t)

	if err := Save(&Config{ServerURL: "ftp://example.com"}); err == nil {
		t.Fatal("Save() = nil, want an error")
	}

	path, _ := Path()
	if _, err := os.Stat(path); err == nil {
		t.Error("Save() wrote a file despite failing validation")
	}
}

func TestLoadRejectsMalformedJSON(t *testing.T) {
	tempConfig(t)
	path, _ := Path()
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("WriteFile() = %v", err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("Load() = nil, want a parse error")
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		cfg     Config
		wantErr bool
	}{
		"zero value":          {cfg: Config{}},
		"valid":               {cfg: Config{ServerURL: "http://localhost:8080", Vaults: []Vault{{Name: "notes", Path: "/home/user/notes"}}}},
		"bad scheme":          {cfg: Config{ServerURL: "ftp://example.com"}, wantErr: true},
		"no host":             {cfg: Config{ServerURL: "https://"}, wantErr: true},
		"relative path":       {cfg: Config{Vaults: []Vault{{Name: "notes", Path: "notes"}}}, wantErr: true},
		"duplicate names":     {cfg: Config{Vaults: []Vault{{Name: "a", Path: "/a"}, {Name: "a", Path: "/b"}}}, wantErr: true},
		"empty name":          {cfg: Config{Vaults: []Vault{{Name: "", Path: "/a"}}}, wantErr: true},
		"name with space":     {cfg: Config{Vaults: []Vault{{Name: "my notes", Path: "/a"}}}, wantErr: true},
		"name with separator": {cfg: Config{Vaults: []Vault{{Name: "work/notes", Path: "/a"}}}, wantErr: true},
		"name starts with -":  {cfg: Config{Vaults: []Vault{{Name: "-force", Path: "/a"}}}, wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tc.cfg.Validate(); (err != nil) != tc.wantErr {
				t.Fatalf("Validate() = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestAddVault(t *testing.T) {
	cfg := &Config{}

	if err := cfg.AddVault(Vault{Name: "notes", Path: "/home/user/notes"}); err != nil {
		t.Fatalf("AddVault() = %v", err)
	}
	if err := cfg.AddVault(Vault{Name: "notes", Path: "/elsewhere"}); err == nil {
		t.Error("AddVault() = nil, want an error for a duplicate name")
	}
	if err := cfg.AddVault(Vault{Name: "work", Path: "relative"}); err == nil {
		t.Error("AddVault() = nil, want an error for a relative path")
	}
	if len(cfg.Vaults) != 1 {
		t.Errorf("len(Vaults) = %d, want 1", len(cfg.Vaults))
	}
}

func TestRemoveVault(t *testing.T) {
	cfg := &Config{Vaults: []Vault{{Name: "notes", Path: "/a"}, {Name: "work", Path: "/b"}}}

	if !cfg.RemoveVault("notes") {
		t.Error("RemoveVault(notes) = false, want true")
	}
	if cfg.RemoveVault("notes") {
		t.Error("RemoveVault(notes) = true on the second call, want false")
	}
	if len(cfg.Vaults) != 1 || cfg.Vaults[0].Name != "work" {
		t.Errorf("Vaults = %+v, want work only", cfg.Vaults)
	}
}

func TestVaultReturnsPointer(t *testing.T) {
	cfg := &Config{Vaults: []Vault{{Name: "notes", Path: "/a"}}}

	v, ok := cfg.Vault("notes")
	if !ok {
		t.Fatal("Vault(notes) not found")
	}
	v.RemoteID = "vault_123"
	if cfg.Vaults[0].RemoteID != "vault_123" {
		t.Error("Vault() returned a copy; updates would be lost")
	}

	if _, ok := cfg.Vault("missing"); ok {
		t.Error("Vault(missing) = true, want false")
	}
}

func TestDirPrecedence(t *testing.T) {
	t.Setenv("POTOK_CONFIG_DIR", "/custom/potok")
	t.Setenv("XDG_CONFIG_HOME", "/xdg")
	if got, _ := Dir(); got != "/custom/potok" {
		t.Errorf("Dir() = %q, want POTOK_CONFIG_DIR to win", got)
	}

	t.Setenv("POTOK_CONFIG_DIR", "")
	if got, _ := Dir(); got != filepath.Join("/xdg", "potok") {
		t.Errorf("Dir() = %q, want $XDG_CONFIG_HOME/potok", got)
	}
}
