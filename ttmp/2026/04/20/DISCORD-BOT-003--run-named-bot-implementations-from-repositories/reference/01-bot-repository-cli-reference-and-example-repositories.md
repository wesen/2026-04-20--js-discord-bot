---
Title: Bot Repository CLI Reference and Example Repositories
Ticket: DISCORD-BOT-003
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
      Note: Bot-oriented CLI entrypoints live here
    - Path: examples/discord-bots/README.md
      Note: Repository of example bot implementations for the new CLI surface
ExternalSources: []
Summary: Quick reference for the named bot repository runner CLI and the example bot repository layout.
LastUpdated: 2026-04-20T14:55:00-04:00
WhatFor: Provide copy/paste CLI examples and repository layout guidance for named bot implementations.
WhenToUse: Use when testing or authoring named bot repositories for this repo.
---

# Bot Repository CLI Reference and Example Repositories

## Goal

Provide quick-reference examples for listing, inspecting, and running named bot implementations from repositories.

## Context

The `bots` command group is bot-oriented, not function-oriented. A bot implementation is a whole JS bot script/package that exports `defineBot(...)`.

## Quick Reference

### List available bots

```bash
discord-bot bots list --bot-repository ./examples/discord-bots
```

### Show one bot

```bash
discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots
```

### Run one bot

```bash
discord-bot bots run knowledge-base \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID"
```

### Run multiple bots in one host

```bash
discord-bot bots run knowledge-base support moderation \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --sync-on-start
```

## Example repository shape

```text
examples/discord-bots/
  knowledge-base/
    index.js
    lib/
      docs.js
  support/
    index.js
  moderation/
    index.js
  announcements.js
```

## Included example bots

- `knowledge-base` — relative `require()` helper, search/article commands, `messageCreate`
- `support` — deferred/edit/follow-up interaction flow, embeds, buttons, `guildCreate`
- `moderation` — embeds, components, ephemeral responses, `messageCreate`
- `announcements` — root-level bot script for direct-file discovery

## Usage notes

- Use `configure({ name: ... })` to set the bot name explicitly.
- Keep slash-command names globally unique across selected bots.
- Use helper modules under bot subdirectories when you want to test relative `require()` behavior.
- Use `--sync-on-start` with `bots run` when you want the selected bots to bulk-overwrite slash commands before opening the gateway.
