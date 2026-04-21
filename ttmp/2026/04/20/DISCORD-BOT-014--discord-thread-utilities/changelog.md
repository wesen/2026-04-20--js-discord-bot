# Changelog

Created `DISCORD-BOT-014` to track the next planned Discord JS feature slice in the roadmap.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Request-scoped Discord capability object will grow with thread helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_channels.go — Thread helpers may share channel-host implementation seams here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Runtime tests should validate the new thread APIs here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/support/index.js — Support-style examples are natural consumers of thread utilities

Implemented `DISCORD-BOT-014`: JavaScript can now fetch, join, leave, and start threads through request-scoped `ctx.discord.threads.*` helpers, the support example bot demonstrates those helpers, and the thread snapshot shape is normalized through the existing channel snapshot infrastructure.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Exposed the thread helper namespace through the request-scoped Discord capability object
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops.go — Wired thread operations into the host ops builder
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_threads.go — Implemented thread fetch/join/leave/start host operations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_helpers.go — Added thread start payload normalization helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_maps.go — Extended normalized channel snapshots with thread-relevant fields
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added runtime coverage and normalization checks for thread helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/support/index.js — Added support-oriented thread utility commands
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Updated example notes and permission guidance for the new thread helpers

Decision for ticket scope: archive/lock management is intentionally deferred to a later focused follow-up instead of widening this ticket beyond fetch/join/leave/start.
