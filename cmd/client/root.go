package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "potok",
	Short: "Potok CLI for syncing Obsidian vaults",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	// rootCmd.AddCommand(syncCmd)
	// rootCmd.AddCommand(addVaultCmd)
	// rootCmd.AddCommand(uploadFileCmd)
	// rootCmd.AddCommand(restoreVaultCmd)
	// rootCmd.AddCommand(listVaultsCmd)
	// rootCmd.AddCommand(setApiUrlCmd)
	// rootCmd.AddCommand(setApiKeyCmd)
}
