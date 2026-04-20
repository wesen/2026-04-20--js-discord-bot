---
Title: Debugging Message and Channel Moderation Flows
Ticket: DISCORD-BOT-010
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
    - Path: internal/jsdiscord/host.go
      Note: Host moderation operations and debug logs for message/channel utilities live here
    - Path: internal/jsdiscord/bot.go
      Note: JS capability object exposes the implemented message/channel moderation methods here
    - Path: examples/discord-bots/moderation/index.js
      Note: Moderation example bot advertises the implemented message/channel utility commands here
ExternalSources: []
Summary: Operator playbook for validating and debugging the implemented DISCORD-BOT-010 message and channel moderation utility APIs.
LastUpdated: 2026-04-20T21:00:00-04:00
WhatFor: Help operators validate and debug message fetch/pin/unpin/listPinned/bulkDelete and channel fetch/topic/slowmode flows.
WhenToUse: Use when moderation utility commands do not behave as expected.
---

# Debugging Message and Channel Moderation Flows

## Goal

Provide a practical runbook for validating and debugging the implemented DISCORD-BOT-010 surfaces:

- `messages.fetch`
- `messages.pin`
- `messages.unpin`
- `messages.listPinned`
- `messages.bulkDelete`
- `channels.fetch`
- `channels.setTopic`
- `channels.setSlowmode`

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

### 1. Confirm the selected commands exist

```bash
GOWORK=off go run ./cmd/discord-bot bots help moderation --bot-repository ./examples/discord-bots
```

Expected command list should include:

- `mod-fetch-message`
- `mod-pin`
- `mod-unpin`
- `mod-list-pins`
- `mod-bulk-delete`
- `mod-fetch-channel`
- `mod-set-topic`
- `mod-set-slowmode`

### 2. Confirm the bot is running in the expected guild/channel

If a command relies on `ctx.channel.id`, make sure you are invoking it in the intended guild text channel.

## Message moderation checklist

### `mod-fetch-message`

Expected inputs:
- one valid message ID in the current channel

What to expect:
- ephemeral confirmation
- embed showing fetched message content
- debug host log for `discord.messages.fetch`

Common failure modes:
- wrong message ID
- bot lacks permission to read message history in the channel

### `mod-pin` / `mod-unpin`

Expected inputs:
- one valid message ID in the current channel

What to expect:
- ephemeral confirmation
- debug host log for `discord.messages.pin` or `discord.messages.unpin`

Common failure modes:
- bot lacks permission to manage messages
- wrong message ID
- channel mismatch

### `mod-list-pins`

What to expect:
- ephemeral summary with pinned message count
- embed listing pinned message IDs/content previews
- debug host log for `discord.messages.listPinned`

Common failure modes:
- bot lacks permission to read pinned/history information in the channel

### `mod-bulk-delete`

Expected inputs:
- comma-separated message IDs in the current channel

What to expect:
- ephemeral confirmation with count
- debug host log for `discord.messages.bulkDelete`

Common failure modes:
- bot lacks permission to manage messages
- IDs are invalid or from the wrong channel
- Discord rejects one or more targeted messages

## Channel utility checklist

### `mod-fetch-channel`

What to expect:
- ephemeral summary of current channel data
- debug host log for `discord.channels.fetch`

### `mod-set-topic`

Expected inputs:
- a topic string

What to expect:
- ephemeral confirmation
- debug host log for `discord.channels.setTopic`

Common failure modes:
- bot lacks permission to manage the channel

### `mod-set-slowmode`

Expected inputs:
- a numeric slowmode value in seconds

What to expect:
- ephemeral confirmation
- debug host log for `discord.channels.setSlowmode`

Common failure modes:
- bot lacks permission to manage the channel
- invalid value for the channel or guild context

## Debug logs to look for

With `--log-level debug`, the host should now emit structured logs for:

- `discord.messages.fetch`
- `discord.messages.pin`
- `discord.messages.unpin`
- `discord.messages.listPinned`
- `discord.messages.bulkDelete`
- `discord.channels.fetch`
- `discord.channels.setTopic`
- `discord.channels.setSlowmode`

Those logs should help answer:

- did JavaScript call the host utility?
- which channel/message IDs were targeted?
- how many message IDs were bulk-deleted?
- which topic/slowmode value was attempted?

## Validation commands

```bash
GOWORK=off go test ./internal/jsdiscord ./internal/bot ./cmd/discord-bot
GOWORK=off go test ./...
GOWORK=off go run ./cmd/discord-bot bots help moderation --bot-repository ./examples/discord-bots
```

## Review starting points

If the commands are present in help output but the live behavior is wrong, review these files first:

- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/host.go`
- `internal/jsdiscord/runtime_test.go`
- `examples/discord-bots/moderation/lib/register-message-moderation-commands.js`
- `examples/discord-bots/moderation/lib/register-channel-moderation-commands.js`
