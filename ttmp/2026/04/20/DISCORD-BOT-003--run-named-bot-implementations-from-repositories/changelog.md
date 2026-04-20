# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-003` to rewrite the `bots` CLI around named bot implementations and to add example bot repositories that exercise the current Discord JS host capabilities.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-003--run-named-bot-implementations-from-repositories/index.md — Ticket index for the bot repository runner work
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-003--run-named-bot-implementations-from-repositories/design-doc/01-bot-repository-runner-architecture-and-implementation-guide.md — Implementation guide for discovery, multi-bot composition, and CLI behavior
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-003--run-named-bot-implementations-from-repositories/reference/01-bot-repository-cli-reference-and-example-repositories.md — Quick-reference CLI and repository examples

## 2026-04-20

Rewrote the `bots` command group around named bot implementations instead of jsverb-like functions, added descriptor-based discovery and multi-bot runtime composition, and created a full example repository of Discord bot implementations.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/bootstrap.go — Repository discovery now scans bot scripts and loads bot descriptors via the local Discord JS host
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command.go — `bots list|help|run <bot...>` now operates on named bot implementations
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor.go — New bot descriptor and script inspection model
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/multihost.go — Multi-bot runtime composition with command routing and event fan-out
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Live host now supports loading multiple bot scripts at once
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Example repository of named bot implementations

## 2026-04-20

Added more operator-facing visibility for `bots run`: a `--print-parsed-values` flag that prints the resolved Discord config plus selected bot descriptors and exits, plus startup/sync logs that show which bot implementations and slash commands were actually loaded.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command.go — `bots run` now accepts `--print-parsed-values`
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/runtime.go — Printed run-request summary includes redacted config and selected bot metadata
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go — Startup logging now shows loaded bot implementations and synced command names
