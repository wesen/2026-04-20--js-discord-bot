---
Title: Implementation and Architecture Guide
Ticket: DISCORD-BOT-001
Status: active
Topics:
    - backend
    - chat
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/discord-bot/commands.go
      Note: Glazed run/validate/sync command implementations
    - Path: cmd/discord-bot/main.go
      Note: CLI entrypoint and signal-aware execution
    - Path: cmd/discord-bot/root.go
      Note: Glazed/Cobra root command
    - Path: internal/bot/bot.go
      Note: Discord session lifecycle and interaction handlers
    - Path: internal/config/config.go
      Note: Shared credential parsing and validation
ExternalSources: []
Summary: Architecture and implementation plan for a simple Go Discord bot with a Glazed-based CLI.
LastUpdated: 2026-04-20T10:04:43.000860095-04:00
WhatFor: Explain the recommended stack, CLI shape, runtime flow, and delivery plan.
WhenToUse: Use this when implementing or reviewing the first version of the bot.
---


# Implementation and Architecture Guide

## Executive Summary

Build a small but production-friendly Discord bot in Go using a Glazed-backed CLI as the entry point. The bot should start with a single primary command, `discord-bot run`, and be able to connect to Discord over the gateway, register slash commands, and respond to a small set of interactions such as `ping` and `echo`.

The recommended approach is intentionally simple: use `github.com/bwmarrin/discordgo` for the Discord API integration, keep configuration in environment variables plus CLI flags, and use Glazed for command parsing, help text, logging setup, and future subcommands.

## Problem Statement

A “simple bot” is easy to overcomplicate. Common failure modes are:

- mixing bot runtime logic with CLI wiring
- hard-coding credentials instead of reading config from the environment
- registering slash commands ad hoc inside event handlers
- making the startup path difficult to test
- adding too much framework before the bot has a stable shape

This guide defines a small, clear structure that keeps the bot maintainable as features grow.

## Goals

- Start the app through a Glazed command tree.
- Keep a single obvious runtime command for the first version.
- Support gateway-based bot behavior rather than a webhook-only design.
- Make credentials explicit and document how to obtain them.
- Keep the first version easy to test locally in one Discord guild.

## Non-Goals

- Full dashboard or web admin UI.
- Multi-tenant bot management.
- Persistent storage for user data.
- Advanced command framework or plugin system.
- Sharding at scale.

## Proposed Solution

### Recommended stack

- **CLI**: Glazed command definitions with Cobra integration.
- **Discord integration**: `discordgo`.
- **Config loading**: environment variables plus command flags.
- **Logging**: structured logging via the Glazed root command setup.
- **State**: in-memory for the first version.

### CLI shape

Use a Glazed root command and keep the first user-facing entry point simple:

```text
discord-bot run
```

Recommended subcommands:

```text
discord-bot run             Start the bot and connect to Discord

discord-bot sync-commands   Register slash commands to a guild or globally

discord-bot validate-config Check whether required credentials are present
```

For the first pass, `run` is the main command. `sync-commands` can be separate if you want command registration to be explicit and repeatable.

### Runtime flow

1. Parse CLI flags using Glazed.
2. Load environment-based configuration.
3. Validate that required credentials are present.
4. Create a Discord session.
5. Register event handlers.
6. Optionally register slash commands.
7. Open the gateway connection.
8. Handle messages and interactions.
9. On shutdown, close the session cleanly.

## Design Decisions

### 1) Use Glazed for the CLI entry point

Glazed gives the project a structured command layer from the beginning.

Benefits:

- consistent flags and help output
- easier future expansion into more commands
- built-in support for logging/help wiring patterns
- a clean place for config validation and diagnostics

Root command responsibilities should include:

- logging initialization
- help system wiring
- command registration
- shared flags such as `--log-level`

### 2) Keep Discord-specific runtime code out of the CLI package

The CLI should only parse options and call into an internal bot package.

Recommended split:

- `cmd/discord-bot/...` for CLI wiring
- `internal/config` for config loading/validation
- `internal/bot` for Discord session lifecycle and event handling
- `internal/commands` for slash-command definitions and registration

This separation keeps the bot easy to test without invoking the CLI.

### 3) Prefer gateway mode for the first version

Gateway-based bots are the simplest fit for common chat workflows:

- listen for messages
- respond to events
- support slash commands

A webhook-only architecture is not the right default here because it adds extra hosting complexity without simplifying the first release.

### 4) Start with guild-scoped slash commands during development

Guild commands propagate much faster than global commands. For development, register commands against a single guild.

Once behavior stabilizes, you can mirror the same command definitions to global registration.

### 5) Keep command registration explicit

Slash commands should not be registered implicitly inside every startup path unless there is a good reason. A separate `sync-commands` command makes it obvious when API state changes.

That also makes it safer to deploy a bot that is already running but should not re-register commands on every restart.

## Suggested Package Layout

```text
cmd/discord-bot/main.go
cmd/discord-bot/root.go
cmd/discord-bot/run.go
cmd/discord-bot/sync_commands.go

internal/config/config.go
internal/bot/bot.go
internal/bot/events.go
internal/bot/shutdown.go
internal/commands/definitions.go
internal/commands/register.go
internal/commands/handlers.go
internal/observability/logging.go
```

### File responsibilities

