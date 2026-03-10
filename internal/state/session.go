package state

import "database/sql"

// Session represents a session record in state.
type Session struct {
	SessionID      string
	TraceID        string
	SpanID         string
	StartTime      int64
	Cwd            string
	PermissionMode string
	StartType      string
	GitBranch      string
	GitRemoteURL   string
	GitRepoName    string
	GitRepoOwner   string
}

// CreateSession inserts a new session record.
func (s *Store) CreateSession(sess Session) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sessions
		(session_id, trace_id, span_id, start_time, cwd, permission_mode, start_type,
		 git_branch, git_remote_url, git_repo_name, git_repo_owner)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.SessionID, sess.TraceID, sess.SpanID, sess.StartTime,
		sess.Cwd, sess.PermissionMode, sess.StartType,
		sess.GitBranch, sess.GitRemoteURL, sess.GitRepoName, sess.GitRepoOwner,
	)
	return err
}

// GetSession retrieves a session record by session ID.
func (s *Store) GetSession(sessionID string) (Session, error) {
	var sess Session
	err := s.db.QueryRow(`
		SELECT session_id, trace_id, span_id, start_time, cwd, permission_mode, start_type,
		       git_branch, git_remote_url, git_repo_name, git_repo_owner
		FROM sessions WHERE session_id = ?`, sessionID,
	).Scan(
		&sess.SessionID, &sess.TraceID, &sess.SpanID, &sess.StartTime,
		&sess.Cwd, &sess.PermissionMode, &sess.StartType,
		&sess.GitBranch, &sess.GitRemoteURL, &sess.GitRepoName, &sess.GitRepoOwner,
	)
	if err == sql.ErrNoRows {
		return Session{}, nil
	}
	return sess, err
}

// IncrementCounter atomically increments a named counter for the session.
func (s *Store) IncrementCounter(sessionID, name string) error {
	_, err := s.db.Exec(`
		INSERT INTO counters (session_id, name, value) VALUES (?, ?, 1)
		ON CONFLICT (session_id, name) DO UPDATE SET value = value + 1`,
		sessionID, name,
	)
	return err
}

// IncrementCounterBy atomically increments a named counter by a given amount.
func (s *Store) IncrementCounterBy(sessionID, name string, amount int64) error {
	_, err := s.db.Exec(`
		INSERT INTO counters (session_id, name, value) VALUES (?, ?, ?)
		ON CONFLICT (session_id, name) DO UPDATE SET value = value + ?`,
		sessionID, name, amount, amount,
	)
	return err
}

// GetCounter retrieves the current value of a named counter.
func (s *Store) GetCounter(sessionID, name string) (int64, error) {
	var value int64
	err := s.db.QueryRow(`
		SELECT value FROM counters WHERE session_id = ? AND name = ?`,
		sessionID, name,
	).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return value, err
}
