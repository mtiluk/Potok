package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

var Prefix = "potok_"

const keyBytes = 32

func GenerateAPIKey() (key string, err error) {
	raw := make([]byte, keyBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("auth: generate key: %w", err)
	}
	key = Prefix + base64.RawURLEncoding.EncodeToString(raw)
	return key, nil
}
