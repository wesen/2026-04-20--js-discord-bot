---
Title: UI DSL Example Sketches for the Knowledge Base Bot
Ticket: DISCORD-BOT-019
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js
      Note: Current command/component/modal wiring provides the baseline to compare against the DSL sketches
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/search.js
      Note: Search result screen is the strongest candidate for a local screen DSL
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/review.js
      Note: Review queue flow shows the best candidate for stateful screen/action consolidation
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/render.js
      Note: Existing render helpers are the likely substrate for a future DSL layer
ExternalSources: []
Summary: Concrete example shapes for several UI DSL approaches applied to the knowledge-base bot.
LastUpdated: 2026-04-21T07:10:00-04:00
WhatFor: Make the UI DSL discussion concrete with side-by-side examples for the knowledge-base bot's main interaction patterns.
WhenToUse: Use when evaluating which UI DSL shape is the best fit before implementing anything.
---

# Goal

Show concrete DSL shapes for the knowledge-base bot's main UI use cases so the design discussion stays grounded.

# Use case matrix

The main UI surfaces in this bot are:

1. **teach form**
2. **search results screen**
3. **review queue screen**
4. **source details sheet**
5. **status mutation actions**
6. **alias-heavy command registration**

# DSL Family 1 — Builder-first

## Teach modal

```js
await ctx.showModal(
  ui.form("knowledge:submit", "Teach the knowledge bot")
    .text("title", "Title", { required: true, min: 3, max: 100 })
    .textarea("summary", "Summary", { required: true, min: 10, max: 300 })
    .textarea("body", "Body", { required: true, min: 20, max: 2000 })
    .text("tags", "Tags (comma-separated)")
    .text("source", "Source URL or note")
)
```

## Search results

```js
return ui.message()
  .ephemeral()
  .content(kbui.searchSummary(view))
  .embed(kbui.entryCard(view.selectedEntry, kbui.searchMeta(view)))
  .row(
    ui.select(searchIds.select)
      .placeholder("Choose a knowledge entry to inspect")
      .options(kbui.entryOptions(view.pageEntries, view.state.selectedId))
  )
  .row(
    ui.button(searchIds.previous, "Previous", "secondary"),
    ui.button(searchIds.next, "Next", "secondary")
  )
  .row(
    ui.button(searchIds.open, "Open", "primary"),
    ui.button(searchIds.source, "Source", "secondary"),
    ui.button(searchIds.export, "Export", "success")
  )
```

## Review queue

```js
return ui.message()
  .ephemeral()
  .content(`Review queue for ${state.status}`)
  .embed(kbui.reviewCard(selectedEntry, state))
  .row(
    ui.select(reviewIds.select)
      .placeholder("Choose an entry to review")
      .options(kbui.entryOptions(entries, state.selectedId))
  )
  .row(
    ui.button(reviewIds.verify, "Verify", "success"),
    ui.button(reviewIds.edit, "Edit", "primary"),
    ui.button(reviewIds.stale, "Mark stale", "secondary"),
    ui.button(reviewIds.reject, "Reject", "danger"),
    ui.button(reviewIds.source, "Source", "secondary")
  )
```

## When this family is best

Use this if the main pain is raw payload verbosity and modal/action-row boilerplate.

# DSL Family 2 — Declarative screens

## Search screen

```js
return kbui.screen("searchResults", {
  title: `Search: ${view.query}`,
  entry: view.selectedEntry,
  results: view.pageEntries,
  selectedId: view.state.selectedId,
  page: view.page,
  pageCount: view.pageCount,
  hasPrevious: view.hasPrevious,
  hasNext: view.hasNext,
  relatedEntries: view.relatedEntries,
})
```

## Review screen

```js
return kbui.screen("reviewQueue", {
  status: state.status,
  limit: state.limit,
  entries,
  selectedId: state.selectedId,
})
```

## Source sheet

