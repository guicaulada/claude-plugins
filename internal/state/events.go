package state

// SpanEvent represents a recorded event for later attachment to a parent span.
type SpanEvent struct {
	SessionID string
	SpanID    string // The span this event belongs to (e.g., session span ID)
	Name      string
	Timestamp int64 // UnixNano
	Attrs     string // JSON-encoded attributes
}

// RecordEvent stores an event for later attachment to a span.
func (s *Store) RecordEvent(e SpanEvent) error {
	_, err := s.db.Exec(`
		INSERT INTO events (session_id, span_id, name, timestamp, attrs)
		VALUES (?, ?, ?, ?, ?)`,
		e.SessionID, e.SpanID, e.Name, e.Timestamp, e.Attrs,
	)
	return err
}

// GetEvents retrieves all recorded events for a given span.
func (s *Store) GetEvents(sessionID, spanID string) ([]SpanEvent, error) {
	rows, err := s.db.Query(`
		SELECT session_id, span_id, name, timestamp, attrs
		FROM events WHERE session_id = ? AND span_id = ?
		ORDER BY timestamp ASC`, sessionID, spanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []SpanEvent
	for rows.Next() {
		var e SpanEvent
		if err := rows.Scan(&e.SessionID, &e.SpanID, &e.Name, &e.Timestamp, &e.Attrs); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
