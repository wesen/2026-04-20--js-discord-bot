---
Title: Slack API and Block Kit Reference
Ticket: slack-backend
Status: active
Topics:
    - slack
    - backend
    - javascript
    - bot-framework
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/jsdiscord/host_dispatch.go
      Note: Current Discord interaction dispatch path to compare against Slack block_actions/view_submission dispatch
    - Path: internal/jsdiscord/host_responses.go
      Note: Current normalized response/edit handling that Slack chat.postMessage/chat.update should emulate
    - Path: internal/jsdiscord/ui_components.go
      Note: Existing transport-neutral-ish UI component builders to map onto Block Kit buttons/actions
    - Path: internal/jsdiscord/ui_form.go
      Note: Existing modal/form abstraction to map onto Slack views.open/view_submission
    - Path: pkg/botcli/runtime_factory.go
      Note: Runtime/module registration point likely affected by adding a Slack backend
ExternalSources:
    - https://api.slack.com/apis
    - https://api.slack.com/web
    - https://api.slack.com/events-api
    - https://api.slack.com/interactivity
    - https://api.slack.com/block-kit
    - https://api.slack.com/reference/block-kit/blocks
    - https://api.slack.com/reference/block-kit/block-elements
    - https://api.slack.com/reference/block-kit/composition-objects
    - https://api.slack.com/surfaces/messages
    - https://api.slack.com/surfaces/modals
    - https://api.slack.com/authentication/oauth-v2
Summary: Reference for adding a Slack backend while preserving the existing JavaScript bot API. Covers Slack Web API, Events API, interactivity, slash commands, modals, Block Kit mapping, state/correlation, auth/scopes, and backend abstraction implications.
LastUpdated: 2026-05-01T16:20:02.03721-07:00
WhatFor: Use when designing or implementing Slack as an additional transport/backend for the bot framework.
WhenToUse: Before changing runtime dispatch, UI response normalization, interaction handling, command registration, or adding Slack-specific adapter code.
---


# Slack API and Block Kit Reference

## Goal

Add a Slack backend/capability to the bot framework such that existing JavaScript bots ideally do **not** need to change. JavaScript should continue to express commands, events, UI messages, buttons, and modals through the current framework API; the Go/runtime layer should adapt those concepts to Slack Web API, Events API, Interactivity, slash commands, and Block Kit.

## Context

The current framework is Discord-oriented in naming and transport, but many concepts map cleanly to Slack:

| Framework concept | Discord today | Slack equivalent |
|---|---|---|
| Slash command | Discord application command interaction | Slack slash command request (`/name`) or app mention/event command parser |
| Component button | Discord component interaction custom ID | Block Kit `button` element `action_id` / `value` in `block_actions` payload |
| Modal | Discord modal submit | Slack modal `views.open`, `view_submission` payload |
| Message response | Interaction response/edit | `chat.postMessage`, `chat.update`, `response_url`, ephemeral responses |
| Channel/user/team IDs | Discord guild/channel/user | Slack team/channel/user IDs |
| Bot token | Discord bot token | Slack bot token (`xoxb-...`) with scopes |
| Event dispatch | Discord gateway events | Slack Events API over HTTP or Socket Mode |

Primary design target: keep the JavaScript API transport-neutral. If JS calls `ui.message().content(...).row(ui.button(...))`, Slack adapter should produce Slack Block Kit, not require JS to call Slack-specific APIs.

## Quick Reference

### Slack platform surfaces

Slack apps interact with users via several surfaces:

| Surface | API/event | Use in this project |
|---|---|---|
| Messages | `chat.postMessage`, `chat.update`, `chat.delete` | Main bot responses and editable adventure/game screens |
| Ephemeral messages | `chat.postEphemeral`, `response_url` with `response_type: ephemeral` | Errors, private confirmations, command validation |
| Slash commands | Slash command HTTP POST | Map framework `command(...)` definitions to Slack `/command` entry points |
| Interactive components | Interactivity HTTP POST, payload type `block_actions` | Buttons/selects from `ui.button` etc. |
| Modals | `views.open`, `views.update`, payloads `view_submission`, `view_closed` | Map `ui.form(...)` and modal handlers |
| Events | Events API (`event_callback`) or Socket Mode | `ready`, message events, app mentions, future event handlers |
| App Home | `views.publish` | Optional later; not required for backend parity |

