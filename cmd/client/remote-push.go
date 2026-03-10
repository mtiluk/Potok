package main

import (
	"github.com/spf13/cobra"
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

		return nil
	},
}
