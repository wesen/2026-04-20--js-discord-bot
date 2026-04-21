---
Title: Discord Knowledge Steward Rich UX and Retrieval Enhancements Diary
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
RelatedFiles:
    - Path: examples/discord-bots/README.md
      Note: Example repository notes for the updated knowledge-base bot UX
    - Path: examples/discord-bots/knowledge-base/index.js
      Note: Bot wiring and interaction changes recorded in the diary
    - Path: examples/discord-bots/knowledge-base/lib/reactions.js
      Note: Reaction promotion implementation recorded in the diary
    - Path: examples/discord-bots/knowledge-base/lib/render.js
      Note: Rich citation rendering implementation recorded in the diary
    - Path: examples/discord-bots/knowledge-base/lib/review.js
      Note: Rich review queue implementation recorded in the diary
    - Path: examples/discord-bots/knowledge-base/lib/search.js
      Note: Search panel, source details, and export implementation recorded in the diary
    - Path: internal/jsdiscord/knowledge_base_runtime_test.go
      Note: Runtime validation evidence referenced in the diary
ExternalSources: []
Summary: |
    Chronological diary for the rich review queue and reaction-based promotion work in the Discord knowledge steward bot.
LastUpdated: 2026-04-20T23:35:00-04:00
WhatFor: Capture the implementation trail, decisions, and validation steps for the post-MVP UX slice.
WhenToUse: Use when continuing the ticket or reviewing how the rich review and reaction flows were built.
---



# Diary

## Goal

This diary records the implementation work for the post-MVP knowledge steward bot slice. It documents how the rich review queue and reaction-based promotion were designed, how they were implemented in modular JavaScript, and how they were validated with runtime tests.

## Step 1: Rich review queue UI

I replaced the command-only review flow with a richer queue experience. The bot now renders a review card with a select menu of candidate entries, action buttons for verify/edit/source/stale/reject, and an edit modal that pre-fills the current entry. Review state is stored per reviewer in the bot's JS store so that selection and button actions can remain stable across interactions.

That changed the review workflow from "look up an id, then run a status command" to a more Discord-native flow where the reviewer can scan the queue, choose an entry, and act on it directly.

### Prompt Context

**User prompt (verbatim):** "Go ahead, task by task, do the rich review and reaction thing"

**Assistant interpretation:** Implement the next ticket slice in order, starting with the rich review UI and then the reaction-based promotion behavior.

**Inferred user intent:** Turn the planned review improvements into a usable Discord interaction flow, and keep the work organized into incremental commits and documentation updates.

**Commit (code):** 5e2db54 — "feat: add rich knowledge review ui"

### What I did

- Added `examples/discord-bots/knowledge-base/lib/review.js` to manage review-state storage, queue rendering, select-menu behavior, button actions, and the edit modal.
- Reworked `examples/discord-bots/knowledge-base/index.js` so `/review` and `/kb-review` return a queue card instead of a plain text list.
- Added select-menu and button component handlers for review selection, verify, edit, source, stale, and reject.
- Updated `internal/jsdiscord/knowledge_base_runtime_test.go` to exercise the review queue UI with a real captured entry.

### Why

The review flow needed to become easier to use inside Discord. The rich UI keeps the reviewer in the channel, reduces the need to copy entry IDs around manually, and makes the selected entry's source and status visible at a glance.

### What worked

- The JS runtime already supported static component handlers and modals, so the richer review workflow stayed in JavaScript.
- Per-user/per-channel review state in the bot store was enough to track selection without adding a new host API.
- The modal API already allowed prefilling values, which made entry editing practical.
- The runtime test could dispatch component interactions directly, so the queue behavior was straightforward to validate.

### What didn't work

- The first version of the review test assumed the queue output would still contain a plain `id:` line. After the richer card rendering, the test had to be updated to read the `Entry ID` embed field instead.
- No runtime or test failures survived the follow-up adjustments, but the test expectations had to be aligned with the richer card shape.

### What I learned

- Static component IDs work well when selection state is stored in the bot store.
- Discord review flows are much easier to understand when the selected entry includes its own source and identity information.
- A review queue can stay readable without trying to encode dynamic entry IDs into component IDs.

### What was tricky to build

The main subtlety was state management. Because the JS runtime uses static component registrations, the review UI could not rely on dynamic component IDs per entry. The fix was to store the reviewer's current selection in the JS store and let the select menu choose the active entry while the buttons operate on that selected entry.

### What warrants a second pair of eyes