### Transport options: Events API vs Socket Mode

Slack supports two common app delivery modes:

| Mode | How it works | Pros | Cons |
|---|---|---|---|
| Events API + Interactivity HTTP endpoints | Slack sends signed HTTP POSTs to public endpoints | Production-standard, simple stateless webhooks | Requires public HTTPS URL/tunnel in dev |
| Socket Mode | App opens websocket to Slack using app-level token (`xapp-...`) | Great local/dev, no public URL required | Requires app-level token and socket client loop |

Recommended implementation path:

1. Start with HTTP endpoints if the existing service already exposes an HTTP server.
2. Consider Socket Mode as a dev/deployment option if local bot running is a strong requirement.
3. Keep event normalization identical regardless of delivery mechanism.

### Authentication and request verification

Slack request verification uses:

- `X-Slack-Signature`
- `X-Slack-Request-Timestamp`
- signing secret
- HMAC-SHA256 base string: `v0:{timestamp}:{raw_body}`
- compare with header value `v0={hex_digest}`

Implementation requirements:

- Read and preserve the **raw body** before parsing form/JSON.
- Reject old timestamps, commonly >5 minutes skew.
- Constant-time compare signatures.
- Never log bot tokens, signing secrets, or full authorization payloads.

Pseudo-code:

```go
base := "v0:" + timestamp + ":" + string(rawBody)
mac := hmac.New(sha256.New, []byte(signingSecret))
mac.Write([]byte(base))
expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
if !hmac.Equal([]byte(expected), []byte(actualHeader)) { reject }
```

### Slack tokens and scopes

Common tokens:

| Token | Prefix | Purpose |
|---|---|---|
| Bot token | `xoxb-` | Web API calls as bot |
| User token | `xoxp-` | User-scoped calls; avoid unless needed |
| App-level token | `xapp-` | Socket Mode and app-level features |

Likely bot scopes:

| Scope | Need |
|---|---|
| `commands` | Receive slash command invocations |
| `chat:write` | Post/update bot messages |
| `chat:write.public` | Post to public channels without joining, if desired |
| `channels:read` | Public channel metadata, if needed |
| `groups:read` | Private channel metadata, if needed |
| `im:read`, `mpim:read` | DM/multi-person DM support, if needed |
| `users:read` | User lookup/display names, if needed |
| `files:write` | File uploads/exports, if mapping Discord attachments |
| `app_mentions:read` | App mention event support |
| `channels:history`, `groups:history`, `im:history`, `mpim:history` | Only if reading message content/history is required |

Keep minimum scopes initially: `commands`, `chat:write`, and interactivity do not themselves require a separate scope beyond app configuration and token use.

### Incoming payload types

#### Slash command payload

Slack slash commands are `application/x-www-form-urlencoded` POSTs with fields like:

```text
token=legacy-verification-token
team_id=T123
team_domain=example
channel_id=C123
channel_name=general
user_id=U123
user_name=alice
command=/adventure-start
text=prompt underwater adventure
response_url=https://hooks.slack.com/commands/...
trigger_id=123.456.abcdef
```

Important behavior:

- Slack expects an acknowledgement within ~3 seconds.
- Immediate response body can post a message, or backend can respond later via `response_url`.
- `trigger_id` is required to open a modal and expires quickly.

Adapter implication:

- Normalize to existing command `ctx` shape: command name, args parsed from `text`, team/guild ID, channel ID, user ID, response handle.
- If existing JS command expects structured slash options, Slack adapter needs an option parser for `text`, e.g. `prompt underwater adventure`, `seed haunted-gate prompt "..."`, or key/value syntax.

#### Interactivity payload: `block_actions`

Slack sends interactive component callbacks as form field `payload=<json>`. Shape excerpt:

