---
Title: "Using the Go-Side UI DSL for Discord Bots"
Slug: "go-side-ui-dsl-for-discord-bots"
Short: "A practical walkthrough for building Discord bot UIs with the Go-backed ui module, including messages, embeds, components, modals, and in-place updates."
Topics:
- discord
- javascript
- bots
- tutorial
- ui
- goja
- components
- modals
- database
- interaction-updates
Commands:
- bots help
- bots run
Flags:
- bot-repository
- bot-token
- application-id
- guild-id
- sync-on-start
- print-parsed-values
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

## What this tutorial is for

This tutorial explains how to build interactive Discord bot UIs with the Go-side `require("ui")` module. The important idea is simple: JavaScript still writes the fluent chains, but Go owns the builder state, the validation rules, and the final response shape.

That gives you the ergonomics of a DSL without the usual cost of a free-form JS object soup.

> [!summary]
> - Use `require("ui")` when you want fluent JavaScript builders with host-enforced validation.
> - Let Go own the payload shape, especially for embeds, components, forms, and response types.
> - Use in-place updates for component clicks so pagers and cards stay in one message thread.

## 1. Start with the mental model

The UI DSL is not a separate framework. It is a host-side builder layer that happens to be exposed to JavaScript.

In practice that means:

- JavaScript decides *what* the bot should render.
- Go decides *whether* the request is valid.
- Discord decides *whether* the final response shape is acceptable.

The middle step is the one that matters. When a JS bot returns a plain object, the host has to guess what the object means. When a bot returns a typed builder output from `require("ui")`, the host already knows whether it is dealing with an embed, a button row, a modal, or a whole message response.

## 2. Use the module from JavaScript

The showcase bot re-exports the Go module from its `lib/ui/index.js` entrypoint:

```js
module.exports = {
  ...require("ui"),
  ...require("./screen"),
}
```

That lets bot code stay small and readable:

```js
const ui = require("./lib/ui")
```

The builders then read naturally:

```js
return ui.message()
  .content("Hello")
  .embed(ui.embed("Greeting").description("Built with Go-side builders"))
  .row(ui.button("hello:ack", "OK", "primary"))
  .build()
```

## 3. Build a message first

The message builder is the outer shell for most interaction responses.

It is the place where you decide:

- content text
- embeds
- components
- ephemeral visibility
- whether the payload is an explicit follow-up or an in-place update

```js
return ui.message()
  .content("Search results for **sqlite**")
  .ephemeral()
  .embed(
    ui.embed("Results")
      .description("Five matching entries")
      .field("Status", "verified", true)
  )
  .build()
```

If a component handler should update the current message instead of creating a new one, the host now does that automatically for component interactions. If you explicitly want a new message, you can opt into a follow-up response.

## 4. Layer embeds and components separately

The `embed()` builder is for layout and structured display. The `button()` and `select()` builders are for interaction.

That split matters because the host can validate each shape independently.

```js
const card = ui.card(selected.title)
  .description(selected.summary)
  .meta("Status", selected.status, true)
  .meta("Category", selected.category, true)

return ui.message()
  .embed(card)
  .row(
    ui.select("demo:article-select")
      .placeholder("Choose an article")
      .optionEntries(pageEntries.map((entry) => ({
        id: entry.id,
        label: entry.title,
        description: entry.status,
      })), selected.id)
  )
  .build()
```

A useful rule of thumb:

- use `embed()` for text the user reads
- use `button()` for direct actions
- use `select()` when the user is choosing among several similar items

### Every component `customId` needs a handler

Rendering a button or select is only half the job. Every interactive `customId` you put into a UI payload must also have a matching registered handler.

```js
command("debug", async (ctx) => {
  return ui.message()
    .ephemeral()
    .content("Debug dashboard")
    .row(ui.button("show-space:debug:member", "Member", "primary"))
    .build()
})

component("show-space:debug:member", async (ctx) => {
  return renderDebugScreen(ctx, "member")
})
```

If you forget the `component("show-space:debug:member", ...)` registration, the message still renders, but the click fails because the bot has no handler for that `customId`.

This is the most common thing to miss when you convert a static response into a real interactive screen.

## 5. Use modal forms for structured input

Modal forms are the right tool when the user needs to type several values at once.

The Go-side DSL keeps the `customId` keys stable so the submit handler receives the expected values in `ctx.values`.

