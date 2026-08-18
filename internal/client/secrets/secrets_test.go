package secrets

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestGetSetDelete(t *testing.T) {
	keyring.MockInit()
	const key, value = APIKey, "pk_live_abc123"

	if _, err := Get(key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() before Set = %v, want ErrNotFound", err)
	}

	if err := Set(key, value); err != nil {
		t.Fatalf("Set() = %v", err)
	}
	got, err := Get(key)
	if err != nil {
		t.Fatalf("Get() = %v", err)
	}
	if got != value {
		t.Errorf("Get() = %q, want %q", got, value)
	}

	if err := Set(key, "replaced"); err != nil {
		t.Fatalf("Set() overwrite = %v", err)
	}
	if got, _ := Get(key); got != "replaced" {
		t.Errorf("Get() after overwrite = %q, want %q", got, "replaced")
	}

	if err := Delete(key); err != nil {
		t.Fatalf("Delete() = %v", err)
	}
	if _, err := Get(key); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after Delete = %v, want ErrNotFound", err)
	}
}

func TestKeysAreIndependent(t *testing.T) {
	keyring.MockInit()

	if err := Set(VaultKeyName("notes"), "one"); err != nil {
		t.Fatalf("Set(notes) = %v", err)
	}
	if err := Set(VaultKeyName("work"), "two"); err != nil {
		t.Fatalf("Set(work) = %v", err)
	}
	if err := Delete(VaultKeyName("notes")); err != nil {
		t.Fatalf("Delete(notes) = %v", err)
	}

	got, err := Get(VaultKeyName("work"))
	if err != nil {
		t.Fatalf("Get(work) = %v", err)
	}
	if got != "two" {
		t.Errorf("Get(work) = %q, want it untouched by the delete", got)
	}
}

func TestVaultKeyIsNamespaced(t *testing.T) {
	if got := VaultKeyName(APIKey); got == APIKey {
		t.Errorf("VaultKey(%q) = %q, want it namespaced away from the API key", APIKey, got)
	}
	if got, want := VaultKeyName("notes"), "vault:notes"; got != want {
		t.Errorf("VaultKey(notes) = %q, want %q", got, want)
	}
}

func TestVaultKeyName(t *testing.T) {
	want := "vault:test"
	got := VaultKeyName("test")

	if got != "vault:test" {
		t.Errorf("Got %v, wanted %v", got, want)
	}
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"key length empty", "", true},
		{"invalid char", "\n", true},
		{"valid key", "example", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
