# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-009` for richer inbound Discord events and moderation/admin host APIs in the JavaScript runtime.

Expanded the task plan for `DISCORD-BOT-009` into concrete implementation slices and added a dedicated diary document. The ticket now breaks event expansion into message lifecycle, reaction, and guild-member phases before moderation/admin host APIs, which makes the work easier to implement and review in smaller commits.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-009--discord-event-expansion-and-moderation-admin-apis/tasks.md — Detailed step-by-step execution plan for event and moderation work
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-009--discord-event-expansion-and-moderation-admin-apis/reference/02-diary.md — Chronological diary for the implementation work
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-009--discord-event-expansion-and-moderation-admin-apis/index.md — Updated ticket index with links to the new diary

Implemented Phase 1A of `DISCORD-BOT-009`: the live bot now forwards `messageUpdate` and `messageDelete` into the JavaScript host, the runtime context exposes `ctx.before` for cached prior message state, host normalization handles update/delete-safe message payloads, and the moderation example bot now demonstrates message edit/delete event logging.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Added discordgo session handlers for `messageUpdate` and `messageDelete`
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Added `DispatchMessageUpdate`, `DispatchMessageDelete`, and update/delete-safe message normalization helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Extended dispatch/context wiring with `before`, `member`, and `reaction` fields for richer event payloads
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added regression coverage for `messageUpdate` and `messageDelete`
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/index.js — Example bot now demonstrates message lifecycle event logging
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Updated example repository notes to mention message lifecycle events

Implemented Phase 1B of `DISCORD-BOT-009`: the live bot now forwards `reactionAdd` and `reactionRemove` into the JavaScript host, the runtime context exposes `ctx.reaction` and reaction-related actor/member payloads, and the moderation example bot now logs reaction activity alongside message lifecycle events.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Added discordgo session handlers and required reaction intent wiring
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Added reaction dispatch methods plus reaction/member/emoji normalization helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added regression coverage for `reactionAdd` and `reactionRemove`
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/index.js — Example bot now demonstrates reaction event logging
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Updated example repository notes to mention reaction event coverage
