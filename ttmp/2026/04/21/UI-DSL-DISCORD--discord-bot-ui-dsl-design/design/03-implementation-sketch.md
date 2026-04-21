---
Title: Implementation Sketch
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
RelatedFiles: []
ExternalSources: []
Summary: "How to build the discord-ui module as a pure-JavaScript library inside the go-go-goja runtime"
LastUpdated: 2026-04-21T08:04:14.256013672-04:00
---

# Implementation Sketch

## Goal

Show how to build the `discord-ui` module as a pure-JavaScript library that works inside the existing go-go-goja runtime, with zero changes to the Go host.

## Architecture

```
examples/discord-bots/knowledge-base/lib/ui/
  index.js          — public API exports
  view.js           — View builder
  embed.js          — Embed builder
  row.js            — Row builder (buttons, select, links)
  select.js         — Select menu builder
  modal.js          — Modal builder
  state.js          — Scoped state helper
  palette.js        — Named color definitions
  utils.js          — truncateLabel, clamp, etc.
```

Or, as a framework-level module:
```
js_modules/discord-ui/
  index.js
  ...
```

Registered via `require("discord-ui")`.

## Builder Pattern

Each builder is a factory function returning an object with chainable methods and a `.raw()` terminal.

```javascript
// view.js
function view() {
  const payload = { content: "", embeds: [], components: [] }
  return {
    content(text) { payload.content = text; return this },
    embed(fn) { payload.embeds.push(fn(embed()).raw()); return this },
    row(fn) { payload.components.push(fn(row()).raw()); return this },
    ephemeral() { payload.ephemeral = true; return this },
    raw() { return payload },
  }
}
```

**Why this pattern?**
- Works in Goja (no ES6 classes required)
- Simple to implement and debug
- IDE-friendly autocomplete (properties on returned object)
- Immutable-ish: each call builds a fresh payload

## Scoped Component IDs

The framework passes `ctx` with `name`, `interaction`, etc. The DSL extracts the current command/component name to auto-prefix IDs.

```javascript
// utils.js
function makeScopedId(ctx, action, scope = "auto") {
  if (scope === "global") return action
  const prefix = ctx.name || ctx.component?.customId || "bot"
  return `${prefix}:${action}`
}
```

In a command handler named `ask`:
```javascript
row().button("Open", "open")  // customId = "ask:open"
```

In a component handler for `"ask:select"`:
```javascript
row().button("Export", "export")  // customId = "ask:select:export"
```

## State Helper

Leverages the existing `ctx.store` API (get/set/delete/keys/namespace).

```javascript
// state.js
function state(ctx, namespace, opts = {}) {
  const key = buildStateKey(ctx, namespace)
  const defaults = opts.defaults || {}
  const limits = opts.limits || {}

  function get() {
    const stored = ctx.store.get(key, null)
    if (!stored || typeof stored !== "object") return { ...defaults }
    return normalize({ ...defaults, ...stored }, limits)
  }

  function set(partial) {
    const next = normalize({ ...get(), ...partial }, limits)
    ctx.store.set(key, next)
    return next
  }

  function clear() {
    ctx.store.delete(key)
  }

  return { get, set, clear }
}

function buildStateKey(ctx, namespace) {
  const g = String(ctx.guild?.id || "dm").trim() || "dm"
  const c = String(ctx.channel?.id || "unknown").trim() || "unknown"
  const u = String(ctx.user?.id || ctx.me?.id || "unknown").trim() || "unknown"
  return `ui.state.${namespace}.${g}.${c}.${u}`
}

function normalize(obj, limits) {
  const out = {}
  for (const [k, v] of Object.entries(obj)) {
    out[k] = clampValue(v, limits[k])
  }
  return out
}

function clampValue(v, limit) {
  if (!limit) return v
  const [min, max] = limit
  const n = Number(v)
  if (!Number.isFinite(n)) return v
  return Math.min(Math.max(n, min), max)
}
```

## Embed Builder

```javascript
// embed.js
const PALETTE = require("./palette")

function embed() {
  const e = {}
  return {
    title(text, fallback = "") { e.title = text || fallback; return this },
    description(text, fallback = "") { e.description = text || fallback; return this },
    color(nameOrHex) { e.color = PALETTE[nameOrHex] || nameOrHex || 0; return this },
    field(name, value, inline = false) {
      if (!e.fields) e.fields = []
      e.fields.push({ name, value: String(value ?? ""), inline })
      return this
    },
    fields(list) {
      if (!e.fields) e.fields = []
      for (const f of list) e.fields.push({ name: f.name, value: String(f.value ?? ""), inline: f.inline })
      return this
    },
    footer(text) { e.footer = { text }; return this },
    image(url) { e.image = { url }; return this },
    thumbnail(url) { e.thumbnail = { url }; return this },
    url(link) { e.url = link; return this },
    timestamp(date) { e.timestamp = date instanceof Date ? date.toISOString() : date; return this },
    author(name, icon, url) { e.author = { name, icon_url: icon, url }; return this },
    raw() { return e },
  }
}
```

## Row Builder

