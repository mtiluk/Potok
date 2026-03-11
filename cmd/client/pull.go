package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/crypto"
	"github.com/michaeltukdev/Potok/internal/preflight"
	"github.com/michaeltukdev/Potok/internal/style"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var pullCmd = &cobra.Command{
	Use:   "pull [name]",
	Short: "Download and decrypt a vault from the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		style.Bold.Fprintln(cmd.OutOrStdout(), "Potok Pull")
		cmd.Println()

		_, err := preflight.Check()
		if err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		apiKey, err := keyring.Get("potok", "api-key")
		if err != nil || apiKey == "" {
			return fmt.Errorf("run 'potok init' first")
		}

		vaultName := args[0]

		if !cfg.HasVault(vaultName) {
			return fmt.Errorf(
				"vault %q not found — run 'potok vault-add' first",
				vaultName,
			)
		}

		vaultPassword, err := keyring.Get("potok", "vault:"+vaultName)
		if err != nil || vaultPassword == "" {
			return fmt.Errorf("no password found for vault %q", vaultName)
		}

		vaultPath := cfg.VaultPath(vaultName)

		c := client.NewClient(cfg.ServerURL, apiKey)

		files, err := c.ListFiles(vaultName)
		if err != nil {
			return fmt.Errorf("list files: %w", err)
		}

		if len(files) == 0 {
			cmd.Println("No files to pull.")
			return nil
		}

		salt, err := c.DownloadFile(vaultName, ".potok/salt")
		if err != nil {
			return fmt.Errorf("download salt: %w", err)
		}

		key, err := crypto.DeriveKey([]byte(vaultPassword), salt)
		if err != nil {
			return fmt.Errorf("derive key: %w", err)
		}

		for _, relPath := range files {
			data, err := c.DownloadFile(vaultName, relPath)
			if err != nil {
				return fmt.Errorf("download %s: %w", relPath, err)
			}

			decrypted, err := crypto.DecryptBytes(key, data)
			if err != nil {
				return fmt.Errorf("decrypt %s: %w", relPath, err)
			}

			absPath := filepath.Join(vaultPath, relPath)
			if err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
				return fmt.Errorf("create dir for %s: %w", relPath, err)
			}

			if err := os.WriteFile(absPath, decrypted, 0644); err != nil {
				return fmt.Errorf("write %s: %w", relPath, err)
			}

			cmd.Printf("  ✓ %s\n", relPath)
		}

		cmd.Printf("\nPulled %d files to %q\n", len(files), vaultPath)
		return nil
	},
}