- Whether the review queue should default to `draft` or `review` in the long term.
- Whether the source button should remain ephemeral or open a richer source detail view.
- Whether the review edit modal needs more fields, such as explicit status notes or canonical aliases.

### What should be done in the future

- Add richer search cards and export-to-channel behavior.
- Decide whether review actions should also refresh the queue message automatically.
- Add more coverage for modal editing and queue state persistence.

### Code review instructions

Start with:

- `examples/discord-bots/knowledge-base/lib/review.js`
- `examples/discord-bots/knowledge-base/index.js`
- `internal/jsdiscord/knowledge_base_runtime_test.go`

Validate with:

- `go test ./internal/jsdiscord -run TestKnowledgeBaseBotUsesSQLiteStoreForCaptureSearchAndReview -count=1`
- `go test ./...`

### Technical details

- Review state is stored via `ctx.store` using a per-user, per-channel, per-guild key.
- The review queue render includes an entry selector and action buttons.
- The edit modal uses the selected entry's current values as defaults.

## Step 2: Reaction-based promotion

I added trusted reaction promotion so the bot can treat specific emoji reactions as a community signal to move candidate knowledge forward. The reaction handler now looks up the captured message, checks configured trusted users or roles, and promotes the entry from `draft` to `review` or from `review` to `verified`.

This completed the second half of the ticket's interaction slice: the bot can now be guided not only by explicit review actions, but also by lightweight social signals from the channel.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** After the rich review UI, add reaction-based promotion so trusted reactions can move candidate knowledge forward.

**Inferred user intent:** Give the bot a low-friction community signal for surfacing useful messages while keeping the promotion policy configurable and visible.

**Commit (code):** a9230c8 — ":art: Work on reactions"

### What I did

- Added `examples/discord-bots/knowledge-base/lib/reactions.js` with the trusted reaction promotion logic.
- Extended `examples/discord-bots/knowledge-base/index.js` with reaction-promotion config fields and the `reactionAdd` event registration.
- Updated the runtime test to dispatch a trusted reaction and confirm that the entry moves into the review queue before being verified from the review UI.

### Why

Reaction-based promotion gives the bot a more IRC-like community memory model: people can reinforce useful information with a lightweight emoji signal instead of always opening a form or running a command.

### What worked

- The runtime already delivered the reaction event context the bot needed: the message, user, guild, channel, and reaction emoji.
- The trust policy was simple enough to express in JavaScript and configure through the bot's run fields.
- The promotion workflow stayed transparent because the bot replies in the channel when a reaction actually changes an entry's status.

### What didn't work

- The first test expectation assumed the reaction handler would always produce a reply record. The promotion logic itself was more important than the reply count, so the test was adjusted to validate the queue status instead of relying on reply bookkeeping.

### What I learned

- Source-linked promotion works well when the captured message id and channel id are preserved in the store.
- A small trust policy is enough for the first version; we do not need a full role engine to prove the workflow.
- The promotion rule is easiest to understand when it advances the entry one step at a time.

### What was tricky to build

The key sharp edge was deciding how to trust a reaction without making the bot too permissive. The current implementation allows configurable trusted users and roles, with a simple emoji allowlist, while keeping the promotion step visible and reversible through the review UI.

### What warrants a second pair of eyes

- Whether the default trusted policy should be empty, allow all non-bot users, or require an explicit allowlist.
- Whether repeated trusted reactions should ever jump straight to verified.
- Whether reaction promotion should also notify a dedicated review channel in addition to the current channel.

### What should be done in the future

- Add richer search and citation rendering.
- Decide whether reaction promotion should carry an audit note into the review queue UI.
- Add export-to-channel from ask results.

### Code review instructions

Start with:

- `examples/discord-bots/knowledge-base/lib/reactions.js`
- `examples/discord-bots/knowledge-base/index.js`
- `internal/jsdiscord/knowledge_base_runtime_test.go`

Validate with:

- `go test ./internal/jsdiscord -run TestKnowledgeBaseBotUsesSQLiteStoreForCaptureSearchAndReview -count=1`
- `go test ./...`

### Technical details

- Promotion emojis default to `🧠,📌` and can be overridden with bot config.
- Trusted reviewer IDs and role IDs are configurable through run fields.
- Promotion currently moves `draft -> review`, `review -> verified`, and `stale/rejected -> review`.

## Step 3: Source citations, rich search cards, and export-from-ask

