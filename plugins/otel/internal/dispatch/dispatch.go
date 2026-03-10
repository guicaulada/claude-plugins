package dispatch

import (
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/debug"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/payload"
)

// HandlerFunc processes a hook event envelope.
type HandlerFunc func(env payload.Envelope) error

// Registry maps hook event names to handler functions.
type Registry struct {
	handlers map[string]HandlerFunc
}

// New creates an empty handler registry.
func New() *Registry {
	return &Registry{
		handlers: make(map[string]HandlerFunc),
	}
}

// Register adds a handler for a hook event name.
func (r *Registry) Register(eventName string, handler HandlerFunc) {
	r.handlers[eventName] = handler
}

// Dispatch routes an envelope to the registered handler.
// Unknown events are logged and skipped.
func (r *Registry) Dispatch(env payload.Envelope) error {
	handler, ok := r.handlers[env.HookEventName]
	if !ok {
		debug.Log("no handler registered for event: %s", env.HookEventName)
		return nil
	}

	debug.Log("dispatching event: %s (session: %s)", env.HookEventName, env.SessionID)
	return handler(env)
}
