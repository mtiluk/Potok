package auth

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	valid, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() error: %v", err)
	}

	if !strings.HasPrefix(valid, Prefix) {
		t.Errorf("GenerateAPIKey() = %v, want %v", valid, Prefix+valid)
	}
}
