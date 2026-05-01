# Changelog

## 2026-05-01

- Initial workspace created


## 2026-05-01

Created custom-kb example bot, implementation guide, diary, and runtime test for SQLite-backed link storage/search.

### Related Files

- /home/manuel/code/wesen/2026-05-01--discord-ai-in-action/examples/discord-bots/custom-kb/index.js — New custom KB bot implementation
- /home/manuel/code/wesen/2026-05-01--discord-ai-in-action/ttmp/2026/05/01/DISCORD-BOT-CUSTOM-KB--custom-kb-discord-bot/design/01-custom-kb-discord-bot-implementation-guide.md — Implementation guide and source notes


## 2026-05-01

Updated the starter tutorial to prefer the Go-side UI DSL for buttons, selects, messages, embeds, and modals instead of raw component payloads.

### Related Files

- /home/manuel/code/wesen/2026-05-01--discord-ai-in-action/pkg/doc/tutorials/building-and-running-discord-js-bots.md — Starter tutorial now teaches UI DSL-first component and modal examples


## 2026-05-01

Added a pointer from the starter tutorial to the dedicated Go-side UI DSL tutorial.

### Related Files

- /home/manuel/code/wesen/2026-05-01--discord-ai-in-action/pkg/doc/tutorials/building-and-running-discord-js-bots.md — UI DSL starter section links to detailed tutorial


## 2026-05-01

Added Makefile bump-glazed target that bumps all github.com/go-go-golems modules discovered by go list, ran it, and upgraded go-go-goja/glazed/geppetto-related dependency graph.

### Related Files

- /home/manuel/code/wesen/2026-05-01--discord-ai-in-action/Makefile — New bump-glazed target for go-go-golems dependency upgrades
- /home/manuel/code/wesen/2026-05-01--discord-ai-in-action/go.mod — go-go-golems dependency versions bumped by make bump-glazed


## 2026-05-01

Ran custom-kb in tmux with updated local credentials; fixed gateway 4014 by deriving Discord intents from registered bot events; verified the bot connected and synced guild commands.

### Related Files

- /home/manuel/code/wesen/2026-05-01--discord-ai-in-action/examples/discord-bots/custom-kb/index.js — Bot now running in tmux session custom-kb-bot
- /home/manuel/code/wesen/2026-05-01--discord-ai-in-action/internal/bot/bot.go — Gateway intents now derive from JS event descriptors so interaction-only bots do not request disallowed privileged intents