```js
return kbui.screen("sourceDetails", {
  entry,
  heading: `Source for ${entry.title}`,
})
```

## When this family is best

Use this if the main pain is that each screen is manually assembled from the same kinds of parts.

# DSL Family 3 — Stateful flow definitions

## Search flow

```js
const searchFlow = kbui.flow("knowledge.search", {
  commandNames: ["ask", "kb-search"],
  autocomplete: [{ command: "ask", option: "query" }, { command: "kb-search", option: "query" }],

  init(ctx) {
    return { query: "", page: 1, selectedId: "", limit: ctx.config.reviewLimit || 5 }
  },

  command(ctx) {
    const query = String(ctx.args.query || "").trim()
    const limit = Number(ctx.config.reviewLimit || 5)
    const results = store.search(ctx.config, query, limit)
    return {
      state: { query, page: 1, selectedId: results[0] && results[0].id || "", limit },
      reply: kbui.searchResults(search.searchView(ctx, store)),
    }
  },

  actions: {
    select(ctx, selectedId) { ... },
    next(ctx) { ... },
    previous(ctx) { ... },
    open(ctx) { ... },
    source(ctx) { ... },
    export(ctx) { ... },
  },
})
```

## Review flow

```js
const reviewFlow = kbui.flow("knowledge.review", {
  commandNames: ["review", "kb-review"],
  actions: {
    select(ctx, entryId) { ... },
    verify(ctx) { ... },
    stale(ctx) { ... },
    reject(ctx) { ... },
    edit(ctx) { ... },
    source(ctx) { ... },
  },
})
```

## When this family is best

Use this if the main pain is not rendering itself, but the separation between state persistence, rendering, and action wiring.

# DSL Family 4 — Domain-specific sugar on top of primitives

This is the approach I would actually start with.

## `kbui.teachForm()`

```js
await ctx.showModal(kbui.teachForm())
```

## `kbui.searchResults(view)`

```js
return kbui.searchResults(view)
```

## `kbui.reviewQueue({ entries, state })`

```js
return kbui.reviewQueue({ entries, state })
```

## `kbui.sourceSheet(entry)`

```js
return kbui.sourceSheet(entry)
```

## `kbui.aliases(...)`

```js
kbui.aliases(command, {
  search: ["ask", "kb-search"],
  article: ["article", "kb-article"],
  review: ["review", "kb-review"],
  recent: ["recent", "kb-recent"],
})
```

# Example before/after sketches

## Before: search command body

```js
const query = String((ctx.args || {}).query || "").trim()
const limit = Number((ctx.config || {}).reviewLimit || 5)
const results = store.search(ctx.config, query, limit)
search.stateFromSearchCommand(ctx, query, limit, results)
return search.buildSearchMessage(search.searchView(ctx, store))
```

## After: command body with local DSL

```js
return kbui.runSearch(ctx, store)
```

or, if we want to keep the flow explicit:

```js
const flow = kbflows.search(ctx, store)
return kbui.searchResults(flow.view)
```

## Before: review select component

```js
const selectedId = firstValue(ctx.values)
if (!selectedId) {
  return { content: "Please choose a knowledge entry from the review dropdown.", ephemeral: true }
}
review.setReviewSelection(ctx, selectedId)
const current = review.currentReviewEntry(ctx, store)
if (!current) {
  return { content: "No review entry is currently available.", ephemeral: true }
}
return review.reviewReply(current, "Selected")
```

## After: review select with local DSL

```js
return kbflows.review.select(ctx, store)
```

# Recommended first implementation target

If we do build this, the best first targets are:

1. `teachForm()`
2. `searchResults(view)`
3. `reviewQueue(view)`

Those three would test whether the DSL actually improves clarity without forcing a full framework rewrite.

# Recommendation

Use:

- **small generic UI builders** for payload composition
- **knowledge-base-local DSL helpers** for screen names and domain semantics

Do **not** start by inventing a large generic cross-bot screen framework.
