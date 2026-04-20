---
Title: Discord Autocomplete and Richer Command Option Metadata Architecture and Implementation Guide
Ticket: DISCORD-BOT-007
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
    - Path: internal/jsdiscord/host.go
      Note: Current option normalization must grow to support richer metadata and autocomplete responses
    - Path: internal/jsdiscord/bot.go
      Note: Builder and runtime dispatch should register autocomplete handlers
    - Path: internal/jsdiscord/runtime_test.go
      Note: Tests should cover focused options and response choice normalization
    - Path: examples/discord-bots/ping/index.js
      Note: Example bot should demonstrate at least one autocomplete-backed command
ExternalSources: []
Summary: Detailed design for autocomplete interactions and richer slash-command option metadata in the JavaScript Discord bot API.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Explain how command options should evolve beyond the current minimal schema and how dynamic autocomplete should be wired.
WhenToUse: Use when implementing or reviewing autocomplete and richer option support.
---

# Discord Autocomplete and Richer Command Option Metadata Architecture and Implementation Guide

## Executive Summary

The current command option schema is intentionally small: type, description, and required-ness. That is enough for basic commands, but not for modern Discord UX. Discord autocomplete introduces a second interaction path for commands, and richer option metadata makes those commands easier to validate and use.

This ticket adds both pieces in one cohesive slice because they depend on the same command spec normalization path.

## Problem Statement

Current gaps include:

- no way to mark an option as `autocomplete: true`
- no way to register a JS autocomplete handler
- no way to emit autocomplete result choices
- no support for useful metadata such as `choices`, `minLength`, `maxLength`, `minValue`, or `maxValue`

Because of these gaps, richer command UX still has to live outside the JS API.

## Scope

This ticket covers:

- autocomplete handler registration and dispatch
- focused-option context shape
- richer option metadata on command specs
- normalization of autocomplete results into Discord choices

Subcommands and localization can come later if needed; they are intentionally left out of the first implementation slice so review stays manageable.

## Proposed JS API

### Command spec

```js
command("kb-search", {
  description: "Search docs",
  options: {
    query: {
      type: "string",
      description: "Search query",
      required: true,
      autocomplete: true,
      minLength: 2,
      maxLength: 100,
    },
    limit: {
      type: "integer",
      description: "Number of results",
      minValue: 1,
      maxValue: 20,
    }
  }
}, async (ctx) => {
  return { content: `query=${ctx.args.query}` }
})
```

### Autocomplete registration

```js
autocomplete("kb-search", "query", async (ctx) => {
  const q = String(ctx.focused.value || "")
  return [
    { name: `Search for ${q}`, value: q },
    { name: "Architecture", value: "architecture" },
  ]
})
```

## Current-State Analysis

`internal/jsdiscord/host.go` already turns command specs into `discordgo.ApplicationCommandOption` values, but today it only sets:

- type
- description
- required

That means most of the implementation work is an extension of an already-good normalization seam. This is the right ticket boundary because the same file that generates command options should also understand which options are autocomplete-enabled.

## Proposed Runtime Model

### Registration model

Add an autocomplete registry keyed by:

- command name
- option name

A simple composite key is enough for v1.

### Dispatch model

When Discord sends `InteractionApplicationCommandAutocomplete`:

1. inspect `ApplicationCommandData()`
2. locate the focused option
3. build a request with both the full option map and the focused option info
4. dispatch to the matching handler
5. normalize the returned array into Discord choices

### Context contract

Autocomplete handlers should receive:

- `ctx.command.name`
- `ctx.args` for the partially filled option set
- `ctx.focused.name`
- `ctx.focused.value`
- `ctx.interaction`, `ctx.user`, `ctx.guild`, `ctx.channel`

## Pseudocode

```text
on autocomplete interaction:
  data = interaction.ApplicationCommandData()
  focused = findFocusedOption(data.Options)
  handler = bot.lookupAutocomplete(data.Name, focused.Name)
  choices = handler({ args: optionMap(data.Options), focused: focused })
  respond(type=AutocompleteResult, data={ choices: normalize(choices) })
```

## Implementation Plan

### Phase 1 — richer spec normalization

Files:

- `internal/jsdiscord/host.go`

Work:

- map `autocomplete` into `ApplicationCommandOption.Autocomplete`
- map static `choices`
- map numeric and string min/max constraints

### Phase 2 — runtime registration and dispatch

Files:

- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/host.go`

Work:

- add `autocomplete(commandName, optionName, handler)`
- add dispatch method for autocomplete interactions
- add result normalization for `ApplicationCommandOptionChoice`

### Phase 3 — descriptors and help visibility

Files:

- `internal/jsdiscord/descriptor.go`
- `internal/botcli/command.go`

Work:

- include autocomplete info in descriptors
- optionally show autocomplete-enabled options in bot help output

### Phase 4 — examples and tests

Files:

- `internal/jsdiscord/runtime_test.go`
- `examples/discord-bots/ping/index.js`
- `examples/discord-bots/README.md`

Work:

- add runtime tests for focused option dispatch
- add one realistic example command

## Test Strategy

1. Command option normalization should preserve `autocomplete`, `choices`, and length/value constraints.
2. Focused-option detection should work for typical top-level command options.
3. Autocomplete handlers should return up to 25 normalized choices.
4. Commands should still execute normally after autocomplete-enabled options are submitted.

## Risks and Tradeoffs

### Risk: nested option trees

Subcommands and groups add a recursive option shape. The first slice can support top-level options first, as long as the design leaves room for recursive traversal later.

### Risk: mixing static choices and autocomplete

Discord treats choices and autocomplete as mutually exclusive per option. The normalization code should reject invalid specs early.

## References

- `internal/jsdiscord/host.go`
- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/runtime_test.go`
- `examples/discord-bots/ping/index.js`