```json
{
  "type": "block_actions",
  "user": { "id": "U123", "username": "alice" },
  "api_app_id": "A123",
  "team": { "id": "T123", "domain": "example" },
  "channel": { "id": "C123", "name": "general" },
  "message": { "ts": "1710000000.000100", "blocks": [] },
  "response_url": "https://hooks.slack.com/actions/...",
  "trigger_id": "123.456.abcdef",
  "actions": [
    {
      "action_id": "adv:choice:0",
      "block_id": "row_0",
      "text": { "type": "plain_text", "text": "Enter the reef" },
      "value": "adv:choice:0",
      "type": "button",
      "action_ts": "1710000001.000200"
    }
  ]
}
```

Adapter implication:

- Dispatch on `actions[0].action_id`, matching existing component custom IDs.
- Provide `ctx.message.content` equivalent by decoding stored message metadata or extracting text from Slack blocks.
- Preserve `response_url`, `channel.id`, and `message.ts` for edits.

#### Modal payloads: `view_submission`

Shape excerpt:

```json
{
  "type": "view_submission",
  "team": { "id": "T123" },
  "user": { "id": "U123" },
  "view": {
    "id": "V123",
    "callback_id": "adv:modal:freeform",
    "private_metadata": "{\"channel_id\":\"C123\",\"message_ts\":\"...\"}",
    "state": {
      "values": {
        "action_block": {
          "action": {
            "type": "plain_text_input",
            "value": "swim toward the glowing trench"
          }
        }
      }
    }
  }
}
```

Adapter implication:

- Dispatch on `view.callback_id` to existing modal handlers.
- Convert `view.state.values` into current `ctx.values` shape (`{ action: "..." }`).
- Use `private_metadata` to carry channel/message correlation because modal submissions do not naturally include the original message context.

### Outgoing Web API calls

Most relevant methods:

| Method | Purpose |
|---|---|
| `chat.postMessage` | Send channel message |
| `chat.update` | Edit message by `channel` + `ts` |
| `chat.postEphemeral` | Private ephemeral message to user in channel |
| `chat.delete` | Delete message |
| `views.open` | Open modal using `trigger_id` |
| `views.update` | Update modal |
| `views.publish` | App Home publish, optional |
| `files.uploadV2` / external upload flow | File export support |

Slack message identity is `channel` + `ts`, not a single snowflake message ID.

### Slack acknowledgements and delayed responses

Slack requires fast acknowledgement:

- Slash commands: respond within ~3 seconds.
- Interactive callbacks: acknowledge within ~3 seconds.
- Events API: return HTTP 200 quickly.

Recommended backend behavior:

1. Verify signature.
2. Parse payload.
3. Create internal normalized interaction context.
4. Immediately ACK if the handler may take time.
5. Run JS handler.
6. Use `response_url`, `chat.postMessage`, or `chat.update` for actual response/edit.

This is close to current Discord `defer` semantics. Implement Slack `ctx.defer()` as immediate ACK plus stored response handle.

### Block Kit basics

Slack messages are arrays of `blocks`. Common block types:

| Block | Purpose |
|---|---|
| `section` | Text with Markdown (`mrkdwn`) or plain text; optional accessory |
| `actions` | Row of buttons/selects/datepickers/etc. |
| `context` | Small metadata text/images |
| `divider` | Horizontal divider |
| `header` | Plain-text header |
| `image` | Image block |
| `input` | Modal/home input only; not normal messages |
| `rich_text` | Slack-native rich text, usually avoid generating manually initially |

Text objects:

```json
{ "type": "mrkdwn", "text": "*Hello* <@U123>" }
{ "type": "plain_text", "text": "Click me", "emoji": true }
```

Button element:

```json
{
  "type": "button",
  "action_id": "adv:choice:0",
  "text": { "type": "plain_text", "text": "Enter the reef", "emoji": true },
  "value": "adv:choice:0",
  "style": "primary"
}
```

Button styles:

- omitted/default = neutral
- `primary` = green emphasis
- `danger` = red emphasis

Limits to account for:

- Max 50 blocks per message.
- `actions` block max 25 elements, but keep much lower for parity with Discord rows.
- Button text max is limited; keep labels concise.
- Section text has a character limit; long framework content may need chunking/truncation.

