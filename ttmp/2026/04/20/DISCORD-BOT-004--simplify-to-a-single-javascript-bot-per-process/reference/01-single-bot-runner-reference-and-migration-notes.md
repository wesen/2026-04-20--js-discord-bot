---
Title: Single-Bot Runner Reference and Migration Notes
Ticket: DISCORD-BOT-004
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
    - Path: internal/botcli/command.go
      Note: Operator-facing `bots` CLI contract should simplify here
    - Path: internal/jsdiscord/host.go
      Note: Live single-bot host after deleting the obsolete multi-host composition layer
ExternalSources: []
Summary: Quick reference for the single-bot-per-process model and migration notes away from multi-bot runtime composition.
LastUpdated: 2026-04-20T15:28:00-04:00
WhatFor: Give maintainers and operators a concise reference for the simplified runner model and the consequences of the design pivot.
WhenToUse: Use when discussing or implementing the simplification from multi-bot runs to single-bot runs.
---

# Single-Bot Runner Reference and Migration Notes

## Goal

Provide a short, operator-facing reference for the decision to run exactly one selected JavaScript bot implementation per process.

## Context

The project briefly added a multi-bot runner where commands like this were possible:

```bash
discord-bot bots run knowledge-base support moderation
```

This ticket proposes undoing that and standardizing on:

```bash
discord-bot bots run knowledge-base
```

If a bot author wants multiple logical capabilities, they should compose them inside one selected JS bot.

## Quick Reference

### Preferred CLI

```bash
discord-bot bots list --bot-repository ./examples/discord-bots
discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots
discord-bot bots run knowledge-base --bot-repository ./examples/discord-bots
```

### Preferred internal JS composition

```js
const { defineBot } = require("discord")
const support = require("./features/support")
const moderation = require("./features/moderation")
const knowledge = require("./features/knowledge")

module.exports = defineBot((api) => {
  support.register(api)
  moderation.register(api)
  knowledge.register(api)
})
```

### Why single-bot is simpler

| Concern | Multi-bot host model | Single-bot host model |
| --- | --- | --- |
| Operator UX | select one-or-many bots | select one bot |
| Runtime ownership | many runtimes per process | one runtime per process |
| Command collisions | must detect/reject | impossible within host-level selection |
| Startup flags | need namespacing | simple single schema |
| Future Glazed integration | awkward | straightforward |

## Migration notes

### Keep

- descriptor/discovery logic
- `bots list`
- `bots help <bot>`
- bot inspection via `describe()`

### Simplify

- `bots run <bot...>` -> `bots run <bot>`
- one selected script -> one runtime
- one selected bot -> one startup config schema

### Move out of host scope

- multi-bot in-process composition
- host-level collision handling between selected bots
- multi-bot startup flag namespacing

### Deletion note

`internal/jsdiscord/multihost.go` was fully deleted once the single-bot model was firm again.

Why delete it instead of merely keeping it unused?

- it encoded a runtime shape the project explicitly decided not to keep
- leaving it around made the codebase imply that host-level multi-bot composition was still a near-term supported option
- it created extra review and maintenance surface for code that was no longer on the intended architecture path
- descriptor inspection and repository discovery already preserved the useful pieces, so the remaining multi-host runtime layer was mostly architectural debt

The practical rule is now simple: one selected bot script maps to one live `Host`, one runtime, and one startup-config schema.

## Usage examples

### Current desired mental model

```bash
discord-bot bots run knowledge-base
```

### Future desired startup field model

```bash
discord-bot bots run knowledge-base --index-path ./docs --read-only --print-parsed-values
```

Where those flags come from the selected bot’s `configure({ run: ... })` schema and are parsed through the Glazed/Cobra stack.

## Related

- `design-doc/01-single-javascript-bot-per-process-architecture-and-implementation-guide.md`
- `tasks.md`
