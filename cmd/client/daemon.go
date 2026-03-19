package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	gosync "sync"
	"syscall"
	"time"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/crypto"
	potoksync "github.com/michaeltukdev/Potok/internal/sync"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var pollInterval time.Duration

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Watch and sync all registered vaults simultaneously",
	Args:  cobra.NoArgs,
	RunE:  runDaemon,
}

func init() {
	daemonCmd.Flags().DurationVar(
		&pollInterval, "interval", 30*time.Second,
		"How often to poll the server for remote changes",
	)
}

type vaultSyncer struct {
	name    string
	watcher *potoksync.Watcher
	handler *potoksync.EventHandler
	poller  *potoksync.Poller
}

func runDaemon(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if len(cfg.Vaults) == 0 {
		return fmt.Errorf("no vaults registered — run 'potok vault-add' first")
	}

	apiKey, err := keyring.Get("potok", "api-key")
	if err != nil || apiKey == "" {
		return fmt.Errorf("run 'potok init' first")
	}

	c := client.NewClient(cfg.ServerURL, apiKey)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	var syncers []vaultSyncer

	for _, vault := range cfg.Vaults {
		vlog := logger.With("vault", vault.Name)

		password, err := keyring.Get("potok", "vault:"+vault.Name)
		if err != nil || password == "" {
			vlog.Error("no password in keyring, skipping")
			continue
		}

		salt, err := c.DownloadFile(vault.Name, ".potok/salt")
		if err != nil {
			vlog.Error("no salt on server — run 'potok push' first", "err", err)
			continue
		}

		vlog.Info("salt downloaded",
			"len", len(salt),
			"hex", fmt.Sprintf("%x", salt),
		)

		encKey, err := crypto.DeriveKey([]byte(password), salt)
		if err != nil {
			vlog.Error("derive key failed", "err", err)
			continue
		}

		if _, err := os.Stat(vault.Path); os.IsNotExist(err) {
			vlog.Error("vault path does not exist", "path", vault.Path)
			continue
		}

		guard := potoksync.NewPullGuard()

		manifest, err := potoksync.LoadManifest(vault.Name)
		if err != nil {
			vlog.Error("load manifest failed", "err", err)
			continue
		}

		watcher, err := potoksync.NewWatcher(
			vault.Path, 500*time.Millisecond, vlog,
		)
		if err != nil {
			vlog.Error("create watcher failed", "err", err)
			continue
		}

		handler := potoksync.NewEventHandler(
			vault.Name, vault.Path, c, encKey, guard, manifest, vlog,
		)

		poller, err := potoksync.NewPoller(
			vault.Name, vault.Path, c, encKey, guard, vlog,
		)
		if err != nil {
			vlog.Error("create poller failed", "err", err)
			watcher.Close()
			continue
		}

		syncers = append(syncers, vaultSyncer{
			name:    vault.Name,
			watcher: watcher,
			handler: handler,
			poller:  poller,
		})

		vlog.Info("watching", "path", vault.Path)
	}

	if len(syncers) == 0 {
		return fmt.Errorf("no vaults could be initialised")
	}

	fmt.Printf(
		"Potok daemon started — watching %d vault(s), polling every %s\n",
		len(syncers), pollInterval,
	)
	fmt.Println("Press Ctrl+C to stop.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg gosync.WaitGroup
	var cfgMu gosync.Mutex
	stopCh := make(chan struct{})

	for _, s := range syncers {
		s := s

		// Tracked watcher goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.watcher.Start()
		}()

		// Event processing goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				case batch, ok := <-s.watcher.Events:
					if !ok {
						return
					}
					s.handler.HandleBatch(batch)

					cfgMu.Lock()
					cfg.UpdateLastSynced(
						s.name,
						time.Now().Format(time.RFC3339),
					)
					_ = cfg.Save()
					cfgMu.Unlock()
				case err, ok := <-s.watcher.Errors:
					if !ok {
						return
					}
					logger.Error("watcher error",
						"vault", s.name, "err", err,
					)
				}
			}
		}()

		// Poll goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(pollInterval)
			defer ticker.Stop()

			s.poller.Poll()

			for {
				select {
				case <-stopCh:
					return
				case <-ticker.C:
					s.poller.Poll()
				}
			}
		}()
	}

	<-sigCh
	fmt.Println("\nShutting down...")

	close(stopCh)
	for _, s := range syncers {
		s.watcher.Close()
	}
	wg.Wait()

	fmt.Println("All watchers stopped.")
	return nil
}
