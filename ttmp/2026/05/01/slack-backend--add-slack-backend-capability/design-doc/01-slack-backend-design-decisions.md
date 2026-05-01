---
Title: Slack Backend Design Decisions
Ticket: slack-backend
Status: active
Topics:
    - slack
    - backend
    - javascript
    - bot-framework
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/jsdiscord/host_dispatch.go
      Note: Existing JS-facing dispatch contracts to preserve while normalizing Slack interactions
    - Path: internal/jsdiscord/host_responses.go
      Note: Existing response/defer/edit behavior to emulate for Slack HTTP responses and Web API updates
    - Path: internal/jsdiscord/ui_components.go
      Note: Existing button/component abstraction to render as Slack Block Kit
    - Path: internal/jsdiscord/ui_form.go
      Note: Existing form/modal abstraction to render as Slack views.open modals
ExternalSources: []
Summary: 'Resolved initial Slack backend design questions: keep Discord naming in JS, generate Slack manifest, use HTTP Events/Interactivity, store adapter state in SQLite, support one command option, and inline file/export content into messages.'
LastUpdated: 2026-05-01T16:45:13.657299-07:00
WhatFor: Use as the implementation decision record for the first Slack backend iteration.
WhenToUse: Before implementing Slack transport, manifest generation, persistence, command parsing, or response/file mapping.
---


# Slack Backend Design Decisions

## Executive Summary

The first Slack backend iteration should preserve the existing JavaScript bot API exactly, including Discord-flavored naming such as `guild`, `channel`, `component`, `customId`, `ctx.defer`, `ctx.edit`, and `ctx.showModal`. Slack support should be implemented as a Go-side transport/backend adapter that normalizes Slack HTTP payloads into the same JS-facing runtime contracts and renders existing normalized UI responses into Slack Block Kit.

The resolved direction is:

1. Retain Discord naming in JavaScript.
2. Generate a Slack app manifest from bot definitions.
3. Use Slack HTTP Events API / slash command / Interactivity endpoints directly, not Socket Mode initially.
4. Move all Slack backend state/correlation into SQLite.
5. For Slack slash commands, support only one option initially.
6. For files/exports, inline content into Slack messages instead of uploading files.

## Problem Statement

The framework currently exposes a Discord-oriented JavaScript API. We want Slack as an additional backend/capability without requiring existing JS bots to branch on platform or rewrite their UI code. Slack differs from Discord in several important ways:

- Slack slash commands provide raw text rather than structured command options.
- Slack interactive messages identify messages by `channel + ts` rather than a single message ID.
- Slack modals require `trigger_id` and use `private_metadata` for correlation.
- Slack requires fast HTTP acknowledgement for slash commands, events, and interactivity.
- Slack app setup commonly happens through an app manifest, not runtime command sync exactly like Discord.
- Slack file uploads require extra scopes and API flows.

The design must make these differences disappear at the JavaScript boundary where practical.

## Proposed Solution

Build a Slack backend adapter in Go that has four core responsibilities:

1. **Ingress normalization**
   - Verify Slack request signatures.
   - Accept HTTP POSTs for slash commands, events, and interactivity.
   - Parse Slack payloads into existing JS-facing command/component/modal/event contexts.
   - Preserve Discord naming in JS context objects for compatibility.

2. **Response rendering**
   - Convert normalized message responses into Slack `chat.postMessage` / `chat.update` calls with Block Kit.
   - Convert framework buttons into Slack Block Kit button elements.
   - Convert framework forms/modals into Slack `views.open` modal views.
   - Inline file/export payloads into messages.

3. **SQLite-backed state/correlation**
   - Store Slack message mapping: normalized message ID/content/context -> Slack `channel`, `ts`.
   - Store response handles needed for deferred/edit flows.
   - Store modal `private_metadata` correlation.
   - Persist enough data that restarts do not break component/modal callbacks for recent messages.

