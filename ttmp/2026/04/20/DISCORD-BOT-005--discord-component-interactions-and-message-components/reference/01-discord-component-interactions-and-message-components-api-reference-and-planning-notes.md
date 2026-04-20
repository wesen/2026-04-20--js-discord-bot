---
Title: Discord Component Interactions and Message Components API Reference and Planning Notes
Ticket: DISCORD-BOT-005
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
      Note: Component payload normalization and interaction response helpers
    - Path: internal/jsdiscord/bot.go
      Note: JS registration surface and dispatch context shape
ExternalSources: []
Summary: Quick reference for the proposed component API and the concrete implementation checklist.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Provide copy/paste-friendly examples for component handlers and payloads.
WhenToUse: Use when implementing or reviewing component support.
---

# Goal

Give maintainers a concrete reference for the proposed JavaScript component API and the implementation slices required on the Go side.

## Quick API Reference

### Register a button handler

```js
component("article:open", async (ctx) => {
  return { content: "opened", ephemeral: true }
})
```

### Register a select-menu handler

```js
component("support:queue", async (ctx) => {
  const selected = ctx.values || []
  return { content: `queue=${selected.join(",")}`, ephemeral: true }
})
```

### Outgoing select-menu payload

```js
{
  type: "actionRow",
  components: [
    {
      type: "select",
      customId: "support:queue",
      placeholder: "Choose a queue",
      minValues: 1,
      maxValues: 1,
      options: [
        { label: "Billing", value: "billing" },
        { label: "Infra", value: "infra", description: "Infrastructure support" }
      ]
    }
  ]
}
```

## Expected Context Fields

- `ctx.component.customId`
- `ctx.component.type`
- `ctx.values`
- `ctx.reply(payload)`
- `ctx.defer(payload)`
- `ctx.edit(payload)`
- `ctx.followUp(payload)`

## File-Level Checklist

- `internal/jsdiscord/bot.go` — registration + dispatch + context
- `internal/jsdiscord/host.go` — Discord interaction branch + payload normalization
- `internal/jsdiscord/descriptor.go` — descriptor parsing
- `internal/jsdiscord/runtime_test.go` — runtime-level coverage
- `examples/discord-bots/ping/index.js` — operator-facing demo
