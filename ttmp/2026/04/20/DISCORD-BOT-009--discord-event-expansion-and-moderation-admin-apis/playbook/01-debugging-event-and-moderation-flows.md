---
Title: Debugging Event and Moderation Flows
Ticket: DISCORD-BOT-009
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/bot/bot.go
      Note: Live Discord session wiring and required intents for event delivery
    - Path: internal/jsdiscord/host.go
      Note: Debug lifecycle logs and host moderation operations live here
    - Path: examples/discord-bots/moderation/index.js
      Note: Example bot that demonstrates the currently implemented DISCORD-BOT-009 surfaces
ExternalSources: []
Summary: Operator playbook for validating and debugging the implemented DISCORD-BOT-009 event and moderation slices.
LastUpdated: 2026-04-20T20:15:00-04:00
WhatFor: Help operators verify gateway intents, inspect event delivery, and debug host moderation operations using the current logging and example bot surfaces.
WhenToUse: Use when a moderation example command or event handler does not behave as expected.
---

# Debugging Event and Moderation Flows

## Goal

Provide a practical runbook for validating the currently implemented DISCORD-BOT-009 slices:

- message lifecycle events
- reaction events
- guild member events
- host moderation member APIs

## Recommended startup command

Run the moderation example bot with debug logs enabled:

```bash
GOWORK=off go run ./cmd/discord-bot bots run moderation \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --sync-on-start \
  --log-level debug
```

## First checks

### 1. Confirm the selected bot and event list

Use help output before starting the live bot:

```bash
GOWORK=off go run ./cmd/discord-bot bots help moderation --bot-repository ./examples/discord-bots
```

Expected events should include:

- `messageCreate`
- `messageUpdate`
- `messageDelete`
- `reactionAdd`
- `reactionRemove`
- `guildMemberAdd`
- `guildMemberUpdate`
- `guildMemberRemove`

### 2. Confirm commands are present

Expected moderation example commands should include:

- `mod-summary`
- `mod-guidelines`
- `mod-add-role`
- `mod-timeout`

## Event debugging checklist

### Message lifecycle events

To exercise message events:

- send `!modping`
- edit a normal user message
- delete a normal user message

What to expect:

- `!modping` should trigger the moderation bot reply through `messageCreate`
- message edits should produce JS log output about `messageUpdate`
- deletes should produce JS log output about `messageDelete`

If message update/delete do not arrive:

- verify the bot is connected to the expected guild
- verify you are testing with non-bot user messages
- confirm the bot process is running with `--log-level debug`
- confirm the moderation bot was the selected bot implementation

### Reaction events

To exercise reaction events:

- add a reaction to a message in the guild
- remove the reaction

What to expect:

- debug logs from the host showing event dispatch
- JS-side moderation logs for `reactionAdd` / `reactionRemove`

If reactions do not arrive:

- verify the bot has the reaction intent enabled in the current build
- confirm the bot has access to the channel where the reaction occurs

### Guild member events

To exercise member events:

- join with a test account or observe a real member join/leave/update
- if possible, modify member roles/nicknames to provoke `guildMemberUpdate`

What to expect:

- JS-side moderation logs for joins, updates, and leaves

If guild member events do not arrive:

- verify the bot application has the necessary member intent enabled in Discord’s developer settings if required for your deployment
- verify the runtime is using the current build with `GuildMembers` intent wired in

## Moderation command debugging checklist

### `mod-add-role`

Expected inputs:

- a valid target user ID
- a valid role ID
- command must be run in a guild

Common failure modes:

- wrong user ID or role ID
- bot lacks permission to manage roles
- bot’s highest role is below the target role in the hierarchy

### `mod-timeout`

Expected inputs:

- a valid target user ID
- a positive duration in seconds
- command must be run in a guild

Common failure modes:

- bot lacks timeout/moderation permission
- bot’s role hierarchy is below the target member
- invalid or unexpected duration value

## Debug logs to look for

With `--log-level debug`, the host now emits lifecycle logs for:

- event/interaction dispatch start
- reply / defer / follow-up / edit / modal actions
- host Discord operations such as:
  - `discord.channels.send`
  - `discord.messages.edit`
  - `discord.messages.delete`
  - `discord.messages.react`
  - `discord.members.addRole`
  - `discord.members.removeRole`
  - `discord.members.timeout`

These logs should help answer:

- did the event reach the host?
- did the host call into JavaScript?
- did JavaScript call a host moderation operation?
- what target IDs were used?

## Known current limitations

- `timeout(...)` currently supports:
  - `durationSeconds`
  - `until` (RFC3339)
  - `clear: true`
- `timeout(...)` does not yet send an audit-log reason
- destructive moderation APIs like `kick`, `ban`, and `unban` are not yet implemented in this slice

## Validation commands

```bash
GOWORK=off go test ./internal/jsdiscord ./internal/bot ./cmd/discord-bot
GOWORK=off go test ./...
GOWORK=off go run ./cmd/discord-bot bots help moderation --bot-repository ./examples/discord-bots
```

## When to escalate to code review

If the bot is connected, the help output shows the expected events/commands, and the debug logs still do not show dispatch or host-operation entries, review these files first:

- `internal/bot/bot.go`
- `internal/jsdiscord/host.go`
- `examples/discord-bots/moderation/index.js`
