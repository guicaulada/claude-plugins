package state

import "database/sql"

// Subagent represents an active subagent span in state.
type Subagent struct {
	AgentID      string
	SessionID    string
	SpanID       string
	ParentSpanID string
	StartTime    int64
	AgentType    string
	AgentName    string
}

// CreateSubagent inserts a new subagent record.
func (s *Store) CreateSubagent(sa Subagent) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO subagents
		(agent_id, session_id, span_id, parent_span_id, start_time, agent_type, agent_name)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sa.AgentID, sa.SessionID, sa.SpanID, sa.ParentSpanID,
		sa.StartTime, sa.AgentType, sa.AgentName,
	)
	return err
}

// GetSubagent retrieves a subagent record by agent_id.
func (s *Store) GetSubagent(agentID string) (Subagent, error) {
	var sa Subagent
	err := s.db.QueryRow(`
		SELECT agent_id, session_id, span_id, parent_span_id, start_time,
		       agent_type, agent_name
		FROM subagents WHERE agent_id = ?`, agentID,
	).Scan(
		&sa.AgentID, &sa.SessionID, &sa.SpanID, &sa.ParentSpanID,
		&sa.StartTime, &sa.AgentType, &sa.AgentName,
	)
	if err == sql.ErrNoRows {
		return Subagent{}, nil
	}
	return sa, err
}

// GetSubagentsByParent retrieves all subagents with the given parent span ID.
func (s *Store) GetSubagentsByParent(sessionID, parentSpanID string) ([]Subagent, error) {
	rows, err := s.db.Query(`
		SELECT agent_id, session_id, span_id, parent_span_id, start_time,
		       agent_type, agent_name
		FROM subagents WHERE session_id = ? AND parent_span_id = ?
		ORDER BY start_time ASC`, sessionID, parentSpanID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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

// DeleteSubagent removes a subagent by agent_id.
func (s *Store) DeleteSubagent(agentID string) error {
	_, err := s.db.Exec(`DELETE FROM subagents WHERE agent_id = ?`, agentID)
	return err
}
