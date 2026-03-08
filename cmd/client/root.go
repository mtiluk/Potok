package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "potok",
	Short: "Potok CLI for syncing Obsidian vaults",
}

func Execute() {
	red := color.New(color.FgRed, color.Bold)

	if err := rootCmd.Execute(); err != nil {
		red.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addVaultCmd)
	rootCmd.SilenceErrors = true
}
