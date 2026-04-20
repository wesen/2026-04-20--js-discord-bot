---
Title: Discord Autocomplete and Richer Command Option Metadata API Reference and Planning Notes
Ticket: DISCORD-BOT-007
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
    - Path: internal/jsdiscord/host.go
      Note: Command option and autocomplete response normalization live here
    - Path: internal/jsdiscord/bot.go
      Note: Autocomplete registration and runtime context shape belong here
ExternalSources: []
Summary: Quick reference for autocomplete registrations, focused option context, and richer command option metadata.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Provide copy/paste-ready API sketches for autocomplete and richer option specs.
WhenToUse: Use when implementing or reviewing autocomplete and richer option support.
---

# Goal

Provide a quick API reference for command option metadata and autocomplete handlers.

## Quick API

```js
command("kb-search", {
  description: "Search docs",
  options: {
    query: {
      type: "string",
      required: true,
      autocomplete: true,
      minLength: 2,
      maxLength: 100,
    },
    scope: {
      type: "string",
      choices: [
        { name: "All docs", value: "all" },
        { name: "Runbooks", value: "runbooks" },
      ]
    }
  }
}, async (ctx) => ({ content: ctx.args.query }))

autocomplete("kb-search", "query", async (ctx) => {
  return [
    { name: "Architecture", value: "architecture" },
    { name: String(ctx.focused.value || ""), value: String(ctx.focused.value || "") },
  ]
})
```

## Expected Context Fields

- `ctx.args`
- `ctx.focused.name`
- `ctx.focused.value`
- `ctx.command.name`

## File-Level Checklist

- `internal/jsdiscord/host.go`
- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/descriptor.go`
- `internal/jsdiscord/runtime_test.go`
