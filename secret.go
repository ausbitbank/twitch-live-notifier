package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	secretStoreVersion = 1
	secretStoreFile    = "secrets.json"
	secretKeyFile      = "secret.key"
	secretKeySize      = 32
)

// secretProtector is the platform-specific at-rest encryption mechanism. On
// Windows it is DPAPI; on macOS/Linux it is AES-GCM with a key held by the
// platform keyring (or a 0600 fallback key file when no keyring is available).
type secretProtector interface {
	Protect(plaintext []byte) ([]byte, error)
	Unprotect(ciphertext []byte) ([]byte, error)
	Description() string
}

type secretFile struct {
	Version int               `json:"version"`
	Entries map[string]string `json:"entries"`
}

type secretStore struct {
	dir       string
	path      string
	entries   map[string]string
	protector secretProtector
}

func openSecretStore(dir string) (*secretStore, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, errors.New("secrets directory is empty")
	}
	protector, err := newPlatformSecretProtector(dir)
	if err != nil {
		return nil, err
	}
	store := &secretStore{
		dir:       dir,
		path:      filepath.Join(dir, secretStoreFile),
		entries:   map[string]string{},
		protector: protector,
	}
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read the secrets file: %w", err)
	}
	var file secretFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, errors.New("the secrets file is not valid JSON")
	}
	if file.Entries != nil {
		store.entries = file.Entries
	}
	return store, nil
}

func (s *secretStore) Set(id, plaintext string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("secret ID is empty")
	}
	ciphertext, err := s.protector.Protect([]byte(plaintext))
	if err != nil {
		return err
	}
	s.entries[id] = base64.StdEncoding.EncodeToString(ciphertext)
	return nil
}

func (s *secretStore) Get(id string) (string, error) {
	value, ok := s.entries[strings.TrimSpace(id)]
	if !ok {
		return "", errors.New("secret was not found")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", errors.New("stored secret is invalid")
	}
	plaintext, err := s.protector.Unprotect(ciphertext)
	if err != nil {
		return "", errors.New("stored secret could not be unlocked")
	}
	return string(plaintext), nil
}

func (s *secretStore) Delete(id string) {
	delete(s.entries, strings.TrimSpace(id))
}

func (s *secretStore) Save() error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return errors.New("could not create the settings folder")
	}
	data, err := json.MarshalIndent(secretFile{Version: secretStoreVersion, Entries: s.entries}, "", "  ")
	if err != nil {
		return errors.New("could not encode secrets")
	}
	data = append(data, '\n')
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return errors.New("could not write the secrets file")
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return errors.New("could not save the secrets file")
	}
	return nil
}

func (s *secretStore) Description() string {
	return s.protector.Description()
}

func randomSecretKey() ([]byte, error) {
	key := make([]byte, secretKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, errors.New("could not generate a secret key")
	}
	return key, nil
}

type aesGCMProtector struct {
	gcm cipher.AEAD
}

func newAESGCMProtector(key []byte) (secretProtector, error) {
	if len(key) != secretKeySize {
		return nil, errors.New("secret key has the wrong length")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("could not initialise secret encryption")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("could not initialise secret encryption")
	}
	return &aesGCMProtector{gcm: gcm}, nil
}

func (p *aesGCMProtector) Protect(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, p.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, errors.New("could not encrypt the secret")
	}
	return p.gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (p *aesGCMProtector) Unprotect(ciphertext []byte) ([]byte, error) {
	nonceSize := p.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("stored secret is truncated")
	}
	return p.gcm.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
}

func (p *aesGCMProtector) Description() string {
	return "aes-gcm"
}

func newKeyFileProtector(dir string) (secretProtector, error) {
	path := filepath.Join(dir, secretKeyFile)
	key, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		key, err = randomSecretKey()
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, errors.New("could not create the settings folder")
		}
		if err := os.WriteFile(path, key, 0o600); err != nil {
			return nil, errors.New("could not write the fallback secret key")
		}
	} else if err != nil {
		return nil, fmt.Errorf("could not read the fallback secret key: %w", err)
	}
	if len(key) != secretKeySize {
		return nil, errors.New("the fallback secret key has the wrong length")
	}
	return newAESGCMProtector(key)
}
