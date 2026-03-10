package main

import (
	"github.com/spf13/cobra"
)

//
// potok remote-list - Lists all vaults available on the server for the current
// user / API key.
//
// Usage Examples:
//   potok remote-list
//
// Output:
//   - Vault name, last updated time, and size for each remote vault
//   - Shows helpful message if no remote vaults exist
//
// Requires:
//   - Server URL and API key (run 'potok init' first)
//

var remoteListCmd = &cobra.Command{
	Use:   "remote-list",
	Short: "Lists vaults available on the server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}
