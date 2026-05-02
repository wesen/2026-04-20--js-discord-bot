# Changelog

## 2026-05-01

- Initial workspace created


## 2026-05-01

Refined implementation proposal: JavaScript owns Discord/game orchestration while Go owns OpenRouter LLM access and controls provider settings such as model, token, base URL, and max tokens.

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/adventure--discord-ascii-choose-your-own-adventure/design/02-implementation-proposal.md — Updated implementation architecture for Go-owned LLM module
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/adventure--discord-ascii-choose-your-own-adventure/tasks.md — Updated Milestone 1 tasks for Go-owned LLM host module


## 2026-05-01

Switched design from YAML-defined scenes to SQLite-defined scenes and JSON LLM contracts because the JS runtime has database support but no native YAML/file access.

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/adventure--discord-ascii-choose-your-own-adventure/design/02-implementation-proposal.md — Revised data model from YAML files to SQLite tables and JSON contracts
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/adventure--discord-ascii-choose-your-own-adventure/tasks.md — Updated Milestone 1 tasks for SQLite-defined scenes and JSON parsing


## 2026-05-01

Step 1: Added Go-owned OpenRouter LLM adapter exposed as adventure_llm.completeJson; focused tests pass (commit f9a47bf).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/openrouter_module.go — OpenRouter adapter implementation
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/openrouter_module_test.go — Adapter tests


## 2026-05-01

Step 2: Added repo-discovered JS adventure bot prototype with SQLite-defined scenes, JSON LLM contracts, buttons, free-form modal flow, and audit persistence (commit 3feb8ef).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/index.js — Adventure bot entrypoint
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/store.js — SQLite persistence layer


## 2026-05-01

Improved adventure interaction flow: loading edits after clicks, modal submits now defer/update original message when possible, LLM can mark final scenes, final messages attach a session JSON export, and multiplayer ownership tests cover non-owner clicks (commit a0f3097).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/index.js — Loading states and final-scene rendering flow
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/adventure_runtime_test.go — Multiplayer ownership and stale-click tests
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/host_responses.go — Modal submissions from component messages now defer/update the original message


## 2026-05-01

Made adventure sessions channel-scoped collaborative: any channel member can advance the active session; loading cards now include previous scene context, selected action/free-form text, and actor name; tests cover another player advancing without breaking the starter's session (commit 83a7177).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/index.js — Collaborative channel session behavior and actor-aware loading calls
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/render.js — Loading message includes previous scene
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/adventure_runtime_test.go — Collaborative multiplayer runtime coverage


## 2026-05-01

Added Goja callback streaming for OpenRouter: adventure_llm.streamJson reads SSE chunks and invokes a JS callback; adventure JS throttles chunk previews into Discord loading-message edits while preserving final JSON parsing (commit 8f88f70).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/index.js — Progress editor sends streamed draft previews to Discord
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/llm.js — Uses streamJson when an onChunk callback is provided
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/openrouter_module.go — OpenRouter SSE streaming and JS callback bridge


## 2026-05-01

Replaced final downloadable JSON export with an in-message coda and history navigation that works on both Discord and Slack (commit 28ee3c2).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/index.js — History/resume can load completed sessions
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/render.js — Final scene now renders coda/lookback instead of files
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/store.js — Added latest active/completed session lookup for final history navigation


## 2026-05-01

Added Slack-specific backfill command to replace old final export messages with coda/lookback Slack messages (commit be743fa).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/cmd/discord-bot/root.go — Registers slack-adventure-coda-backfill command
- /Users/kball/git/go-go-golems/discord-bot/cmd/discord-bot/slack_adventure_coda.go — Backfill command for old Slack adventure export endings
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/adventure--discord-ascii-choose-your-own-adventure/reference/02-diary.md — Recorded Step 8


## 2026-05-01

Changed completed-session history navigation so the return button says Coda, making it clear users can jump back to the ending after scrolling backward (commit 09d0466).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/render.js — History return button label for completed adventures


## 2026-05-01

Added coda storyboard image generation from the full adventure using Go-owned OpenRouter image generation with Gemini Flash Image (commits da9c47f, 0402dba).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/engine.js — Generates full-story storyboard at coda time
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/render.js — Attaches storyboard output to coda
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/openrouter_module.go — Added adventure_llm.generateImage and image response parsing
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/slack_backend.go — Uploads generated image files to Slack thread


## 2026-05-01

History scenes and coda lookback now show the action that led to each scene and who took it (commit 02ca618).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/index.js — Passes actor into engine inputs
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/engine.js — Carries actor through choice/freeform and terminal endings
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/render.js — Renders action/actor in history scenes and coda lookback
- /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/lib/store.js — Loads scene action metadata from audit records

