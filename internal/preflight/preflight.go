package preflight

import (
	"fmt"

	"github.com/michaeltukdev/Potok/internal/config"
	"github.com/zalando/go-keyring"
)

type Status struct {
	URL    string
	APIKey string
}

func Check() (*Status, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("server URL not set — run `potok init` first")
	}

	apiKey, err := keyring.Get("potok", "api-key")
	if err != nil || apiKey == "" {
		return nil, fmt.Errorf("API key not found — run `potok init` first")
	}

	return &Status{
		URL:    cfg.ServerURL,
		APIKey: apiKey,
	}, nil
}
