# Changelog

Created `DISCORD-BOT-013` to track the next planned Discord JS feature slice in the roadmap.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Request-scoped Discord capability object will grow with message history and listing helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_messages.go — Message history/listing host operations will live alongside existing message moderation operations here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_helpers.go — Message list option normalization will live here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Runtime tests should validate the new message history/listing APIs here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-message-moderation-commands.js — The moderation example should demonstrate message history/listing helpers here

Implemented the Phase 1 core of `DISCORD-BOT-013`: JavaScript can now list bounded message history through `ctx.discord.messages.list(...)`, the host supports a narrow `before`/`after`/`around`/`limit` payload shape, and the moderation example bot now includes `mod-list-messages`.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Exposed the message history/listing helper through the request-scoped Discord capability object
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_helpers.go — Added bounded message list option normalization
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_messages.go — Implemented message listing host operations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added runtime coverage for message listing helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-message-moderation-commands.js — Added the example message history command
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Updated example notes and visibility guidance for the new message history helper
