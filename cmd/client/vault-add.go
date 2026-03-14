package main

import (
	"fmt"

	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/prompt"
	"github.com/michaeltukdev/Potok/internal/style"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

//
// potok vault add - Register a local folder as a vault and store its encryption
// password in the OS keyring.
//
// Usage Examples:
//   potok vault add notes
//   potok vault add notes --path ~/Documents/Obsidian/Notes
//   potok vault add notes --upload
//
// Inputs:
//   - Prompts:
//  	- Vault path (if --path not provided)
//   	- Vault password (input hidden; stored in OS keyring)
//   - Flags:
//	    --path string          Path to the local vault folder
//		--upload               Encrypt and upload after adding
//      --password-stdin       Read vault password from stdin (non-interactive)
//

var addVaultCmd = &cobra.Command{
	Use:   "vault-add",
	Short: "Select a Vault to be backed up securely!",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		style.Bold.Fprintln(cmd.OutOrStdout(), "Registering vault...")
		cmd.Println()

		vaultName, err := prompt.PromptVaultName()
		if err != nil {
			return fmt.Errorf("vault name error: %w", err)
		}

		if cfg.HasVault(vaultName) {
			return fmt.Errorf("vault %q is already registered", vaultName)
		}

		style.Dim.Print("Select vault path: ")
		vaultPath, err := prompt.SelectVaultPath()
		if err != nil {
			return fmt.Errorf("vault path error: %w", err)
		}
		fmt.Println(vaultPath)

		vaultPassword, err := prompt.Secret(style.Dim.Sprint("Vault password (input hidden): "))
		if err != nil {
			return fmt.Errorf("read vault password: %w", err)
		}

		if vaultPassword == "" {
			return fmt.Errorf("password cannot be empty")
		}

		passwordConfirmation, err := prompt.Secret(style.Dim.Sprint("Confirm password (input hidden): "))
		if err != nil {
			return fmt.Errorf("read password confirmation: %w", err)
		}

		if vaultPassword != passwordConfirmation {
			return fmt.Errorf("passwords do not match")
		}

		keyringUser := "vault:" + vaultName
		if err := keyring.Set("potok", keyringUser, vaultPassword); err != nil {
			return fmt.Errorf("failed to store password in keyring: %w", err)
		}

		cfg.AddVault(config.VaultInfo{
			Name: vaultName,
			Path: vaultPath,
		})

		if err := cfg.Save(); err != nil {
			_ = keyring.Delete("potok", keyringUser)
			return fmt.Errorf("failed to save config: %w", err)
		}

		// 6. Summary
		cmd.Println()
		cmd.Println("Vault registered successfully!")
		cmd.Printf("- Config: %s\n", config.Path())
		cmd.Printf("- Password: stored in OS keyring (service=\"potok\", user=%q)\n", keyringUser)
		cmd.Println()
		style.Dim.Fprintln(cmd.OutOrStdout(), "Use 'potok push "+vaultName+"' to upload or 'potok pull "+vaultName+"' to download.")

		return nil
	},
}
