---
Title: DSL Use Cases — Before/After
Ticket: UI-DSL-DISCORD
Status: active
Topics:
    - discord-bots
    - ui-dsl
    - framework-design
    - js-goja
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/discord-bots/knowledge-base/lib/render.js
      Note: Rendering functions - before/after examples
    - Path: examples/discord-bots/knowledge-base/lib/review.js
      Note: Review queue - before/after examples
    - Path: examples/discord-bots/knowledge-base/lib/search.js
      Note: Search UI - before/after examples
ExternalSources: []
Summary: Concrete before/after comparisons for every major UI pattern in the knowledge-base bot
LastUpdated: 2026-04-21T08:04:14.256013672-04:00
WhatFor: ""
WhenToUse: ""
---


# DSL Use Cases — Before/After

## Goal

Show concrete before/after comparisons for every major UI pattern in the knowledge-base bot. Each use case is derived from actual code in `examples/discord-bots/knowledge-base/`.

---

## Use Case 1: Knowledge Announcement (Simple Embed)

**Source:** `lib/render.js` — `knowledgeAnnouncement()`

### Before
```javascript
function knowledgeAnnouncement(entry, verb) {
  const action = verb || "Saved"
  return {
    content: `${action} knowledge entry **${entry.title}** (${entry.status}).`,
    embeds: [knowledgeEmbed(entry)],
  }
}

function knowledgeEmbed(entry) {
  return {
    title: entry.title || "Untitled knowledge",
    description: entry.summary || entry.body || "",
    color: statusColor(entry.status),
    fields: [
      { name: "Status", value: String(entry.status || "draft"), inline: true },
      { name: "Confidence", value: confidenceLabel(entry.confidence), inline: true },
      { name: "Canonical source", value: canonicalSourceLabel(entry), inline: false },
      { name: "Tags", value: formatTags(entry.tags), inline: false },
      { name: "Aliases", value: formatAliases(entry.aliases), inline: false },
      { name: "Source citation", value: formatSourceCitation(entry), inline: false },
      { name: "Source details", value: formatSourceDetails(entry), inline: false },
    ],
  }
}
```

### After (DSL)
```javascript
const { view, embed } = require("discord-ui")

function knowledgeAnnouncement(entry, verb) {
  return view()
    .content(`${verb || "Saved"} knowledge entry **${entry.title}** (${entry.status}).`)
    .embed(e => e
      .title(entry.title, "Untitled knowledge")
      .description(entry.summary || entry.body)
      .color(entry.status)
      .field("Status", entry.status, true)
      .field("Confidence", confidenceLabel(entry.confidence), true)
      .field("Canonical source", canonicalSourceLabel(entry))
      .field("Tags", formatTags(entry.tags))
      .field("Aliases", formatAliases(entry.aliases))
      .field("Source citation", formatSourceCitation(entry))
      .field("Source details", formatSourceDetails(entry))
    )
}
```

**Improvement:** Eliminated nested object construction. Named color lookup handles `statusColor()` automatically.

---

## Use Case 2: Review Queue with Select + Buttons

**Source:** `lib/review.js` — `buildQueueMessage()` + `buildReviewComponents()`

### Before
```javascript
function buildQueueMessage(entries, state) {
  const items = Array.isArray(entries) ? entries : []
  const selectedId = String(state && state.selectedId || "").trim()
  const selectedEntry = items.find((entry) => entry && entry.id === selectedId) || items[0] || null
  const queueMeta = { status: state && state.status ? state.status : "draft", limit: state && state.limit ? state.limit : 5 }
  const components = buildReviewComponents(items, { ...queueMeta, selectedId: selectedEntry && selectedEntry.id })
  return {
    content: items.length > 0
      ? `Review queue for ${queueMeta.status} (${items.length} entr${items.length === 1 ? "y" : "ies"})`
      : `No ${queueMeta.status} entries found.`,
    embeds: [renderSelectedEntryCard(selectedEntry, queueMeta)],
    components,
    ephemeral: true,
  }
}

function buildReviewComponents(entries, state) {
  const items = Array.isArray(entries) ? entries.slice(0, 25) : []
  const selectedId = String(state && state.selectedId || "").trim()
  const selectOptions = items.map((entry) => ({
    label: truncateLabel(entry.title || entry.slug || entry.id || "Untitled"),
    value: entry.id,
    description: truncateLabel(`${entry.status} · ${confidenceLabel(entry.confidence)} · ${formatTagsShort(entry.tags)}`),
    default: entry.id === selectedId,
  }))

  const components = []
  if (selectOptions.length > 0) {
    components.push({
      type: "actionRow",
      components: [{ type: "select", customId: REVIEW_COMPONENTS.select, placeholder: "Choose a knowledge entry to review", options: selectOptions }],
    })
  }
  components.push({
    type: "actionRow",
    components: [
      { type: "button", style: "success", label: "Verify", customId: REVIEW_COMPONENTS.verify },
      { type: "button", style: "secondary", label: "Edit", customId: REVIEW_COMPONENTS.edit },
      { type: "button", style: "secondary", label: "Source", customId: REVIEW_COMPONENTS.source },
      { type: "button", style: "primary", label: "Stale", customId: REVIEW_COMPONENTS.stale },
      { type: "button", style: "danger", label: "Reject", customId: REVIEW_COMPONENTS.reject },
    ],
  })
  return components
}
```

