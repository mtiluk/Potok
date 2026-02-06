package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/michaeltukdev/Potok/internal/prompt"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

// potok init - Initialises Potok on this device by collecting the required configuration settings.
//
// Usage Examples:
//
// 	potok init
//	potok init --url http://localhost:8080
//
// Inputs:
//
// 	- Prompts:
//		- Server URL (if --url not provided)
//		- API key (if not already stored)
//	- Flags:
//		--url string   Potok server base URL
//

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialises Potok on this device by collecting the required configuration settings.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Potok setup (device initialisation)")
		cmd.Println()

		flagURL, err := cmd.Flags().GetString("url")
		if err != nil {
			return fmt.Errorf("read --url flag: %w", err)
		}

		var url string
		if strings.TrimSpace(flagURL) != "" {
			url = strings.TrimSpace(flagURL)
		} else {
			url, err = prompt.InputDefault("Server URL", "http://localhost:8080")
			if err != nil {
				return fmt.Errorf("read server url: %w", err)
			}
		}

		url = strings.TrimSpace(url)
		url = strings.TrimSuffix(url, "/")

		apiKey, err := prompt.Secret("API key (input hidden): ")
		if err != nil {
			return fmt.Errorf("read api key: %w", err)
		}

		resp, err := client.MakeAuthenticatedRequest(apiKey, url+"/me")
		if err != nil {
			return fmt.Errorf("authenticate with server: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf(
				"authentication failed (status %d): check API key and ensure the server is running",
				resp.StatusCode,
			)
		}

		var me struct {
			Username string `json:"username"`
			ID       int    `json:"id"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
			return fmt.Errorf("decode /me response: %w", err)
		}

		cmd.Println("Authentication request success!")

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		cfg.APIURL = url
		cfg.Username = me.Username

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		if err := keyring.Set("potok", "api-key", apiKey); err != nil {
			return fmt.Errorf("save api key to keyring: %w", err)
		}

		cmd.Println()
		cmd.Println("Initialized.")
		cmd.Printf("Config saved (api_url): %s\n", cfg.APIURL)
		cmd.Printf("Signed in as: %s\n", cfg.Username)
		cmd.Println(`API key stored in OS keyring (service="potok", user="api-key").`)

		return nil
	},
}

func init() {
	initCmd.Flags().String("url", "", "Potok server base URL")
}
