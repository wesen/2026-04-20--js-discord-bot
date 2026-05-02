---
Title: Slack Backend Setup and Smoke Test
Ticket: slack-backend
Status: active
Topics:
    - slack
    - backend
    - javascript
    - bot-framework
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/discord-bot/slack_commands.go
      Note: Commands used by the setup and smoke-test playbook
    - Path: examples/discord-bots/adventure/index.js
      Note: Example unchanged JavaScript bot for Slack smoke testing
    - Path: internal/jsdiscord/slack_backend.go
      Note: Slack backend behavior validated by smoke-test steps
ExternalSources: []
Summary: Repeatable setup and smoke-test procedure for generated Slack manifest, HTTP Slack backend, and unchanged JavaScript bot scripts.
LastUpdated: 2026-05-01T16:54:49.835099-07:00
WhatFor: Use to create/install a Slack app from the generated manifest and smoke test the Slack HTTP backend.
WhenToUse: After changing Slack backend routing, manifest generation, Block Kit rendering, or deployment configuration.
---


# Slack Backend Setup and Smoke Test

## Purpose

Smoke test the Slack backend implementation with an unchanged JavaScript bot script. This validates the end-to-end shape:

1. Generate Slack app manifest.
2. Install/configure Slack app.
3. Expose local HTTP server through HTTPS tunnel.
4. Run `slack-serve`.
5. Invoke slash command and click Block Kit buttons.
6. Confirm message edits and SQLite state are working.

## Environment Assumptions

Required:

- A Slack workspace where you can create/install apps.
- Public HTTPS URL for the local server, e.g. ngrok or cloudflared.
- `SLACK_BOT_TOKEN` from the installed Slack app.
- `SLACK_SIGNING_SECRET` from the Slack app basic information page.
- Bot script that uses the existing JS `discord`/`ui` APIs.

Optional but useful:

- `OPENROUTER_API_KEY` if testing the adventure bot's LLM flow.
- SQLite CLI for inspecting adapter state.

## Commands

### 1. Run tests locally

```bash
go test ./... -count=1
```

Expected: all packages pass.

### 2. Start HTTPS tunnel

Example with ngrok:

```bash
ngrok http 8080
```

Record the HTTPS forwarding URL, e.g.:

```text
https://abc123.ngrok-free.app
```

### 3. Generate Slack manifest

```bash
go run ./cmd/discord-bot slack-manifest \
  --bot-script ./examples/discord-bots/adventure/index.js \
  --base-url https://abc123.ngrok-free.app \
  > /tmp/slack-adventure-manifest.json
```

Review command URLs:

```bash
grep -E 'slack/(commands|interactivity|events)' /tmp/slack-adventure-manifest.json
```

Expected URLs:

```text
https://abc123.ngrok-free.app/slack/commands
https://abc123.ngrok-free.app/slack/interactivity
https://abc123.ngrok-free.app/slack/events
```

### 4. Create/install Slack app

In Slack app management:

1. Create app from manifest.
2. Paste `/tmp/slack-adventure-manifest.json`.
3. Install app to workspace.
4. Copy bot token into environment as `SLACK_BOT_TOKEN`.
5. Copy signing secret into environment as `SLACK_SIGNING_SECRET`.
6. Invite the bot into the channel if needed:

```text
/invite @your-bot-name
```

### 5. Run Slack backend

```bash
export SLACK_BOT_TOKEN=xoxb-...
export SLACK_SIGNING_SECRET=...

go run ./cmd/discord-bot slack-serve \
  --bot-script ./examples/discord-bots/adventure/index.js \
  --listen-addr :8080 \
  --slack-state-db ./var/slack-adventure.sqlite
```

Expected log line:

```text
starting slack backend addr=:8080
```

### 6. Smoke test slash command

In Slack channel:

```text
/adventure-start underwater adventure
```

Expected:

- Bot posts a message in the channel.
- Message is rendered with Slack Block Kit buttons.
- No JavaScript changes are required.
- If the bot script uses `ctx.edit`, the same Slack message should update rather than posting a new public message each time.

### 7. Smoke test button interactivity

Click one of the rendered buttons.

Expected:

- Slack sends `block_actions` to `/slack/interactivity`.
- Backend dispatches the existing JS component handler by `action_id` / custom ID.
- Bot updates the original message via `chat.update`.

### 8. Smoke test modal interactivity

If the bot has a modal button, click it and submit.

Expected:

- Backend calls `views.open` using the Slack `trigger_id`.
- Modal submission dispatches existing JS modal handler by `callback_id`.
- `ctx.values` contains plain input values keyed by action ID.

### 9. Inspect SQLite state

```bash
sqlite3 ./var/slack-adventure.sqlite '.tables'
sqlite3 ./var/slack-adventure.sqlite 'select team_id, channel_id, message_ts, substr(content,1,80) from slack_messages order by updated_at desc limit 5;'
sqlite3 ./var/slack-adventure.sqlite 'select kind, callback_id, action_id, created_at from slack_interactions order by created_at desc limit 10;'
```

Expected tables:

```text
slack_interactions
slack_messages
slack_modal_contexts
```

## Exit Criteria

Success means:

- `go test ./... -count=1` passes.
- Generated manifest installs in Slack.
- `slack-serve` accepts signed Slack requests.
- Slash command dispatches to unchanged JS command handler.
- Buttons dispatch to unchanged JS component handlers.
- Modals dispatch to unchanged JS modal handlers, if tested.
- Public command responses and later edits update the same Slack message.
- SQLite contains message and interaction records.

## Notes

### One-option command limitation

Slack raw slash command text is mapped to one JS option. The mapper prefers option names in this order:

1. `prompt`
2. `text`
3. `query`
4. `input`
5. `message`
6. alphabetical fallback

For the adventure bot, this means:

```text
/adventure-start underwater adventure
```

normalizes to:

```json
{ "prompt": "underwater adventure" }
```

### Files/exports

Slack file upload is intentionally not implemented in the first version. Normalized response `files` are appended to message text as inline code blocks and may be truncated.

### Common failures

| Symptom | Likely cause |
|---|---|
| `invalid slack signature` | Wrong signing secret, changed tunnel/body, or request timestamp outside tolerance |
| Slack says command failed immediately | Endpoint not reachable or app manifest URL mismatch |
| `channel_not_found` / `not_in_channel` | Bot not invited to channel or missing write permission |
| `missing_scope` | Add needed scope and reinstall Slack app |
| Modal does not open | `trigger_id` expired or `views.open` failed |
| Edits create new messages | Message timestamp was not recorded; inspect `slack_messages` and backend logs |
