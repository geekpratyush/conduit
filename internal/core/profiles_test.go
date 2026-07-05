package core

import (
	"path/filepath"
	"testing"

	"github.com/geekpratyush/conduit/internal/plugin"
)

func TestConnectionStoreSeedsSamples(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.json")
	s, err := OpenConnectionStore(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if len(s.List()) == 0 {
		t.Fatal("expected seeded sample profiles on first open")
	}
	// Samples must never carry a plaintext secret.
	for _, c := range s.List() {
		if c.Secret != "" {
			t.Fatalf("sample %q has a non-empty Secret; profiles must store vault refs only", c.ID)
		}
	}
}

func TestConnectionStoreCRUDAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.json")
	s, err := OpenConnectionStore(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	cfg := plugin.ConnectionConfig{
		ID: "my-api", Name: "My API", Protocol: "rest",
		URL: "https://api.example.com", Auth: plugin.AuthBearer,
		Secret: "vault:rest/my-api-token", // a vault ref, not the token itself
	}
	if err := s.Save(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Reload from disk with a fresh instance.
	s2, err := OpenConnectionStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got, err := s2.Get("my-api")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.URL != cfg.URL || got.Secret != cfg.Secret {
		t.Fatalf("roundtrip mismatch: got %+v", got)
	}

	if err := s2.Delete("my-api"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s2.Get("my-api"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestConnectionStoreSaveRequiresID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.json")
	s, _ := OpenConnectionStore(path)
	if err := s.Save(plugin.ConnectionConfig{Name: "no id"}); err == nil {
		t.Fatal("expected error saving profile with empty ID")
	}
}
