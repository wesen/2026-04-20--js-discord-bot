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


## 2026-04-20

Added a maintainer-facing analysis of the Glazed env-loading gotcha and listed the docs that should be clarified upstream.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/analysis/01-glazed-documentation-feedback-and-env-loading-gotcha.md — Documentation feedback writeup for Glazed maintainers


## 2026-04-20

Smoke-tested the bot in a detached tmux session, confirmed gateway connection, and verified slash-command sync for the development guild.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/cmd/discord-bot/commands.go — run and sync-commands paths exercised during the tmux smoke test
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/cmd/discord-bot/main.go — Signal-aware CLI entrypoint used in the tmux run session
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Discord gateway connection and slash-command registration verified by smoke test
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/playbook/01-local-validation-and-smoke-test-checklist.md — Repeatable validation checklist now includes the tmux run workflow


## 2026-04-20

Wrote three future-facing notes in the Obsidian vault — a reusable jsverbs integration playbook plus two project reports about a JavaScript API and jsverbs support for this Discord bot — then copied those notes into the ticket `sources/` directory with `cp` for durable ticket-local reference.

### Related Files

- /home/manuel/code/wesen/obsidian-vault/Projects/2026/04/20/ARTICLE - Playbook - Adding jsverbs to Arbitrary Go Glazed Tools.md — Reusable vault note explaining how to add jsverbs to arbitrary Go + Glazed CLIs
- /home/manuel/code/wesen/obsidian-vault/Projects/2026/04/20/PROJ - JS Discord Bot - Building a Discord Bot with a JavaScript API.md — Vault project report for the JavaScript-hosted Discord bot direction
- /home/manuel/code/wesen/obsidian-vault/Projects/2026/04/20/PROJ - JS Discord Bot - Adding jsverbs Support.md — Vault project report for adding jsverbs support to the Discord bot
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/sources/ARTICLE - Playbook - Adding jsverbs to Arbitrary Go Glazed Tools.md — Ticket-local copy of the reusable integration playbook
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/sources/PROJ - JS Discord Bot - Building a Discord Bot with a JavaScript API.md — Ticket-local copy of the JS API project report
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/sources/PROJ - JS Discord Bot - Adding jsverbs Support.md — Ticket-local copy of the jsverbs-support project report


## 2026-04-20

Moved the host-side jsverbs bot CLI into this repository so `discord-bot` now exposes a local `bots list|run|help` surface, while `go-go-goja` is used only as the imported engine/jsverbs dependency. Also added a local `examples/bots` repository and bot-CLI tests here.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/cmd/discord-bot/root.go — Root command now mounts the local `bots` subcommand group
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/bootstrap.go — Local repository bootstrap and duplicate detection for jsverbs bot repositories
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command.go — Local `discord-bot bots list|run|help` command surface
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/runtime.go — Local runtime-backed verb execution built on imported `go-go-goja` packages
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command_test.go — Local smoke coverage for the moved bot CLI
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/bots/README.md — Local real-example repository for testing the moved bot CLI
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/go.mod — Adds the `go-go-goja` dependency via import/replace for local development

