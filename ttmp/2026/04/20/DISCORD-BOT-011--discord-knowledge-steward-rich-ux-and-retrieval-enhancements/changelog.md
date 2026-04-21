---
Title: Discord Knowledge Steward Rich UX and Retrieval Enhancements Changelog
Ticket: DISCORD-BOT-011
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: reference
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: >
  Change log for the rich review, reaction promotion, citation, search, and export ticket.
LastUpdated: 2026-04-20T23:15:00-04:00
WhatFor: Record major changes and decisions for the ticket.
WhenToUse: Use when reviewing the implementation history of the rich UX slice.
---

# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-011` to collect the knowledge steward bot's next UX and retrieval slice: rich review UI, reaction-based promotion, source citation UX, rich search UI, and export-to-channel behavior.

Added the design guide, implementation sketches, and diary for the ticket. The documents now explain the interaction model, the review-state approach, the reaction trust policy, and the future search/export work.

Implemented the first two slices of the ticket in JavaScript: the review queue now uses a select menu, buttons, and an edit modal, and the bot now promotes candidate knowledge from trusted reactions with visible replies. The runtime test covers both flows against a real SQLite store.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js — Bot wiring for the rich review queue and reaction promotion
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/review.js — Review state, queue rendering, action handling, and edit modal helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/reactions.js — Trusted reaction promotion helpers
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/knowledge_base_runtime_test.go — Runtime validation for review and reaction behavior
- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-011--discord-knowledge-steward-rich-ux-and-retrieval-enhancements/tasks.md — Updated task list showing the completed and remaining slices
