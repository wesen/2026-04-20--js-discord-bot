# Changelog

Created `DISCORD-BOT-011` to track the next moderation/admin-oriented Discord JS API slice after message and channel moderation utilities. This ticket focuses on guild metadata lookup and role inspection helpers so JavaScript moderation bots can inspect the current guild and available roles before taking administrative actions.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — The request-scoped Discord capability object exposes the new guild and role lookup helpers here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_maps.go — Normalized guild and role snapshots live here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Runtime tests validate the new guild and role lookup helpers here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/index.js — The moderation example bot demonstrates the new lookup helpers

Implemented the Phase 1 core of `DISCORD-BOT-011`: JavaScript can now fetch guild snapshots and list/fetch guild roles through request-scoped `ctx.discord.guilds.*` and `ctx.discord.roles.*` helpers, the host normalizes guild/role payloads into plain JS maps, and the moderation example bot now includes `mod-fetch-guild`, `mod-list-roles`, and `mod-fetch-role`.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Exposed guild and role lookup helpers through the request-scoped Discord capability object
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops.go — Wired guild and role lookup operations into the host ops builder
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_guilds.go — Implemented guild lookup host operations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_roles.go — Implemented role list/fetch host operations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_maps.go — Added normalized guild and role snapshot helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added runtime coverage for guild and role lookup helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-guild-role-lookup-commands.js — Added example commands for the new lookup surface
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Updated example notes and permission guidance for the new lookup helpers
