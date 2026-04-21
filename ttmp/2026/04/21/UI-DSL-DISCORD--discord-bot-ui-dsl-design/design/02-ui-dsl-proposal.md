---
Title: UI DSL Proposal
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
    - Path: examples/discord-bots/knowledge-base/index.js
      Note: Main bot definition with raw payload patterns to replace
    - Path: examples/discord-bots/knowledge-base/lib/render.js
      Note: Rendering functions with hand-crafted embeds
    - Path: examples/discord-bots/knowledge-base/lib/review.js
      Note: Review queue state + component building
    - Path: examples/discord-bots/knowledge-base/lib/search.js
      Note: Search state + component building
    - Path: internal/jsdiscord/bot.go
      Note: Framework runtime that dispatches to JS handlers
ExternalSources: []
Summary: Core specification for a Discord UI DSL to replace raw payload construction
LastUpdated: 2026-04-21T08:04:14.256013672-04:00
WhatFor: ""
WhenToUse: ""
---


# UI DSL Proposal — Discord Bot UI Builder

## Goal

Design a JavaScript UI DSL that eliminates boilerplate when building Discord bot responses, components, modals, and stateful interactions. The DSL should feel native to the existing `defineBot` API, require no Go-side changes, and provide escape hatches to raw Discord payloads.

## Context

The `examples/discord-bots/knowledge-base/` bot demonstrates several recurring pain points:

1. **Raw payload construction** — Every response is a hand-crafted `{ content, embeds, components }` object.
2. **Scattered component IDs** — String constants for `customId` live in separate files from their handlers.
3. **Repetitive state management** — Search and review each implement ~100 lines of nearly identical load/save/normalize logic.
4. **Deep nesting in modals** — 5 action rows × 1 text input each = 40+ lines of structural noise.
5. **Ephemeral/reply branching** — Error paths and success paths repeat similar embed shapes.
6. **Command alias duplication** — Pairs like `ask`/`kb-search` have identical bodies.

## Design Principles

| Principle | Rationale |
|-----------|-----------|
| **Composable** | Small building blocks combine into complex UIs without deep nesting. |
| **Chainable** | Fluent API reads top-to-bottom, matches Discord's visual layout. |
| **Scoped** | Component IDs and state keys are auto-namespaced by command/context. |
| **Escape hatch** | Every builder exposes `.raw()` for custom Discord API features. |
| **Zero Go changes** | Implemented entirely as a JS module (`require("discord-ui")`). |
| **Type-friendly** | Even without TS, the chainable pattern gives IDE autocomplete hints. |

## Core Abstractions

### 1. `View` — A complete Discord interaction response

```javascript
const { view } = require("discord-ui")

view()
  .content("Hello, world")
  .embed(embed => embed.title("Status").field("Uptime", "2h 30m"))
  .row(row => row.button("Confirm", "confirm").button("Cancel", "cancel", { style: "danger" }))
  .ephemeral()
```

**Builder methods:**
- `.content(text)` — message content
- `.embed(fn)` — adds one embed via `Embed` builder
- `.embeds(...fns)` — adds multiple embeds
- `.row(fn)` — adds an action row via `Row` builder
- `.rows(...fns)` — adds multiple rows
- `.ephemeral()` — sets ephemeral flag
- `.raw()` — returns `{ content, embeds, components, ephemeral }`

### 2. `Embed` — A Discord embed

```javascript
const { embed } = require("discord-ui")

embed()
  .title("Knowledge Entry")
  .description(entry.summary)
  .color("verified")           // named color from palette
  .field("Status", entry.status, true)
  .field("Confidence", "80%", true)
  .field("Tags", entry.tags.join(", "))
  .footer(`Page ${page}/${total}`)
  .raw()
```

**Builder methods:**
- `.title(text)`, `.description(text)`, `.url(link)`, `.timestamp(date)`
- `.color(nameOrHex)` — supports named palette + raw hex
- `.field(name, value, inline?)` — adds a field
- `.fields([...])` — bulk add
- `.footer(text)`, `.image(url)`, `.thumbnail(url)`, `.author(name, icon?, url?)`
- `.raw()` — returns Discord embed object

**Named color palette:**
```javascript
const PALETTE = {
  primary:   0x5865F2,   // blurple
  success:   0x57F287,   // green
  danger:    0xED4245,   // red
  warning:   0xFEE75C,   // yellow
  muted:     0x95A5A6,   // gray
  info:      0x3498DB,   // blue
  verified:  0x57F287,
  stale:     0x95A5A6,
  rejected:  0xED4245,
  draft:     0xFEE75C,
}
```

### 3. `Row` — An action row of components

```javascript
const { row } = require("discord-ui")

row()
  .button("Verify", "verify", { style: "success" })
  .button("Edit", "edit")
  .button("Source", "source")
  .button("Stale", "stale", { style: "primary" })
  .button("Reject", "reject", { style: "danger" })
```

