package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Vault struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Client) ListVaults() ([]Vault, error) {
	resp, err := c.request(http.MethodGet, "/vaults", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}

	var vaults []Vault
	if err := json.NewDecoder(resp.Body).Decode(&vaults); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return vaults, nil
}
