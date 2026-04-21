---
Title: Knowledge Base UI DSL Brainstorm and Design Options
Ticket: DISCORD-BOT-019
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
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js
      Note: Main bot wiring currently mixes domain logic with raw Discord response composition and component routing
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/search.js
      Note: Search view state and raw component/embed assembly show the current UI composition style
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/review.js
      Note: Review queue UI flow shows repeated selection/state/render patterns
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/render.js
      Note: Shared embed helpers exist, but higher-level page/screen composition is still manual
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/support/index.js
      Note: Simpler bot useful as a contrast point for deciding how much DSL power is actually needed
ExternalSources: []
Summary: Analyze the current knowledge-base bot UI composition style and propose layered DSL options with concrete example shapes.
LastUpdated: 2026-04-21T07:10:00-04:00
WhatFor: Capture a design brainstorm for making the knowledge-base bot UI code more elegant without losing runtime clarity.
WhenToUse: Use when deciding whether and how to introduce a UI DSL for Discord bot views, forms, actions, and stateful flows.
---

# Goal

Analyze the current `examples/discord-bots/knowledge-base/` bot and propose UI DSL directions that would make the code more elegant, less repetitive, and easier to grow.

# Current architecture map

The knowledge-base bot currently has four relevant layers:

1. **Top-level command and interaction registration** in `examples/discord-bots/knowledge-base/index.js`
2. **Feature-specific state/view helpers** in `lib/search.js` and `lib/review.js`
3. **Shared rendering helpers** in `lib/render.js`
4. **Domain/storage/capture logic** in `lib/store.js`, `lib/capture.js`, and `lib/reactions.js`

That is already better than one giant file, but the UI layer still feels too low-level for a bot of this size.

# What feels awkward today

## 1. Command alias duplication leaks into UI flow code

### Problem
The same interaction flow is often wired twice with near-identical bodies.

### Where to look
- `examples/discord-bots/knowledge-base/index.js:113`
- `examples/discord-bots/knowledge-base/index.js:132`
- `examples/discord-bots/knowledge-base/index.js:151`
- `examples/discord-bots/knowledge-base/index.js:173`
- `examples/discord-bots/knowledge-base/index.js:195`
- `examples/discord-bots/knowledge-base/index.js:218`
- `examples/discord-bots/knowledge-base/index.js:241`
- `examples/discord-bots/knowledge-base/index.js:255`

### Example
```js
command("ask", { ... }, async (ctx) => {
  store.ensure(ctx.config)
  const query = String((ctx.args || {}).query || "").trim()
  const limit = Number((ctx.config || {}).reviewLimit || 5)
  const results = store.search(ctx.config, query, limit)
  search.stateFromSearchCommand(ctx, query, limit, results)
  return search.buildSearchMessage(search.searchView(ctx, store))
})

command("kb-search", { ... }, async (ctx) => {
  store.ensure(ctx.config)
  const query = String((ctx.args || {}).query || "").trim()
  const limit = Number((ctx.config || {}).reviewLimit || 5)
  const results = store.search(ctx.config, query, limit)
  search.stateFromSearchCommand(ctx, query, limit, results)
  return search.buildSearchMessage(search.searchView(ctx, store))
})
```

### Why it matters
The UI code is noisier than it needs to be, and repeated flow registration hides the actual UX structure.

### Cleanup sketch
```js
ui.aliasCommand(["ask", "kb-search"], searchFlow.command({
  option: "query",
  autocomplete: true,
}))

ui.aliasCommand(["article", "kb-article"], articleFlow.command({
  option: "name",
  autocomplete: true,
}))
```

## 2. View rendering is still raw Discord object assembly

### Problem
The code is full of raw `content`, `embeds`, and `components` objects. The render helpers reduce some duplication, but the composition still feels like hand-building JSON instead of expressing screens.

### Where to look
- `examples/discord-bots/knowledge-base/lib/search.js:159`
- `examples/discord-bots/knowledge-base/lib/search.js:188`
- `examples/discord-bots/knowledge-base/lib/review.js:327`
- `examples/discord-bots/knowledge-base/lib/review.js:361`
- `examples/discord-bots/knowledge-base/index.js:499`

