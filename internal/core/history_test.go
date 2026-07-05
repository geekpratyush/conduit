package core

import (
	"path/filepath"
	"testing"
)

func TestHistoryAddSearchReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.db")
	h, err := OpenHistory(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer h.Close()

	id, err := h.Add(HistoryEntry{
		Protocol: "rest",
		Title:    "GET /users",
		Summary:  "GET https://api.example.com/users 200",
		Detail:   `{"method":"GET","url":"https://api.example.com/users"}`,
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := h.Add(HistoryEntry{
		Protocol: "kafka",
		Title:    "produce orders",
		Summary:  "produced to topic orders",
		Detail:   `{"topic":"orders"}`,
	}); err != nil {
		t.Fatalf("add 2: %v", err)
	}

	// FTS search should find the REST entry by a term in its summary.
	res, err := h.Search("users", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(res) != 1 || res[0].ID != id {
		t.Fatalf("search users: got %d rows %+v", len(res), res)
	}

	// Empty query returns most-recent-first (2 rows).
	all, err := h.Search("", 10)
	if err != nil {
		t.Fatalf("search all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(all))
	}

	// Favorite toggle + replay fetch.
	if err := h.SetFavorite(id, true); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	got, err := h.Get(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.Favorite {
		t.Fatal("favorite flag not persisted")
	}
	if got.Detail == "" {
		t.Fatal("replay detail missing")
	}
}
