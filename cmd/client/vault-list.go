package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/spf13/cobra"
)

//
// potok vault-list - Lists all vaults registered locally on this device.
//
// Usage Examples:
//   potok vault-list
//
// Output:
//   - Name, local path, and last sync time for each vault
//   - Shows "never" if vault has not been synced
//   - Shows helpful message if no vaults are registered
//

var listVaultsCmd = &cobra.Command{
	Use:   "vault-list",
	Short: "Lists vaults known locally",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Vaults) == 0 {
			cmd.Println("No vaults registered. Run 'potok vault-add' to get started.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Name\tPath\tLast Synced")
		fmt.Fprintln(w, "----\t----\t-----------")

		for _, v := range cfg.Vaults {
			synced := "never"
			if v.LastSynced != "" {
				synced = v.LastSynced
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", v.Name, v.Path, synced)
		}
		w.Flush()

		return nil
	},
}
