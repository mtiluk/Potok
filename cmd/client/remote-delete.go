package main

import (
	"github.com/spf13/cobra"
)

//
// potok remote-delete - Delete a vault from the server. Local config is not
// affected.
//
// Usage Examples:
//   potok remote-delete notes
//   potok remote-delete notes --yes
//
// Inputs:
//   - Args:
//     - Vault name (required, positional)
//   - Flags:
//     --yes    Skip confirmation prompt
//
// Requires:
//   - Server URL and API key (run 'potok init' first)
//

var remoteDeleteCmd = &cobra.Command{
	Use:   "remote-delete [name]",
	Short: "Delete a vault from the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}
