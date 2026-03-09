package prompt

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/michaeltukdev/Potok/internal/style"
	"github.com/ncruces/zenity"
)

var vaultNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

func PromptVaultName() (string, error) {
	for {
		name, err := Input(style.Dim.Sprint("Vault name (a-z, 0-9, hyphens, underscores): "))
		if err != nil {
			return "", fmt.Errorf("vault name error: %w", err)
		}

		name = strings.TrimSpace(name)

		if name == "" {
			fmt.Println("Vault name cannot be empty.")
			continue
		}

		if !vaultNameRegex.MatchString(name) {
			fmt.Println("Only letters, numbers, hyphens, and underscores allowed. Must start with a letter or number.")
			continue
		}

		return name, nil
	}
}

func SelectVaultPath() (string, error) {
	vaultDir, err := zenity.SelectFile(
		zenity.Title("Select vault to back up"),
		zenity.Directory(),
	)
	if err != nil {
		if errors.Is(err, zenity.ErrCanceled) {
			return "", fmt.Errorf("selection canceled")
		}
		return "", fmt.Errorf("selecting vault: %w", err)
	}

	if vaultDir == "" {
		return "", fmt.Errorf("no valid directory selected")
	}

	info, err := os.Stat(vaultDir)
	if err != nil {
		return "", fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("selected path is not a directory: %s", vaultDir)
	}

	return vaultDir, nil
}