### Example
```js
return {
  content: `Found ${view.allResults ? view.allResults.length : entries.length} knowledge entr...`,
  embeds: [renderSearchResultCard(selectedEntry, { ... })],
  components: buildSearchComponents(entries, state, {
    hasPrevious: Boolean(view.hasPrevious),
    hasNext: Boolean(view.hasNext),
  }),
  ephemeral: true,
}
```

### Why it matters
The important thing is “render the search results screen”, not “build three action rows and one embed array by hand”.

### Cleanup sketch
```js
return kbui.searchResults(view)
```

with the DSL expanding internally to a Discord payload.

## 3. Screen state and interaction routing are conceptually linked, but encoded separately

### Problem
The code stores view state in one place, renders components in another, and routes components in a third. The pieces are reasonable, but they are not obviously one screen definition.

### Where to look
- `examples/discord-bots/knowledge-base/lib/search.js:1-157`
- `examples/discord-bots/knowledge-base/lib/search.js:159-238`
- `examples/discord-bots/knowledge-base/index.js:412-474`
- `examples/discord-bots/knowledge-base/lib/review.js:1-129`
- `examples/discord-bots/knowledge-base/lib/review.js:327-380`
- `examples/discord-bots/knowledge-base/index.js:338-410`

### Example
```js
search.stateFromSearchCommand(ctx, query, limit, results)
return search.buildSearchMessage(search.searchView(ctx, store))
```

and then later:

```js
component(search.SEARCH_COMPONENTS.next, async (ctx) => {
  store.ensure(ctx.config)
  search.shiftSearchPage(ctx, 1)
  const view = search.searchView(ctx, store)
  ...
})
```

### Why it matters
The mental model is already “this is one stateful screen with actions”. The current API makes the author reconstruct that model manually.

### Cleanup sketch
```js
const searchScreen = ui.screen("knowledge.search", {
  load(ctx) { ... },
  render(view) { ... },
  actions: {
    select(ctx, value) { ... },
    next(ctx) { ... },
    previous(ctx) { ... },
    open(ctx) { ... },
    source(ctx) { ... },
    export(ctx) { ... },
  },
})
```

## 4. Modal construction is structurally repetitive

### Problem
The teach modal and review edit modal are both built from repeated `actionRow` + `textInput` blocks.

### Where to look
- `examples/discord-bots/knowledge-base/index.js:499`
- `examples/discord-bots/knowledge-base/lib/review.js:130`

### Example
```js
{
  type: "actionRow",
  components: [
    {
      type: "textInput",
      customId: "title",
      label: "Title",
      style: "short",
      required: true,
      minLength: 3,
      maxLength: 100,
    },
  ],
}
```

### Why it matters
The UI intent is “a form with fields”, but the code says “five nested raw component rows”.

### Cleanup sketch
```js
return ui.form("knowledge:submit", "Teach the knowledge bot")
  .text("title", "Title").required().min(3).max(100)
  .textarea("summary", "Summary").required().min(10).max(300)
  .textarea("body", "Body").required().min(20).max(2000)
  .text("tags", "Tags (comma-separated)")
  .text("source", "Source URL or note")
```

## 5. Shared rendering exists, but it is one layer too low

### Problem
`render.js` already centralizes embed structure, but it still mostly returns raw Discord payload fragments. It is a useful helper layer, not yet a view DSL.

### Where to look
- `examples/discord-bots/knowledge-base/lib/render.js:55`
- `examples/discord-bots/knowledge-base/lib/render.js:72`
- `examples/discord-bots/knowledge-base/lib/render.js:99`
- `examples/discord-bots/knowledge-base/lib/render.js:114`
- `examples/discord-bots/knowledge-base/lib/render.js:164`

### Example
```js
function knowledgeAnnouncement(entry, verb) {
  const action = verb || "Saved"
  return {
    content: `${action} knowledge entry **${entry.title}** (${entry.status}).`,
    embeds: [knowledgeEmbed(entry)],
  }
}
```

