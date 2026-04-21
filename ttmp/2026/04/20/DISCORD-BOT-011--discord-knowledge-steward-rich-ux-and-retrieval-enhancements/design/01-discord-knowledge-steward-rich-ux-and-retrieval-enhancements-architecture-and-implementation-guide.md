---
Title: Discord Knowledge Steward Rich UX and Retrieval Enhancements Architecture and Implementation Guide
Ticket: DISCORD-BOT-011
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/discord-bots/knowledge-base/index.js
      Note: Bot wiring for review, search, export, and reaction promotion
    - Path: examples/discord-bots/knowledge-base/lib/review.js
      Note: Review queue UI, selection state, and edit modal helpers
    - Path: examples/discord-bots/knowledge-base/lib/search.js
      Note: Search panel, source citation, and export helpers
    - Path: examples/discord-bots/knowledge-base/lib/render.js
      Note: Rich embed and source citation rendering helpers
    - Path: examples/discord-bots/knowledge-base/lib/reactions.js
      Note: Trusted reaction promotion helpers
    - Path: internal/jsdiscord/knowledge_base_runtime_test.go
      Note: Runtime coverage for review, search, source, and export flows
ExternalSources: []
Summary: |
    Architecture and implementation guide for the post-MVP knowledge steward bot slice that adds a rich review queue, reaction-based promotion, stronger source citation UX, richer search cards, and export-to-channel behavior.
LastUpdated: 2026-04-20T23:35:00-04:00
WhatFor: Explain how to evolve the knowledge steward bot beyond the MVP into a more interactive review and retrieval experience.
WhenToUse: Use when designing or implementing the next UX-heavy knowledge steward bot slice.
---



# Discord Knowledge Steward Rich UX and Retrieval Enhancements Architecture and Implementation Guide

## Executive Summary

The knowledge steward bot already has a usable MVP: it can listen to chat, capture candidate knowledge into SQLite, search entries, and change status through commands. The next step is to make that workflow much easier to use in Discord itself.

This ticket groups the features that matter most for the next iteration:

1. a rich review queue UI with buttons and modals,
2. reaction-based promotion for trusted knowledge signals,
3. stronger source citation presentation,
4. a richer search result experience, and
5. the ability to export an answer or knowledge card back into the channel from `/ask`.

The key goal is not just “more features.” It is to make knowledge stewardship feel like a native Discord workflow instead of a pile of commands. Reviewers should be able to scan, select, verify, edit, and reject entries quickly. Readers should see citations and status in a way that makes the bot trustworthy. And the bot should use low-friction signals, such as reactions, to help useful chat messages move from draft to canonical knowledge.

## Problem Statement and Scope

The MVP proved that the bot can persist and retrieve knowledge. It did not yet solve the day-to-day usability problems that make a knowledge steward pleasant to use:

- command-only review is slow,
- entry selection is awkward without a visual queue,
- source attribution is stored but not presented richly,
- search results are functional but plain, and
- answers cannot yet be exported cleanly back into the conversation.

This ticket is about those interaction layers.

### In scope

- Rich review queue responses with action buttons and modal editing.
- Trusted reaction promotion for candidate knowledge entries.
- Source citation presentation inside review/search/article cards.
- Improved search result cards and result selection UX.
- Exporting an ask result back into the current channel.
- Documentation, tests, and example updates for the new flows.

### Out of scope for this slice

- A new storage backend.
- Full-text search engines or embeddings.
- Merge/duplicate-resolution workflows beyond what is needed to support the review queue.
- Major Go host changes, unless a future feature makes them necessary.

## Current-State Analysis

### The MVP already provides the storage and status primitives

The bot currently uses the go-go-goja database module and a SQLite store written in JavaScript. That means the review and retrieval layer can build on durable entry records instead of inventing a parallel state system.

The store already supports:

- loading the database,
- inserting candidate entries,
- updating entries,
- changing status,
- fetching a single entry,
- listing entries by status,
- recent queries, and
- search.

That is enough to build richer interactions on top.

### The runtime already supports the interaction primitives we need

