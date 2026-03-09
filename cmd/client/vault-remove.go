package main

import (
	"fmt"

	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/style"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

//
// potok vault remove - Remove a vault from local config and optionally delete
// its encryption password from the OS keyring.
//
// Usage Examples:
//   potok vault-remove notes
//   potok vault-remove notes --keep-password
//
// Inputs:
//   - Args:
//     - Vault name (required, positional)
//

var removeVaultCmd = &cobra.Command{
	Use:   "vault-remove",
	Short: "Remove local vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if !cfg.HasVault(vaultName) {
			return fmt.Errorf("vault %q is not registered", vaultName)
		}

		keyringUser := "vault:" + vaultName
		if err := keyring.Delete("potok", keyringUser); err != nil {
			cmd.Printf("Warning: could not remove password from keyring: %v\n", err)
		}

		cfg.RemoveVault(vaultName)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		cmd.Printf("Vault %q removed from local config.\n", vaultName)
		style.Dim.Fprintln(cmd.OutOrStdout(), "Remote vault not affected. Use 'potok remote-delete "+vaultName+"' to delete from server.")

		return nil
	},
}