### Why it matters
This is a good foundation, but the next abstraction should be “screen/panel/form/action set”, not just “object-returning helper function”.

# DSL design goals

Any proposed UI DSL should:

- keep the underlying Discord payload visible enough for debugging
- reduce repetition in forms, screens, and action rows
- unify stateful screens with their interaction handlers
- make aliases cheap
- stay usable by simpler bots, not just the knowledge-base bot
- avoid inventing a mini React clone inside the Discord bot runtime

# Design options

## Option A — Builder DSL for payloads

This is the smallest and safest option. It does not try to solve state or routing. It just makes payload construction less verbose.

### Shape
```js
ui.message()
  .content("Saved knowledge entry **X**")
  .ephemeral()
  .embed(
    ui.embed("Saved")
      .description("Knowledge stored")
      .field("Status", "draft", { inline: true })
  )
```

### Good for
- modals
- embeds
- action rows
- buttons/selects

### Pros
- low-risk
- easy to adopt incrementally
- works for every bot in the repo

### Cons
- does not solve stateful screen routing
- still leaves search/review logic fragmented across files

## Option B — Declarative screen DSL

This option models a screen as data plus a renderer.

### Shape
```js
kbui.screen("searchResults", {
  state: { query, page, selectedId },
  entry: view.selectedEntry,
  results: view.pageEntries,
  actions: ["previous", "next", "open", "source", "export"],
})
```

Internally that becomes the current `content + embeds + components` payload.

### Good for
- reusable panels like search/review/result-detail
- avoiding bespoke `buildSearchMessage` / `buildQueueMessage`

### Pros
- closer to the actual UX model
- easier to standardize button bars and selectors
- keeps render output declarative

### Cons
- still needs a separate action-registration story
- may become “JSON but different” if not designed carefully

## Option C — Stateful flow DSL

This is the most ambitious option. It treats search/review/teach as named flows with state, rendering, and actions together.

### Shape
```js
const searchFlow = ui.flow("knowledge.search", {
  init(ctx) { ... },
  load(ctx) { ... },
  render(view) { ... },
  actions: {
    select(ctx, payload) { ... },
    next(ctx) { ... },
    previous(ctx) { ... },
    open(ctx) { ... },
    source(ctx) { ... },
    export(ctx) { ... },
  },
})
```

### Good for
- the knowledge-base bot specifically
- future bots with review/triage workflows

### Pros
- matches the real architecture best
- state, render, and actions live together
- makes screen flows easier to reason about

### Cons
- larger implementation cost
- more framework-like
- likely overkill for simpler bots such as `support` or `announcements`

## Option D — Domain-specific DSL on top of small primitives

This is the option I would recommend.

Use a small general-purpose UI layer, then build a knowledge-specific DSL on top of it.

### Shape
```js
const ui = require("discord-ui")
const kbui = require("./lib/kb-ui")

return kbui.searchResults(view)
return kbui.reviewQueue(queue)
await ctx.showModal(kbui.teachForm(entry))
```

### Why this is better
- generic primitives stay reusable
- knowledge-base-specific ergonomics live near the bot
- the abstraction level can match the domain instead of trying to force every bot into one huge framework

# Recommended layered architecture

## Layer 1 — Generic Discord UI primitives

A small helper module should expose:

- `message()`
- `embed()`
- `actions()` / `row()`
- `button()`
- `select()`
- `form()`
- `field()` / `textarea()`

### Example
```js
ui.message()
  .ephemeral()
  .content("No review entry is currently selected.")
```

## Layer 2 — Flow helpers for stateful screens

A small flow helper should handle:

- keying state
- loading state
- saving state
- generating stable custom IDs from a flow namespace

### Example
```js
const reviewFlow = ui.flow("knowledge.review", {
  stateKey: (ctx) => `knowledge.review.state.${ctx.guild.id}.${ctx.channel.id}.${ctx.user.id}`,
})
```

## Layer 3 — Knowledge-specific screens/forms

This should live inside `examples/discord-bots/knowledge-base/lib/` and expose:

- `kbui.teachForm()`
- `kbui.searchResults(view)`
- `kbui.reviewQueue(view)`
- `kbui.sourceSheet(entry)`
- `kbui.entryCard(entry, meta)`

