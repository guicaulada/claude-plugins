# otel

OpenTelemetry traces, metrics, and logs from Claude Code hook lifecycle events.

A Claude Code plugin that provides in-depth observability of your Claude Code sessions — traces with full span hierarchy, enriched metrics with rich dimensions, and log records for every hook event.

## Features

- **Traces**: Session → Prompt → Tool/Subagent span hierarchy with parent-child linking
- **Metrics**: 14 metrics (counters + histograms) with dimensions for project, branch, file type, tool, and more
- **Logs**: 18 event types as OTel log records with trace correlation
- **Enrichment**: Git context (branch, repo, owner), file metadata (extension, language via go-enry), line diffs
- **Interruption handling**: Orphaned spans from interrupted operations are exported with error status
- **Span events**: Timeline annotations on parent spans for child lifecycle events

## Installation

### From marketplace

```bash
claude plugin marketplace add https://github.com/guicaulada/claude-plugins
claude plugin install otel@guicaulada-plugins
```

### Using `--plugin-dir` (development)

```bash
git clone https://github.com/guicaulada/claude-plugins.git
cd claude-plugins/plugins/otel
make build
claude --plugin-dir .
```

## Configuration

The plugin reuses Claude Code's OTel environment variables. Set `CLAUDE_CODE_ENABLE_TELEMETRY=1` or `OTEL_PLUGIN_ENABLED=1` to enable.

### Required

| Variable | Description |
|---|---|
| `CLAUDE_CODE_ENABLE_TELEMETRY` | Master switch. Plugin respects this unless `OTEL_PLUGIN_ENABLED` overrides. |
| `OTEL_PLUGIN_ENABLED` | Plugin-specific toggle. `1` enables even if telemetry is off. `0` disables. |

### OTel Export

The plugin reads standard `OTEL_EXPORTER_OTLP_*` env vars automatically. Plugin-specific overrides:

