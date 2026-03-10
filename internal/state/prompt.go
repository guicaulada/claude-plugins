package state

import "database/sql"

// Prompt represents an active prompt span in state.
type Prompt struct {
	ID          int64
	SessionID   string
	SpanID      string
	StartTime   int64
	PromptIndex int
}

// CreatePrompt inserts a new prompt record.
func (s *Store) CreatePrompt(p Prompt) error {
	_, err := s.db.Exec(`
		INSERT INTO prompts (session_id, span_id, start_time, prompt_index)
		VALUES (?, ?, ?, ?)`,
		p.SessionID, p.SpanID, p.StartTime, p.PromptIndex,
	)
	return err
}

// GetCurrentPrompt retrieves the most recent prompt for a session.
func (s *Store) GetCurrentPrompt(sessionID string) (Prompt, error) {
	var p Prompt
	err := s.db.QueryRow(`
		SELECT id, session_id, span_id, start_time, prompt_index
		FROM prompts WHERE session_id = ?
		ORDER BY id DESC LIMIT 1`, sessionID,
	).Scan(&p.ID, &p.SessionID, &p.SpanID, &p.StartTime, &p.PromptIndex)
	if err == sql.ErrNoRows {
		return Prompt{}, nil
	}
	return p, err
}

// GetPromptCount returns the number of prompts for a session.
func (s *Store) GetPromptCount(sessionID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM prompts WHERE session_id = ?`, sessionID,
	).Scan(&count)
	return count, err
}

// DeletePrompt removes a prompt by ID.
func (s *Store) DeletePrompt(id int64) error {
	_, err := s.db.Exec(`DELETE FROM prompts WHERE id = ?`, id)
	return err
}
