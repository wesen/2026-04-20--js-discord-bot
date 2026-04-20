# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-004` to document the decision to simplify the runtime back to a single JavaScript bot implementation per process. The design package explains why multi-bot composition should move out of the host/runtime layer and into the selected bot implementation itself when needed.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-004--simplify-to-a-single-javascript-bot-per-process/index.md — Ticket index for the simplification work
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-004--simplify-to-a-single-javascript-bot-per-process/design-doc/01-single-javascript-bot-per-process-architecture-and-implementation-guide.md — Detailed design and implementation guide for the simplification
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-004--simplify-to-a-single-javascript-bot-per-process/reference/01-single-bot-runner-reference-and-migration-notes.md — Operator-facing reference and migration notes

Implemented the first cleanup slice for `DISCORD-BOT-004`: the live host now loads exactly one selected JavaScript bot script, `discord-bot bots run` accepts exactly one bot selector, parsed-values output reflects a single selected bot, and the multi-host composition layer is no longer used in the main runtime path. The host also now errors clearly when no JavaScript bot script is configured, which matches the decision that live bot execution should happen through `bots run` or an explicit `--bot-script` selection instead of an old fallback example bot.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Simplified the live host to one selected bot script and removed multi-host usage from the main path
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command.go — Changed `bots run` back to a single-bot command contract
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/runtime.go — Simplified run request handling and parsed-values output to one selected bot
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command_test.go — Updated tests to reflect the single-bot runner model

Implemented the next `DISCORD-BOT-004` slice for bot startup config: bot descriptors now parse `configure({ run: ... })`, `bots help <bot>` prints the available runtime fields, `bots run <bot>` resolves those fields through a dynamic Glazed/Cobra parser after bot selection, and the resulting values are injected into JavaScript as `ctx.config`. The first example rollout updates `knowledge-base` to demonstrate optional `index_path` and `read_only` startup config and shows the parsed runtime config in `--print-parsed-values` output.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor.go — Added typed parsing for `run` metadata and normalized runtime field names
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Added `ctx.config` to the runtime context and dispatch contract
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Carried runtime config into live dispatch requests
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/run_schema.go — Added two-stage parsing and dynamic Glazed/Cobra run-field parsing
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command.go — Wired `bots run` through the new pre-parse + dynamic runtime-config parse flow
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/runtime.go — Added runtime-config propagation and parsed-values reporting
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js — Demonstrates `configure({ run: ... })` and `ctx.config`

Added a focused postmortem document for the ping bot `/search` failure. The analysis explains that the bug was caused by the example bot assuming a browser-style `setTimeout` global in a module-based Goja runtime, and that the confusing `promise rejected: map[]` log line came from lossy promise rejection formatting rather than from the command router itself. The postmortem also records the follow-up fixes: switching the example bot to `require("timer")`, improving promise rejection diagnostics, and adding clearer interaction debug context.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-004--simplify-to-a-single-javascript-bot-per-process/analysis/01-ping-bot-search-failure-postmortem.md — Maintainer-facing bug report and postmortem for the ping bot search failure
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js — Example bot fix that switched from `setTimeout` to the timer module
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Improved rejected-promise diagnostics so JavaScript errors are no longer reduced to `map[]`
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Added more contextual interaction dispatch logging and error wrapping

Completed the next cleanup and observability slice for `DISCORD-BOT-004`: the obsolete `internal/jsdiscord/multihost.go` and `internal/jsdiscord/multihost_test.go` files were deleted, descriptor coverage was moved into a dedicated `descriptor_test.go`, and the live single-bot host gained richer lifecycle debug logs for defer/reply/edit/follow-up/modal flows plus request-scoped Discord operations. This keeps the codebase aligned with the single-bot architecture decision and makes it easier to trace which script and interaction path produced a given Discord action.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Added structured lifecycle debug logging for interaction responses and host Discord operations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor_test.go — Preserved descriptor and fallback-name coverage after removing the multi-host test file
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/multihost.go — Deleted obsolete multi-bot runtime composition code
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/multihost_test.go — Deleted obsolete multi-bot runtime tests
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-004--simplify-to-a-single-javascript-bot-per-process/reference/01-single-bot-runner-reference-and-migration-notes.md — Added an explicit explanation for why the multi-host layer was deleted instead of merely left unused