The JS runtime already exposes:

- `command(...)`
- `event(...)`
- `component(...)`
- `modal(...)`
- `autocomplete(...)`

The component and modal support is especially important here, because the review queue can stay inside Discord instead of forcing reviewers back into text commands.

The Discord event dispatcher also already sends enough context to make reaction-based promotion work:

- `ctx.message`
- `ctx.user`
- `ctx.guild`
- `ctx.channel`
- `ctx.member`
- `ctx.reaction`
- `ctx.config`
- `ctx.discord`

That means the bot can inspect the reacted message, find the corresponding knowledge entry, and update its status directly from the event handler.

### The current bot example is already the right place to extend

The `knowledge-base` example is the canonical home for the knowledge steward bot. It already owns the SQLite store, the capture logic, and the review/search commands. The new UX features should remain in that example and continue to use modular JS files for capture, review, reactions, rendering, and store access.

## Gap Analysis

### 1. Rich review UI is missing

The MVP review flow is still command-first. Reviewers can list entries and change status, but they do not yet get a queue card with a selector, action buttons, or an edit modal.

What is missing:

- a visual queue card,
- a way to select one entry from a list,
- buttons for verify / edit / source / stale / reject,
- a modal that edits the selected entry, and
- stateful selection handling across component clicks.

### 2. Reaction-based promotion is missing

The bot already listens to reactions in the runtime, but the example bot does not yet use them to promote candidate knowledge.

What is missing:

- a trusted emoji list,
- a trusted reviewer policy,
- reaction-to-entry lookup,
- a promotion rule that changes status, and
- a visible channel acknowledgement when the promotion happens.

### 3. Source citation UX is too plain

The data model carries source metadata, but the presentation is still mostly text-only. Reviewers and readers need to see source attribution in a structured way so they can judge the trustworthiness of the entry quickly.

### 4. Search results are not yet rich enough

Search works, but it does not yet present a polished answer card with status, citations, and an export path.

### 5. Exporting an answer back into the channel is missing

Users should be able to take a useful `/ask` result and post it back into the channel, preserving the citations and the canonical entry identity.

## Proposed Solution

## Rich Review Queue

The review queue should become a rich Discord message with:

- one selected entry card,
- a select menu of queued entries,
- action buttons,
- an edit modal, and
- ephemeral acknowledgements for each action.

### Selection model

The cleanest low-friction model is to keep a per-user, per-channel review state in the bot store:

- current status filter,
- queue limit,
- selected entry id.

When a reviewer opens `/review`, the bot stores the current view state and preselects the first entry. When the user changes the select menu, the selected entry id changes in that view state. Buttons then act on the currently selected entry.

That lets the queue stay stable without trying to invent dynamic button IDs for every entry.

### Review card contents

A selected entry card should include:

- entry id,
- title,
- summary/body,
- status,
- confidence,
- tags,
- aliases,
- source attribution,
- version number,
- queue status/limit footer.

The queue message itself should also include a selector and action buttons.

### Review actions

The buttons should support:

- `Verify` — mark the selected entry verified,
- `Edit` — open a modal to edit the selected entry,
- `Source` — show the current source details,
- `Stale` — mark the selected entry stale,
- `Reject` — reject the selected entry.

The modal should prefill the entry fields so small corrections are quick.

## Reaction-Based Promotion

Reaction promotion should be conservative and transparent.

### Policy

A reaction should promote a candidate only when:

1. the reaction emoji is one of the configured promotion emojis,
2. the reactor is trusted by config or role membership, and
3. the reaction points at a message that maps to a known candidate entry.

### Promotion lifecycle

A simple progression works well for the first version:

- `draft` -> `review`
- `review` -> `verified`
- `stale` -> `review`
- `rejected` -> `review`

This makes a trusted reaction useful both as a first-step endorsement and as a second-step confirmation.

### Visibility

If a promotion happens, the bot should reply in the channel with a short summary and a structured embed showing the updated entry.

## Source Citation UX

Source attribution should be visible in both review and search results.

Suggested presentation:

