# Changelog

## 2026-04-21

- Initial workspace created


## 2026-04-21

Uploaded design document, diary, and Discord API source to reMarkable at /ai/2026/04/21/DISCORD-BOT-020

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-020--discord-interaction-types-demo-bot-with-user-commands-message-commands-and-subcommands/design/01-design-and-implementation-guide.md — Design and implementation guide


## 2026-04-21

Step 2: Added userCommand, messageCommand, subcommand to JS bot API runtime (commit 08b92cb)

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — Added commandType
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor.go — Added CommandDescriptor.Type
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime.go — Exposed userCommand


## 2026-04-21

Step 3: Updated command sync and interaction dispatch for user/message/subcommand types (commit 09fc236)

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_commands.go — Added command type mapping and sub_command/sub_command_group option types
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_dispatch.go — Three-way dispatch for user


## 2026-04-21

Step 4: Created interaction-types demo bot (commit e16bda2)

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/interaction-types/index.js — Demo bot with hello


## 2026-04-21

Step 5: Updated docs and finalized design document with implementation log (commit bfdc56d, b8fcf87)

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Listed interaction-types bot
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/pkg/doc/topics/discord-js-bot-api-reference.md — Added userCommand


## 2026-04-21

Step 6: Uploaded final docs and demo bot source to reMarkable at /ai/2026/04/21/DISCORD-BOT-020


## 2026-04-21

Changed subcommands from admin/kick/ban to fun/roll/coin for safer demo; committed all ticket docs (commit eaf8258)

