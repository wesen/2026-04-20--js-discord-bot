# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-007` for autocomplete support and richer command option metadata in the JavaScript Discord bot API.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Command option normalization and autocomplete responses live here
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — JS builder/runtime layer that needs autocomplete registrations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Tests should prove focused option dispatch and response choices

Implemented the first autocomplete and richer-option slice: added JS autocomplete registrations, focused-option dispatch, autocomplete response normalization, command option support for `autocomplete`, static `choices`, and length/value constraints, plus runtime tests and an example autocomplete command.