```javascript
// row.js
function row() {
  const components = []
  return {
    button(label, action, opts = {}) {
      const btn = {
        type: "button",
        style: opts.style || "secondary",
        label,
        customId: opts.scope ? action : makeScopedId(opts.ctx, action),
      }
      if (opts.emoji) btn.emoji = opts.emoji
      if (opts.disabled) btn.disabled = true
      if (opts.url) { btn.style = "link"; btn.url = opts.url; delete btn.customId }
      components.push(btn)
      return this
    },
    link(label, url) {
      components.push({ type: "button", style: "link", label, url })
      return this
    },
    select(placeholder, action, fn, opts = {}) {
      const sel = selectBuilder(placeholder, action, opts.ctx)
      fn(sel)
      components.push(sel.raw())
      return this
    },
    raw() { return { type: "actionRow", components } },
  }
}
```

## Select Builder

```javascript
// select.js
function selectBuilder(placeholder, action, ctx) {
  const payload = {
    type: "select",
    customId: makeScopedId(ctx, action),
    placeholder,
    options: [],
  }
  return {
    option(label, value, opts = {}) {
      payload.options.push({
        label: truncateLabel(label),
        value,
        description: opts.description ? truncateLabel(opts.description) : undefined,
        default: opts.default || false,
      })
      return this
    },
    options(list) {
      for (const o of list) {
        payload.options.push({
          label: truncateLabel(o.label),
          value: o.value,
          description: o.description ? truncateLabel(o.description) : undefined,
          default: o.default || false,
        })
      }
      return this
    },
    raw() { return payload },
  }
}
```

## Modal Builder

```javascript
// modal.js
function modal(customId, title) {
  const payload = { customId, title, components: [] }
  return {
    short(id, label, opts = {}) {
      payload.components.push(textInputRow(id, label, "short", opts))
      return this
    },
    paragraph(id, label, opts = {}) {
      payload.components.push(textInputRow(id, label, "paragraph", opts))
      return this
    },
    raw() { return payload },
  }
}

function textInputRow(id, label, style, opts = {}) {
  const input = { type: "textInput", customId: id, label, style, required: opts.required || false }
  if (opts.minLength) input.minLength = opts.minLength
  if (opts.maxLength) input.maxLength = opts.maxLength
  if (opts.value) input.value = opts.value
  if (opts.placeholder) input.placeholder = opts.placeholder
  return { type: "actionRow", components: [input] }
}
```

## Aliases Helper

```javascript
// index.js — aliases export
function aliases(registerFn, names, spec, handler) {
  for (const name of names) {
    registerFn(name, spec, handler)
  }
}
```

Usage inside `defineBot`:
```javascript
const { aliases } = require("discord-ui")

module.exports = defineBot(({ command, event, component, modal, configure }) => {
  aliases(command, ["ask", "kb-search"], {
    description: "Search the shared knowledge base",
    options: { query: { type: "string", description: "Search query", required: true, autocomplete: true } }
  }, async (ctx) => {
    // handler
  })
})
```

## Error Helper

```javascript
// index.js
function error(message) {
  return { content: message, ephemeral: true }
}
```

## Validation Layer (Optional)

Add a `validate` flag to builders that asserts Discord limits:

```javascript
function view(opts = {}) {
  const payload = { content: "", embeds: [], components: [] }
  return {
    row(fn) {
      if (opts.validate && payload.components.length >= 5) {
        throw new Error("Discord allows max 5 action rows")
      }
      payload.components.push(fn(row()).raw())
      return this
    },
    // ...
  }
}
```

This is off by default (production) and on during development/tests.

## Integration Path

### Option A: Bot-local module (immediate)
Place `lib/ui/` inside the knowledge-base bot. No framework changes.

### Option B: Framework module (recommended)
Add `js_modules/discord-ui/` to the framework's module path. Bots can `require("discord-ui")`.

### Option C: NPM-style package (future)
If the framework supports `package.json` + `node_modules`, publish as a package.

## Testing Strategy

```javascript
// test/ui/view.test.js
const { view, embed, row } = require("discord-ui")

function test() {
  const payload = view()
    .content("Hello")
    .embed(e => e.title("T").color("verified"))
    .row(r => r.button("OK", "ok", { style: "success" }))
    .raw()

  assert(payload.content === "Hello")
  assert(payload.embeds[0].title === "T")
  assert(payload.embeds[0].color === 0x57F287)
  assert(payload.components[0].components[0].customId === "ok")
}
```

## Migration Path for knowledge-base Bot

1. Create `lib/ui/` with all builder modules
2. Replace `lib/render.js` functions incrementally
3. Replace `lib/review.js` component building
4. Replace `lib/search.js` component building
5. Replace modals in `index.js`
6. Replace state management in `lib/review.js` and `lib/search.js`
7. Replace command aliases with `aliases()` helper
8. Delete `SEARCH_COMPONENTS` and `REVIEW_COMPONENTS` constants

Estimated effort: 2-3 hours for a full rewrite. The bot shrinks from ~650 lines to ~250 lines.

## Related Files

- `examples/discord-bots/knowledge-base/lib/render.js` — rendering to replace
- `examples/discord-bots/knowledge-base/lib/review.js` — review UI to replace
- `examples/discord-bots/knowledge-base/lib/search.js` — search UI to replace
- `examples/discord-bots/knowledge-base/index.js` — main bot to simplify