| Variable | Description |
|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint (read by SDK automatically) |
| `OTEL_EXPORTER_OTLP_HEADERS` | Auth headers (read by SDK automatically) |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | Protocol: `http/protobuf`, `http/json`, `grpc` |
| `OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT` | Plugin-specific endpoint override |
| `OTEL_PLUGIN_EXPORTER_OTLP_HEADERS` | Plugin-specific headers override |
| `OTEL_PLUGIN_EXPORTER_OTLP_PROTOCOL` | Plugin-specific protocol override |
| `OTEL_PLUGIN_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE` | Temporality override: `delta` or `cumulative` (overrides `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE`). See [Metrics Temporality](#metrics-temporality) |
| `OTEL_RESOURCE_ATTRIBUTES` | Custom resource attributes (read by SDK) |

### Privacy & Detail Control

| Variable | Description | Default |
|---|---|---|
| `OTEL_LOG_USER_PROMPTS` | Include prompt content, task descriptions, compact instructions in logs | `0` (disabled) |
| `OTEL_LOG_TOOL_DETAILS` | Include tool input details (bash commands, file paths, patterns, URLs) in logs | `0` (disabled) |
| `OTEL_PLUGIN_METRICS_INCLUDE_HIGH_CARDINALITY` | Include high-cardinality attributes (`cwd`, VCS context) on metrics | `0` (disabled) |

### Auth & Headers

| Variable | Description |
|---|---|
| `otelHeadersHelper` | Script in Claude settings that outputs JSON headers (e.g., `{"Authorization": "Bearer ..."}`) |
| `CLAUDE_CODE_OTEL_HEADERS_HELPER_DEBOUNCE_MS` | Header refresh interval (default: 1740000ms / 29 min). Shared across sessions. |

### Debug

| Variable | Description |
|---|---|
| `OTEL_PLUGIN_DEBUG` | Set to `1` to enable debug logging to `$TMPDIR/claude-code-otel-plugin/debug.log` |

## Signals

### Traces

Span hierarchy:

```
session (root)
├── prompt
│   ├── tool:Read
│   ├── tool:Edit
│   ├── tool:Bash
│   └── tool:Agent
│       └── agent:Explore
│           ├── tool:Bash
│           └── tool:Read
├── prompt
│   └── ...
└── (exported at SessionEnd with aggregated attributes)
```

Session span attributes: `prompt_count`, `tool_count`, `error_count`, `subagent_count`, `lines_added`, `lines_removed`, `commit_count`, `branch_count`, `repo_count`, `interrupted_count`, `notification_count`, `compact_count`, `task_count`, `permission_request_count`, VCS context.

### Metrics

| Metric | Type | Unit | Key Dimensions |
|---|---|---|---|
| `claude_code.sessions` | Counter | `{session}` | start_type, cwd, vcs.* |
| `claude_code.session.duration` | Histogram | `s` | start_type, end_reason, cwd, vcs.* |
| `claude_code.prompts` | Counter | `{prompt}` | cwd, vcs.* |
| `claude_code.prompt.duration` | Histogram | `s` | cwd, vcs.* |
| `claude_code.tool.uses` | Counter | `{use}` | tool.name, tool.success, cwd, file.*, vcs.* |
| `claude_code.tool.duration` | Histogram | `s` | tool.name, tool.success, cwd, file.*, vcs.* |
| `claude_code.errors` | Counter | `{error}` | tool.name, error.is_interrupt, vcs.* |
| `claude_code.lines_changed` | Counter | `{line}` | lines_changed.type (added/removed), cwd, file.*, vcs.* |
| `claude_code.subagents` | Counter | `{agent}` | agent.type, agent.name, vcs.* |
| `claude_code.subagent.duration` | Histogram | `s` | agent.type, agent.name, vcs.* |
| `claude_code.compacts` | Counter | `{compact}` | trigger |
| `claude_code.notifications` | Counter | `{notification}` | notification.type |
| `claude_code.tasks` | Counter | `{task}` | — |
| `claude_code.permission_requests` | Counter | `{request}` | — |

### Events (Logs)

All 18 Claude Code hook events are exported as OTel log records with `trace_id`/`span_id` correlation:

`session.start`, `session.end`, `prompt.submit`, `prompt.stop`, `tool.start`, `tool.end`, `tool.error`, `permission.request`, `agent.start`, `agent.stop`, `notification`, `task.completed`, `instructions.loaded`, `config.change`, `worktree.create`, `worktree.remove`, `teammate.idle`, `compact`

## Architecture

The plugin is a Go binary invoked by Claude Code hooks. Each hook event is a separate process invocation:

1. Read JSON payload from stdin
2. Parse common fields and dispatch by `hook_event_name`
3. Read/write session state in SQLite (`$TMPDIR/claude-code-otel-plugin/<session_id>/state.db`)
4. Export OTel signals (traces, metrics, logs) via OTLP HTTP
5. Exit 0 (never interfere with Claude Code)

Key design decisions:
- **No long-running process** — each hook invocation is independent
- **Delta temporality for all metrics** — see [Metrics Temporality](#metrics-temporality)
- **SQLite with WAL mode** for cross-process state sharing (span correlation, counters)
- **BatchSpanProcessor** for efficient bulk export at SessionEnd
- **Shared header cache** in `$TMPDIR/claude-code-otel-plugin/headers.json` across sessions
- **Graceful degradation** — errors are swallowed silently, debug log when `OTEL_PLUGIN_DEBUG=1`
- **`service.name: claude-code-otel-plugin`** distinguishes from built-in `claude-code` signals

### Metrics Temporality

This plugin **requires delta temporality** to produce correct metrics. Claude Code sets `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta` by default, so this works out of the box. **Do not override this to `cumulative`** — it will silently break all metrics.

**Why?** Each hook invocation is a short-lived process with a fresh OTel MeterProvider. With cumulative temporality, every counter always reports `1` and every histogram always contains a single observation, because there is no in-process state carried over between invocations. Prometheus-style queries like `rate()`, `increase()`, and `histogram_quantile()` require monotonically increasing cumulative values, which short-lived processes cannot provide. Delta temporality reports only what happened in each invocation (`+1`, `+500ms`), and the backend aggregates the deltas correctly.

If your backend requires cumulative temporality, use an intermediary to convert delta to cumulative before ingestion — for example [Grafana Alloy](https://grafana.com/blog/how-to-use-opentelemetry-and-grafana-alloy-to-convert-delta-to-cumulative-at-scale/) or the [OTel Collector](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/deltatocumulativeprocessor).

## Development

```bash
make build      # Build for current platform
make build-all  # Cross-compile for all 4 targets
make test       # Run tests
make vet        # Run go vet
make lint       # Run golangci-lint
make clean      # Remove build artifacts
```

## License

[Apache-2.0](LICENSE)
