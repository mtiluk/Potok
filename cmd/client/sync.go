// cmd/sync.go
package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/crypto"
	potoksync "github.com/michaeltukdev/Potok/internal/sync"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var syncCmd = &cobra.Command{
	Use:   "sync [vault-name]",
	Short: "Watch a vault for changes and sync them",
	Args:  cobra.ExactArgs(1),
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	vaultName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	vaultPath := cfg.VaultPath(vaultName)
	if vaultPath == "" {
		return fmt.Errorf(
			"vault %q not found in config — run 'potok vault-add' first",
			vaultName,
		)
	}

	apiKey, err := keyring.Get("potok", "api-key")
	if err != nil || apiKey == "" {
		return fmt.Errorf("run 'potok init' first")
	}

	vaultPassword, err := keyring.Get("potok", "vault:"+vaultName)
	if err != nil || vaultPassword == "" {
		return fmt.Errorf(
			"no password found for vault %q — run 'potok vault-add' first",
			vaultName,
		)
	}

	c := client.NewClient(cfg.ServerURL, apiKey)

	salt, err := c.DownloadFile(vaultName, ".potok/salt")
	if err != nil {
		return fmt.Errorf(
			"download salt: %w (run 'potok push %s' first to initialise the vault)",
			err, vaultName,
		)
	}

	encKey, err := crypto.DeriveKey([]byte(vaultPassword), salt)
	if err != nil {
		return fmt.Errorf("derive key: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	watcher, err := potoksync.NewWatcher(
		vaultPath, 500*time.Millisecond, logger,
	)
	if err != nil {
		return fmt.Errorf("start watcher: %w", err)
	}
	defer watcher.Close()

	guard := potoksync.NewPullGuard()

	manifest, err := potoksync.LoadManifest(vaultName)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	handler := potoksync.NewEventHandler(
		vaultName, vaultPath, c, encKey, guard, manifest, logger,
	)

	fmt.Printf("Syncing vault %q at %s\n", vaultName, vaultPath)
	fmt.Println("Press Ctrl+C to stop.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go watcher.Start()

	for {
		select {
		case batch, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			handler.HandleBatch(batch)

			cfg.UpdateLastSynced(vaultName, time.Now().Format(time.RFC3339))
			_ = cfg.Save()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logger.Error("watcher error", "err", err)

		case <-sigCh:
			fmt.Println("\nShutting down...")
			return nil
		}
	}
}
