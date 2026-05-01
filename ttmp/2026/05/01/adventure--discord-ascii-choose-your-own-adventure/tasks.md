# Tasks

## TODO

- [x] Milestone 1: create examples/discord-bots/adventure/index.js as a repo-discovered defineBot entrypoint with configure metadata and runtime config keys
- [x] Milestone 1: add haunted-gate seed adventure in JS seed data and seed it into SQLite with tone, constraints, initial stats, inventory/flag vocabulary, and opening scene prompt
- [x] Milestone 1: implement SQLite-backed store abstraction for seeds, sessions, scenes, choices, and audit logs using require("database")
- [x] Milestone 1: implement canonical adventure engine helpers for state creation, turn advancement, effect validation/clamping, stale turn checks, and audit logging
- [x] Milestone 1: implement Go-owned OpenRouter LLM host module exposed to JS as a narrow function; Go controls API key, model, max tokens, base URL, and provider settings, defaulting to anthropic/claude-3.5-haiku
- [x] Milestone 1: implement prompt assembly and JSON response parsing for scene_patch generation
- [x] Milestone 1: implement ASCII scene card renderer that fits Discord message limits and produces choice buttons plus Try something else
- [x] Milestone 1: implement /adventure-start command to create a session, call the LLM, validate scene JSON, persist state, and render the first scene
- [x] Milestone 1: implement choice component handler with custom ID parsing, owner/session/turn validation, LLM next-scene generation, and message update
- [x] Milestone 1: implement free-form modal flow: Try something else button opens modal, modal text is interpreted into validated action JSON, then advances scene
- [x] Milestone 1: persist per-session audit logs in SQLite containing player inputs, raw LLM responses, parsed JSON, validation results, and applied effects
- [x] Milestone 1: add README/run instructions for adventure bot, including OpenRouter API key setup, Go-owned LLM defaults, and discord-bot run command
