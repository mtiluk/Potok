package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func newStore(t *testing.T) *FileStore {
	t.Helper()
	return &FileStore{Dir: t.TempDir()}
}

func TestLoadReturnsNotFoundWhenMissing(t *testing.T) {
	store := newStore(t)

	_, err := store.Load()

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load() = %v, want ErrNotFound", err)
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	store := newStore(t)
	synced := time.Date(2026, time.March, 4, 12, 0, 0, 0, time.UTC)
	want := &Config{
		ServerURL: "https://potok.example.com",
		Vaults: []Vault{
			{Name: "notes", Path: filepath.Join(store.Dir, "notes"), LastSyncedAt: &synced},
			{Name: "work", Path: filepath.Join(store.Dir, "work")},
		},
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if got.ServerURL != want.ServerURL {
		t.Errorf("ServerURL = %q, want %q", got.ServerURL, want.ServerURL)
	}
	if len(got.Vaults) != len(want.Vaults) {
		t.Fatalf("len(Vaults) = %d, want %d", len(got.Vaults), len(want.Vaults))
	}
	if got.Vaults[0].LastSyncedAt == nil || !got.Vaults[0].LastSyncedAt.Equal(synced) {
		t.Errorf("Vaults[0].LastSyncedAt = %v, want %v", got.Vaults[0].LastSyncedAt, synced)
	}
	if got.Vaults[1].LastSyncedAt != nil {
		t.Errorf("Vaults[1].LastSyncedAt = %v, want nil for a never-synced vault", got.Vaults[1].LastSyncedAt)
	}
}

func TestSaveOverwritesExistingConfig(t *testing.T) {
	store := newStore(t)
	first := &Config{ServerURL: "https://one.example.com"}
	if err := store.Save(first); err != nil {
		t.Fatalf("Save(first) = %v", err)
	}

	second := &Config{ServerURL: "https://two.example.com"}
	if err := store.Save(second); err != nil {
		t.Fatalf("Save(second) = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if got.ServerURL != second.ServerURL {
		t.Errorf("ServerURL = %q, want %q", got.ServerURL, second.ServerURL)
	}

	entries, err := os.ReadDir(store.Dir)
	if err != nil {
		t.Fatalf("ReadDir() = %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("config directory holds %v, want config.json only", names)
	}
}

func TestSaveUsesRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file modes are not meaningful on Windows")
	}
	store := newStore(t)
	if err := store.Save(&Config{ServerURL: "https://potok.example.com"}); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	info, err := os.Stat(store.Path())
	if err != nil {
		t.Fatalf("Stat() = %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config.json mode = %o, want 600", perm)
	}
}

func TestSaveRejectsInvalidConfig(t *testing.T) {
	store := newStore(t)

	err := store.Save(&Config{ServerURL: "ftp://example.com"})

	if err == nil {
		t.Fatal("Save() = nil, want an error for an invalid config")
	}
	if _, statErr := os.Stat(store.Path()); statErr == nil {
		t.Error("Save() wrote a file despite failing validation")
	}
}

func TestLoadRejectsMalformedJSON(t *testing.T) {
	store := newStore(t)
	if err := os.WriteFile(store.Path(), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("WriteFile() = %v", err)
	}

	if _, err := store.Load(); err == nil {
		t.Fatal("Load() = nil, want an error for malformed JSON")
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		cfg     Config
		wantErr bool
	}{
		"zero value": {
			cfg: Config{},
		},
		"valid": {
			cfg: Config{
				ServerURL: "http://localhost:8080",
				Vaults:    []Vault{{Name: "notes", Path: "/home/user/notes"}},
			},
		},
		"unsupported url scheme": {
			cfg:     Config{ServerURL: "ftp://example.com"},
			wantErr: true,
		},
		"url without a host": {
			cfg:     Config{ServerURL: "https://"},
			wantErr: true,
		},
		"relative vault path": {
			cfg:     Config{Vaults: []Vault{{Name: "notes", Path: "notes"}}},
			wantErr: true,
		},
		"duplicate vault names": {
			cfg: Config{Vaults: []Vault{
				{Name: "notes", Path: "/a"},
				{Name: "notes", Path: "/b"},
			}},
			wantErr: true,
		},
		"empty vault name": {
			cfg:     Config{Vaults: []Vault{{Name: "", Path: "/a"}}},
			wantErr: true,
		},
		"vault name with a space": {
			cfg:     Config{Vaults: []Vault{{Name: "my notes", Path: "/a"}}},
			wantErr: true,
		},
		"vault name with a path separator": {
			cfg:     Config{Vaults: []Vault{{Name: "work/notes", Path: "/a"}}},
			wantErr: true,
		},
		"vault name starting with a dash": {
			cfg:     Config{Vaults: []Vault{{Name: "-force", Path: "/a"}}},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
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
	if len(cfg.Vaults) != 1 {
		t.Fatalf("len(Vaults) = %d, want 1", len(cfg.Vaults))
	}

	if err := cfg.AddVault(Vault{Name: "notes", Path: "/somewhere/else"}); err == nil {
		t.Error("AddVault() = nil, want an error for a duplicate name")
	}
	if len(cfg.Vaults) != 1 {
		t.Errorf("len(Vaults) = %d after a rejected add, want 1", len(cfg.Vaults))
	}
}

func TestRemoveVault(t *testing.T) {
	cfg := &Config{Vaults: []Vault{
		{Name: "notes", Path: "/a"},
		{Name: "work", Path: "/b"},
	}}

	if removed := cfg.RemoveVault("notes"); !removed {
		t.Error("RemoveVault(notes) = false, want true")
	}
	if removed := cfg.RemoveVault("notes"); removed {
		t.Error("RemoveVault(notes) = true on the second call, want false")
	}
	if len(cfg.Vaults) != 1 || cfg.Vaults[0].Name != "work" {
		t.Errorf("Vaults = %+v, want work only", cfg.Vaults)
	}
}

func TestVaultLookup(t *testing.T) {
	cfg := &Config{Vaults: []Vault{{Name: "notes", Path: "/a"}}}

	got, ok := cfg.Vault("notes")
	if !ok {
		t.Fatal("Vault(notes) not found")
	}
	got.RemoteID = "vault_123"
	if cfg.Vaults[0].RemoteID != "vault_123" {
		t.Error("Vault() returned a copy, want a pointer into the slice")
	}

	if _, ok := cfg.Vault("missing"); ok {
		t.Error("Vault(missing) = true, want false")
	}
}

func TestDirPrefersPotokConfigDir(t *testing.T) {
	t.Setenv("POTOK_CONFIG_DIR", "/custom/potok")
	t.Setenv("XDG_CONFIG_HOME", "/xdg")

	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir() = %v", err)
	}
	if got != "/custom/potok" {
		t.Errorf("Dir() = %q, want /custom/potok", got)
	}
}

func TestDirFallsBackToXDG(t *testing.T) {
	t.Setenv("POTOK_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "/xdg")

	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir() = %v", err)
	}
	if want := filepath.Join("/xdg", "potok"); got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}
