package main

import (
	"os"

	"github.com/michaeltukdev/Potok/internal/style"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "potok",
	Short: "Potok CLI for syncing Obsidian vaults",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		style.Red.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addVaultCmd)
	rootCmd.SilenceErrors = true
}