### After (DSL)
```javascript
const { view, embed, row } = require("discord-ui")

function buildQueueMessage(entries, state) {
  const items = Array.isArray(entries) ? entries : []
  const selectedId = String(state?.selectedId || "").trim()
  const selectedEntry = items.find(e => e?.id === selectedId) || items[0] || null
  const status = state?.status || "draft"
  const limit = state?.limit || 5

  return view()
    .content(items.length > 0
      ? `Review queue for ${status} (${items.length} entr${items.length === 1 ? "y" : "ies"})`
      : `No ${status} entries found.`)
    .embed(e => renderSelectedEntryCard(selectedEntry, { status, limit }, e))
    .row(r => r.select("Choose a knowledge entry to review", "select", sel =>
      sel.options(items.map(entry => ({
        label: entry.title || entry.slug || entry.id || "Untitled",
        value: entry.id,
        description: `${entry.status} · ${confidenceLabel(entry.confidence)} · ${formatTagsShort(entry.tags)}`,
        default: entry.id === selectedId,
      })))
    ))
    .row(r => r
      .button("Verify", "verify", { style: "success" })
      .button("Edit", "edit")
      .button("Source", "source")
      .button("Stale", "stale", { style: "primary" })
      .button("Reject", "reject", { style: "danger" })
    )
    .ephemeral()
}
```

**Improvement:** No more `REVIEW_COMPONENTS` constant object. Component IDs are auto-scoped. No more manual `actionRow` wrapping. Select options are mapped declaratively.

---

## Use Case 3: Search Results with Pagination

**Source:** `lib/search.js` — `buildSearchMessage()` + `buildSearchComponents()`

### Before
```javascript
function buildSearchMessage(view) {
  const state = normalizeSearchState(view && view.state)
  const entries = Array.isArray(view && view.pageEntries) ? view.pageEntries : []
  const selectedEntry = view && view.selectedEntry ? view.selectedEntry : null
  const query = String(view && view.query || state.query || "").trim()
  if (!selectedEntry) {
    return { content: `No knowledge entries matched ${query}.`, ephemeral: true }
  }
  return {
    content: `Found ${view.allResults ? view.allResults.length : entries.length} knowledge entr${(view.allResults ? view.allResults.length : entries.length) === 1 ? "y" : "ies"} for ${query}.`,
    embeds: [renderSearchResultCard(selectedEntry, { query, total: view.allResults ? view.allResults.length : entries.length, position: view.selectedIndex, page: view.page, pageCount: view.pageCount, relatedEntries: view.relatedEntries })],
    components: buildSearchComponents(entries, state, { hasPrevious: Boolean(view.hasPrevious), hasNext: Boolean(view.hasNext) }),
    ephemeral: true,
  }
}

function buildSearchComponents(entries, state, controls) {
  const items = Array.isArray(entries) ? entries.slice(0, 25) : []
  const selectedId = String(state && state.selectedId || "").trim()
  const selectOptions = items.map((entry) => ({
    label: truncateLabel(entry.title || entry.slug || entry.id || "Untitled"),
    value: entry.id,
    description: truncateLabel(`${entry.status} · ${confidenceLabel(entry.confidence)} · ${formatTagsShort(entry.tags)}`),
    default: entry.id === selectedId,
  }))

  const components = []
  if (selectOptions.length > 0) {
    components.push({ type: "actionRow", components: [{ type: "select", customId: SEARCH_COMPONENTS.select, placeholder: "Choose a knowledge entry to inspect", options: selectOptions }] })
  }
  components.push({ type: "actionRow", components: [
    { type: "button", style: "secondary", label: "Previous", customId: SEARCH_COMPONENTS.previous },
    { type: "button", style: "secondary", label: "Next", customId: SEARCH_COMPONENTS.next },
  ]})
  components.push({ type: "actionRow", components: [
    { type: "button", style: "primary", label: "Open", customId: SEARCH_COMPONENTS.open },
    { type: "button", style: "secondary", label: "Source", customId: SEARCH_COMPONENTS.source },
    { type: "button", style: "success", label: "Export", customId: SEARCH_COMPONENTS.export },
  ]})
  return components
}
```

