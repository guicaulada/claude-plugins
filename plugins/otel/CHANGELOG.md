# Changelog

## [0.3.0](https://github.com/guicaulada/claude-plugins/compare/otel-v0.2.3...otel-v0.3.0) (2026-03-12)


### ⚠ BREAKING CHANGES

* **otel:** metric names changed — dashboards need updating

### Features

* **otel:** gate high-cardinality metric attributes ([ce3204c](https://github.com/guicaulada/claude-plugins/commit/ce3204c8f313182811878f97d27cc7cc00182f9c))


### Bug Fixes

* **otel:** accept common boolean formats for env vars ([6ab73a5](https://github.com/guicaulada/claude-plugins/commit/6ab73a50e2cb309dcc4d05dfdc2f5334ac725d0f))
* **otel:** add custom histogram bucket boundaries ([a008cd4](https://github.com/guicaulada/claude-plugins/commit/a008cd4e0405b5e56910228bdfc74ed543643620))
* **otel:** add descriptions to all metric instruments ([cbe5015](https://github.com/guicaulada/claude-plugins/commit/cbe50159c332dcc676edc6a6a64c22a239b5b82f))
* **otel:** add descriptive body to all log events ([108ca8b](https://github.com/guicaulada/claude-plugins/commit/108ca8b40fce15e31c062200600335594bc8536b))
* **otel:** add permission_requests counter metric ([3c7f3ac](https://github.com/guicaulada/claude-plugins/commit/3c7f3acfb67baf66407e6eff0e57448a9cd320fb))
* **otel:** add post-join containment check for git ref path ([4ecdfc9](https://github.com/guicaulada/claude-plugins/commit/4ecdfc9cee46a4e9e5e2797cbb71a3d48bc01b5c))
* **otel:** add WithTelemetrySDK to resource attributes ([6f26abe](https://github.com/guicaulada/claude-plugins/commit/6f26abece2122983638784a0671818d9f6e91198))
* **otel:** add WithUnit to all metric instruments ([c50a744](https://github.com/guicaulada/claude-plugins/commit/c50a7440ad1a8790fa805991c0ba5a225acb6302))
* **otel:** derive prompt index from persistent counter ([89086fa](https://github.com/guicaulada/claude-plugins/commit/89086fac1b376634b99c6e067a034a4fe2f98be5))
* **otel:** include all counters in session span ([92ed449](https://github.com/guicaulada/claude-plugins/commit/92ed449a010f5a69793fc70547e56544a6436f29))
* **otel:** map log severity levels by event type ([3cfba2a](https://github.com/guicaulada/claude-plugins/commit/3cfba2aab8158e613681628be219c2e5124b6f44))
* **otel:** namespace lines_changed type attribute ([234f6fe](https://github.com/guicaulada/claude-plugins/commit/234f6fe324fc7aa49b05ef30b33862af64850237))
* **otel:** remove redundant duration_ms span attribute ([9e06a4e](https://github.com/guicaulada/claude-plugins/commit/9e06a4e1a6f09f1b5beb3ece7a288045df30fdfb))
* **otel:** support richer log body distinct from event name ([f7b2988](https://github.com/guicaulada/claude-plugins/commit/f7b2988a0762640a84f706ae9d4e9ec9c57cf912))
* **otel:** use parseBool consistently for all boolean env vars ([44efe19](https://github.com/guicaulada/claude-plugins/commit/44efe195b075218351306d3fc68e107edf16b4e1))
* **otel:** use seconds for duration metrics per UCUM standard ([8b6c296](https://github.com/guicaulada/claude-plugins/commit/8b6c296c96b53dd49617125260a414c0d704bc6c))
* **otel:** use typed log attributes instead of string maps ([07b0e39](https://github.com/guicaulada/claude-plugins/commit/07b0e3911e618af71f928e9310668d0081a7a78a))
* **otel:** validate git ref path before following ([f16c94e](https://github.com/guicaulada/claude-plugins/commit/f16c94ecf82ffd14a9076e0b26a6716105acd7e6))


### Code Refactoring

* **otel:** rename counter metrics to plural nouns ([81de585](https://github.com/guicaulada/claude-plugins/commit/81de585a90cc6f6437ebccaa48f9169779796d74))

## [0.2.3](https://github.com/guicaulada/claude-plugins/compare/otel-v0.2.2...otel-v0.2.3) (2026-03-11)


### Bug Fixes

* **otel:** end orphaned spans at parent end time ([751f46e](https://github.com/guicaulada/claude-plugins/commit/751f46ed78e1afad6c1e70b990be03a691670853))

## [0.2.2](https://github.com/guicaulada/claude-plugins/compare/otel-v0.2.1...otel-v0.2.2) (2026-03-11)


### Features

* **otel:** add skill name to Skill tool spans ([bbba4ad](https://github.com/guicaulada/claude-plugins/commit/bbba4ad642a775c5a1c374012f7e8ed1c14606f3))

## [0.2.1](https://github.com/guicaulada/claude-plugins/compare/otel-v0.2.0...otel-v0.2.1) (2026-03-11)


### Features

* add marketplace component to release-please, complete metadata ([7181fd1](https://github.com/guicaulada/claude-plugins/commit/7181fd1c14eea74086b6031758f72ee00634784c))

## [0.2.0](https://github.com/guicaulada/claude-plugins/compare/otel-v0.1.0...otel-v0.2.0) (2026-03-11)


### Features

* convert to plugin marketplace with Release Please ([1271790](https://github.com/guicaulada/claude-plugins/commit/1271790b83f4274bf2276b962949ef0b644c0c2b))
* **otel:** commit pre-built binaries for plugin distribution ([c59ae4f](https://github.com/guicaulada/claude-plugins/commit/c59ae4f988d4e2b54c38712d2d274849c1ecb4e7))
