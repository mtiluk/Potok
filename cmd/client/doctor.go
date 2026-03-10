package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/style"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

//
// potok doctor - Run diagnostics to verify Potok is configured correctly.
//
// Usage Examples:
//   potok doctor
//
// Checks:
//   - Config file is readable
//   - Server URL is set
//   - API key is present in OS keyring
//   - Server is reachable
//   - Registered vault paths exist
//   - Vault passwords are present in OS keyring
//   - File permissions are correct
//

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics to verify Potok is configured correctly",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		w := cmd.OutOrStdout()
		style.Bold.Fprintln(w, "Potok Doctor")
		cmd.Println()

		r := &report{}

		// -- Config --
		cmd.Println("Config")

		cfg, err := config.Load()
		if err != nil {
			r.fail(w, "Config file readable", "Failed to load config — run 'potok init'")
			cmd.Println()
			cmd.Printf("Summary: %d passed, %d failed, %d skipped\n", r.passed, r.failed, r.skipped)
			if r.failed > 0 {
				os.Exit(1)
			}
			return nil
		}
		r.pass(w, "Config file readable", config.Path())

		serverReachable := false
		if cfg.ServerURL == "" {
			r.fail(w, "Server URL set", "Run 'potok init'")
		} else {
			r.pass(w, "Server URL set", cfg.ServerURL)
		}

		cmd.Println()

		// -- Server --
		cmd.Println("Server")

		if cfg.ServerURL != "" {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(cfg.ServerURL)
			if err != nil {
				r.fail(w, "Server reachable", "Check URL or server status")
			} else {
				resp.Body.Close()
				r.pass(w, "Server reachable", fmt.Sprintf("HTTP %d", resp.StatusCode))
				serverReachable = true
			}
		} else {
			r.skip(w, "Server reachable", "Skipped (no server URL)")
		}

		apiKey, err := keyring.Get("potok", "api-key")
		if err != nil || apiKey == "" {
			r.fail(w, "API key in keyring", "Run 'potok init'")
		} else {
			r.pass(w, "API key in keyring", "Found")
		}

		if !serverReachable {
			r.skip(w, "API key valid", "Skipped (server unreachable)")
		} else if apiKey == "" {
			r.skip(w, "API key valid", "Skipped (no API key)")
		} else {
			// TODO: make an authenticated request to validate the key
			r.skip(w, "API key valid", "Not yet implemented")
		}

		cmd.Println()

		// -- Vaults --
		cmd.Println("Vaults")

		if len(cfg.Vaults) == 0 {
			style.Dim.Fprintln(w, "  No vaults registered")
		}

		for _, v := range cfg.Vaults {
			cmd.Printf("  %s\n", v.Name)

			info, err := os.Stat(v.Path)
			if err != nil {
				r.fail(w, "  Local path exists", v.Path+" — path missing or moved")
				r.skip(w, "  Path is a directory", "Skipped (path missing)")
				r.skip(w, "  Path is readable", "Skipped (path missing)")
			} else {
				r.pass(w, "  Local path exists", v.Path)

				if !info.IsDir() {
					r.fail(w, "  Path is a directory", "Path is a file, not a directory")
				} else {
					r.pass(w, "  Path is a directory", "")
				}

				f, err := os.Open(v.Path)
				if err != nil {
					r.fail(w, "  Path is readable", "Check permissions")
				} else {
					f.Close()
					r.pass(w, "  Path is readable", "")
				}
			}

			_, err = keyring.Get("potok", "vault:"+v.Name)
			if err != nil {
				r.fail(w, "  Password in keyring", "Re-run 'potok vault-add'")
			} else {
				r.pass(w, "  Password in keyring", "Found")
			}
		}

		cmd.Println()
		cmd.Printf("Summary: %d passed, %d failed, %d skipped\n", r.passed, r.failed, r.skipped)

		if r.failed > 0 {
			os.Exit(1)
		}

		return nil
	},
}

type report struct {
	passed  int
	failed  int
	skipped int
}

func (r *report) pass(w io.Writer, name, detail string) {
	style.Green.Fprintf(w, "  ✓ %-30s %s\n", name, detail)
	r.passed++
}

func (r *report) fail(w io.Writer, name, fix string) {
	style.Red.Fprintf(w, "  ✗ %-30s %s\n", name, fix)
	r.failed++
}

func (r *report) skip(w io.Writer, name, reason string) {
	style.Dim.Fprintf(w, "  ○ %-30s %s\n", name, reason)
	r.skipped++
}