### After (DSL)
```javascript
const { view, row } = require("discord-ui")

function buildSearchMessage(view) {
  const state = normalizeSearchState(view?.state)
  const entries = Array.isArray(view?.pageEntries) ? view.pageEntries : []
  const selectedEntry = view?.selectedEntry || null
  const query = String(view?.query || state.query || "").trim()

  if (!selectedEntry) {
    return view().content(`No knowledge entries matched ${query}.`).ephemeral()
  }

  const total = view.allResults?.length || entries.length

  return view()
    .content(`Found ${total} knowledge entr${total === 1 ? "y" : "ies"} for ${query}.`)
    .embed(e => renderSearchResultCard(selectedEntry, {
      query, total, position: view.selectedIndex,
      page: view.page, pageCount: view.pageCount, relatedEntries: view.relatedEntries,
    }, e))
    .row(r => r.select("Choose a knowledge entry to inspect", "select", sel =>
      sel.options(entries.slice(0, 25).map(entry => ({
        label: entry.title || entry.slug || entry.id || "Untitled",
        value: entry.id,
        description: `${entry.status} · ${confidenceLabel(entry.confidence)} · ${formatTagsShort(entry.tags)}`,
        default: entry.id === state.selectedId,
      })))
    ))
    .row(r => r
      .button("Previous", "previous", { disabled: !view.hasPrevious })
      .button("Next", "next", { disabled: !view.hasNext })
    )
    .row(r => r
      .button("Open", "open", { style: "primary" })
      .button("Source", "source")
      .button("Export", "export", { style: "success" })
    )
    .ephemeral()
}
```

**Improvement:** No `SEARCH_COMPONENTS` constant. Pagination buttons get `disabled` state declaratively. The whole function is ~60% shorter.

---

## Use Case 4: Modal Form (Teach Modal)

**Source:** `index.js` — `buildTeachModal()`

### Before
```javascript
function buildTeachModal() {
  return {
    customId: "knowledge:submit",
    title: "Teach the knowledge bot",
    components: [
      { type: "actionRow", components: [{ type: "textInput", customId: "title", label: "Title", style: "short", required: true, minLength: 3, maxLength: 100 }] },
      { type: "actionRow", components: [{ type: "textInput", customId: "summary", label: "Summary", style: "paragraph", required: true, minLength: 10, maxLength: 300 }] },
      { type: "actionRow", components: [{ type: "textInput", customId: "body", label: "Body", style: "paragraph", required: true, minLength: 20, maxLength: 2000 }] },
      { type: "actionRow", components: [{ type: "textInput", customId: "tags", label: "Tags (comma-separated)", style: "short", required: false, maxLength: 200 }] },
      { type: "actionRow", components: [{ type: "textInput", customId: "source", label: "Source URL or note", style: "short", required: false, maxLength: 300 }] },
    ],
  }
}
```

### After (DSL)
```javascript
const { modal } = require("discord-ui")

function buildTeachModal() {
  return modal("knowledge:submit", "Teach the knowledge bot")
    .short("title", "Title", { required: true, minLength: 3, maxLength: 100 })
    .paragraph("summary", "Summary", { required: true, minLength: 10, maxLength: 300 })
    .paragraph("body", "Body", { required: true, minLength: 20, maxLength: 2000 })
    .short("tags", "Tags (comma-separated)", { maxLength: 200 })
    .short("source", "Source URL or note", { maxLength: 300 })
}
```

