---
Title: Investigation Diary
Ticket: DISCORD-BOT-CUSTOM-KB
Status: active
Topics: [backend, chat, javascript, goja, sqlite]
DocType: reference
Intent: long-term
Summary: Chronological diary for implementing the custom SQLite-backed KB link bot.
LastUpdated: 2026-05-01T15:25:00-04:00
---

# Investigation Diary

## 2026-05-01

- Created docmgr ticket `DISCORD-BOT-CUSTOM-KB` for a custom KB Discord bot.
- Inspected the existing `knowledge-base` bot, especially `lib/store.js`, to reuse the SQLite store-factory pattern.
- Read UI DSL docs and examples: `using-the-go-side-ui-dsl-for-discord-bots.md`, `ui-showcase/index.js`, and prior UI DSL ticket docs.
- Implemented `examples/discord-bots/custom-kb/` with SQLite link storage, add/search/list commands, modal input, select results, and refresh button.
- Added runtime test `internal/jsdiscord/custom_kb_runtime_test.go` for save/search behavior.
- Hit `GoError: Invalid module` while loading `require("database")`. The same error occurred in the existing knowledge-base SQLite test, confirming this was runtime composition rather than the new bot.
- Inspected `../corporate-headquarters/go-go-goja` and compared it to the pinned `go-go-goja v0.4.12` in this repo. The newer branch auto-selects default modules for a plain builder and adds a `db` alias; v0.4.12 requires explicit host-access module opt-in.
- Fixed the Discord runtime factory sites by adding `WithModules(engine.DefaultRegistryModulesNamed("database"))` in `internal/jsdiscord/host.go`, `internal/jsdiscord/helpers_test.go`, and `pkg/botcli/runtime_factory.go`.
- Restored the bot to use the Go-side `require("ui")` module directly.
- Validation now passes for `go test ./internal/jsdiscord -run TestCustomKB -count=1`, the existing knowledge-base SQLite test, and `go test ./pkg/botcli ./internal/jsdiscord -count=1`.
- Added `Makefile` target `bump-glazed`, modeled after `../corporate-headquarters/pinocchio`, but generalized to bump every `github.com/go-go-golems/*` module discovered by `go list -m all`.
- Ran `make bump-glazed`; this upgraded `github.com/go-go-golems/glazed` from `v1.2.3` to `v1.2.6`, `github.com/go-go-golems/go-go-goja` from `v0.4.12` to `v0.4.15`, and refreshed transitive dependencies via `go mod tidy`.
- Re-ran `go test ./pkg/botcli ./internal/jsdiscord -count=1` and `go run ./cmd/discord-bot bots help custom-kb --bot-repository ./examples/discord-bots`; both passed.
- Updated the local ignored `.envrc` with the provided Discord application credentials and ran the bot in tmux session `custom-kb-bot`.
- First live run failed with Discord gateway close `4014: Disallowed intent(s)` because the generic bot host requested privileged member/message-content intents even though `custom-kb` is interaction-only.
- Fixed `internal/bot/bot.go` so gateway intents are derived from the JS bot event descriptors: interaction-only bots request only `Guilds`, while message/member/reaction bots request the extra intents they actually need.
- Re-ran the bot in tmux; it connected as the expected new application user, synced `/kb-add`, `/kb-link`, `/kb-search`, and `/kb-list` to the development guild, initialized the SQLite DB, and stayed running.
- Verified the registered guild commands through Discord's REST API using the bot token from the environment, without printing the token.

## Next follow-up

- If/when this repo upgrades to the newer `go-go-goja`, decide whether to rely on the new auto-default behavior or keep explicit `database` opt-in for a safer Discord bot sandbox.
