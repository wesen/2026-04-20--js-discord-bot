# Changelog

## 2026-04-21

- Initial workspace created


## 2026-04-21

Created ticket with design-doc and diary. Wrote comprehensive 31KB analysis/design/implementation guide covering Discord API primer, architecture diagrams, pseudocode, rendering logic, file output structure, and 7-phase implementation plan.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-ARCHIVE-BOT--discord-archive-helper-bot-thread-message-markdown-downloader/design-doc/01-analysis-design-implementation-guide.md — Main deliverable


## 2026-04-21

Uploaded bundled PDF to reMarkable: /ai/2026/04/21/DISCORD-ARCHIVE-BOT/DISCORD-ARCHIVE-BOT Guide

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-ARCHIVE-BOT--discord-archive-helper-bot-thread-message-markdown-downloader/design-doc/01-analysis-design-implementation-guide.md — Uploaded as PDF bundle


## 2026-04-21

Updated embedded docs: added runtime environment warning to tutorial and API reference; added message history pagination example; added framework-at-a-glance section.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/pkg/doc/tutorials/building-and-running-discord-js-bots.md — Added NOT Node.js constraints box and 3-sentence overview


## 2026-04-21

Uploaded updated guide v2 to reMarkable: /ai/2026/04/21/DISCORD-ARCHIVE-BOT/DISCORD-ARCHIVE-BOT Guide v2.pdf

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-ARCHIVE-BOT--discord-archive-helper-bot-thread-message-markdown-downloader/design-doc/01-analysis-design-implementation-guide.md — Rewritten for defineBot framework


## 2026-04-21

Design doc v3: changed thread archive to messageCommand (right-click message in thread); added input type reference section explaining no calendar widgets available; added before_message_id option for channel archive as time anchor; updated fetchAllMessages helper.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-ARCHIVE-BOT--discord-archive-helper-bot-thread-message-markdown-downloader/design-doc/01-analysis-design-implementation-guide.md — v3 changes


## 2026-04-21

Updated input type reference: mapped all 11 Discord ApplicationCommandOptionTypes against framework support; confirmed ATTACHMENT (type 11) is not yet supported; confirmed no date/calendar types exist in Discord's API at all; cited official Discord docs as source.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-ARCHIVE-BOT--discord-archive-helper-bot-thread-message-markdown-downloader/design-doc/01-analysis-design-implementation-guide.md — Input type accuracy pass


## 2026-04-21

Identified 6 framework gaps during design review; spun off into ticket DISCORD-BOT-FRAMEWORK-GAPS for future implementation (ATTACHMENT, channel_types, permissions, nsfw, contexts, localizations).

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK-GAPS--discord-js-bot-framework-missing-ui-surface-types-command-metadata/design-doc/01-framework-gaps-analysis-implementation-plan.md — Related framework gaps ticket


## 2026-04-21

Implemented archive-helper bot scaffold: fetcher.js (pagination + before_message_id), renderer.js (markdown with frontmatter), index.js (/archive-channel command + Archive Thread messageCommand + ready event + default_limit config). Verified with bots list/help. Committed e6a2e1c.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/archive-helper/index.js — Main bot file


## 2026-04-21

Completed all implementation tasks: error handling added to all async operations (e6a2e1c, 3b9cf33, 33d3258), README written (a5ef38e), Go build validates. Bot is ready for testing in a live Discord server.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/archive-helper/README.md — Usage documentation


## 2026-04-21

Fixed framework error logging: unregistered events (messageCreate, messageUpdate, etc.) no longer log ERROR when a bot doesn't handle them. Commit 684ee7e in internal/jsdiscord/bot.go.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go — dispatchEvent now returns empty results instead of panicking for unregistered events


## 2026-04-21

Fixed thread starter messages appearing empty: enriched framework messageMap with timestamp, type, messageReference, attachments, embeds, mentions, editedTimestamp, referencedMessage. Bot now resolves THREAD_STARTER_MESSAGE (type 21) placeholders by fetching the actual starter message from the parent channel. Renderer updated for real timestamps, attachments, embeds, and edit badges. Commit a3a0362.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_maps.go — Enriched messageMap with 8 new fields