**Builder methods:**
- `.button(label, action, opts?)` — `opts: { style, emoji, disabled, url }`
- `.select(placeholder, action, optionsFn)` — dropdown
- `.link(label, url)` — link button (no customId)
- `.raw()` — returns `{ type: "actionRow", components: [...] }`

### 4. `Select` — A select menu (used inside a row)

```javascript
row().select("Choose entry", "select", (sel) =>
  sel.option("Entry A", "id-a", { description: "draft · 80%" })
     .option("Entry B", "id-b", { description: "verified · 95%", default: true })
     .options(entries.map(e => ({ label: e.title, value: e.id, description: e.status })))
)
```

### 5. `Modal` — A modal form

```javascript
const { modal } = require("discord-ui")

modal("knowledge:submit", "Teach the knowledge bot")
  .short("title", "Title", { required: true, minLength: 3, maxLength: 100 })
  .paragraph("summary", "Summary", { required: true, minLength: 10, maxLength: 300 })
  .paragraph("body", "Body", { required: true, minLength: 20, maxLength: 2000 })
  .short("tags", "Tags (comma-separated)", { maxLength: 200 })
  .short("source", "Source URL or note", { maxLength: 300 })
```

**Builder methods:**
- `.short(id, label, opts?)` — single-line text input
- `.paragraph(id, label, opts?)` — multi-line text input
- `.raw()` — returns full modal payload

### 6. `State` — Scoped key-value persistence

```javascript
const { state } = require("discord-ui")

// In a command handler:
const s = state(ctx, "search")  // auto-scoped by guild+channel+user
s.set({ query, limit, page: 1, selectedId: entries[0]?.id })

// In a component handler:
const s = state(ctx, "search")
const { query, page } = s.get()
s.set({ page: page + 1 })
```

**API:**
- `state(ctx, namespace)` — returns scoped store
- `.get(key?)` — get all or one key
- `.set(partial)` — merge-update
- `.clear()` — remove all

## Scoped Component IDs

Instead of manually namespacing `customId` strings, the DSL auto-prefixes based on context:

```javascript
// In handler for command "ask":
row().button("Open", "open")        // customId = "ask:open"
row().select("Pick", "select", ...) // customId = "ask:select"

// In handler for component "ask:select":
row().button("Export", "export")    // customId = "ask:select:export"
```

**Manual override:**
```javascript
row().button("Open", "open", { scope: "global" }) // customId = "open"
```

This eliminates the `SEARCH_COMPONENTS` / `REVIEW_COMPONENTS` constant objects entirely.

## Alias Commands

For duplicate command pairs, a helper:

```javascript
const { aliases } = require("discord-ui")

aliases(command, ["ask", "kb-search"], {
  description: "Search the shared knowledge base",
  options: { query: { type: "string", description: "Search query", required: true, autocomplete: true } }
}, async (ctx) => { /* handler */ })
```

Expands to two `command()` registrations with the same spec and handler.

## Error Response Helper

```javascript
const { error } = require("discord-ui")

return error("No knowledge entry found for 'foo'.")
// => { content: "No knowledge entry found for 'foo'.", ephemeral: true }
```

## Complete Example: Knowledge Announcement

### Before (current)
```javascript
return {
  content: `${verb} knowledge entry **${entry.title}** (${entry.status}).`,
  embeds: [{
    title: entry.title || "Untitled knowledge",
    description: entry.summary || entry.body || "",
    color: statusColor(entry.status),
    fields: [
      { name: "Status", value: String(entry.status || "draft"), inline: true },
      { name: "Confidence", value: confidenceLabel(entry.confidence), inline: true },
      { name: "Tags", value: formatTags(entry.tags), inline: false },
    ],
  }],
}
```

### After (DSL)
```javascript
const { view, embed } = require("discord-ui")

return view()
  .content(`${verb} knowledge entry **${entry.title}** (${entry.status}).`)
  .embed(e => e
    .title(entry.title, "Untitled knowledge")
    .description(entry.summary || entry.body)
    .color(entry.status)
    .field("Status", entry.status, true)
    .field("Confidence", confidenceLabel(entry.confidence), true)
    .field("Tags", formatTags(entry.tags))
  )
```

## Module Structure

```
discord-ui/
  index.js          — exports all builders
  view.js           — View builder
  embed.js          — Embed builder
  row.js            — Row + Select + Button builders
  modal.js          — Modal builder
  state.js          — State helper
  palette.js        — Named colors
  utils.js          — truncateLabel, formatters
```

## Open Questions

1. Should the DSL validate Discord limits (25 select options, 5 action rows, 4000 chars)?
2. Should state auto-sync with component interactions (e.g., `state(ctx, "search").page.inc()`)?
3. Should embeds support Markdown templates (`` .description`Status: ${status}` ``)?
4. How should the DSL handle multi-select menus vs single-select?

## Related Files

- `examples/discord-bots/knowledge-base/index.js` — current bot definition
- `examples/discord-bots/knowledge-base/lib/render.js` — current rendering logic
- `examples/discord-bots/knowledge-base/lib/review.js` — review state + components
- `examples/discord-bots/knowledge-base/lib/search.js` — search state + components
