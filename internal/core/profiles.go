package core

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"

	"github.com/geekpratyush/conduit/internal/plugin"
)

// ConnectionStore persists saved connection profiles as JSON at
// ~/.conduit/connections.json. Profiles never hold plaintext secrets: a
// profile's Secret field carries a vault reference (a key into the
// CredentialVault), so the on-disk file is safe to sync or inspect.
//
// The store seeds a set of deletable public sample endpoints on first use so a
// fresh install has something to click.
type ConnectionStore struct {
	path string

	mu       sync.RWMutex
	profiles map[string]plugin.ConnectionConfig // keyed by ID
}

// ErrNotFound is returned when a profile ID is unknown.
var ErrNotFound = errors.New("connection profile not found")

// OpenConnectionStore loads the store at path (creating and seeding it with
// public samples if the file does not yet exist).
func OpenConnectionStore(path string) (*ConnectionStore, error) {
	s := &ConnectionStore{path: path, profiles: make(map[string]plugin.ConnectionConfig)}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		s.seedSamples()
		if err := s.persist(); err != nil {
			return nil, err
		}
		return s, nil
	} else if err != nil {
		return nil, err
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *ConnectionStore) load() error {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var list []plugin.ConnectionConfig
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &list); err != nil {
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profiles = make(map[string]plugin.ConnectionConfig, len(list))
	for _, c := range list {
		s.profiles[c.ID] = c
	}
	return nil
}

func (s *ConnectionStore) persist() error {
	s.mu.RLock()
	list := make([]plugin.ConnectionConfig, 0, len(s.profiles))
	for _, c := range s.profiles {
		list = append(list, c)
	}
	s.mu.RUnlock()
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })

	blob, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, blob, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Save inserts or updates a profile (by ID) and persists.
func (s *ConnectionStore) Save(c plugin.ConnectionConfig) error {
	if c.ID == "" {
		return errors.New("connection profile requires an ID")
	}
	s.mu.Lock()
	s.profiles[c.ID] = c
	s.mu.Unlock()
	return s.persist()
}

// Get returns the profile for id.
func (s *ConnectionStore) Get(id string) (plugin.ConnectionConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.profiles[id]
	if !ok {
		return plugin.ConnectionConfig{}, ErrNotFound
	}
	return c, nil
}

// Delete removes a profile and persists. Deleting an unknown ID is a no-op.
func (s *ConnectionStore) Delete(id string) error {
	s.mu.Lock()
	_, existed := s.profiles[id]
	delete(s.profiles, id)
	s.mu.Unlock()
	if !existed {
		return nil
	}
	return s.persist()
}

// List returns all profiles sorted by name.
func (s *ConnectionStore) List() []plugin.ConnectionConfig {
	s.mu.RLock()
	out := make([]plugin.ConnectionConfig, 0, len(s.profiles))
	for _, c := range s.profiles {
		out = append(out, c)
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// seedSamples populates deletable public endpoints so a fresh install is
// immediately explorable. These carry no secrets.
func (s *ConnectionStore) seedSamples() {
	samples := []plugin.ConnectionConfig{
		{ID: "sample-rest-httpbin", Name: "Sample · httpbin (REST)", Protocol: "rest",
			URL: "https://httpbin.org/get", Auth: plugin.AuthNone},
		{ID: "sample-graphql-countries", Name: "Sample · Countries (GraphQL)", Protocol: "graphql",
			URL: "https://countries.trevorblades.com/", Auth: plugin.AuthNone},
		{ID: "sample-sse-wikimedia", Name: "Sample · Wikimedia firehose (SSE)", Protocol: "sse",
			URL: "https://stream.wikimedia.org/v2/stream/recentchange", Auth: plugin.AuthNone},
		{ID: "sample-sftp-rebex", Name: "Sample · Rebex test server (SFTP)", Protocol: "sftp",
			Host: "test.rebex.net", Port: 22, Username: "demo", Auth: plugin.AuthBasic,
			Extras: map[string]string{"sample": "true"}},
		{ID: "sample-grpc-grpcbin", Name: "Sample · grpcb.in (gRPC)", Protocol: "grpc",
			Host: "grpcb.in", Port: 9000, Auth: plugin.AuthNone},
	}
	for _, c := range samples {
		s.profiles[c.ID] = c
	}
}