### Mapping existing `ui` to Slack Block Kit

Current JS should remain like:

```js
return ui.message()
  .content("Choose your path")
  .row(
    ui.button("adv:choice:0", "Enter the reef", "primary"),
    ui.button("adv:choice:1", "Check oxygen", "secondary")
  )
  .build()
```

Slack adapter should normalize to:

```json
{
  "channel": "C123",
  "text": "Choose your path",
  "blocks": [
    {
      "type": "section",
      "text": { "type": "mrkdwn", "text": "Choose your path" }
    },
    {
      "type": "actions",
      "block_id": "row_0",
      "elements": [
        {
          "type": "button",
          "action_id": "adv:choice:0",
          "text": { "type": "plain_text", "text": "Enter the reef", "emoji": true },
          "value": "adv:choice:0",
          "style": "primary"
        },
        {
          "type": "button",
          "action_id": "adv:choice:1",
          "text": { "type": "plain_text", "text": "Check oxygen", "emoji": true },
          "value": "adv:choice:1"
        }
      ]
    }
  ]
}
```

Recommended style mapping:

| Framework style | Slack style |
|---|---|
| `primary` | `primary` |
| `danger` | `danger` |
| `secondary` | omitted |
| unknown | omitted |

### Preserving message content for stale-turn checks

Existing JS may inspect `ctx.message.content`, e.g. parsing `Turn N` from rendered text. Slack interactive payloads include `message.blocks` and sometimes top-level `message.text`, but Block Kit text may not round-trip exactly.

Options:

1. **Best**: store canonical normalized message content server-side keyed by Slack `channel + ts` when posting/updating. Populate `ctx.message.content` from that store during interactions.
2. Fallback: reconstruct text from `message.blocks` section text, but this is lossy.
3. Encode small state in button `value`, e.g. `custom_id|turn=3`, but this may require changing semantics and risks JS-visible differences.

Recommendation: implement an adapter-side message state table/cache for Slack posted messages.

### Modal mapping from framework form API

Framework form:

```js
await ctx.showModal(
  ui.form("adv:modal:freeform", "Try something else")
    .textarea("action", "What do you try?").required().min(2).max(800)
    .build()
)
```

Slack `views.open` payload:

```json
{
  "trigger_id": "123.456.abcdef",
  "view": {
    "type": "modal",
    "callback_id": "adv:modal:freeform",
    "title": { "type": "plain_text", "text": "Try something else" },
    "submit": { "type": "plain_text", "text": "Submit" },
    "close": { "type": "plain_text", "text": "Cancel" },
    "private_metadata": "{\"channel_id\":\"C123\",\"message_ts\":\"171...\"}",
    "blocks": [
      {
        "type": "input",
        "block_id": "action_block",
        "label": { "type": "plain_text", "text": "What do you try?" },
        "element": {
          "type": "plain_text_input",
          "action_id": "action",
          "multiline": true,
          "min_length": 2,
          "max_length": 800
        }
      }
    ]
  }
}
```

### Attachments/files/export mapping

Discord-style response payloads may include `files: [{ name, content, contentType }]`.

Slack options:

1. Upload a file with `files.uploadV2` or Slack's external upload flow.
2. If small JSON/text, include as code block in message thread or snippet-like upload.
3. If file upload scope is unavailable, degrade gracefully with a message saying export is unavailable.

Recommended initial approach:

- Support message content and blocks first.
- Add file support behind `files:write` scope.
- For adventure exports, upload JSON file to the channel and include message text.

### Rate limits and retries

Slack Web API methods are rate-limited per method/workspace/app. Responses may include:

- HTTP `429`
- `Retry-After` header
- JSON `{ "ok": false, "error": "ratelimited" }` in some contexts

Adapter should:

- Honor `Retry-After`.
- Queue or coalesce rapid `chat.update` calls.
- For streaming/progress edits, throttle more aggressively than Discord if needed.
- Log method, team, channel, and retry delay, but not tokens or full payload secrets.

### Error response patterns

