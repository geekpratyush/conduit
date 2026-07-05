// Package security holds Conduit's credential vault and certificate manager.
// Secrets are encrypted at rest with AES-256-GCM under a key derived from the
// user's master password via PBKDF2, and are never written in plaintext.
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// pbkdf2Iters is the PBKDF2 work factor used for master-key derivation.
	pbkdf2Iters = 200_000
	keyLen      = 32 // AES-256
	saltLen     = 16
	vaultMagic  = "GNGV1" // file format marker
)

// ErrLocked is returned when an operation needs an unlocked vault.
var ErrLocked = errors.New("vault: locked")

// ErrBadPassword is returned when decryption fails (wrong master password or
// tampered file).
var ErrBadPassword = errors.New("vault: incorrect master password or corrupt vault")

// vaultFile is the on-disk envelope: a salt + AES-GCM nonce + ciphertext of the
// JSON-encoded secret map. The magic and salt are stored in the clear; the
// secrets are not.
type vaultFile struct {
	Magic      string `json:"magic"`
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

// CredentialVault stores named secrets encrypted at rest. It is safe for
// concurrent use. Callers reference a secret by key (a "vault ref"); saved
// connection profiles store only these refs, never the secret value.
type CredentialVault struct {
	path string

	mu      sync.RWMutex
	key     []byte            // derived key while unlocked; nil when locked
	salt    []byte            // persisted salt (stable across unlocks)
	secrets map[string]string // plaintext secrets, only present while unlocked
}

// NewVault returns a vault backed by the file at path. The vault starts locked;
// call Unlock (existing file) or Initialize (new file) before use.
func NewVault(path string) *CredentialVault {
	return &CredentialVault{path: path}
}

// Exists reports whether a vault file is already present on disk.
func (v *CredentialVault) Exists() bool {
	_, err := os.Stat(v.path)
	return err == nil
}

// Locked reports whether the vault currently holds no derived key.
func (v *CredentialVault) Locked() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.key == nil
}

// Initialize creates a brand-new, empty vault protected by password, failing if
// one already exists. It leaves the vault unlocked.
func (v *CredentialVault) Initialize(password string) error {
	if v.Exists() {
		return fmt.Errorf("vault: already initialized at %s", v.path)
	}
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}
	v.mu.Lock()
	v.salt = salt
	v.key = deriveKey(password, salt)
	v.secrets = make(map[string]string)
	v.mu.Unlock()
	return v.persist()
}

// Unlock derives the key from password and loads+decrypts the existing vault.
func (v *CredentialVault) Unlock(password string) error {
	raw, err := os.ReadFile(v.path)
	if err != nil {
		return err
	}
	var vf vaultFile
	if err := json.Unmarshal(raw, &vf); err != nil {
		return fmt.Errorf("vault: parse: %w", err)
	}
	if vf.Magic != vaultMagic {
		return errors.New("vault: unrecognised file format")
	}
	key := deriveKey(password, vf.Salt)
	plain, err := aesGCMOpen(key, vf.Nonce, vf.Ciphertext)
	if err != nil {
		return ErrBadPassword
	}
	secrets := make(map[string]string)
	if len(plain) > 0 {
		if err := json.Unmarshal(plain, &secrets); err != nil {
			return fmt.Errorf("vault: decode secrets: %w", err)
		}
	}
	v.mu.Lock()
	v.salt = vf.Salt
	v.key = key
	v.secrets = secrets
	v.mu.Unlock()
	return nil
}

// Lock discards the derived key and cached secrets from memory.
func (v *CredentialVault) Lock() {
	v.mu.Lock()
	v.key = nil
	v.secrets = nil
	v.mu.Unlock()
}

// Put stores (or replaces) a secret and persists the vault.
func (v *CredentialVault) Put(key, secret string) error {
	v.mu.Lock()
	if v.key == nil {
		v.mu.Unlock()
		return ErrLocked
	}
	v.secrets[key] = secret
	v.mu.Unlock()
	return v.persist()
}

// Get returns the secret for key and whether it was present.
func (v *CredentialVault) Get(key string) (string, bool, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.key == nil {
		return "", false, ErrLocked
	}
	s, ok := v.secrets[key]
	return s, ok, nil
}

// Delete removes a secret and persists the vault. Removing an absent key is a
// no-op.
func (v *CredentialVault) Delete(key string) error {
	v.mu.Lock()
	if v.key == nil {
		v.mu.Unlock()
		return ErrLocked
	}
	delete(v.secrets, key)
	v.mu.Unlock()
	return v.persist()
}

// Keys returns the sorted-insensitive list of stored secret keys (not values).
func (v *CredentialVault) Keys() ([]string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.key == nil {
		return nil, ErrLocked
	}
	keys := make([]string, 0, len(v.secrets))
	for k := range v.secrets {
		keys = append(keys, k)
	}
	return keys, nil
}

// persist encrypts the current secret map and atomically writes the vault file.
func (v *CredentialVault) persist() error {
	v.mu.RLock()
	if v.key == nil {
		v.mu.RUnlock()
		return ErrLocked
	}
	plain, err := json.Marshal(v.secrets)
	key := append([]byte(nil), v.key...)
	salt := append([]byte(nil), v.salt...)
	v.mu.RUnlock()
	if err != nil {
		return err
	}

	nonce, ct, err := aesGCMSeal(key, plain)
	if err != nil {
		return err
	}
	vf := vaultFile{Magic: vaultMagic, Salt: salt, Nonce: nonce, Ciphertext: ct}
	blob, err := json.Marshal(vf)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(v.path), 0o700); err != nil {
		return err
	}
	tmp := v.path + ".tmp"
	if err := os.WriteFile(tmp, blob, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, v.path)
}

func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iters, keyLen, sha256.New)
}

func aesGCMSeal(key, plaintext []byte) (nonce, ciphertext []byte, err error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	return nonce, gcm.Seal(nil, nonce, plaintext, nil), nil
}

func aesGCMOpen(key, nonce, ciphertext []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
