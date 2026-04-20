# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-010` for the next Discord moderation/admin utility slice after member moderation. This ticket focuses on message inspection/moderation helpers and small channel utility APIs rather than broader guild-wide admin CRUD.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-010--discord-message-and-channel-moderation-utilities/index.md — Ticket index for the new moderation utility work
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-010--discord-message-and-channel-moderation-utilities/design-doc/01-discord-message-and-channel-moderation-utilities-architecture-and-implementation-guide.md — Detailed design and phase ordering
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-010--discord-message-and-channel-moderation-utilities/reference/01-discord-message-and-channel-moderation-utilities-api-reference-and-planning-notes.md — Quick API sketches and normalized payload direction
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-010--discord-message-and-channel-moderation-utilities/tasks.md — Detailed task list for phased implementation

Implemented Phase 1 of `DISCORD-BOT-010`: JavaScript can now fetch a message, pin/unpin a message, and list pinned messages through `ctx.discord.messages.*`. The moderation example bot now demonstrates those utilities, and it was split into focused registration modules so it remains readable as the admin surface grows.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Exposed message fetch/pin/unpin/listPinned through the request-scoped Discord capability object
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Implemented host message moderation operations and normalized fetched/pinned message payloads
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added runtime coverage for the Phase 1 message moderation utilities
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/index.js — Refactored the moderation example into composed registration modules
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-message-moderation-commands.js — Added example commands for message fetch/pin/unpin/listPinned
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Updated example repository notes for the new Phase 1 utilities and the in-bot composition split

Implemented Phase 2 of `DISCORD-BOT-010`: JavaScript can now bulk delete messages through `ctx.discord.messages.bulkDelete(...)`, the host accepts practical message-ID list payload forms and emits structured logs for the destructive operation, and the moderation example bot now includes a `mod-bulk-delete` command.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Exposed `bulkDelete` through the request-scoped Discord message capability surface
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Implemented bulk-delete host operations and message-ID payload normalization
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added runtime coverage for message bulk deletion
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-message-moderation-commands.js — Added the example bulk-delete command
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Updated example notes to mention the bulk-delete utility
