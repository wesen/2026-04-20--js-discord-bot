---
Title: Local Validation and Smoke Test Checklist
Ticket: DISCORD-BOT-001
Status: active
Topics:
    - backend
    - chat
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/discord-bot/commands.go
      Note: Validate/sync/run commands exercised by the smoke tests
    - Path: cmd/discord-bot/main.go
      Note: Commands and signal handling used in the smoke tests
    - Path: internal/bot/bot.go
      Note: Gateway connection and slash-command behavior validated by the playbook
ExternalSources: []
Summary: Repeatable local validation steps for the Discord bot CLI and gateway flow.
LastUpdated: 2026-04-20T10:11:00-04:00
WhatFor: Verify configuration loading, command syncing, and startup behavior before a release or review.
WhenToUse: Use after changing CLI flags, env loading, or bot behavior.
---


# Local Validation and Smoke Test Checklist

## Purpose

Validate the Discord bot CLI locally before inviting the bot to a server or merging changes. This playbook checks that the Glazed command tree, environment loading, config validation, and slash-command sync path all work together.

## Environment Assumptions

- `direnv` or your shell has already exported the Discord environment variables.
- A private test guild is available for command sync and smoke tests.
- The bot token has not been revoked.
- The Discord application already has a bot user.

## Commands

### 1) Load the environment

```bash
direnv allow
```

Expected result:
- `.envrc` values are exported into the shell session.

### 2) Confirm the CLI builds

```bash
go test ./...
```

Expected result:
- All packages compile and the test run succeeds.

### 3) Validate configuration only

```bash
go run ./cmd/discord-bot validate-config
```

Expected result:
- A row or table indicates the configuration is valid.
- No token secret is printed in full.

### 4) Sync slash commands to the development guild

```bash
go run ./cmd/discord-bot sync-commands
```

Expected result:
- `/ping` and `/echo` are registered.
- The scope is the configured guild when `DISCORD_GUILD_ID` is present.

### 5) Start the bot

```bash
go run ./cmd/discord-bot run
```

Expected result:
- The process connects to Discord.
- The logs show a ready/connected message.
- `/ping` replies with `pong` in the test guild.
- `/echo` returns the supplied text.

## Exit Criteria

The bot is considered locally ready when:

- `go test ./...` passes.
- `validate-config` succeeds using the shell environment.
- `sync-commands` registers the expected commands.
- `run` connects cleanly and responds to `/ping`.

## Failure Modes

- **Missing token or application ID**: the validate command should fail fast with a clear message.
- **Guild ID missing**: `sync-commands` may fall back to global registration; double-check before running against production.
- **Discord reconnect errors**: inspect the logs and confirm the token has not been revoked.

## Related

- `design-doc/01-implementation-and-architecture-guide.md`
- `reference/02-discord-credentials-and-setup.md`
