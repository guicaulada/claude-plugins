package payload

import "encoding/json"

// Envelope represents the common fields present in every hook event payload.
// Event-specific fields are captured in RawEvent for deferred parsing.
type Envelope struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	HookEventName  string `json:"hook_event_name"`

	// Subagent context (present when hook fires inside a subagent)
	AgentID   string `json:"agent_id,omitempty"`
	AgentType string `json:"agent_type,omitempty"`

	// RawEvent holds the full JSON for event-specific parsing.
	RawEvent json.RawMessage `json:"-"`
}

// Parse unmarshals raw JSON into an Envelope, preserving the full payload
// in RawEvent for event-specific field extraction.
func Parse(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, err
	}
	env.RawEvent = json.RawMessage(data)
	return env, nil
}
