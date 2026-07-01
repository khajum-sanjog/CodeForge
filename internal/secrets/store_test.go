package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "codeforge-secrets-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	secretsPath := filepath.Join(tempDir, "secrets.enc")

	// Set env var for testing to bypass prompt or random key generation if needed
	// we want to ensure custom key is generated/reused correctly.
	os.Setenv("CODEFORGE_MASTER_KEY", "12345678901234567890123456789012") // 32 bytes key
	defer os.Unsetenv("CODEFORGE_MASTER_KEY")

	store, err := LoadStore(secretsPath)
	if err != nil {
		t.Fatalf("failed to load secrets store: %v", err)
	}

	// Test Set and Get
	err = store.Set("MY_SECRET", "super-secret-value")
	if err != nil {
		t.Fatalf("failed to set secret: %v", err)
	}

	val, err := store.Get("MY_SECRET")
	if err != nil {
		t.Fatalf("failed to get secret: %v", err)
	}
	if val != "super-secret-value" {
		t.Errorf("expected 'super-secret-value', got %q", val)
	}

	// Test List
	err = store.Set("ANOTHER_SECRET", "another-value")
	if err != nil {
		t.Fatalf("failed to set secret: %v", err)
	}

	keys := store.List()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "ANOTHER_SECRET" || keys[1] != "MY_SECRET" {
		t.Errorf("unexpected key list: %+v", keys)
	}

	// Test Delete
	err = store.Delete("ANOTHER_SECRET")
	if err != nil {
		t.Fatalf("failed to delete secret: %v", err)
	}

	_, err = store.Get("ANOTHER_SECRET")
	if err == nil {
		t.Error("expected error getting deleted secret, but got nil")
	}

	// Reload store
	store2, err := LoadStore(secretsPath)
	if err != nil {
		t.Fatalf("failed to reload secrets store: %v", err)
	}

	val, err = store2.Get("MY_SECRET")
	if err != nil {
		t.Fatalf("failed to get secret after reload: %v", err)
	}
	if val != "super-secret-value" {
		t.Errorf("expected 'super-secret-value' after reload, got %q", val)
	}
}
