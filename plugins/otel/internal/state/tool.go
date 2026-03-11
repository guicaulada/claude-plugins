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

// GetLatestToolByName retrieves the most recently created tool with the given name.
func (s *Store) GetLatestToolByName(sessionID, toolName string) (Tool, error) {
	var t Tool
	err := s.db.QueryRow(`
		SELECT tool_use_id, session_id, span_id, parent_span_id, start_time,
		       tool_name, file_path, file_snapshot
		FROM tools WHERE session_id = ? AND tool_name = ?
		ORDER BY start_time DESC LIMIT 1`, sessionID, toolName,
	).Scan(
		&t.ToolUseID, &t.SessionID, &t.SpanID, &t.ParentSpanID,
		&t.StartTime, &t.ToolName, &t.FilePath, &t.FileSnapshot,
	)
	if err == sql.ErrNoRows {
		return Tool{}, nil
	}
	return t, err
}

// GetToolsByParent retrieves all tools with the given parent span ID.
func (s *Store) GetToolsByParent(sessionID, parentSpanID string) ([]Tool, error) {
	rows, err := s.db.Query(`
		SELECT tool_use_id, session_id, span_id, parent_span_id, start_time,
		       tool_name, file_path, file_snapshot
		FROM tools WHERE session_id = ? AND parent_span_id = ?
		ORDER BY start_time ASC`, sessionID, parentSpanID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tools []Tool
	for rows.Next() {
		var t Tool
		if err := rows.Scan(
			&t.ToolUseID, &t.SessionID, &t.SpanID, &t.ParentSpanID,
			&t.StartTime, &t.ToolName, &t.FilePath, &t.FileSnapshot,
		); err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

// DeleteTool removes a tool by tool_use_id.
func (s *Store) DeleteTool(toolUseID string) error {
	_, err := s.db.Exec(`DELETE FROM tools WHERE tool_use_id = ?`, toolUseID)
	return err
}
