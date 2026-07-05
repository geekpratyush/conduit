package core

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite" // pure-Go, cgo-free SQLite driver (FTS5 enabled)
)

// HistoryEntry is one recorded interaction (a REST call, a query, a produced
// message, …). Summary/Detail are indexed for full-text search.
type HistoryEntry struct {
	ID        int64
	Protocol  string
	Title     string
	Summary   string    // short, searchable one-liner (e.g. "GET /users 200")
	Detail    string    // full replayable payload (JSON), searchable
	Favorite  bool
	CreatedAt time.Time
}

// HistoryStore persists interactions to SQLite with an FTS5 full-text index:
// full-text search, favorites, and one-click replay. It uses the pure-Go
// modernc driver so static, CGO_ENABLED=0 builds keep working.
type HistoryStore struct {
	db *sql.DB
}

// OpenHistory opens (creating if needed) the history database at path and
// ensures the schema exists.
func OpenHistory(path string) (*HistoryStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: serialise writers
	h := &HistoryStore{db: db}
	if err := h.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return h, nil
}

func (h *HistoryStore) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS history (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    protocol   TEXT NOT NULL,
    title      TEXT NOT NULL,
    summary    TEXT NOT NULL,
    detail     TEXT NOT NULL,
    favorite   INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL
);
CREATE VIRTUAL TABLE IF NOT EXISTS history_fts USING fts5(
    title, summary, detail,
    content='history', content_rowid='id'
);
CREATE TRIGGER IF NOT EXISTS history_ai AFTER INSERT ON history BEGIN
    INSERT INTO history_fts(rowid, title, summary, detail)
    VALUES (new.id, new.title, new.summary, new.detail);
END;
CREATE TRIGGER IF NOT EXISTS history_ad AFTER DELETE ON history BEGIN
    INSERT INTO history_fts(history_fts, rowid, title, summary, detail)
    VALUES('delete', old.id, old.title, old.summary, old.detail);
END;
`
	_, err := h.db.Exec(schema)
	return err
}

// Add records a new entry and returns its assigned ID.
func (h *HistoryStore) Add(e HistoryEntry) (int64, error) {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	res, err := h.db.Exec(
		`INSERT INTO history(protocol, title, summary, detail, favorite, created_at)
		 VALUES(?,?,?,?,?,?)`,
		e.Protocol, e.Title, e.Summary, e.Detail, boolToInt(e.Favorite), e.CreatedAt.UnixMilli(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Search returns entries matching an FTS5 query (most recent first). An empty
// query returns the most recent entries unfiltered.
func (h *HistoryStore) Search(query string, limit int) ([]HistoryEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	var (
		rows *sql.Rows
		err  error
	)
	if query == "" {
		rows, err = h.db.Query(
			`SELECT id, protocol, title, summary, detail, favorite, created_at
			 FROM history ORDER BY created_at DESC LIMIT ?`, limit)
	} else {
		rows, err = h.db.Query(
			`SELECT h.id, h.protocol, h.title, h.summary, h.detail, h.favorite, h.created_at
			 FROM history_fts f JOIN history h ON h.id = f.rowid
			 WHERE history_fts MATCH ? ORDER BY h.created_at DESC LIMIT ?`, query, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HistoryEntry
	for rows.Next() {
		var (
			e  HistoryEntry
			fav int
			ms int64
		)
		if err := rows.Scan(&e.ID, &e.Protocol, &e.Title, &e.Summary, &e.Detail, &fav, &ms); err != nil {
			return nil, err
		}
		e.Favorite = fav != 0
		e.CreatedAt = time.UnixMilli(ms)
		out = append(out, e)
	}
	return out, rows.Err()
}

// SetFavorite toggles the favorite flag on an entry.
func (h *HistoryStore) SetFavorite(id int64, fav bool) error {
	_, err := h.db.Exec(`UPDATE history SET favorite=? WHERE id=?`, boolToInt(fav), id)
	return err
}

// Get returns a single entry by ID (for replay).
func (h *HistoryStore) Get(id int64) (HistoryEntry, error) {
	var (
		e   HistoryEntry
		fav int
		ms  int64
	)
	err := h.db.QueryRow(
		`SELECT id, protocol, title, summary, detail, favorite, created_at
		 FROM history WHERE id=?`, id).
		Scan(&e.ID, &e.Protocol, &e.Title, &e.Summary, &e.Detail, &fav, &ms)
	if err != nil {
		return HistoryEntry{}, err
	}
	e.Favorite = fav != 0
	e.CreatedAt = time.UnixMilli(ms)
	return e, nil
}

// Close releases the database handle.
func (h *HistoryStore) Close() error {
	if h.db == nil {
		return nil
	}
	return h.db.Close()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
