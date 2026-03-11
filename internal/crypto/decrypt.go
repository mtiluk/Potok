package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// DecryptBytes decrypts data produced by EncryptBytes.
// Input: nonce (12B) || ciphertext
func DecryptBytes(key, data []byte) ([]byte, error) {
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