```js
await ctx.showModal(
  ui.form("feedback:submit", "Feedback Form")
    .text("title", "Title")
    .required()
    .textarea("feedback", "Your feedback")
    .required()
    .build()
)
```

In the modal submit handler:

```js
modal("feedback:submit", async (ctx) => {
  const title = String((ctx.values || {}).title || "").trim()
  const feedback = String((ctx.values || {}).feedback || "").trim()

  return ui.message()
    .ephemeral()
    .content("Thanks for your feedback!")
    .embed(ui.embed(title || "Feedback").description(feedback || "(no content)"))
    .build()
})
```

### Why the field API is customId-first

The modal field `customId` becomes the key in the submitted value map. That is why the field builder uses `text(customId, label)` and `textarea(customId, label)` rather than the other way around.

If you get this wrong, the modal opens fine but the submit handler sees default or empty values.

## 6. Let the host own in-place updates

A Discord component click is usually not a brand-new conversation. It is a mutation of the current screen.

For that reason the host now treats component interactions as update-in-place by default. That means:

- a search pager edits the existing message
- a product card view updates the same card thread
- a review screen stays anchored to one message

This is what the user expects when they click the next page or choose another result.

Use a new follow-up message only when the interaction really should branch into a fresh thread of output.

## 7. Separate transient screen state from durable data

The UI DSL is about interaction shape, not storage strategy.

Use `ctx.store` or a small flow helper when you only need per-runtime state for the current screen.

Use `require("database")` when the state must survive restarts or be shared across the bot’s long-term data model.

A good division looks like this:

- `ctx.store` — current pager position, selected item, active screen state
- `require("database")` — knowledge entries, review queues, persisted application data

In the showcase bot, the screen helpers keep track of pagination and selection state, while the knowledge-base bot uses `require("database")` for durable SQLite-backed records.

### The persistence rule

If the information is part of the UI session, it can live in `ctx.store`.

If the information is part of the bot’s memory, use `require("database")`.

That distinction keeps UI code lightweight without pretending transient state is durable.

## 8. Common mistakes

### Passing a raw object instead of a builder

This is the most common mistake when you first adopt the DSL:

```js
ui.message().embed({ title: "raw object" }) // wrong
```

Use the builder object instead:

```js
ui.message().embed(ui.embed("Title").description("...")).build()
```

### Nesting rows

`ui.pager()` already returns a row. Pass it to `ui.message().row(...)` and let the host flatten it.

### Treating component clicks as new messages by default

That makes interactive UIs noisy. Prefer in-place updates.

### Forgetting to register a component handler

A `ui.button("some:id", ...)` or `ui.select("some:id")` call does not automatically create the handler. You still need a matching `component("some:id", async (ctx) => { ... })` registration.

### Using `ctx.store` for durable state

`ctx.store` is not a database. It is just session-level scratch state.

## 9. Example: a browsable card gallery

```js
function renderProducts(ctx, products, selected) {
  return ui.message()
    .ephemeral()
    .content("Product catalog")
    .embed(
      ui.card(selected.name)
        .description(selected.description)
        .meta("Price", `$${selected.price.toFixed(2)}`, true)
        .meta("Stock", String(selected.stock), true)
    )
    .row(
      ui.select("catalog:select")
        .placeholder("Choose a product")
        .optionEntries(products.map((p) => ({
          id: p.id,
          label: p.name,
          description: `$${p.price.toFixed(2)}`,
        })), selected.id)
    )
    .row(ui.pager("catalog:prev", "catalog:next", { hasPrevious: true, hasNext: true }))
    .build()
}
```

This is the pattern to aim for:

- one message
- one embed describing the current screen
- one select for choosing a record
- one pager row for navigation
- one response type that updates in place

## 10. Where to look in the repository

The best examples live here:

- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ui-showcase/index.js`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ui-showcase/lib/ui/index.js`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/ui_module.go`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/ui_message.go`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/ui_components.go`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/ui_form.go`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_responses.go`

## 11. The main lesson

The point of a Go-side DSL is not to make JavaScript less capable. It is to give JavaScript a nicer surface while giving the host enough control to prevent malformed payloads, wrong-parent method calls, and noisy interaction behavior.

The bot author sees fluent code.
The host sees typed builders.
Discord sees valid payloads.

That is the division of labor worth preserving.
