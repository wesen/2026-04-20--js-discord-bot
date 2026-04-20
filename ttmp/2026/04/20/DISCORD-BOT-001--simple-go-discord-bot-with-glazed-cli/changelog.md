# Changelog

## 2026-04-20

- Initial workspace created
- Added a detailed design guide for the Go Discord bot with a Glazed CLI
- Added a credential/setup reference describing the Discord values needed for local development and bot invites
- Added a diary entry and related all ticket documents for easier navigation

## 2026-04-20

Implemented the Go module, Glazed CLI root and subcommands, Discord config/session helpers, and a local validation playbook; corrected the Glazed env-loading path so .envrc values load through the built-in env source.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/cmd/discord-bot/commands.go — run/validate-config/sync-commands implementations and env-loading fix
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/cmd/discord-bot/main.go — Signal-aware entrypoint for the bot CLI
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/cmd/discord-bot/root.go — Glazed/Cobra root command and logging/help wiring
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Discord session lifecycle and slash-command handling
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/config/config.go — Shared config parsing and validation
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/playbook/01-local-validation-and-smoke-test-checklist.md — Repeatable local smoke-test commands

