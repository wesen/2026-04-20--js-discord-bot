# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-002` to move the sandbox-style JS host functionality into `js-discord-bot` and start a real Discord-specific JavaScript bot API implementation here.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-002--move-sandbox-here-and-implement-real-javascript-discord-bot-api/index.md — New ticket index for the move-and-implement work
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-002--move-sandbox-here-and-implement-real-javascript-discord-bot-api/design-doc/01-sandbox-move-and-discord-javascript-api-architecture-guide.md — Architecture plan for the local Discord JS bot runtime
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-002--move-sandbox-here-and-implement-real-javascript-discord-bot-api/reference/01-discord-javascript-bot-api-reference-and-example-script.md — API reference and example script for the local Discord JS bot API

## 2026-04-20

Ported the sandbox-style runtime-local JS host layer into `internal/jsdiscord`, renamed the JS entrypoint to `require("discord")`, and integrated the live Discord host so slash commands and `ready` events can dispatch into JavaScript handlers in this repository.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime.go — Runtime-scoped registrar and `require("discord")` module loader
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — JS bot builder, dispatch contract, context object, store, logging, and async settlement
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Live Discord host bridge for command sync and interaction/event dispatch
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Bot runtime now optionally loads and runs a JavaScript bot script
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/config/config.go — New `bot-script` setting for loading a JavaScript bot definition
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/js-bots/ping.js — First local JavaScript Discord bot example

## 2026-04-20

Expanded the local Discord JS bot API with richer response payloads and broader live event coverage. JavaScript handlers can now use embeds, action-row/button components, deferred interaction responses, response edits, follow-up messages, and additional `guildCreate` / `messageCreate` events.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — Response normalization now supports embeds, components, deferred/edit/follow-up flows, and additional live event dispatch paths
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — JS context now exposes `message`, `followUp`, and `edit` in addition to `reply` and `defer`
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Live Discord host now registers `guildCreate` and `messageCreate` handlers and enables the required intents
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go — Added tests for richer payload normalization and message-event context wiring
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/js-bots/ping.js — Example bot now demonstrates embeds, deferred edit/follow-up flows, and message event handling
