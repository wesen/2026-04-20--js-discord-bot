---
Title: Discord Knowledge Steward Rich UX and Retrieval Enhancements Tasks
Ticket: DISCORD-BOT-011
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
  Work plan for the next knowledge steward bot slice: rich review UI, reaction-based promotion, source citation UX, rich search UI, and export from ask results.
LastUpdated: 2026-04-20T22:56:00-04:00
WhatFor: Track implementation tasks for the post-MVP UX and retrieval work.
WhenToUse: Use when building the next interaction and retrieval layer for the bot.
---

# Tasks

## Ticket setup

- [x] Create the ticket workspace
- [ ] Add a design guide for the rich UX and retrieval slice
- [ ] Add an implementation guide / API sketch doc
- [ ] Add a short diary once implementation begins

## Review UI

- [ ] Add a rich review queue card with action buttons
- [ ] Support Verify / Edit / Stale / Reject / Merge / Open Source actions from the queue
- [ ] Add a modal for editing existing entries, tags, and source attribution
- [ ] Add queue pagination and status filters

## Reaction-based promotion

- [ ] Define trusted reaction signals for promoting candidate knowledge
- [ ] Promote or queue candidates when trusted users react with the chosen emoji(s)
- [ ] Make reaction promotion visible in the channel or review queue
- [ ] Add runtime tests for the reaction-based promotion path

## Source citation UX

- [ ] Render source citations in search and article cards
- [ ] Add jump-link / open-source behavior for source messages or threads
- [ ] Add related-entry hints and canonical-source highlighting
- [ ] Make source metadata easy to read in the review UI

## Rich search UI

- [ ] Replace plain search results with richer cards and status badges
- [ ] Add pagination / next-page support for larger result sets
- [ ] Add autocomplete for tags, aliases, and article names
- [ ] Improve result ranking so verified entries appear first

## Export from ask results

- [ ] Add an export-to-channel action from ask result cards
- [ ] Support posting a selected answer or knowledge card into the current channel
- [ ] Preserve citations when exporting the result
- [ ] Add a private/ephemeral preview before the export is posted publicly

## Validation and rollout

- [ ] Update the example README and help text for the new interactions
- [ ] Add runtime and CLI tests for the new flows
- [ ] Validate with `go test ./...`
- [ ] Validate named-bot help / run commands for the updated example

## Notes

- Build on `DISCORD-BOT-010`; this ticket is the richer follow-up, not a rewrite.
- Keep the implementation modular in JavaScript unless a Go host change is strictly required.
- Favor visible and source-backed interactions over hidden automation.