Web API response shape:

```json
{ "ok": true, "channel": "C123", "ts": "1710000000.000100", "message": { } }
```

Error shape:

```json
{ "ok": false, "error": "channel_not_found" }
```

Common errors:

| Error | Meaning / likely fix |
|---|---|
| `invalid_auth` | Bad/missing token |
| `not_authed` | Missing auth header |
| `missing_scope` | Add OAuth scope and reinstall app |
| `channel_not_found` | Bot not in channel or wrong ID |
| `not_in_channel` | Invite bot or use `chat:write.public` for public channels |
| `invalid_arguments` | Payload malformed; inspect blocks with Block Kit Builder |
| `trigger_expired` | Modal opened too late; must call `views.open` quickly |
| `message_not_found` | Wrong `channel`/`ts` for update/delete |

### OAuth/install model

For a single-workspace/internal app, static env config may be enough:

```text
SLACK_BOT_TOKEN=xoxb-...
SLACK_SIGNING_SECRET=...
SLACK_APP_TOKEN=xapp-...   # only for Socket Mode
```

For multi-workspace SaaS:

- Implement OAuth v2 install flow.
- Store bot token per `team_id` / enterprise install.
- Select token based on incoming payload `team.id` / `team_id`.
- Handle token rotation if enabled.

Initial backend can support single-workspace env token while keeping token lookup interface ready for multi-team later.

## Backend abstraction implications

### Recommended Go package shape

Potential package layout:

```text
internal/slackbot/
  server.go             # HTTP handlers for slash, events, interactivity
  verify.go             # signing secret verification
  client.go             # Web API client wrapper
  normalize.go          # Slack payload -> framework ctx
  render.go             # normalized response -> Slack message/blocks
  commands.go           # slash command registration/docs if applicable
  state.go              # message correlation channel+ts -> normalized content/context
```

Or, if the existing Discord package should generalize first:

```text
internal/transport/
  types.go              # normalized command/component/modal/event contracts
  responses.go          # normalized response model
internal/discord/...
internal/slack/...
```

### Keep JS unchanged by normalizing at boundaries

The Slack adapter should provide these existing concepts to JS:

| Existing JS expectation | Slack adapter responsibility |
|---|---|
| `ctx.user.id` | Slack `user.id` |
| `ctx.channel.id` | Slack `channel.id` |
| `ctx.guild.id` | Slack `team.id` (alias as guild/workspace) |
| `ctx.args` | Parse slash command text into option-like object |
| `ctx.message.content` | Restore stored normalized content for `channel+ts` |
| `ctx.defer()` | ACK within 3 seconds and store response handle |
| `ctx.edit(message)` | `chat.update` if original message exists; otherwise `response_url` / post |
| `ctx.showModal(form)` | `views.open` using `trigger_id` |
| `component(customID, handler)` | Match Slack `action_id` |
| `modal(callbackID, handler)` | Match Slack `view.callback_id` |

### Response routing rules

Suggested normalized response behavior:

| Handler situation | Slack behavior |
|---|---|
| Slash command immediate public response | `response_url` or `chat.postMessage` |
| Slash command deferred then edit | post initial placeholder via response_url, then `chat.update` with stored `ts` |
| Button interaction updating same message | `chat.update(channel, message.ts)` |
| Button interaction ephemeral error | `response_url` with ephemeral response or `chat.postEphemeral` |
| Modal submission updating source message | use `private_metadata` channel/message_ts then `chat.update` |

### Command registration caveat

Discord slash commands can be programmatically registered by the bot. Slack slash commands are configured in the Slack app manifest/admin UI/API. Options are not structured like Discord options; Slack gives a raw text string.

Possible approaches:

1. Generate a Slack app manifest from JS command definitions.
2. Use one slash command, e.g. `/bot`, with subcommands mapped to JS command names.
3. Create one Slack slash command per JS command where feasible.

For "JS unchanged", option parsing is the biggest gap. Existing Discord commands with structured `options` need a Slack text parser.

Recommended parser behavior:

