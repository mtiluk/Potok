package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/crypto"
	"github.com/michaeltukdev/Potok/internal/preflight"
	"github.com/michaeltukdev/Potok/internal/style"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

//
// potok push - Encrypt and upload a vault to the server.
//
// Usage Examples:
//   potok push notes
//
// Inputs:
//   - Args:
//     - Vault name (required, positional)
//
// Behaviour:
//   - Reads vault path from local config
//   - Reads encryption password from OS keyring
//   - Creates vault on server if it doesn't exist
//   - Encrypts all files locally then uploads
//   - Updates last synced time in local config
//
// Requires:
//   - Vault registered locally (run 'potok vault-add' first)
//   - Server URL and API key (run 'potok init' first)
//

var pushCmd = &cobra.Command{
	Use:   "push [name]",
	Short: "Encrypt and upload a vault to the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		style.Bold.Fprintln(cmd.OutOrStdout(), "Potok Push")
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

		created, err := c.CreateVault(vaultName)
		if err != nil {
			return fmt.Errorf("create vault: %w", err)
		}

		if created {
			cmd.Println("Created remote vault:", vaultName)
		}

		salt, err := c.DownloadFile(vaultName, ".potok/salt")
		if err != nil {
			salt, err = crypto.GenerateSalt()
			if err != nil {
				return fmt.Errorf("generate salt: %w", err)
			}

			if err := c.UploadFile(vaultName, ".potok/salt", salt); err != nil {
				return fmt.Errorf("upload salt: %w", err)
			}

			cmd.Println("Generated new encryption salt")
		}

		key, err := crypto.DeriveKey([]byte(vaultPassword), salt)
		if err != nil {
			return fmt.Errorf("derive key: %w", err)
		}

		files, err := client.WalkVault(vaultPath)
		if err != nil {
			return fmt.Errorf("walk vault: %w", err)
		}

		if len(files) == 0 {
			cmd.Println("No files to push.")
			return nil
		}

		for _, relPath := range files {
			absPath := filepath.Join(vaultPath, relPath)

			encrypted, err := crypto.EncryptFile(key, absPath)
			if err != nil {
				return fmt.Errorf("encrypt %s: %w", relPath, err)
			}

			fmt.Println("uploadsss")

			if err := c.UploadFile(vaultName, relPath, encrypted); err != nil {
				return fmt.Errorf("upload %s: %w", relPath, err)
			}

			cmd.Printf("  ✓ %s\n", relPath)
		}

		cfg.UpdateLastSynced(vaultName, time.Now().Format(time.RFC3339))
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		cmd.Printf("\nPushed %d files to %q\n", len(files), vaultName)
		return nil
	},
}
