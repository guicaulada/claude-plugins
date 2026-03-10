package events

import (
	"context"
	"encoding/json"

	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

// bestParentSpanID returns the most specific active parent span:
// current prompt > session. Used for informational events that
// don't have their own parent context.
func bestParentSpanID(store *state.Store, env payload.Envelope, sess state.Session) string {
	if prompt, err := store.GetCurrentPrompt(env.SessionID); err == nil && prompt.SessionID != "" {
		return prompt.SpanID
	}
	return sess.SpanID
}

// PermissionRequest

type permissionRequestEvent struct {
	ToolName string `json:"tool_name"`
}

func HandlePermissionRequest(env payload.Envelope) error {
	var event permissionRequestEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse PermissionRequest: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.tool.name"] = event.ToolName
	provider.EmitEvent("claude_code.permission.request", sess.TraceID, sess.SpanID, logAttrs)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "permission.request",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"tool.name": event.ToolName}),
	})

	debug.Log("permission request: session=%s tool=%s", env.SessionID, event.ToolName)
	return nil
}

// Notification

type notificationEvent struct {
	NotificationType string `json:"notification_type"`
	Title            string `json:"title"`
	Message          string `json:"message"`
}

func HandleNotification(env payload.Envelope) error {
	var event notificationEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse Notification: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	_ = store.IncrementCounter(env.SessionID, "notification_count")

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.notification.type"] = event.NotificationType
	if event.Title != "" {
		logAttrs["claude_code.notification.title"] = event.Title
	}
	if cfg.LogToolDetails && event.Message != "" {
		logAttrs["claude_code.notification.message"] = event.Message
	}
	provider.EmitEvent("claude_code.notification", sess.TraceID, sess.SpanID, logAttrs)

	provider.CounterAdd(ctx, "claude_code.notification.count", 1,
		attribute.String("claude_code.notification.type", event.NotificationType),
	)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "notification",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"type": event.NotificationType}),
	})

	debug.Log("notification: session=%s type=%s", env.SessionID, event.NotificationType)
	return nil
}

// TaskCompleted

type taskCompletedEvent struct {
	TaskID          string `json:"task_id"`
	TaskSubject     string `json:"task_subject"`
	TaskDescription string `json:"task_description"`
	TeammateName    string `json:"teammate_name"`
	TeamName        string `json:"team_name"`
}

func HandleTaskCompleted(env payload.Envelope) error {
	var event taskCompletedEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse TaskCompleted: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	_ = store.IncrementCounter(env.SessionID, "task_count")

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.task.id"] = event.TaskID
	logAttrs["claude_code.task.subject"] = event.TaskSubject
	if cfg.LogUserPrompts && event.TaskDescription != "" {
		logAttrs["claude_code.task.description"] = event.TaskDescription
	}
	if event.TeammateName != "" {
		logAttrs["claude_code.task.teammate_name"] = event.TeammateName
	}
	if event.TeamName != "" {
		logAttrs["claude_code.task.team_name"] = event.TeamName
	}
	provider.EmitEvent("claude_code.task.completed", sess.TraceID, sess.SpanID, logAttrs)

	provider.CounterAdd(ctx, "claude_code.task.count", 1)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "task.completed",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"task.subject": event.TaskSubject}),
	})

	debug.Log("task completed: session=%s task=%s subject=%s", env.SessionID, event.TaskID, event.TaskSubject)
	return nil
}

// InstructionsLoaded

type instructionsLoadedEvent struct {
	FilePath        string `json:"file_path"`
	MemoryType      string `json:"memory_type"`
	LoadReason      string `json:"load_reason"`
	TriggerFilePath string `json:"trigger_file_path"`
	ParentFilePath  string `json:"parent_file_path"`
}

