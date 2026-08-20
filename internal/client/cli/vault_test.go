package cli

// func obsidianVault(t *testing.T) string {
// 	t.Helper()
// 	dir := t.TempDir()
// 	if err := os.Mkdir(filepath.Join(dir, ".obsidian"), 0o755); err != nil {
// 		t.Fatalf("Mkdir(.obsidian) = %v", err)
// 	}
// 	return dir
// }

// func TestVaultAdd(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		path       string
// 		passphrase string
// 		wantErr    bool
// 	}{
// 		{"key length empty", "", true},
// 	}

// 	// Take path
// 	// Resolve it to absolute
// 	// Check directory exists and is readable
// 	// If does not contain .obsidian warn
// 	// Prompt for a passphrase twice
// 	// Store under vault:<name>
// 	// Append to config
// }