4. **Manifest generation**
   - Generate a Slack app manifest from bot metadata/command definitions.
   - Include slash command declarations, interactivity URL, events URL, bot scopes, and display metadata.

## Design Decisions

### 1. Retain Discord naming in JavaScript

Decision: keep the JS API as-is, including Discord-oriented names.

Examples that remain valid in Slack-backed bots:

```js
ctx.guild.id      // Slack team/workspace ID
ctx.channel.id    // Slack channel ID
ctx.user.id       // Slack user ID
component("x:y", handler)
modal("x:modal", handler)
ui.button("x:y", "Click", "primary")
```

Rationale:

- Avoids platform branches in user bot scripts.
- Preserves existing examples and tests.
- Keeps this as a backend capability rather than a JS API migration.

Adapter mapping:

| JS field | Slack source |
|---|---|
| `ctx.guild.id` | `team_id` / `team.id` |
| `ctx.guild.name` | `team_domain` when available |
| `ctx.channel.id` | `channel_id` / `channel.id` |
| `ctx.user.id` | `user_id` / `user.id` |
| `ctx.message.id` | synthetic `slack:{channel}:{ts}` |
| `ctx.message.content` | SQLite-stored normalized content, fallback block text extraction |

### 2. Generate Slack app manifest

Decision: generate a Slack app manifest rather than requiring all setup to be hand-authored.

Rationale:

- Slack does not have Discord-equivalent runtime slash-command sync semantics.
- A generated manifest makes setup reproducible and reviewable.
- The manifest can be versioned or printed by CLI.

Initial manifest should include:

- `display_information.name`
- bot user display name
- slash command(s)
- interactivity request URL
- event subscription request URL
- minimal OAuth bot scopes

Open implementation choice: whether manifest generation is a CLI command, a `--print-slack-manifest` run mode, or both.

### 3. Use HTTP Events/Interactivity, not Socket Mode

Decision: first implementation uses HTTP endpoints:

- `/slack/commands`
- `/slack/interactivity`
- `/slack/events`

Rationale:

- Matches production Slack deployment model.
- Avoids app-level Socket Mode token requirement.
- Keeps operational behavior explicit: signed HTTP requests in, Web API calls out.

Implications:

- Local development will need a public HTTPS tunnel such as ngrok/cloudflared or equivalent.
- Handlers must ACK within Slack's ~3 second deadline.
- Long-running JS work should use deferred response/edit semantics.

### 4. Move all backend state into SQLite

Decision: Slack adapter correlation/state should be SQLite-backed.

Rationale:

- Slack component and modal callbacks need durable message/context correlation.
- In-memory state would break across restarts.
- Existing project already uses SQLite in JS examples; Go-side backend state can follow the same deployment assumption.

Suggested tables:

```sql
CREATE TABLE slack_messages (
  id TEXT PRIMARY KEY,              -- synthetic slack:{team}:{channel}:{ts}
  team_id TEXT NOT NULL,
  channel_id TEXT NOT NULL,
  message_ts TEXT NOT NULL,
  thread_ts TEXT,
  content TEXT NOT NULL DEFAULT '',
  response_json TEXT NOT NULL DEFAULT '{}',
  context_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(team_id, channel_id, message_ts)
);

CREATE TABLE slack_interactions (
  id TEXT PRIMARY KEY,
  team_id TEXT NOT NULL,
  channel_id TEXT,
  user_id TEXT,
  kind TEXT NOT NULL,               -- slash_command|block_actions|view_submission|event
  callback_id TEXT,
  action_id TEXT,
  trigger_id TEXT,
  response_url TEXT,
  message_ts TEXT,
  raw_payload_json TEXT NOT NULL,
  acked_at TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE slack_modal_contexts (
  id TEXT PRIMARY KEY,
  team_id TEXT NOT NULL,
  channel_id TEXT,
  message_ts TEXT,
  callback_id TEXT NOT NULL,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  expires_at TEXT
);
```

### 5. Only allow one slash command option initially

Decision: Slack command parser supports one option only for the first version.