```text
/adventure-start prompt underwater adventure
/adventure-start prompt="underwater adventure" seed=haunted-gate
/adventure-start --prompt "underwater adventure" --seed haunted-gate
```

Normalize all into:

```json
{ "prompt": "underwater adventure", "seed": "haunted-gate" }
```

### Block Kit Builder validation

When debugging `invalid_blocks` or `invalid_arguments`, paste generated JSON into Slack Block Kit Builder:

- https://app.slack.com/block-kit-builder/

Keep adapter tests with golden JSON for common framework UI payloads:

- plain content only
- content + one row buttons
- multiple rows
- ephemeral message
- modal with textarea
- file/export response

## Usage Examples

### Example: normalized button response to Slack blocks

Input normalized response:

```json
{
  "content": "```\n╔═ Reef Gate\nTurn 0\n\nYou hover outside the glowing reef.\n\nHP: 8  OXYGEN: 6\n```",
  "components": [
    [
      { "type": "button", "customId": "adv:choice:0", "label": "Enter", "style": "primary" },
      { "type": "button", "customId": "adv:choice:1", "label": "Surface", "style": "secondary" }
    ]
  ]
}
```

Slack output:

```json
{
  "text": "Reef Gate Turn 0",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "```\n╔═ Reef Gate\nTurn 0\n\nYou hover outside the glowing reef.\n\nHP: 8  OXYGEN: 6\n```"
      }
    },
    {
      "type": "actions",
      "block_id": "row_0",
      "elements": [
        {
          "type": "button",
          "action_id": "adv:choice:0",
          "value": "adv:choice:0",
          "style": "primary",
          "text": { "type": "plain_text", "text": "Enter", "emoji": true }
        },
        {
          "type": "button",
          "action_id": "adv:choice:1",
          "value": "adv:choice:1",
          "text": { "type": "plain_text", "text": "Surface", "emoji": true }
        }
      ]
    }
  ]
}
```

### Example: Slack adapter context for component

```json
{
  "platform": "slack",
  "interactionType": "component",
  "customId": "adv:choice:0",
  "user": { "id": "U123", "username": "alice" },
  "guild": { "id": "T123", "name": "example" },
  "channel": { "id": "C123", "name": "general" },
  "message": {
    "id": "C123:1710000000.000100",
    "content": "stored normalized message content",
    "slack": { "channel": "C123", "ts": "1710000000.000100" }
  },
  "responseHandle": {
    "response_url": "https://hooks.slack.com/actions/...",
    "channel": "C123",
    "message_ts": "1710000000.000100",
    "trigger_id": "123.456.abcdef"
  }
}
```

### Example: Slack app manifest fragments

```yaml
display_information:
  name: Go Bot Framework
features:
  bot_user:
    display_name: go-bot
    always_online: false
  slash_commands:
    - command: /adventure-start
      url: https://example.com/slack/commands
      description: Start an adventure
      usage_hint: 'prompt="underwater adventure"'
      should_escape: false
oauth_config:
  scopes:
    bot:
      - commands
      - chat:write
      - files:write
settings:
  interactivity:
    is_enabled: true
    request_url: https://example.com/slack/interactivity
  event_subscriptions:
    request_url: https://example.com/slack/events
    bot_events:
      - app_mention
  socket_mode_enabled: false
```

## Resolved design questions

Initial answers are captured in the design decision doc:

1. Retain Discord naming in JavaScript; adapt Slack in Go/backend code.
2. Generate a Slack app manifest.
3. Use HTTP Events API / slash command / Interactivity endpoints, not Socket Mode initially.
4. Store Slack backend correlation/state in SQLite.
5. Support only one slash-command option initially; map the full Slack text to that option.
6. Inline files/exports into messages instead of using Slack file upload initially.

## Related

- Ticket index: `ttmp/2026/05/01/slack-backend--add-slack-backend-capability/index.md`
- Tasks: `ttmp/2026/05/01/slack-backend--add-slack-backend-capability/tasks.md`
- Slack API overview: https://api.slack.com/apis
- Slack Block Kit reference: https://api.slack.com/block-kit
- Slack Interactivity: https://api.slack.com/interactivity
