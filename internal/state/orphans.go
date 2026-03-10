package state

// GetOrphanedPrompts returns prompts that were never closed (no Stop).
func (s *Store) GetOrphanedPrompts(sessionID string) ([]Prompt, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, span_id, start_time, prompt_index
		FROM prompts WHERE session_id = ?
		ORDER BY start_time ASC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []Prompt
	for rows.Next() {
		var p Prompt
		if err := rows.Scan(&p.ID, &p.SessionID, &p.SpanID, &p.StartTime, &p.PromptIndex); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}
	return prompts, rows.Err()
}

// GetOrphanedTools returns tools that were never closed (no PostToolUse/PostToolUseFailure).
func (s *Store) GetOrphanedTools(sessionID string) ([]Tool, error) {
	rows, err := s.db.Query(`
		SELECT tool_use_id, session_id, span_id, parent_span_id, start_time,
		       tool_name, file_path, file_snapshot
		FROM tools WHERE session_id = ?
		ORDER BY start_time ASC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

// GetOrphanedSubagents returns subagents that were never closed (no SubagentStop).
func (s *Store) GetOrphanedSubagents(sessionID string) ([]Subagent, error) {
	rows, err := s.db.Query(`
		SELECT agent_id, session_id, span_id, parent_span_id, start_time,
		       agent_type, agent_name
		FROM subagents WHERE session_id = ?
		ORDER BY start_time ASC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []Subagent
	for rows.Next() {
		var sa Subagent
		if err := rows.Scan(
			&sa.AgentID, &sa.SessionID, &sa.SpanID, &sa.ParentSpanID,
			&sa.StartTime, &sa.AgentType, &sa.AgentName,
		); err != nil {
			return nil, err
		}
		agents = append(agents, sa)
	}
	return agents, rows.Err()
}
