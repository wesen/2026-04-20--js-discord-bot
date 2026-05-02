---
Title: Implementation Proposal
Ticket: adventure
Status: active
Topics:
    - discord
    - game
    - adventure
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/discord-bots/ui-showcase/index.js
      Note: Reference for commands, buttons/components, and modal flows
    - Path: examples/discord-bots/unified-demo/index.js
      Note: Small bot showing configure metadata and runtime config access
    - Path: internal/jsdiscord/ui_module.go
      Note: UI builders available through require("ui")
    - Path: internal/jsdiscord/bot_compile.go
      Note: Handler context shape and dispatch request fields
ExternalSources:
    - https://openrouter.ai/docs/quickstart
Summary: "Proposal for an OpenRouter-backed, YAML-schema-driven ASCII choose-your-own-adventure Discord bot."
LastUpdated: 2026-05-01T13:19:44-07:00
WhatFor: "Use as the implementation plan for the first adventure prototype."
WhenToUse: "Before coding the adventure bot, schema, LLM prompt contract, or persistence layer."
---

# Implementation Proposal

## Goal

Build a Discord choose-your-own-adventure bot that renders atmospheric ASCII scenes, accepts both structured button choices and controlled free-form player input, and uses OpenRouter-backed LLM calls to generate structured scene patches that are persisted in SQLite while the bot engine remains the source of truth.

Default LLM provider/model:

- Provider: **OpenRouter**
- Default model: **Anthropic Claude 3.5 Haiku** via OpenRouter, e.g. `anthropic/claude-3.5-haiku`
- Runtime config should allow overriding the model without code changes.

## Core Principles

1. **SQLite is the game contract**
   - Adventures, scenes, turns, choices, effects, interpreted free-form actions, and audit records are represented as SQLite rows and JSON columns managed through `require("database")`.
   - LLM output is parsed as structured JSON by the Go-owned LLM layer or JS adapter, validated, normalized, and inserted into SQLite before affecting state.

2. **The engine owns canonical state**
   - The LLM may propose narration, ASCII art, choices, flags, and effects.
   - The engine applies only valid effects to the canonical session state.
   - Inventory, stats, flags, scene ID, turn count, player/session ownership, and audit history live outside the LLM.

3. **Hybrid interaction model**
   - Buttons are the primary path for common choices.
   - A “Try something else…” button opens a Discord modal for free-form input.
   - Free-form text is interpreted by the LLM into a structured action proposal, then validated.

4. **Cheap by default, configurable later**
   - Use OpenRouter with Haiku as the default low-cost model.
   - Expose model, max tokens, temperature, and API key config through runtime config / environment.

5. **JavaScript owns game orchestration; Go owns LLM access**
   - First implementation should be a repo-discovered JS bot using existing framework APIs for Discord commands, buttons, modals, rendering, and persistence.
   - OpenRouter calls should be implemented in Go and exposed to JS as a narrow host function/module.
   - Go, not JS, controls provider settings such as model, max tokens, base URL, and token/API key. JS supplies only the prompt payload and receives structured text/results.

## Proposed File Layout

```text
examples/discord-bots/adventure/
  index.js                  # defineBot entrypoint
  README.md                 # run instructions
  lib/
    engine.js               # state machine, validation/clamping, render helpers
    llm.js                  # JS adapter around Go-owned LLM module
    schema.js               # validators / normalize helpers for scene patch objects
    store.js                # SQLite persistence via require("database")
    render.js               # ASCII scene card rendering
    prompts.js              # system/developer prompts for structured scene generation
    seeds.js                # seed adventure data as JS objects, inserted into SQLite
```

## Discord Bot Surface

Prefer grouped slash commands if framework support is comfortable; otherwise use flat names for the prototype.

### Commands

| Command | Purpose |
|---|---|
| `/adventure-start` | Start a new adventure session. Options: adventure seed, mode, visibility. |
| `/adventure-resume` | Resume an existing session. |
| `/adventure-state` | Debug/admin view of canonical YAML/JSON state. Ephemeral by default. |
| `/adventure-reset` | Clear active session for the user/channel. |

Possible later grouped shape:

- `/adventure start`
- `/adventure resume`
- `/adventure state`
- `/adventure reset`

### Components

Custom IDs should encode a compact action:

```text
adv:choice:<sessionId>:<turn>:<choiceId>
adv:freeform:<sessionId>:<turn>
adv:resume:<sessionId>
```

The handler must reject stale interactions when the turn in the custom ID does not match canonical session state.

### Modals

The free-form modal should collect:

- `action`: required text, e.g. “I whisper the name carved on the key.”
- optional later fields: tone, target, spoken words, etc.

Modal custom ID:

```text
adv:modal:freeform:<sessionId>:<turn>
```

## Runtime Configuration

Use `configure({ run: ... })` metadata and `ctx.config` so botcli/framework can pass values in.

Suggested config keys:

| Key | Default | Purpose |
|---|---:|---|
| `session_db_path` | `./examples/discord-bots/adventure/data/adventure.sqlite` | SQLite persistence path used by JS `require("database")`. |
| `debug_yaml` | `false` | Include generated YAML in ephemeral/debug output. |

LLM provider configuration should **not** be exposed to JS through `ctx.config`. Go should own:

| Go-owned setting | Default | Purpose |
|---|---:|---|
| `OPENROUTER_API_KEY` / secret config | required for real LLM mode | OpenRouter API key/token. |
| OpenRouter base URL | `https://openrouter.ai/api/v1` | API base. |
| Model | `anthropic/claude-3.5-haiku` | Cheap default model. |
| Temperature | `0.7` | Generation creativity. |
| Max tokens | `1200` | Cost/latency cap. |

OpenRouter request headers should be assembled in Go and should include:

```http
Authorization: Bearer $OPENROUTER_API_KEY
Content-Type: application/json
HTTP-Referer: <optional project URL>
X-Title: discord-bot-adventure
```

## Go-Owned LLM Host Module

Because the JS runtime has no `fetch`, `http`, `process`, `fs`, or npm package access, real OpenRouter integration belongs in Go. Expose a narrow JS module, for example `require("llm")` or `require("adventure_llm")`.

Suggested JS-facing shape:

```js
const llm = require("adventure_llm")

const result = await llm.completeYaml({
  purpose: "scene_patch",
  system: prompts.sceneSystemPrompt(),
  user: prompts.sceneUserPrompt({ seed, session, input, recentHistory }),
})
```

The JS caller may provide:

- `purpose`: bounded enum such as `scene_patch` or `interpret_action` for logging/policy.
- `system`: prompt text.
- `user`: prompt text or structured prompt payload serialized by JS.
- optional lightweight metadata: session ID, turn, adventure ID.

The JS caller must **not** provide:

- model name;
- max tokens;
- API key/token;
- OpenRouter base URL;
- arbitrary headers;
- provider selection.

The Go module should return a simple result:

```js
{
  ok: true,
  text: "scene_patch:\n  scene:\n    ...",
  provider: "openrouter",
  usage: { promptTokens: 123, completionTokens: 456, totalTokens: 579 }
}
```

Errors should be safe for Discord display/logging:

```js
{
  ok: false,
  error: "LLM request failed",
  retryable: true
}
```

## Data Model Sketch

### SQLite tables

Use `require("database")` from JS. Store structured fields in columns where useful and flexible objects as JSON text.

```sql
CREATE TABLE IF NOT EXISTS adventure_seeds (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  genre TEXT NOT NULL DEFAULT '',
  tone TEXT NOT NULL DEFAULT '',
  initial_stats_json TEXT NOT NULL DEFAULT '{}',
  inventory_vocab_json TEXT NOT NULL DEFAULT '[]',
  flag_vocab_json TEXT NOT NULL DEFAULT '[]',
  constraints_json TEXT NOT NULL DEFAULT '{}',
  opening_prompt TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS adventure_sessions (
  id TEXT PRIMARY KEY,
  adventure_id TEXT NOT NULL,
  owner_user_id TEXT NOT NULL,
  guild_id TEXT NOT NULL DEFAULT '',
  channel_id TEXT NOT NULL DEFAULT '',
  thread_id TEXT NOT NULL DEFAULT '',
  mode TEXT NOT NULL DEFAULT 'solo',
  turn INTEGER NOT NULL DEFAULT 0,
  current_scene_id TEXT NOT NULL DEFAULT '',
  stats_json TEXT NOT NULL DEFAULT '{}',
  inventory_json TEXT NOT NULL DEFAULT '[]',
  flags_json TEXT NOT NULL DEFAULT '{}',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS adventure_scenes (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  turn INTEGER NOT NULL,
  title TEXT NOT NULL,
  ascii_art TEXT NOT NULL DEFAULT '',
  narration TEXT NOT NULL DEFAULT '',
  engine_notes_json TEXT NOT NULL DEFAULT '{}',
  raw_patch_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS adventure_choices (
  id TEXT PRIMARY KEY,
  scene_id TEXT NOT NULL,
  choice_id TEXT NOT NULL,
  label TEXT NOT NULL,
  requires_json TEXT NOT NULL DEFAULT '{}',
  proposed_effects_json TEXT NOT NULL DEFAULT '{}',
  next_hint TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  UNIQUE(scene_id, choice_id)
);

CREATE TABLE IF NOT EXISTS adventure_audit (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  turn INTEGER NOT NULL,
  kind TEXT NOT NULL,
  input_json TEXT NOT NULL DEFAULT '{}',
  llm_request_json TEXT NOT NULL DEFAULT '{}',
  llm_response_text TEXT NOT NULL DEFAULT '',
  parsed_json TEXT NOT NULL DEFAULT '{}',
  validation_json TEXT NOT NULL DEFAULT '{}',
  applied_effects_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);
```

### Canonical session state

Canonical state lives in `adventure_sessions`, with JSON columns for flexible state:

```json
{
  "id": "adv_abc123",
  "adventure_id": "haunted-gate",
  "owner_user_id": "123",
  "guild_id": "456",
  "channel_id": "789",
  "mode": "solo",
  "turn": 4,
  "current_scene_id": "gate-whispers-004",
  "stats": { "hp": 8, "sanity": 6 },
  "inventory": ["iron_key"],
  "flags": { "opened_gate": true, "spirit_curious": true }
}
```

### LLM scene patch contract

The Go-owned LLM module should ask OpenRouter for JSON, not YAML. JS validates the returned object and persists scenes/choices into SQLite.

```json
{
  "scene_patch": {
    "scene": {
      "id": "gate-whispers-004",
      "title": "The Whispering Gate",
      "ascii_art": ".--.\n|  |\n|__|",
      "narration": "The gate answers in a voice like wet stone.",
      "choices": [
        {
          "id": "ask_name",
          "label": "Ask the gate its name",
          "next_hint": "dialogue"
        },
        {
          "id": "use_key",
          "label": "Use the iron key",
          "requires": { "inventory": ["iron_key"] },
          "proposed_effects": { "flags": { "attempted_key": true } }
        }
      ]
    },
    "engine_notes": {
      "mood": "eerie, restrained",
      "continuity": "The spirit is curious, not hostile yet."
    }
  }
}
```

### Free-form interpretation contract

For modal text, ask the LLM to interpret the player action as JSON:

```json
{
  "interpreted_action": {
    "summary": "Player tries to respectfully parley with the gate spirit.",
    "kind": "dialogue",
    "target": "gate_spirit",
    "risk": "low",
    "matched_choice_id": "ask_name",
    "proposed_effects": { "flags": { "spirit_respected": true } },
    "response_hint": "The spirit should reward politeness with a clue."
  }
}
```

Then request/generate the next scene patch using the interpreted action plus canonical state.

## Validation and Guardrails

Minimum prototype validation:

- Parse JSON safely.
- Require top-level `scene_patch.scene` or `interpreted_action`.
- Require scene title, narration, and 2–4 choices.
- Limit ASCII art and narration length for Discord message size.
- Allow only safe stat deltas, e.g. `-3 <= hp delta <= +3` per turn.
- Reject unknown inventory mutations unless defined in seed adventure or generated under an allowed namespace.
- Reject stale component clicks by matching session ID and turn.
- Never let LLM change owner, guild/channel, turn count, mode, or session ID.
- Store raw LLM response, parsed JSON, validation result, and applied patch in `adventure_audit`.

## Prompting Strategy

Use a stable system prompt that says:

- You are generating content for a Discord ASCII choose-your-own-adventure game.
- Return only JSON matching the requested schema.
- The engine owns canonical state; do not claim state changes outside proposed effects.
- Keep scenes short enough for Discord.
- Offer 2–4 concrete choices.
- Include one optional free-form affordance when useful.
- Maintain continuity with the provided history and seed constraints.

For each turn, provide:

- seed adventure definition;
- canonical session state;
- recent history summary;
- player input / choice;
- current scene;
- validation constraints.

## First Milestone: Playable Prototype

Deliverables:

1. Add `examples/discord-bots/adventure/index.js`.
2. Add a minimal seed adventure in `lib/seeds.js` and seed it into SQLite.
3. Implement SQLite-backed session/store module using `require("database")`.
4. Implement Go-owned LLM host module for OpenRouter and expose a narrow JS function such as `require("adventure_llm").completeYaml(...)`.
5. Implement JS LLM adapter that calls the Go host module, with mock fallback only for tests/dev when the host module is unavailable.
6. Implement `/adventure-start`.
6. Render generated ASCII scene card with buttons.
7. Implement choice component handler.
8. Implement “Try something else…” modal flow.
9. Persist audit log per session in `adventure_audit`.
10. Add README instructions for running with `OPENROUTER_API_KEY`.

## Second Milestone: Hardening

- Move schema validation to a clearer validator layer.
- Add command autocomplete for seed adventures and sessions.
- Add group/voting mode.
- Add thread creation mode.
- Add admin/debug export of session/scene JSON from SQLite.
- Add tests for validator, renderer, custom ID parsing, stale turn rejection, and OpenRouter response parsing.
- Consider expanding Go-native modules only where needed:
  - keep OpenRouter HTTP client in Go permanently;
  - optionally move robust YAML parsing/validation to Go if JS validation becomes brittle;
  - keep durable SQLite storage in JS via `require("database")` unless stronger transactional semantics are needed;
  - optionally move deterministic state transition enforcement to Go later.

## Open Questions

- Should the Go LLM module be generic (`require("llm")`) or adventure-specific (`require("adventure_llm")`)?
- Which JSON fields should remain flexible JSON columns versus normalized SQLite columns?
- What exact default model slug should we pin after checking current OpenRouter model names and pricing?
- Should the first UI be flat slash commands or grouped subcommands?
