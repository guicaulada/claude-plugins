package state

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store manages session state in SQLite.
type Store struct {
	db  *sql.DB
	dir string
}

// Open creates or opens the SQLite database for a session.
func Open(sessionID string) (*Store, error) {
	dir := filepath.Join(os.TempDir(), "claude-code-otel-plugin", sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create state dir: %w", err)
	}

	dbPath := filepath.Join(dir, "state.db")
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	s := &Store{db: db, dir: dir}
	if err := s.createSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Cleanup checkpoints WAL and removes all state files for the session.
func (s *Store) Cleanup() error {
	_, _ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err := s.db.Close(); err != nil {
		return err
	}
	return os.RemoveAll(s.dir)
}

func (s *Store) createSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS sessions (
    session_id TEXT PRIMARY KEY,
    trace_id TEXT NOT NULL,
    span_id TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    cwd TEXT,
    permission_mode TEXT,
    start_type TEXT,
    git_branch TEXT,
    git_remote_url TEXT,
    git_repo_name TEXT,
    git_repo_owner TEXT
);

CREATE TABLE IF NOT EXISTS prompts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    span_id TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    prompt_index INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS tools (
    tool_use_id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    span_id TEXT NOT NULL,
    parent_span_id TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    tool_name TEXT,
    file_path TEXT,
    file_snapshot BLOB
);

CREATE TABLE IF NOT EXISTS subagents (
    agent_id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    span_id TEXT NOT NULL,
    parent_span_id TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    agent_type TEXT,
    agent_name TEXT
);

CREATE TABLE IF NOT EXISTS counters (
    session_id TEXT NOT NULL,
    name TEXT NOT NULL,
    value INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (session_id, name)
);

CREATE TABLE IF NOT EXISTS cache (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);`

	_, err := s.db.Exec(schema)
	return err
}
