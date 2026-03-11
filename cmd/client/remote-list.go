package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/preflight"
	"github.com/michaeltukdev/Potok/internal/style"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
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
		style.Bold.Fprintln(cmd.OutOrStdout(), "Potok Remote Vaults")
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
			return fmt.Errorf("Run potok init first")
		}

		client := client.NewClient(cfg.ServerURL, apiKey)
		vaults, err := client.ListVaults()

		if len(vaults) == 0 {
			cmd.Println("No remote vaults.")
			return nil
		}

		// TODO: Implement correct syncing later
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Name\tCreated At\tLast Synced")
		fmt.Fprintln(w, "----\t----\t-----------")

		for _, v := range vaults {
			synced := "never"

			fmt.Fprintf(w, "%s\t%s\t%s\n", v.Name, v.CreatedAt, synced)
		}
		w.Flush()

		return nil
	},
}