- a source field with guild, channel, message id, jump URL, and note,
- a source button that shows the original source details,
- a “canonical” or “verified” badge in search results,
- a clear footer or field for review state.

## Rich Search UI

Search results should become answer cards instead of plain text lists.

Suggested shape:

- status badge,
- confidence,
- source citation,
- related entries,
- pagination or next-page support,
- result ranking that prefers verified content.

The current search behavior can stay as the retrieval engine, but the rendering and interaction layer should be upgraded.

## Export to Channel from Ask Results

The `/ask` flow should be able to post a selected answer into the current channel.

The export action should:

1. keep the answer ephemeral until the user confirms,
2. post a channel-visible summary when the user exports it,
3. preserve citations, and
4. make it obvious that the message came from a knowledge entry or synthesized answer.

## Implementation Plan

### Phase 1 — rich review UI

- Add a review state helper in JavaScript.
- Render queue cards with a select menu and buttons.
- Add button handlers for verify, edit, source, stale, and reject.
- Add a modal for editing the selected entry.
- Add runtime tests for the review queue and modal flow.

### Phase 2 — reaction-based promotion

- Add config for trusted emojis and trusted users/roles.
- Add a reactionAdd event handler.
- Map the reacted message to an entry via source metadata.
- Promote the entry status according to the configured promotion rule.
- Add runtime tests for trusted and untrusted reaction events.

### Phase 3 — source citation UX and rich search

- Improve the embed/card render helpers for citations and status.
- Upgrade search results to richer cards.
- Add autocomplete for tags, aliases, and article names if needed.
- Add tests for the richer rendering shapes.

### Phase 4 — export to channel

- Add an export action from `/ask` results.
- Post the selected answer into the current channel.
- Preserve citations in the exported message.
- Add tests and README updates.

## Testing Strategy

The test plan should cover each user-visible interaction:

1. **Review queue test** — capture an entry, open the review queue, select the entry, and verify or edit it.
2. **Reaction promotion test** — capture an entry, react with a trusted emoji, and confirm that the status advances.
3. **Source rendering test** — confirm that the selected review card includes entry id and source details.
4. **Search render test** — confirm that the richer search card carries status and source information.
5. **Export test** — once implemented, confirm that an ask result can be posted to the channel with citations intact.

## Risks, Alternatives, and Open Questions

### Risks

- A review queue with too many controls can become noisy.
- Reaction promotion can be too permissive if trusted-user policy is not configured carefully.
- Rich search cards can become too dense if too many fields are shown at once.
- Exporting answers back into the channel can create duplicate content if it is used too aggressively.

### Alternatives considered

- **Dynamic buttons per entry** — rejected because the current JS component registration is static.
- **Command-only review** — simpler, but not nearly as usable.
- **Automatic promotion without trust checks** — easy to build, but not aligned with the stewardship model.

### Open questions

- Should the queue default to `draft` or `review` in the long term?
- Should trusted reactions promote to `review` only, or should repeated trusted reactions promote straight to `verified`?
- Should export-to-channel post a normal message, a thread reply, or a rich embed card?
- Which fields should be visible in the rich search card versus hidden behind a button or modal?

## References

### Repository files

- `examples/discord-bots/knowledge-base/index.js`
- `examples/discord-bots/knowledge-base/lib/review.js`
- `examples/discord-bots/knowledge-base/lib/reactions.js`
- `examples/discord-bots/knowledge-base/lib/store.js`
- `examples/discord-bots/knowledge-base/lib/render.js`
- `examples/discord-bots/knowledge-base/lib/capture.js`
- `internal/jsdiscord/knowledge_base_runtime_test.go`
- `examples/discord-bots/README.md`

### Related docs

- `ttmp/2026/04/20/DISCORD-BOT-011--discord-knowledge-steward-rich-ux-and-retrieval-enhancements/reference/01-discord-knowledge-steward-rich-ux-and-retrieval-enhancements-api-sketches.md`
- `ttmp/2026/04/20/DISCORD-BOT-011--discord-knowledge-steward-rich-ux-and-retrieval-enhancements/reference/02-diary.md`
