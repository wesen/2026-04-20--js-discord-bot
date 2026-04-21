---
Title: Discord Knowledge Steward Bot Tasks
Ticket: DISCORD-BOT-010
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: task-list
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: >
  Work plan for a Discord bot that listens to chat, records candidate knowledge, and lets the community verify, search, and maintain a shared knowledge base.
LastUpdated: 2026-04-20T22:20:00-04:00
WhatFor: Track implementation and documentation tasks for the knowledge steward bot.
WhenToUse: Use when executing the bot build or checking progress on the ticket.
---

# Tasks

## Documentation

- [x] Create ticket workspace and scaffold the ticket documents
- [x] Gather evidence from the current JS runtime, examples, and CLI help system
- [x] Write the design guide
- [x] Write the implementation guide
- [x] Write the investigation diary
- [x] Relate the design and implementation docs to the most relevant source files
- [x] Validate the ticket with `docmgr doctor --ticket DISCORD-BOT-010 --stale-after 30`
- [x] Upload the document bundle to reMarkable and verify the remote listing

## Phase 1 — passive capture

- [ ] Decide the initial storage backend for knowledge entries and drafts
- [ ] Add a message listener that detects candidate knowledge from `messageCreate`
- [ ] Scope capture to opted-in channels or guilds so the bot remains transparent
- [ ] Record source metadata for every candidate: guild, channel, message, author, timestamp, and jump URL
- [ ] Emit a visible audit breadcrumb when a message becomes a draft knowledge entry
- [ ] Add reaction-based promotion signals for useful messages

## Phase 2 — curation and review

- [ ] Add `/remember` or `/teach` entry points for explicit capture
- [ ] Add a modal-based editor for title, summary, tags, and source attribution
- [ ] Add a review queue for draft entries
- [ ] Add verify / ignore / stale / merge actions for knowledge cards
- [ ] Preserve history when an entry is edited or promoted
- [ ] Add a duplicate-detection pass before creating a new entry

## Phase 3 — retrieval and synthesis

- [ ] Add `/ask` for conversational retrieval
- [ ] Add `/search` and `/article` for direct lookup
- [ ] Add source citation rendering so answers show where knowledge came from
- [ ] Add autocomplete for tags, aliases, and article names
- [ ] Add a `recent` or `queue` view for operators

## Phase 4 — maintenance and safety

- [ ] Add staleness tracking and revalidation flows
- [ ] Add export support for the curated knowledge corpus
- [ ] Add tests for the capture, review, and retrieval paths
- [ ] Add smoke checks for the named-bot runner flow
- [ ] Document permissions, intent requirements, and transparency expectations

## MVP implementation

- [x] Add a SQLite-backed knowledge store helper in `examples/discord-bots/knowledge-base/lib/` using `require("database")`
- [x] Refactor the knowledge-base bot into modular JS files for capture, store, rendering, and commands
- [x] Implement the MVP capture/review/retrieval flow: passive capture, `/teach`, `/kb-search`, `/kb-article`, and a review queue
- [x] Add runtime coverage and example README updates for the SQLite-backed knowledge steward bot

## Notes

- Keep the first implementation small enough to ship as a named bot under `examples/discord-bots/`.
- Prefer human-readable knowledge cards with sources and status over hidden machine state.
- If the store format changes later, migrate the entry schema rather than weakening the audit trail.
