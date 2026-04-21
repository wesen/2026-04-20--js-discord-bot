# Changelog

Created `DISCORD-BOT-012` to track the next moderation/admin-oriented Discord JS API slice after guild and role lookup helpers. This ticket focuses on read-only member fetch/list helpers so JavaScript moderation bots can inspect member details before applying existing moderation operations.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — The request-scoped Discord capability object exposes the new member lookup helpers here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_members.go — Member lookup and moderation operations live here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Runtime tests validate the new member lookup helpers here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-member-moderation-commands.js — The moderation example demonstrates the new member lookup helpers

Implemented the Phase 1 core of `DISCORD-BOT-012`: JavaScript can now fetch one guild member and list a small page of members through request-scoped `ctx.discord.members.fetch(...)` and `ctx.discord.members.list(...)`, the host normalizes member payloads into plain JS maps, and the moderation example bot now includes `mod-fetch-member` and `mod-list-members`.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Exposed member fetch/list helpers through the request-scoped Discord capability object
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_helpers.go — Added list option normalization for member lookup paging
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_members.go — Implemented member fetch/list host operations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added runtime coverage for member lookup helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-member-moderation-commands.js — Added example commands for the new member lookup surface
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Updated example notes and visibility guidance for the new member lookup helpers
