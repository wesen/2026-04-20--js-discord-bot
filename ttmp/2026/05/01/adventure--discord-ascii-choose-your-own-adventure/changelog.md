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

