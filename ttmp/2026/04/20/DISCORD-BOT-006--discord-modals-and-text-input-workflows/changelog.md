# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-006` for modal presentation and modal submit handling in the JavaScript Discord bot API.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Modal response and submit handling entrypoint
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — JS builder/runtime dispatch layer that needs modal surfaces
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js — Example bot should demonstrate a realistic modal workflow

Implemented the first modal slice: added `ctx.showModal(...)`, modal handler registration and descriptor parsing, modal payload normalization, modal submit dispatch with `ctx.values`, runtime tests, and an example modal workflow.