**Improvement:** 28 lines → 7 lines. No nesting. No `actionRow`/`textInput` repetition.

---

## Use Case 5: State Management (Search State)

**Source:** `lib/search.js` — state load/save/normalize (~80 lines)

### Before
```javascript
const DEFAULT_SEARCH_LIMIT = 5
const MAX_SEARCH_RESULTS = 50

function searchStateKey(ctx) {
  const guildId = String((ctx.guild && ctx.guild.id) || "dm").trim() || "dm"
  const channelId = String((ctx.channel && ctx.channel.id) || "unknown-channel").trim() || "unknown-channel"
  const userId = String((ctx.user && ctx.user.id) || (ctx.me && ctx.me.id) || "unknown-user").trim() || "unknown-user"
  return `knowledge.search.state.${guildId}.${channelId}.${userId}`
}

function loadSearchState(ctx) {
  const state = ctx.store.get(searchStateKey(ctx), null)
  if (!state || typeof state !== "object") {
    return { query: "", limit: DEFAULT_SEARCH_LIMIT, page: 1, selectedId: "" }
  }
  return normalizeSearchState(state)
}

function saveSearchState(ctx, state) {
  const normalized = normalizeSearchState(state)
  ctx.store.set(searchStateKey(ctx), normalized)
  return normalized
}

function normalizeSearchState(state) {
  const source = state || {}
  return {
    query: String(source.query || "").trim(),
    limit: clampLimit(source.limit, DEFAULT_SEARCH_LIMIT),
    page: clampPage(source.page, 1),
    selectedId: String(source.selectedId || "").trim(),
  }
}

function clampLimit(limit, fallback) { /* ... */ }
function clampPage(page, fallback) { /* ... */ }
```

### After (DSL)
```javascript
const { state } = require("discord-ui")

// In command handler:
const s = state(ctx, "search", {
  defaults: { query: "", limit: 5, page: 1, selectedId: "" },
  limits: { limit: [1, 25], page: [1, 10] }
})
s.set({ query, limit, page: 1, selectedId: entries[0]?.id })

// In component handler:
const s = state(ctx, "search")
const { page } = s.get()
s.set({ page: page + 1 })
```

**Improvement:** ~80 lines → ~5 lines. Auto-scoped keys. Default values. Clamp validation.

---

## Use Case 6: Command Aliases

**Source:** `index.js` — `ask`/`kb-search`, `article`/`kb-article`, etc.

### Before
```javascript
command("ask", {
  description: "Search the shared knowledge base",
  options: { query: { type: "string", description: "Search query", required: true, autocomplete: true } },
}, async (ctx) => { /* ... */ })

command("kb-search", {
  description: "Search the shared knowledge base",
  options: { query: { type: "string", description: "Search query", required: true, autocomplete: true } },
}, async (ctx) => { /* ... */ })
```

### After (DSL)
```javascript
const { aliases } = require("discord-ui")

aliases(command, ["ask", "kb-search"], {
  description: "Search the shared knowledge base",
  options: { query: { type: "string", description: "Search query", required: true, autocomplete: true } },
}, async (ctx) => { /* ... */ })
```

**Improvement:** DRY. One registration block per conceptual command.

---

## Use Case 7: Error Response

**Source:** `index.js` — scattered ephemeral error returns

### Before
```javascript
if (!entry) {
  return { content: `No knowledge entry found for ${(ctx.args || {}).name}.`, ephemeral: true }
}
```

### After (DSL)
```javascript
const { error } = require("discord-ui")

if (!entry) {
  return error(`No knowledge entry found for ${ctx.args?.name}.`)
}
```

**Improvement:** Consistent ephemeral error formatting.

---

## Summary: Lines of Code Comparison

| Pattern | Before (approx lines) | After (approx lines) | Reduction |
|---------|----------------------|----------------------|-----------|
| Knowledge embed | 25 | 12 | 52% |
| Review queue + components | 45 | 28 | 38% |
| Search message + components | 50 | 30 | 40% |
| Teach modal | 28 | 7 | 75% |
| Search state management | 80 | 5 | 94% |
| Command aliases (×4 pairs) | 32 | 8 | 75% |
| **Total** | **260** | **90** | **65%** |

These are conservative estimates. The DSL also eliminates cognitive overhead: no more remembering Discord's nested payload structure.