I extended the knowledge steward bot beyond review and reaction promotion by adding richer source citation rendering, search-result panels, and export-from-ask behavior. The `/ask` and `/kb-search` commands now return a structured search card with a select menu and action buttons. The selected result shows its entry id, status, confidence, tags, aliases, source citation, and source details. Users can open the full entry, inspect the source, or export the selected answer back into the channel with citations intact.

That brought the bot much closer to the "shared knowledge steward" behavior described in the ticket: the bot is no longer just storing and reviewing knowledge, it is also surfacing the source, answering in a richer format, and helping users broadcast a curated answer back into the conversation.

### Prompt Context

**User prompt (verbatim):** "continue"

**Assistant interpretation:** Continue the ticket work by implementing the next outstanding UX slice: source citation UX, rich search cards, and the export-to-channel interaction.

**Inferred user intent:** Keep advancing the post-MVP Discord knowledge steward and preserve the same commit-and-diary workflow while finishing the remaining high-value user-facing interactions.

**Commit (code):** 86669e3 — "feat: add search export and source citation ux"

### What I did

- Added `examples/discord-bots/knowledge-base/lib/search.js` to manage search-state storage, result selection, source lookup, and export behavior.
- Expanded `examples/discord-bots/knowledge-base/lib/render.js` so knowledge cards now render source citations, source details, and richer search cards.
- Updated `examples/discord-bots/knowledge-base/index.js` so `/ask` and `/kb-search` return a richer search panel with select, open, source, and export actions.
- Updated `examples/discord-bots/README.md` so the example bot notes mention the richer search/export UX.
- Expanded `internal/jsdiscord/knowledge_base_runtime_test.go` to validate source citations, search selection, source lookup, open, and export follow-up behavior.

### Why

This slice fills the biggest remaining usability gap after review and reaction promotion. Without richer source citation and export behavior, the bot could store good knowledge but still feel like a passive archive. The new search panel turns it into an interactive steward that can surface a result, explain where it came from, and share it back into the channel when appropriate.

### What worked

- Reusing the existing SQLite-backed store kept the search/export work in JavaScript instead of requiring more Go host APIs.
- The search state could be modeled with the same per-user/per-channel storage approach as review state.
- Rendering source details in the embed made the search result much more trustworthy and easier to audit.
- The component handlers for open/source/export stayed small because the heavy lifting lived in `search.js` and `render.js`.

### What didn't work

- The first search test assumed the search-state key would be shared automatically between the command and component dispatches. In the runtime test harness, the command needed the same guild/channel/user context as the component interaction.
- The first search result card did not include an `Entry ID` field, which made it awkward to drive the follow-up component interactions. Adding the id to the result card fixed that.

### What I learned

- Rich search UX benefits from showing the source details directly in the card instead of hiding them behind a separate command.
- Export-to-channel is easiest to reason about when the component flow clearly separates the ephemeral preview from the public follow-up.
- The same general pattern works for both review and search panels: store per-context selection state, render a result card, and let buttons operate on the selected entry.

### What was tricky to build

The trickiest part was the interaction between state and context. The search panel only works correctly if the search command and the component interactions resolve to the same state key. In production Discord that happens naturally because the interactions all carry the same guild, channel, and user identifiers, but the runtime test had to be updated to pass those values explicitly.

### What warrants a second pair of eyes

- Whether the export button should post a normal message, a thread reply, or an embedded answer card by default.
- Whether the search result panel should paginate instead of capping at five results.
- Whether the source details should be collapsed into a single field or split into multiple smaller fields for readability.

### What should be done in the future

- Add pagination / next-page support for larger search result sets.
- Add autocomplete for tags, aliases, and article names.
- Decide whether the search panel should support related-entry hints or canonical-source highlighting.

### Code review instructions

Start with:

- `examples/discord-bots/knowledge-base/lib/search.js`
- `examples/discord-bots/knowledge-base/lib/render.js`
- `examples/discord-bots/knowledge-base/index.js`
- `internal/jsdiscord/knowledge_base_runtime_test.go`

Validate with:

- `go test ./internal/jsdiscord -run TestKnowledgeBaseBotUsesSQLiteStoreForCaptureSearchAndReview -count=1`
- `go test ./...`

### Technical details

- Search state is stored via `ctx.store` using a per-user, per-channel, per-guild key.
- The search panel uses a select menu plus Open / Source / Export buttons.
- Export posts a public follow-up with the rendered knowledge card and then edits the ephemeral interaction reply with a confirmation.