- `main.go`: bootstraps the root command.
- `root.go`: sets up Glazed/Cobra, logging, and help.
- `run.go`: runs the bot.
- `sync_commands.go`: registers slash commands.
- `config.go`: loads and validates environment/flags.
- `bot.go`: owns Discord session lifecycle.
- `events.go`: message and interaction handlers.
- `definitions.go`: command definitions.
- `register.go`: API calls to create or update commands.
- `handlers.go`: command dispatch logic.

## Credentials and Configuration

The bot should require the following configuration at minimum:

- `DISCORD_BOT_TOKEN`
- `DISCORD_APPLICATION_ID`

Recommended optional values:

- `DISCORD_GUILD_ID` for faster development registration
- `DISCORD_LOG_LEVEL`
- `DISCORD_COMMAND_PREFIX` if you support prefix commands later
- `DISCORD_ENV` or `APP_ENV` for environment-specific behavior

### Validation rules

- Token must not be empty.
- Application ID must not be empty.
- Guild ID is required only for guild-scoped command syncing.
- If you support presence or privileged intents, require the corresponding developer portal settings.

### Secrets handling

- Read tokens from environment variables first.
- Optionally support a local `.env` file during development.
- Never commit tokens to the repository.
- Never print the full token in logs.

## Discord Feature Model

### Minimal bot features

Start with only a few predictable behaviors:

- `ping` command returns `pong`
- `echo` command repeats user input
- a basic `help` command or help text via Glazed
- simple message event logging

### Interaction strategy

Use slash commands as the default interaction model. They are easier to discover and safer than parsing arbitrary chat text.

You can still add message-based triggers later, but they should be treated as optional extensions.

## Event Handling Model

The bot should register a small set of handlers:

- `Ready` or session-open logging
- `InteractionCreate` for slash commands
- `MessageCreate` only if you intentionally support message-based triggers
- `Disconnect`/shutdown handling for clean exit

Handler logic should be small and dispatch into pure functions whenever possible.

### Suggested pattern

- handlers parse Discord events
- handler functions call service-like helpers
- helpers return response text, embeds, or errors
- the Discord layer performs the actual reply/edit operations

This keeps the command logic testable without a live Discord session.

## Slash Command Registration

### Recommended command lifecycle

1. Define command metadata in code.
2. Sync commands with a dedicated command.
3. Use guild-scoped registration for development.
4. Switch to global registration once the command set stabilizes.

### Example commands

- `/ping`
- `/echo text:<string>`
- `/about`

If you later add admin commands, keep them separate from the basic utility commands.

## Glazed Command Implementation Sketch

A simple root command can look like this:

```go
type RunSettings struct {
    Token         string `glazed:"token"`
    ApplicationID string `glazed:"application-id"`
    GuildID       string `glazed:"guild-id"`
    SyncCommands  bool   `glazed:"sync-commands"`
}
```

The `run` command should expose flags that map to these settings, while still allowing environment variables to fill the same values.

Useful flag examples:

- `--token`
- `--application-id`
- `--guild-id`
- `--sync-commands`
- `--log-level`

## Example Startup Sequence

```text
1. discord-bot run
2. Load flags and env vars
3. Validate config
4. Create Discord session
5. Register handlers
6. Optionally sync slash commands
7. Open the gateway
8. Wait for SIGINT/SIGTERM
9. Close session cleanly
```

## Observability

For the first version, logging is enough.

Log these events:

- startup/shutdown
- config validation failures
- command sync results
- handler errors
- Discord reconnect/disconnect events

Keep logs structured and avoid logging secrets or user content unless needed for debugging.

## Testing Strategy

### Unit tests

- config validation
- command metadata generation
- command dispatch helpers
- response formatting

### Integration tests

- startup path with stubbed config
- command registration using mocked HTTP clients where practical
- a manual smoke test in a private guild

### Manual smoke test

1. Invite the bot to a private server.
2. Run `discord-bot sync-commands`.
3. Run `discord-bot run`.
4. Execute `/ping`.
5. Confirm logs show the interaction and reply.

## Alternatives Considered

### Plain Cobra without Glazed

Rejected because the user specifically wants Glazed commands as the starting CLI layer, and Glazed adds a structured help/output model that is useful even for a small bot.

### Discord interactions only via HTTP webhooks

Rejected for the first version because it adds deployment overhead and is less straightforward than a gateway bot for chat automation.

### Prefix commands only

Rejected as the primary interface because slash commands are the modern Discord default and are easier for users to discover.

## Implementation Plan

### Phase 1: CLI and config foundation

- create the Glazed root command
- add `run`, `sync-commands`, and `validate-config`
- implement config loading and validation
- wire logging and help

### Phase 2: Discord session and basic behavior

- create a Discord session wrapper
- connect to the gateway
- add `/ping`
- add basic message logging

### Phase 3: command syncing and cleanup

- implement guild command registration
- add `echo` and `about`
- refine error handling and shutdown

### Phase 4: test and document

- add unit tests for config and handlers
- create a private guild smoke test checklist
- document deployment and invite URLs

## Open Questions

- Should the bot support prefix commands at all, or stay slash-command only?
- Should command registration happen in a separate subcommand or as a startup flag?
- Do you want a `.env` file for local development, or environment variables only?
- Should the first release be guild-scoped only, or should it support global commands from day one?

## References

- `reference/02-discord-credentials-and-setup.md`
- `reference/01-diary.md`
- Glazed command conventions from the project skill guidance
