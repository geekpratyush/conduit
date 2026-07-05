package security

import (
	"path/filepath"
	"testing"
)

func TestVaultRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	v := NewVault(path)

	if v.Exists() {
		t.Fatal("new vault should not exist yet")
	}
	if err := v.Initialize("master-pass"); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if !v.Exists() {
		t.Fatal("vault file should exist after initialize")
	}
	if v.Locked() {
		t.Fatal("vault should be unlocked after initialize")
	}
	if err := v.Put("rest/token", "s3cr3t"); err != nil {
		t.Fatalf("put: %v", err)
	}

	// Re-open with a fresh instance and the correct password.
	v2 := NewVault(path)
	if err := v2.Unlock("master-pass"); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	got, ok, err := v2.Get("rest/token")
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if got != "s3cr3t" {
		t.Fatalf("got %q, want %q", got, "s3cr3t")
	}
}

func TestVaultWrongPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	v := NewVault(path)
	if err := v.Initialize("correct"); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if err := v.Put("k", "v"); err != nil {
		t.Fatalf("put: %v", err)
	}

	v2 := NewVault(path)
	if err := v2.Unlock("wrong"); err != ErrBadPassword {
		t.Fatalf("expected ErrBadPassword, got %v", err)
	}
}

func TestVaultLockedOps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	v := NewVault(path)
	if err := v.Initialize("pw"); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	v.Lock()
	if !v.Locked() {
		t.Fatal("should be locked")
	}
	if err := v.Put("k", "v"); err != ErrLocked {
		t.Fatalf("expected ErrLocked, got %v", err)
	}
	if _, _, err := v.Get("k"); err != ErrLocked {
		t.Fatalf("expected ErrLocked on get, got %v", err)
	}
}
