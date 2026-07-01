package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// SecretStore manages AES-256-GCM encrypted secrets saved to a file.
type SecretStore struct {
	mu       sync.RWMutex
	filePath string
	secrets  map[string]string // key -> encrypted base64 payload
	key      []byte
}

// ResolvePath replaces standard home directory shortcuts like "~" with the absolute home directory path.
func ResolvePath(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// GetMasterKey retrieves the encryption key from an env variable, the master.key file, or generates a new one.
func GetMasterKey() ([]byte, error) {
	// 1. Check environment variable
	if envKey := os.Getenv("CODEFORGE_MASTER_KEY"); envKey != "" {
		// If base64 encoded
		if dec, err := base64.StdEncoding.DecodeString(envKey); err == nil && len(dec) == 32 {
			return dec, nil
		}
		// If raw 32 bytes
		if len(envKey) == 32 {
			return []byte(envKey), nil
		}
	}

	// 2. Check local config directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(home, ".codeforge")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	keyPath := filepath.Join(configDir, "master.key")
	if data, err := os.ReadFile(keyPath); err == nil {
		trimmed := strings.TrimSpace(string(data))
		if dec, err := base64.StdEncoding.DecodeString(trimmed); err == nil && len(dec) == 32 {
			return dec, nil
		}
		if len(trimmed) == 32 {
			return []byte(trimmed), nil
		}
	}

	// 3. Generate a new key and write it with 0600 permissions
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(keyPath, []byte(encodedKey), 0600); err != nil {
		return nil, err
	}

	return key, nil
}

// LoadStore loads the encrypted secret file or initializes a new one.
func LoadStore(path string) (*SecretStore, error) {
	resolvedPath := ResolvePath(path)
	key, err := GetMasterKey()
	if err != nil {
		return nil, err
	}

	store := &SecretStore{
		filePath: resolvedPath,
		secrets:  make(map[string]string),
		key:      key,
	}

	// Read existing secrets if the file exists
	if _, err := os.Stat(resolvedPath); err == nil {
		data, err := os.ReadFile(resolvedPath)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &store.secrets); err != nil {
				return nil, err
			}
		}
	}

	return store, nil
}

// Set encrypts and updates a secret key-value pair.
func (s *SecretStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encryptedVal, err := s.encrypt(value)
	if err != nil {
		return err
	}

	s.secrets[key] = encryptedVal
	return s.save()
}

// Get decrypts and returns a secret value by key.
func (s *SecretStore) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	encryptedVal, ok := s.secrets[key]
	if !ok {
		return "", errors.New("secret key not found")
	}

	return s.decrypt(encryptedVal)
}

// Delete removes a secret key from the store.
func (s *SecretStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.secrets[key]; !ok {
		return errors.New("secret key not found")
	}

	delete(s.secrets, key)
	return s.save()
}

// List returns a sorted slice of all secret keys.
func (s *SecretStore) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.secrets))
	for k := range s.secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *SecretStore) encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *SecretStore) decrypt(cryptoText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (s *SecretStore) save() error {
	data, err := json.MarshalIndent(s.secrets, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Atomic write using temp file and rename
	tmpFile, err := os.CreateTemp(dir, "secrets-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpFile.Name(), s.filePath)
}
