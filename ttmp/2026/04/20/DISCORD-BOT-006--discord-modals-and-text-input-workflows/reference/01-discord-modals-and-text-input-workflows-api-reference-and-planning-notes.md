---
Title: Discord Modals and Text Input Workflows API Reference and Planning Notes
Ticket: DISCORD-BOT-006
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
      Note: Modal and text-input normalization should be added here
    - Path: internal/jsdiscord/bot.go
      Note: `showModal` and `modal(...)` runtime contract lives here
ExternalSources: []
Summary: Quick reference for modal payloads, text inputs, and modal submit context shapes.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Provide copy/paste-ready examples for modal implementation.
WhenToUse: Use when implementing or reviewing modal support.
---

# Goal

Provide a quick contract reference for modal presentation and modal submit handling.

## Quick API

### Open a modal

```js
await ctx.showModal({
  customId: "feedback:submit",
  title: "Feedback",
  components: [
    {
      type: "actionRow",
      components: [
        {
          type: "textInput",
          customId: "summary",
          label: "Summary",
          style: "short",
          required: true,
          minLength: 5,
          maxLength: 100,
        }
      ]
    }
  ]
})
```

### Receive a submit

```js
modal("feedback:submit", async (ctx) => {
  const summary = ctx.values.summary
  return { content: `Got: ${summary}`, ephemeral: true }
})
```

## Expected Context Fields

- `ctx.modal.customId`
- `ctx.values`
- `ctx.reply`, `ctx.defer`, `ctx.edit`, `ctx.followUp`

## File-Level Checklist

- `internal/jsdiscord/host.go`
- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/runtime_test.go`
- `examples/discord-bots/ping/index.js`
