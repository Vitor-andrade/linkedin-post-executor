// Package secret provides authenticated encryption for data at rest (the
// OAuth tokens stored in SQLite). It uses AES-256-GCM with a key that is
// auto-generated on first run and persisted in a 0600 file, keeping the app
// zero-config and local-first (see the security section of the architecture
// plan).
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const keySize = 32 // AES-256

// Cipher encrypts and decrypts small secrets with AES-256-GCM.
type Cipher struct {
	aead cipher.AEAD
}

// LoadOrCreate returns a Cipher backed by the key at path, generating and
// persisting a new random key (0600) if the file does not yet exist.
func LoadOrCreate(path string) (*Cipher, error) {
	key, err := loadOrCreateKey(path)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt seals plaintext, returning nonce||ciphertext.
func (c *Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return c.aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt opens a nonce||ciphertext blob produced by Encrypt.
func (c *Cipher) Decrypt(blob []byte) ([]byte, error) {
	ns := c.aead.NonceSize()
	if len(blob) < ns {
		return nil, errors.New("secret: ciphertext too short")
	}
	return c.aead.Open(nil, blob[:ns], blob[ns:], nil)
}

func loadOrCreateKey(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err == nil {
		key, derr := hex.DecodeString(strings.TrimSpace(string(raw)))
		if derr != nil || len(key) != keySize {
			return nil, fmt.Errorf("secret: invalid key file %s", path)
		}
		return key, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// First run: generate a fresh key and persist it with tight permissions.
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, err
		}
	}
	if err := os.WriteFile(path, []byte(hex.EncodeToString(key)), 0o600); err != nil {
		return nil, err
	}
	return key, nil
}