Rationale:

- Slack slash commands provide a raw text string, unlike Discord structured options.
- Supporting arbitrary Discord-style option schemas would require a more complex parser and UX design.
- One option covers many simple bots and can unblock Slack backend work.

Mapping:

- If a JS command has zero options, ignore Slack text or pass it as empty.
- If a JS command has one option, map the full Slack `text` into that option.
- If a JS command declares more than one option, manifest generation or startup should warn/error for Slack backend.

Example:

JS command definition:

```js
command("adventure-start", {
  options: {
    prompt: { type: "string", required: false }
  }
}, handler)
```

Slack invocation:

```text
/adventure-start underwater adventure
```

Normalized args:

```json
{ "prompt": "underwater adventure" }
```

For the current adventure bot, this suggests either:

- Slack-specific manifest exposes only a single free-text `prompt`, while seed/mode use defaults.
- Or the command must be refactored later if multi-option Slack support becomes necessary.

### 6. Inline files/exports into messages

Decision: do not implement Slack file uploads initially. Put file/export content into messages.

Rationale:

- Avoids `files:write` scope and upload API complexity.
- Keeps first Slack backend focused on command, message, component, modal parity.
- Works well enough for small JSON/text exports.

Behavior:

- If normalized response contains `files`, append each file as a message section/code block.
- Truncate safely if Slack message/block limits are exceeded.
- Include filename and content type as text.

Example rendering:

```text
Export: adventure-abc.json (application/json)
```json
{ ... }
```
```

If content is too large:

```text
Export adventure-abc.json was too large to inline. Showing first 3500 characters:
```json
...
```
```

## Alternatives Considered

### Rename JS API to platform-neutral names

Rejected for first iteration. It would be cleaner long-term but violates the goal that JavaScript should not have to change.

### Socket Mode first

Rejected for first iteration. Socket Mode is attractive for local dev but adds app-level token requirements and differs from production HTTP deployment.

### In-memory Slack correlation state

Rejected. It would fail after restarts and make interactive messages brittle.

### Full Discord-style option parser for Slack text

Rejected for first iteration. Raw text parsing has many edge cases; one-option support is intentionally simple.

### Slack file uploads

Rejected for first iteration. Inline message content is simpler and avoids additional scopes/API flows.

## Implementation Plan

1. Add Slack backend package skeleton.
2. Implement Slack signature verification for raw HTTP requests.
3. Implement SQLite schema/migrations for Slack adapter state.
4. Implement HTTP routes for:
   - slash commands
   - interactivity
   - events URL verification / event callbacks
5. Normalize Slack slash commands into existing command dispatch contexts.
6. Implement one-option Slack command text mapping.
7. Normalize `block_actions.action_id` into existing component dispatch.
8. Normalize `view_submission.callback_id` and state values into existing modal dispatch.
9. Render normalized framework messages to Slack Block Kit.
10. Implement `ctx.defer`, `ctx.edit`, and `ctx.showModal` behavior for Slack.
11. Generate Slack app manifests from bot command metadata and backend URLs.
12. Add tests:
    - signature verification
    - slash command normalization
    - one-option parsing
    - button response Block Kit rendering
    - modal rendering/submission parsing
    - SQLite message correlation
    - inline file rendering/truncation
13. Add docs/playbook for creating/installing a Slack app from the generated manifest.

## Open Questions

No initial design questions remain from the reference doc. Future implementation questions:

- Exact CLI shape for manifest generation.
- Whether Slack backend should live alongside `internal/jsdiscord` initially or behind a renamed platform-neutral runtime package.
- SQLite database path/config naming for Slack backend state.
- Exact truncation policy for large inline file exports.

## References

- Slack API and Block Kit Reference: `ttmp/2026/05/01/slack-backend--add-slack-backend-capability/reference/01-slack-api-and-block-kit-reference.md`
- Ticket index: `ttmp/2026/05/01/slack-backend--add-slack-backend-capability/index.md`
