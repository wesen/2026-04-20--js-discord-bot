# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-005` for Discord component interactions and richer message component payloads in the JavaScript bot API.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Current interaction dispatch and component normalization entrypoint
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — JS builder/runtime dispatch layer that needs component registrations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js — Example bot that should grow component-driven workflows

Implemented the first component slice: added `component(customId, handler)`, routed Discord message component interactions into JavaScript, expanded outgoing component normalization with select menus, added runtime tests, and updated the example bot to demonstrate button/select flows.