This is the layer that makes the bot source elegant.

# Proposed DSL examples by use case

## Use case 1 — Teach modal

### Current style
- built manually in `index.js`
- one raw `actionRow` per field

### Proposed DSL
```js
await ctx.showModal(
  kbui.teachForm()
    .title("Teach the knowledge bot")
    .titleField("title", { required: true, min: 3, max: 100 })
    .summaryField("summary", { required: true, min: 10, max: 300 })
    .bodyField("body", { required: true, min: 20, max: 2000 })
    .tagsField("tags")
    .sourceField("source")
)
```

## Use case 2 — Search results screen

### Proposed DSL
```js
return kbui.searchResults(view)
```

Expanded mentally into:
- result card embed
- result selector dropdown
- previous/next pager row
- open/source/export action row
- standard ephemeral wrapper

### Slightly more explicit version
```js
return ui.screen("knowledge.search")
  .ephemeral()
  .content(kbui.searchSummary(view))
  .card(kbui.entryCard(view.selectedEntry, kbui.searchMeta(view)))
  .select("entry", view.pageEntries, {
    selected: view.state.selectedId,
    placeholder: "Choose a knowledge entry to inspect",
  })
  .pager({ previous: view.hasPrevious, next: view.hasNext })
  .actions([
    ui.button("open", "Open", "primary"),
    ui.button("source", "Source", "secondary"),
    ui.button("export", "Export", "success"),
  ])
```

## Use case 3 — Review queue

### Proposed DSL
```js
return kbui.reviewQueue({
  status,
  limit,
  entries,
  selectedId,
})
```

### Expanded version
```js
return ui.screen("knowledge.review")
  .ephemeral()
  .content(`Review queue for ${status}`)
  .card(kbui.reviewCard(selectedEntry, { status, limit }))
  .select("entry", entries, {
    selected: selectedId,
    placeholder: "Choose an entry to review",
  })
  .actions([
    ui.button("verify", "Verify", "success"),
    ui.button("stale", "Mark stale", "secondary"),
    ui.button("reject", "Reject", "danger"),
    ui.button("edit", "Edit", "primary"),
    ui.button("source", "Source", "secondary"),
  ])
```

## Use case 4 — Source details sheet

### Proposed DSL
```js
return kbui.sourceSheet(entry)
```

### Benefit
This makes “show source” a named UI concept instead of another ad hoc `content + embed` object.

## Use case 5 — Inline confirmations for mutations

This bot does not currently have strong confirmation flows, but the DSL should make them easy.

### Proposed DSL
```js
return ui.confirmation({
  title: "Verify knowledge entry",
  body: `Promote **${entry.title}** to verified?`,
  confirmId: actions.verify(entry.id),
  cancelId: actions.cancel(entry.id),
})
```

## Use case 6 — Alias registration

### Proposed DSL
```js
kbui.registerAliases(command, {
  search: ["ask", "kb-search"],
  article: ["article", "kb-article"],
  review: ["review", "kb-review"],
  recent: ["recent", "kb-recent"],
})
```

# Recommended implementation path

## Phase 1 — non-invasive helper extraction

Do not change runtime behavior yet.

Add a bot-local helper module like:
```text
examples/discord-bots/knowledge-base/lib/ui/
  primitives.js
  forms.js
  screens.js
  kb-ui.js
```

Refactor only:
- teach modal
- search result screen
- review queue screen

## Phase 2 — alias and action registration helpers

Reduce top-level `index.js` noise by introducing helpers for:
- alias command groups
- component routing tables
- modal routing tables

## Phase 3 — optional generic runtime extraction

Only after the bot-local DSL proves useful, decide whether a reusable runtime module belongs in the broader repo.

# Recommendation

I would **not** start with a big generic “Discord UI framework”.

I would start with:

1. a **small generic builder layer** for payloads/forms/components
2. a **knowledge-base-local DSL** for search/review/teach/source screens

That gives the best balance of:
- elegance
- local clarity
- low implementation risk
- future reusability if the patterns stabilize
