# Changelog

## [0.2.4](https://github.com/guicaulada/claude-plugins/compare/marketplace-v0.2.3...marketplace-v0.2.4) (2026-03-11)


### Features

* **otel:** add skill name to Skill tool spans ([bbba4ad](https://github.com/guicaulada/claude-plugins/commit/bbba4ad642a775c5a1c374012f7e8ed1c14606f3))

## [0.2.3](https://github.com/guicaulada/claude-plugins/compare/marketplace-v0.2.2...marketplace-v0.2.3) (2026-03-11)


### Bug Fixes

* use ./plugins/otel relative path, remove pluginRoot ([503c05e](https://github.com/guicaulada/claude-plugins/commit/503c05eaea99a592befbf9a5391b911449e4ed11))

## [0.2.2](https://github.com/guicaulada/claude-plugins/compare/marketplace-v0.2.1...marketplace-v0.2.2) (2026-03-11)


### Bug Fixes

* use bare plugin name with pluginRoot instead of relative path ([8e0f240](https://github.com/guicaulada/claude-plugins/commit/8e0f240ce4e1438b3af8ea0a12c2e9c65f217dde))

## [0.2.1](https://github.com/guicaulada/claude-plugins/compare/marketplace-v0.2.0...marketplace-v0.2.1) (2026-03-11)


### Features

* add all 9 remaining event handlers and config support ([c711793](https://github.com/guicaulada/claude-plugins/commit/c71179317a2cc0ebbd6715605db8bdb6e45f00ed))
* add git context, file enrichment, and line diff computation ([1616229](https://github.com/guicaulada/claude-plugins/commit/161622973a614109ce51daae2a9bda7c1baa6f52))
* add handler entry point with config and debug logging ([83f422e](https://github.com/guicaulada/claude-plugins/commit/83f422e92a7d45a936e12e71e018d8131f3bb3aa))
* add marketplace component to release-please, complete metadata ([7181fd1](https://github.com/guicaulada/claude-plugins/commit/7181fd1c14eea74086b6031758f72ee00634784c))
* add remaining config support and tool detail sanitization ([6f83e7a](https://github.com/guicaulada/claude-plugins/commit/6f83e7a1e20dd0b689520a3319aaec3473b3cdd5))
* add sensitive attrs gated by OTEL_LOG_* flags ([40deb23](https://github.com/guicaulada/claude-plugins/commit/40deb2314621abf6a32e557fc271248da126daeb))
* add span events to session trace timeline ([ba60d8d](https://github.com/guicaulada/claude-plugins/commit/ba60d8dc3af5d7994045e3e69d08f68d3216cad0))
* add vcs attributes to all spans, track branch/repo counts ([a4f6a1e](https://github.com/guicaulada/claude-plugins/commit/a4f6a1e5a6872e8a1839ba4c86378645dd1952c6))
* **config:** add otelHeadersHelper support from Claude settings ([8f2a8a9](https://github.com/guicaulada/claude-plugins/commit/8f2a8a9524718a518b6de9ab9395fcd315460bfa))
* convert to plugin marketplace with Release Please ([1271790](https://github.com/guicaulada/claude-plugins/commit/1271790b83f4274bf2276b962949ef0b644c0c2b))
* **dispatch:** add event handler registry and routing ([e601394](https://github.com/guicaulada/claude-plugins/commit/e6013947a873a60a6f665838008f6479080de5c0))
* enrich all metrics with full dimensions ([61fd859](https://github.com/guicaulada/claude-plugins/commit/61fd859c65d870d0e6517458eef6e6163b30dff3))
* enrich log records with duration, file info, and error details ([294079c](https://github.com/guicaulada/claude-plugins/commit/294079ce91072b948f4434932e21c6517bc072ac))
* ensure all logs and metrics have complete attributes ([dc96201](https://github.com/guicaulada/claude-plugins/commit/dc962018adf986df3155a50da6cdcf25ed84dd3a))
* **events:** add all trace-related event handlers ([aa9354a](https://github.com/guicaulada/claude-plugins/commit/aa9354a859ac1abc0899da9c10cff23031fad81f))
* **events:** add OTel log records and metrics to all handlers ([7fcba05](https://github.com/guicaulada/claude-plugins/commit/7fcba0547715174b3fee32d8a591b81970d3a562))
* export orphaned prompt spans on interruption ([872018c](https://github.com/guicaulada/claude-plugins/commit/872018c8bb960854b0e4f43480c310e5da69a25c))
* export orphaned spans for interrupted tool calls and subagents ([398b71f](https://github.com/guicaulada/claude-plugins/commit/398b71fe28038ad6586a0932a7f83894265bd3bc))
* **fileinfo:** add file path, extension, language, and diff utils ([c7749ad](https://github.com/guicaulada/claude-plugins/commit/c7749ad49bac86b7a0655ba227cc8fe2a231dbe0))
* **git:** add git context extraction with branch, remote, SHA ([2522ca6](https://github.com/guicaulada/claude-plugins/commit/2522ca649ad663612a67865e8be3e2d1827fbacd))
* log tool input details when OTEL_LOG_TOOL_DETAILS is set ([2d77a69](https://github.com/guicaulada/claude-plugins/commit/2d77a697c2b9025337ef76929cdaf98deb997b01))
* **otel:** add LoggerProvider, MeterProvider, and signal helpers ([3dd8c34](https://github.com/guicaulada/claude-plugins/commit/3dd8c348127fa259ed1688a6b577c0d74dfc8236))
* **otel:** add TracerProvider setup and span builder ([792e66d](https://github.com/guicaulada/claude-plugins/commit/792e66dfad72e64cfd24e99fe41ad5d5554d5da7))
* **otel:** commit pre-built binaries for plugin distribution ([c59ae4f](https://github.com/guicaulada/claude-plugins/commit/c59ae4f988d4e2b54c38712d2d274849c1ecb4e7))
* **payload:** add envelope struct for hook event parsing ([82dc80d](https://github.com/guicaulada/claude-plugins/commit/82dc80da00eee394f7b16df5d4f77efc5c163958))
* record span events for all informational handlers ([b7f5a4e](https://github.com/guicaulada/claude-plugins/commit/b7f5a4eeb37dd6837b841eb6ee64015fe37aaf1a))
* shared header cache with CLAUDE_CODE_OTEL_HEADERS_HELPER_DEBOUNCE_MS ([7f5cb08](https://github.com/guicaulada/claude-plugins/commit/7f5cb0899662eed60dafb079e6d595af82190c9d))
* **state:** add prompt, tool, and subagent CRUD operations ([a810086](https://github.com/guicaulada/claude-plugins/commit/a8100863f6ce72c9c5145f1c36bd5a40511ba436))
* **state:** add SQLite state store with session CRUD ([616dcb7](https://github.com/guicaulada/claude-plugins/commit/616dcb720ef6cfe25eb26628e5d4e4f5c1a04b71))


### Bug Fixes

* add golangci-lint v2 config, fix all lint warnings ([a179892](https://github.com/guicaulada/claude-plugins/commit/a179892366a7ae92a246b9decf6fdf165dead0cb))
* add vcs dimensions to lines_changed metric ([dccc28b](https://github.com/guicaulada/claude-plugins/commit/dccc28b987ba55566dd19a428f337becefafe176))
* attach span events to correct parent spans ([198d6b3](https://github.com/guicaulada/claude-plugins/commit/198d6b3f5264f815963caa255fbd75d87aba69d9))
* cache OTel headers in state for SessionEnd export ([31d9dea](https://github.com/guicaulada/claude-plugins/commit/31d9dea0b9d75d6eb3eeacccf9f2d812079a5656))
* clean event names, add debug logging, session start/end events ([9f746bc](https://github.com/guicaulada/claude-plugins/commit/9f746bc753569a7b9a44c36e7b33d68992db5665))
* correct release-please-action SHA pin to v4.4.0 ([053a4f1](https://github.com/guicaulada/claude-plugins/commit/053a4f1bd93fa0fcd54508969781f15de1ac914d))
* include all counters in session end debug log ([2fbc28c](https://github.com/guicaulada/claude-plugins/commit/2fbc28cbf36ff8370a940e99bb4576a1937c2922))
* include error message in tool error log records ([cb38399](https://github.com/guicaulada/claude-plugins/commit/cb38399a44da4ab5f81d69d4958fec76da0907df))
* increase otelHeadersHelper timeout from 5s to 30s ([c1c646f](https://github.com/guicaulada/claude-plugins/commit/c1c646f6d42624ecec06c3cb077deec4afc36887))
* load and cache OTel headers on first event, not just SessionStart ([3674b84](https://github.com/guicaulada/claude-plugins/commit/3674b843da2ac45b3ffe769701ff3cfc86dc2b23))
* make build also copies to platform-specific binary name ([a01845b](https://github.com/guicaulada/claude-plugins/commit/a01845b4fefb3a6a34b7db3bf1c7ca544ce0bca3))
* make SubagentStop sync to prevent event race with PostToolUse ([55a55e1](https://github.com/guicaulada/claude-plugins/commit/55a55e157ae671602e6a58fb7f0deaeec6dd6885))
* make UserPromptSubmit sync to prevent parent race condition ([a8aa9d6](https://github.com/guicaulada/claude-plugins/commit/a8aa9d65ce781058c408440df2b18b891c8d9263))
* only run otelHeadersHelper once at SessionStart ([b6b05a0](https://github.com/guicaulada/claude-plugins/commit/b6b05a066342f622a46412b682bef6a913befdef))
* **otel:** let SDK read standard OTEL_EXPORTER_* env vars natively ([7cf1e2a](https://github.com/guicaulada/claude-plugins/commit/7cf1e2a12320342ddfe9ff8ca7a5681539dc7b1d))
* **otel:** make session root span a true root with fixed ID generator ([5c38a57](https://github.com/guicaulada/claude-plugins/commit/5c38a577739ea315c9acac5110beb8f00721950e))
* **otel:** use ChildContext to preserve span IDs for parent linking ([e150bfc](https://github.com/guicaulada/claude-plugins/commit/e150bfcf07aa372432ce99346aa0530dbf8816cc))
* parent subagent spans under Agent tool, preserve all span IDs ([1816a35](https://github.com/guicaulada/claude-plugins/commit/1816a359411cc63f7d145dfd13cc62f464d6a200))
* remove invalid exclude-dirs from golangci-lint v2 config ([686718e](https://github.com/guicaulada/claude-plugins/commit/686718e820e2a4efab489b49da9c34addfee28be))
* remove path relativization, use stored permission_mode ([2d744e5](https://github.com/guicaulada/claude-plugins/commit/2d744e5953e1e59aa21a1f07d2c137d27d651531))
* remove permission_mode from session span, add debug logging ([74806df](https://github.com/guicaulada/claude-plugins/commit/74806df8f246acad3470fa36ee46c67e8973ae65))
* remove self-referencing session.start/end span events ([aa686b8](https://github.com/guicaulada/claude-plugins/commit/aa686b88371d9ba116215919a3d26972cdef9282))
* remove tool name sanitization, wire cardinality controls ([a5fa9fa](https://github.com/guicaulada/claude-plugins/commit/a5fa9fa07d08ff839d627161789b82095747330c))
* resolve all CI golangci-lint errcheck warnings ([7564c4c](https://github.com/guicaulada/claude-plugins/commit/7564c4c66c26f70e97db04d7ab8845abb68c534b))
* resolve all golangci-lint warnings ([4be16f6](https://github.com/guicaulada/claude-plugins/commit/4be16f68ebf82fd60ff3cd2a996484e5943a069a))
* resolve remaining CI errcheck warnings ([80b0b9a](https://github.com/guicaulada/claude-plugins/commit/80b0b9a83088ce066cac55d65765417b8e547278))
* restructure main defer for proper panic recovery ([de7bb89](https://github.com/guicaulada/claude-plugins/commit/de7bb892fb445713341ea70f92f2208c99c28eef))
* retry schema creation for concurrent SQLITE_BUSY ([633fd37](https://github.com/guicaulada/claude-plugins/commit/633fd37d12b24408a5b902460b0254b4da60d280))
* session permission_mode, end_reason, error attrs, line diffs ([bdb7c88](https://github.com/guicaulada/claude-plugins/commit/bdb7c886e434b730f8a13eac4ce839cc61429804))
* set 30s timeout on SessionEnd hook ([a88e6ec](https://github.com/guicaulada/claude-plugins/commit/a88e6ec4adbabb6b8e6df2406e68735ab9de1a61))
* set body and severity on log records, add header debug logging ([22f6cb4](https://github.com/guicaulada/claude-plugins/commit/22f6cb4d5c9ff5e2954fe6ac597cc9a3bdfdee17))
* set OTEL_EXPORTER_OTLP_HEADERS env var from cached headers ([8c0e628](https://github.com/guicaulada/claude-plugins/commit/8c0e6288938845a046077235516fb4d772d8340b))
* span events match parent-child hierarchy ([3b76c3a](https://github.com/guicaulada/claude-plugins/commit/3b76c3a0e7931b474f2385b26464d5fa2f749fbc))
* use cached OTel headers for all event handlers ([e1bf2c1](https://github.com/guicaulada/claude-plugins/commit/e1bf2c1c8918d02b9639f3de437aa2e065ad20a0))
* use correct relative/absolute paths in release-please config ([9e1e0ec](https://github.com/guicaulada/claude-plugins/commit/9e1e0ec390b93b2ad501a3f43921c232561ec400))


### Performance Improvements

* switch to BatchSpanProcessor for efficient bulk export ([64ac1a4](https://github.com/guicaulada/claude-plugins/commit/64ac1a4fa063e10f17b039be03dcff2286212070))
