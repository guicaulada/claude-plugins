package state

import "database/sql"

// Tool represents an active tool span in state.
type Tool struct {
	ToolUseID    string
	SessionID    string
	SpanID       string
	ParentSpanID string
	StartTime    int64
	ToolName     string
	FilePath     string
	FileSnapshot []byte
}

// CreateTool inserts a new tool record.
func (s *Store) CreateTool(t Tool) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO tools
		(tool_use_id, session_id, span_id, parent_span_id, start_time, tool_name, file_path, file_snapshot)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ToolUseID, t.SessionID, t.SpanID, t.ParentSpanID,
		t.StartTime, t.ToolName, t.FilePath, t.FileSnapshot,
	)
	return err
}

// GetTool retrieves a tool record by tool_use_id.
func (s *Store) GetTool(toolUseID string) (Tool, error) {
	var t Tool
	err := s.db.QueryRow(`
		SELECT tool_use_id, session_id, span_id, parent_span_id, start_time,
		       tool_name, file_path, file_snapshot
		FROM tools WHERE tool_use_id = ?`, toolUseID,
	).Scan(
		&t.ToolUseID, &t.SessionID, &t.SpanID, &t.ParentSpanID,
		&t.StartTime, &t.ToolName, &t.FilePath, &t.FileSnapshot,
	)
	if err == sql.ErrNoRows {
		return Tool{}, nil
	}
	return t, err
}

// DeleteTool removes a tool by tool_use_id.
func (s *Store) DeleteTool(toolUseID string) error {
	_, err := s.db.Exec(`DELETE FROM tools WHERE tool_use_id = ?`, toolUseID)
	return err
}
