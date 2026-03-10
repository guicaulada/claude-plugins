package payload

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantEvent string
		wantSID   string
		wantErr   bool
	}{
		{
			name: "session start",
			input: `{
				"session_id": "abc123",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd": "/home/user/project",
				"permission_mode": "default",
				"hook_event_name": "SessionStart",
				"source": "startup"
			}`,
			wantEvent: "SessionStart",
			wantSID:   "abc123",
		},
		{
			name: "with subagent context",
			input: `{
				"session_id": "abc123",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd": "/home/user/project",
				"permission_mode": "bypassPermissions",
				"hook_event_name": "PostToolUse",
				"agent_id": "agent-1",
				"agent_type": "Explore"
			}`,
			wantEvent: "PostToolUse",
			wantSID:   "abc123",
		},
		{
			name:    "invalid json",
			input:   `{not valid}`,
			wantErr: true,
		},
		{
			name:      "empty object",
			input:     `{}`,
			wantEvent: "",
			wantSID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := Parse([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if env.HookEventName != tt.wantEvent {
				t.Errorf("HookEventName = %q, want %q", env.HookEventName, tt.wantEvent)
			}
			if env.SessionID != tt.wantSID {
				t.Errorf("SessionID = %q, want %q", env.SessionID, tt.wantSID)
			}
			if env.RawEvent == nil {
				t.Error("RawEvent should not be nil")
			}
		})
	}
}

func TestParseSubagentFields(t *testing.T) {
	input := `{
		"session_id": "s1",
		"cwd": "/tmp",
		"permission_mode": "default",
		"hook_event_name": "PostToolUse",
		"agent_id": "agent-abc",
		"agent_type": "Explore"
	}`

	env, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.AgentID != "agent-abc" {
		t.Errorf("AgentID = %q, want %q", env.AgentID, "agent-abc")
	}
	if env.AgentType != "Explore" {
		t.Errorf("AgentType = %q, want %q", env.AgentType, "Explore")
	}
}
