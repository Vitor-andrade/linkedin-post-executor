package secret

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	c, err := LoadOrCreate(filepath.Join(t.TempDir(), "key"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	plain := []byte("a-linkedin-access-token")
	blob, err := c.Encrypt(plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Contains(blob, plain) {
		t.Fatal("ciphertext leaks the plaintext")
	}

	got, err := c.Decrypt(blob)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Errorf("round-trip mismatch: got %q", got)
	}
}

func TestDecryptRejectsTampering(t *testing.T) {
	c, _ := LoadOrCreate(filepath.Join(t.TempDir(), "key"))
	blob, _ := c.Encrypt([]byte("secret"))
	blob[len(blob)-1] ^= 0xff // flip a bit in the ciphertext
	if _, err := c.Decrypt(blob); err == nil {
		t.Error("expected authentication failure on tampered ciphertext")
	}
}

func TestKeyPersistsAcrossLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "key")
	c1, _ := LoadOrCreate(path)
	blob, _ := c1.Encrypt([]byte("payload"))

	c2, err := LoadOrCreate(path) // reuses the persisted key file
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, err := c2.Decrypt(blob)
	if err != nil {
		t.Fatalf("decrypt with reloaded key: %v", err)
	}
	if string(got) != "payload" {
		t.Errorf("got %q, want %q", got, "payload")
	}
}