func HandleInstructionsLoaded(env payload.Envelope) error {
	var event instructionsLoadedEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse InstructionsLoaded: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	_ = store.IncrementCounter(env.SessionID, "instructions_loaded_count")

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.instructions.file_path"] = event.FilePath
	logAttrs["claude_code.instructions.memory_type"] = event.MemoryType
	logAttrs["claude_code.instructions.load_reason"] = event.LoadReason
	if event.TriggerFilePath != "" {
		logAttrs["claude_code.instructions.trigger_file_path"] = event.TriggerFilePath
	}
	if event.ParentFilePath != "" {
		logAttrs["claude_code.instructions.parent_file_path"] = event.ParentFilePath
	}
	provider.EmitEvent("claude_code.instructions.loaded", sess.TraceID, sess.SpanID, logAttrs)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "instructions.loaded",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"file_path": event.FilePath, "memory_type": event.MemoryType}),
	})

	debug.Log("instructions loaded: session=%s file=%s type=%s reason=%s",
		env.SessionID, event.FilePath, event.MemoryType, event.LoadReason)
	return nil
}

// ConfigChange

type configChangeEvent struct {
	Source   string `json:"source"`
	FilePath string `json:"file_path"`
}

func HandleConfigChange(env payload.Envelope) error {
	var event configChangeEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse ConfigChange: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.config.source"] = event.Source
	if event.FilePath != "" {
		logAttrs["claude_code.config.file_path"] = event.FilePath
	}
	provider.EmitEvent("claude_code.config.change", sess.TraceID, sess.SpanID, logAttrs)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "config.change",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"source": event.Source}),
	})

	debug.Log("config change: session=%s source=%s", env.SessionID, event.Source)
	return nil
}

// WorktreeCreate

type worktreeCreateEvent struct {
	Name string `json:"name"`
}

func HandleWorktreeCreate(env payload.Envelope) error {
	var event worktreeCreateEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse WorktreeCreate: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.worktree.name"] = event.Name
	provider.EmitEvent("claude_code.worktree.create", sess.TraceID, sess.SpanID, logAttrs)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "worktree.create",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"name": event.Name}),
	})

	debug.Log("worktree create: session=%s name=%s", env.SessionID, event.Name)
	return nil
}

// WorktreeRemove

type worktreeRemoveEvent struct {
	WorktreePath string `json:"worktree_path"`
}

func HandleWorktreeRemove(env payload.Envelope) error {
	var event worktreeRemoveEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse WorktreeRemove: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.worktree.path"] = event.WorktreePath
	provider.EmitEvent("claude_code.worktree.remove", sess.TraceID, sess.SpanID, logAttrs)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "worktree.remove",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"path": event.WorktreePath}),
	})

	debug.Log("worktree remove: session=%s path=%s", env.SessionID, event.WorktreePath)
	return nil
}

// TeammateIdle

type teammateIdleEvent struct {
	TeammateName string `json:"teammate_name"`
	TeamName     string `json:"team_name"`
}

func HandleTeammateIdle(env payload.Envelope) error {
	var event teammateIdleEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse TeammateIdle: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.teammate.name"] = event.TeammateName
	logAttrs["claude_code.teammate.team_name"] = event.TeamName
	provider.EmitEvent("claude_code.teammate.idle", sess.TraceID, sess.SpanID, logAttrs)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "teammate.idle",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"teammate.name": event.TeammateName, "team.name": event.TeamName}),
	})

	debug.Log("teammate idle: session=%s teammate=%s team=%s", env.SessionID, event.TeammateName, event.TeamName)
	return nil
}

// PreCompact

type preCompactEvent struct {
	Trigger            string `json:"trigger"`
	CustomInstructions string `json:"custom_instructions"`
}

func HandlePreCompact(env payload.Envelope) error {
	var event preCompactEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse PreCompact: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	_ = store.IncrementCounter(env.SessionID, "compact_count")

	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	sess, _ := store.GetSession(env.SessionID)
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.compact.trigger"] = event.Trigger
	if cfg.LogUserPrompts && event.CustomInstructions != "" {
		logAttrs["claude_code.compact.custom_instructions"] = event.CustomInstructions
	}
	provider.EmitEvent("claude_code.compact", sess.TraceID, sess.SpanID, logAttrs)

	provider.CounterAdd(ctx, "claude_code.compact.count", 1,
		attribute.String("claude_code.compact.trigger", event.Trigger),
	)

	parentSpanID := bestParentSpanID(store, env, sess)
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "compact",
		Timestamp: currentTimestamp(),
		Attrs:     marshalAttrs(map[string]string{"trigger": event.Trigger}),
	})

	debug.Log("pre compact: session=%s trigger=%s", env.SessionID, event.Trigger)
	return nil
}

