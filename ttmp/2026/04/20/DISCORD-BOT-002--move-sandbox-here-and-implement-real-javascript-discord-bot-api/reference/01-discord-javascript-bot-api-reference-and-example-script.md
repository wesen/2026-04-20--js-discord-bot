---
Title: Discord JavaScript Bot API Reference and Example Script
Ticket: DISCORD-BOT-002
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
    - Path: internal/jsdiscord/bot.go
      Note: JS-visible API and runtime context contract
    - Path: internal/jsdiscord/host.go
      Note: Response normalization and Discord command metadata mapping
    - Path: examples/js-bots/ping.js
      Note: First example script using the real local Discord JS API
ExternalSources: []
Summary: Quick reference for the local `require("discord")` API, command specs, context fields, and example bot script.
LastUpdated: 2026-04-20T14:13:00-04:00
WhatFor: Give maintainers a copy/paste-friendly view of the Discord JS bot API that now lives in this repository.
WhenToUse: Use when authoring or reviewing JavaScript bot scripts for this repo.
---

# Discord JavaScript Bot API Reference and Example Script

## Goal

Provide a compact, copy/paste-friendly reference for the local JavaScript Discord bot API.

## Context

This API is hosted in `js-discord-bot`, not `go-go-goja`.

Entry point:

```js
const { defineBot } = require("discord")
```

## Quick Reference

### Definition API

| API | Purpose |
| --- | --- |
| `defineBot(builderFn)` | Define one Discord bot script |
| `command(name, spec?, handler)` | Register a slash command |
| `event(name, handler)` | Register an event handler such as `ready` |
| `configure(options)` | Record bot metadata |

### Supported command spec fields in the first slice

| Field | Purpose |
| --- | --- |
| `description` | Slash-command description |
| `options` | Option definitions keyed by option name |

### Supported option fields in the first slice

| Field | Purpose |
| --- | --- |
| `type` | `string`, `integer`, `boolean`, `number`, `user`, `channel`, `role`, `mentionable` |
| `description` | Discord option description |
| `required` | Whether the option is required |

### Runtime context

| Field | Purpose |
| --- | --- |
| `ctx.args` | Parsed slash-command options as a simple object |
| `ctx.options` | Alias for `ctx.args` |
| `ctx.command` | Basic command metadata |
| `ctx.interaction` | Basic interaction metadata |
| `ctx.message` | Message metadata for `messageCreate` handlers |
| `ctx.user` | Invoking user |
| `ctx.guild` | Guild metadata when present |
| `ctx.channel` | Channel metadata when present |
| `ctx.me` | Current bot user |
| `ctx.metadata` | Host-provided metadata such as `scriptPath` |
| `ctx.reply(payload)` | Send the initial interaction response, or a channel reply for message events |
| `ctx.defer(payload?)` | Send a deferred initial interaction response |
| `ctx.edit(payload)` | Edit the deferred/original interaction response |
| `ctx.followUp(payload)` | Send an interaction follow-up message |
| `ctx.store.*` | Runtime-local in-memory store |
| `ctx.log.info/debug/warn/error(msg, fields)` | Structured logging |

### Supported response payload shapes

Supported return or reply payloads now include:

```js
"pong"
{ content: "pong" }
{ content: "secret pong", ephemeral: true }
{
  content: "pong",
  embeds: [{ title: "Pong", description: "From JavaScript" }],
  components: [{
    type: "actionRow",
    components: [{ type: "button", style: "link", label: "Docs", url: "https://example.com" }]
  }]
}
```

## Usage Example

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "ping-bot", runtime: "js" })

  command("ping", {
    description: "Reply with pong from the JavaScript Discord bot API"
  }, async () => {
    return {
      content: "pong",
      embeds: [{ title: "Pong", description: "JavaScript handled this slash command." }],
    }
  })

  command("echo", {
    description: "Echo text back from JavaScript",
    options: {
      text: {
        type: "string",
        description: "Text to echo back",
        required: true,
      }
    }
  }, async (ctx) => {
    await ctx.defer({ ephemeral: true })
    await ctx.edit({ content: ctx.args.text })
    await ctx.followUp({ content: "Follow-up from JavaScript", ephemeral: true })
  })

  event("ready", async (ctx) => {
    ctx.log.info("js discord bot connected", {
      user: ctx.me && ctx.me.username,
      script: ctx.metadata && ctx.metadata.scriptPath,
    })
  })

  event("messageCreate", async (ctx) => {
    if ((ctx.message && ctx.message.content || "").trim() === "!pingjs") {
      await ctx.reply({ content: "pong from messageCreate" })
    }
  })
})
```

## Intended local usage

```bash
export DISCORD_BOT_SCRIPT=./examples/js-bots/ping.js
GOWORK=off go run ./cmd/discord-bot sync-commands
GOWORK=off go run ./cmd/discord-bot run
```

## Related

- `design-doc/01-sandbox-move-and-discord-javascript-api-architecture-guide.md`
- `reference/02-diary.md`
