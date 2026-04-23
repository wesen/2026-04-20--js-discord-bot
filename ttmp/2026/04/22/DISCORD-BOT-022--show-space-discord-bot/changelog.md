# Changelog

## 2026-04-22

- Created the DISCORD-BOT-022 workspace for the Show Space Discord Bot.
- Captured the developer spec verbatim in the implementation guide.
- Mapped the bot to the repository’s embedded JavaScript runtime instead of external Discord.js/discord.py stacks.
- Identified the current runtime capabilities needed for Phase 1: command handlers, permission checks, channel sends, pin/unpin/listPinned, and runtime configuration.
- Flagged the Phase 2 persistence mismatch between the spec’s Postgres preference and the repository’s current `require("database")` SQLite-oriented runtime module.
- Step 1 complete: added `examples/discord-bots/show-space/index.js` as a discoverable bot shell and verified `bots list` shows `show-space` (commit 1616557).
