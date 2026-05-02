---
Title: Diary
Ticket: adventure
Status: active
Topics:
    - discord
    - game
    - adventure
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/discord-bots/adventure/README.md
      Note: Run instructions for adventure bot (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/index.js
      Note: Adventure bot command/component/modal entrypoint (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/lib/engine.js
      Note: Adventure state advancement and LLM orchestration (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/lib/render.js
      Note: ASCII scene card renderer and UI payloads (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/lib/schema.js
      Note: JSON parsing and validation for LLM contracts (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/lib/store.js
      Note: SQLite store for seeds
    - Path: internal/jsdiscord/host.go
      Note: Registers OpenRouter module in default JS bot host (commit f9a47bf)
    - Path: internal/jsdiscord/openrouter_module.go
      Note: Go-owned OpenRouter adapter exposed to JS as adventure_llm.completeJson (commit f9a47bf)
    - Path: internal/jsdiscord/openrouter_module_test.go
      Note: Focused tests for OpenRouter adapter request/response behavior (commit f9a47bf)
    - Path: pkg/botcli/runtime_factory.go
      Note: Registers OpenRouter module in botcli-created runtimes (commit f9a47bf)
ExternalSources: []
Summary: Implementation diary for the adventure Discord bot milestone.
LastUpdated: 2026-05-01T13:33:30-07:00
WhatFor: Track implementation steps, commits, validation commands, failures, and review notes for the adventure ticket.
WhenToUse: Use when reviewing or continuing the adventure implementation.
---



# Diary

## Goal

Capture the implementation of Milestone 1 for the adventure Discord bot: a Go-owned OpenRouter adapter plus a JS framework bot that uses SQLite-defined scenes, Discord buttons/modals, and structured JSON LLM contracts.

## Step 1: Add Go-Owned OpenRouter Adapter

This step implemented the narrow host-side LLM adapter that JavaScript can call without controlling provider details. The new `adventure_llm` module exposes `completeJson(input)`, while Go/process configuration owns the OpenRouter API key, base URL, model, max tokens, temperature, and request headers.

The implementation is intentionally small: JS supplies prompt text and lightweight metadata, and Go returns a safe `{ ok, text, provider, usage }` result or a safe error object. This establishes the boundary requested before building the JS adventure app.

### Prompt Context

**User prompt (verbatim):** "OK looks good lets try it. Build openrouter adapter first, then the JS app. Keep a diary and commit as you go"

**Assistant interpretation:** Implement the OpenRouter host adapter as the first focused commit, then continue toward the JS app, while recording a docmgr diary and committing incremental work.

**Inferred user intent:** Keep provider secrets/model controls in Go, unblock JS adventure development with a narrow LLM function, and maintain an auditable implementation trail.

**Commit (code):** f9a47bfb59258055389b7ab1a5d71d8b7183dd2e — "Add OpenRouter LLM module"

### What I did

- Added `internal/jsdiscord/openrouter_module.go`.
- Registered `require("adventure_llm")` in default JS bot hosts via `internal/jsdiscord/host.go`.
- Registered the same module in `pkg/botcli/runtime_factory.go` for botcli-created runtimes.
- Added `internal/jsdiscord/openrouter_module_test.go` with an `httptest` OpenRouter-compatible server.
- Verified JS cannot override model/max tokens through the call payload; tests assert the request uses Go/env-owned values.

### Why

- The JS runtime has no `fetch`, `http`, `process`, npm packages, or file access.
- OpenRouter calls require HTTP and secret-bearing headers, which should stay in Go.
- The adventure JS app needs a simple LLM boundary before implementing prompts and scene generation.

### What worked

- `go test ./internal/jsdiscord -run TestOpenRouter -count=1` passed.
- The adapter returns safe, Discord-displayable error objects when the API key is missing.
- The adapter sends OpenRouter chat completion requests to `/chat/completions` with `Authorization`, `Content-Type`, and `X-Title` headers.

### What didn't work

- Broader test commands exposed pre-existing/ambient failures around example bots requiring modules:

```text
go test ./internal/jsdiscord -count=1
--- FAIL: TestKnowledgeBaseBotUsesSQLiteStoreForCaptureSearchAndReview (0.00s)
    knowledge_base_runtime_test.go:14: require bot script: GoError: Invalid module at github.com/dop251/goja_nodejs/require.(*RequireModule).require-fm (native)
...
FAIL github.com/go-go-golems/discord-bot/internal/jsdiscord
```

```text
go test ./pkg/framework ./pkg/botcli ./cmd/discord-bot -count=1
ok   github.com/go-go-golems/discord-bot/pkg/framework 0.302s
--- FAIL: TestNewBotsCommandExposesRunHelp (0.01s)
    command_test.go:26: Received unexpected error:
    discover bots: inspect bot script /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/knowledge-base/index.js: load js bot script: GoError: Invalid module at github.com/dop251/goja_nodejs/require.(*RequireModule).require-fm (native)
FAIL github.com/go-go-golems/discord-bot/pkg/botcli
FAIL github.com/go-go-golems/discord-bot/cmd/discord-bot
```

The focused OpenRouter tests pass; the broader failures appear tied to existing module loading for example bots rather than the OpenRouter adapter itself.

### What I learned

- The JS framework can consume Go-native modules through `engine.RuntimeModuleRegistrar` and `reg.RegisterNativeModule(...)`.
- Returning a plain object from Go is enough for JS `await` usage, even though the current implementation is synchronous/blocking inside the runtime call.
- Keeping the module result shape simple will make the JS app easier to test and degrade gracefully.

### What was tricky to build

- The API boundary needed to avoid exposing model/token/provider knobs to JS while still allowing tests to control Go-owned settings. I used process environment variables for Go-owned configuration so the settings remain outside `ctx.config` and outside the JS call payload.
- The adapter returns safe errors instead of throwing, because Discord handlers can display or log those errors without tripping the interaction response lifecycle.

### What warrants a second pair of eyes

- Whether the adapter should be synchronous as implemented or should return a Goja promise / run HTTP off the runtime owner goroutine.
- Whether environment variable names are the right Go-owned configuration surface.
- Whether `OPENROUTER_MODEL` should remain overrideable by process env or be hard-pinned to Haiku for the prototype.
- Whether `adventure_llm` should become a generic `llm` module later.

### What should be done in the future

- Add JS app integration using `require("adventure_llm").completeJson(...)`.
- Consider adding JSON-mode / provider-specific response format options once OpenRouter model compatibility is verified.
- Add metrics/logging around latency and token usage.

### Code review instructions

- Start with `internal/jsdiscord/openrouter_module.go`, especially `OpenRouterRegistrar`, `openRouterLoader`, and `(*openRouterClient).complete`.
- Check default registration in `internal/jsdiscord/host.go` and `pkg/botcli/runtime_factory.go`.
- Validate with:

```bash
go test ./internal/jsdiscord -run TestOpenRouter -count=1
```

### Technical details

JS-facing call shape:

```js
const llm = require("adventure_llm")
const result = await llm.completeJson({
  purpose: "scene_patch",
  system: "Return JSON only.",
  user: "Generate the next scene.",
  metadata: { sessionId, turn }
})
```

Go-owned configuration:

- `OPENROUTER_API_KEY`
- `OPENROUTER_BASE_URL` defaulting to `https://openrouter.ai/api/v1`
- `OPENROUTER_MODEL` defaulting to `anthropic/claude-3.5-haiku`
- `OPENROUTER_MAX_TOKENS` defaulting to `1200`
- `OPENROUTER_TEMPERATURE` defaulting to `0.7`
- `OPENROUTER_HTTP_REFERER`
- `OPENROUTER_APP_TITLE` defaulting to `discord-bot-adventure`

## Step 2: Add JS Adventure Bot Prototype

This step built the first playable JavaScript adventure bot on top of the existing Discord framework. The bot is repo-discovered, declares slash commands/components/modals, stores seeds/sessions/scenes/choices/audit rows in SQLite, and calls the Go-owned `adventure_llm` module for JSON scene generation and free-form action interpretation.

The prototype keeps dynamic component IDs intentionally simple because current component dispatch is exact-match based. To still protect against stale clicks, rendered scene content includes the turn number and component handlers compare the clicked message's turn against the canonical session turn before advancing.

### Prompt Context

**User prompt (verbatim):** "I updated the underlying code. Try again"

**Assistant interpretation:** Re-run validation after framework dependency/code updates, then continue implementing the JS adventure app.

**Inferred user intent:** Confirm the framework now supports the needed modules and complete the first JS milestone after the Go OpenRouter adapter.

**Commit (code):** 3feb8efcc0adbfdfac3751b93b82412907e89530 — "Add adventure Discord bot prototype"

### What I did

- Added `examples/discord-bots/adventure/index.js` as a repo-discovered `defineBot` entrypoint.
- Added `lib/seeds.js` with the `haunted-gate` seed as JS data.
- Added `lib/store.js` with SQLite schema and persistence for seeds, sessions, scenes, choices, and audit logs.
- Added `lib/schema.js` with JSON extraction/parsing and validators for scene patches and interpreted free-form actions.
- Added `lib/prompts.js` for scene and free-form interpretation prompts.
- Added `lib/llm.js` as the JS adapter around `require("adventure_llm").completeJson(...)`.
- Added `lib/engine.js` for session advancement, effect clamping/application, LLM orchestration, and audit writes.
- Added `lib/render.js` for ASCII scene cards, buttons, state output, and errors.
- Added `examples/discord-bots/adventure/README.md` with run instructions and OpenRouter configuration notes.

### Why

- The ticket design moved from YAML files to SQLite-defined scenes because the JS runtime has database support but no filesystem/YAML implementation.
- JS can own Discord orchestration and SQLite state while Go owns OpenRouter/provider access.
- A small prototype validates the full command/button/modal loop before hardening.

### What worked

- Bot discovery/help now succeeds:

```bash
go run ./cmd/discord-bot bots help adventure --bot-repository ./examples/discord-bots --output json
```

- Focused OpenRouter tests still pass:

```bash
go test ./internal/jsdiscord -run TestOpenRouter -count=1
```

- Broader tests pass after the underlying framework update:

```bash
go test ./pkg/framework ./pkg/botcli ./cmd/discord-bot -count=1
go test ./... -count=1
```

### What didn't work

- Before the user's framework update, bot discovery failed with:

```text
error building command: discover bots: inspect bot script /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/index.js: load js bot script: GoError: Invalid module at github.com/dop251/goja_nodejs/require.(*RequireModule).require-fm (native)
exit status 1
```

- After the update, the same discovery command succeeded.

### What I learned

- The current component registration model is exact-match, so the prototype uses fixed IDs like `adv:choice:0` rather than dynamic `session:turn` IDs.
- Stale click protection can still be approximated by parsing the turn number from the clicked message content and comparing it with canonical SQLite session state.
- `require("database")` is enough for a durable JS-owned store with JSON columns.

### What was tricky to build

- The LLM output needs to be treated as untrusted text even though the prompt asks for JSON. `schema.js` extracts fenced or raw JSON, parses it, normalizes choices, clamps lengths, and returns validation errors instead of throwing.
- Discord interaction response flow differs between slash commands, components, and modals. The bot uses `ctx.defer`/`ctx.edit` around LLM work and direct returns for lightweight state/error responses.
- Since JS does not receive the model/provider settings, `lib/llm.js` only handles request/result shape and never attempts to configure OpenRouter.

### What warrants a second pair of eyes

- Stale component handling: parsing `Turn N` from message content is pragmatic but brittle. A framework-level wildcard/dynamic component routing feature would be cleaner.
- The OpenRouter Go call is synchronous from JS's perspective; long LLM latency could block the runtime owner.
- SQLite schema uses JSON text columns heavily; this is flexible but less queryable than fully normalized state.
- The first prototype has no automated JS-level tests yet.

### What should be done in the future

- Add runtime tests for the adventure bot dispatch flow using mocked `adventure_llm` responses.
- Consider wildcard component/modal routing or custom ID prefixes in the framework.
- Add explicit JSON response-format support in the OpenRouter adapter if supported by the selected model/provider.
- Add group/voting mode and thread-based sessions.

### Code review instructions

- Start with `examples/discord-bots/adventure/index.js` to understand the command/component/modal flow.
- Review `examples/discord-bots/adventure/lib/store.js` for persistence/schema correctness.
- Review `examples/discord-bots/adventure/lib/engine.js` and `lib/schema.js` for validation and state transition safety.
- Validate with:

```bash
go run ./cmd/discord-bot bots help adventure --bot-repository ./examples/discord-bots --output json
go test ./... -count=1
```

### Technical details

Core runtime command:

```bash
export DISCORD_BOT_TOKEN=...
export DISCORD_APPLICATION_ID=...
export DISCORD_GUILD_ID=...
export OPENROUTER_API_KEY=...

GOWORK=off go run ./cmd/discord-bot bots adventure run \
  --bot-repository ./examples/discord-bots \
  --sync-on-start \
  --session-db-path ./examples/discord-bots/adventure/data/adventure.sqlite
```

Primary files:

- `examples/discord-bots/adventure/index.js`
- `examples/discord-bots/adventure/lib/store.js`
- `examples/discord-bots/adventure/lib/engine.js`
- `examples/discord-bots/adventure/lib/schema.js`
- `examples/discord-bots/adventure/lib/render.js`

## Step 7: Replace Download Export with Coda and Lookback

This step changed the ending experience so the bot no longer attaches a downloadable JSON export when an adventure ends. Instead, the final message now includes an in-message coda and a concise turn-by-turn lookback, with navigation buttons to scroll through prior scenes.

The change is transport-friendly: Discord and Slack both render normal message content and buttons, while Slack did not need file-upload support. Completed sessions can now still serve history navigation because history/resume lookup includes completed sessions, not only active ones.

### Prompt Context

**User prompt (verbatim):** "For both slack and discord, instead of a json file that can be downloadable, create a coda and a scrollable look back through the adventure."

**Assistant interpretation:** Replace final JSON file export with a message-native ending summary and history navigation that works across Discord and Slack.

**Inferred user intent:** Make the ending UX readable and interactive in both transports without relying on attachments/downloads.

**Commit (code):** 28ee3c210358d4fd329c157fa71dd3aca840e970 — "Replace adventure export with coda"

### What I did

- Updated final scene rendering in `examples/discord-bots/adventure/lib/render.js`:
  - removed `files: [{ adventure-...json }]`,
  - added a `Coda` section from the ending summary,
  - added a `Look back` section listing scene turns/titles,
  - added a `← Previous` button on final scenes.
- Updated `examples/discord-bots/adventure/lib/store.js` with `findLatestSessionInChannel`, including active and completed sessions.
- Updated `examples/discord-bots/adventure/index.js` so history navigation and `/adventure-resume` can load completed sessions.

### Why

- Slack file upload was intentionally deferred, and inline/downloadable JSON is not a great end-user coda.
- The adventure already stores scenes and snapshots, so a scrollable history is more useful than a raw export.
- Completed adventures previously stopped being discoverable by history handlers because `finishSession` changes status to `completed`.

### What worked

Validation passed:

```bash
go test ./internal/jsdiscord -run TestAdventure -count=1
go test ./... -count=1
```

### What didn't work

- No command failures in this step. The main issue was behavioral: final sessions became `completed`, so history buttons needed completed-session lookup to remain useful.

### What I learned

- Ending UX should use the same primitive set as normal gameplay—message content and buttons—so it works across Slack and Discord.
- Final-session history requires lookup semantics distinct from gameplay actions: choices should require active sessions, but history can read completed sessions.

### What was tricky to build

- Avoiding accidental action availability on completed sessions. Only resume/history were changed to allow completed sessions; choice/freeform flows still require active sessions.
- Keeping the final coda short enough for Discord/Slack message limits while still useful.

### What warrants a second pair of eyes

- Confirm whether `/adventure-resume` should show completed sessions when no active session exists, or whether there should be a distinct `/adventure-history` command later.
- Review final coda truncation lengths for Slack Block Kit section limits and Discord message limits.

### What should be done in the future

- Add a focused test for final scene rendering without files and with history navigation.
- Consider richer lookback pages, e.g. a dedicated summary screen separate from scene-by-scene history.

### Code review instructions

- Review `sceneMessage` and `codaContent` in `examples/discord-bots/adventure/lib/render.js`.
- Review `findLatestSessionInChannel` in `examples/discord-bots/adventure/lib/store.js`.
- Review `showHistory` and `/adventure-resume` lookup changes in `examples/discord-bots/adventure/index.js`.
- Validate with:

```bash
go test ./internal/jsdiscord -run TestAdventure -count=1
go test ./... -count=1
```

### Technical details

Final scene responses now return a normal UI message rather than a plain object with `files`, so both transports render the same coda/history affordance.

## Step 8: Slack Backfill for Old Export Endings

This step added a focused maintenance command to update old Slack adventure ending messages that still contain the previous JSON export representation. The command reads completed sessions from the adventure SQLite database, finds matching Slack messages in the Slack adapter SQLite database, and rewrites those messages with the new coda/lookback format.

The command is dry-run by default. Applying changes requires `--apply`, which calls Slack `chat.update` and then updates the stored Slack message content so future history button clicks still have the right message text context.

### Prompt Context

**User prompt (verbatim):** "Can we do a thing where we modify old adventure threads that have the export and change them to the code? In slack in particular"

**Assistant interpretation:** Add a Slack-specific backfill tool to replace old final adventure export messages with the new coda/lookback ending.

**Inferred user intent:** The user has existing Slack adventure messages with the old export UX and wants to migrate them in place instead of leaving stale endings around.

**Commit (code):** be743fa7514d36823548e138c378eac907479a51 — "Add Slack adventure coda backfill"

### What I did

- Added top-level command `slack-adventure-coda-backfill`.
- The command accepts:
  - `--adventure-db`
  - `--slack-state-db`
  - `--slack-bot-token`
  - `--apply`
- It scans completed adventure sessions and final scenes.
- It finds Slack messages whose stored content appears to contain the old export/session ID.
- It renders a coda plus lookback and a `← Previous` Block Kit button.
- In dry-run mode it prints matched messages without changing Slack.
- In apply mode it calls Slack `chat.update` and updates local `slack_messages.content`.

### Why

- The new final coda UX only affects new endings; existing Slack messages are already posted.
- Slack message updates require `channel + ts`, which the Slack adapter stores in SQLite.
- A dry-run-first command is safer than automatically mutating old messages.

### What worked

Validation passed:

```bash
go test ./cmd/discord-bot -count=1
go test ./... -count=1
go run ./cmd/discord-bot slack-adventure-coda-backfill --help
```

### What didn't work

- No command failures after adding the local helper for first-non-empty strings. Initial compile failed because the command tried to use a helper that existed in another package, not `cmd/discord-bot`:

```text
cmd/discord-bot/slack_adventure_coda.go:178:54: undefined: firstNonEmpty
cmd/discord-bot/slack_adventure_coda.go:192:15: undefined: firstNonEmpty
```

### What I learned

- The Slack adapter's message store makes post-hoc migration possible as long as the old message content includes enough session identity to match.
- Because old export messages included `adventure-<session>.json` or the session ID in JSON, matching can be done conservatively from stored content.

### What was tricky to build

- The command must not update arbitrary adventure messages; it searches for session/export identity in the stored Slack message content.
- The rendered message has to include `Turn N` and a history button so existing history navigation can continue from the migrated final message.

### What warrants a second pair of eyes

- Matching logic is intentionally heuristic: it looks for `adventure-<session>.json`, JSON session ID, or the raw session ID in stored Slack content. Review before running `--apply` broadly.
- The Go coda renderer duplicates some JS rendering logic; if the JS coda evolves, this command may need updates.

### What should be done in the future

- Add a dedicated migration marker/metadata field to Slack message state for easier future backfills.
- Consider a JS-owned coda renderer exposed through a maintenance command so Go does not duplicate adventure presentation logic.

### Code review instructions

- Review `cmd/discord-bot/slack_adventure_coda.go`.
- Review root command wiring in `cmd/discord-bot/root.go`.
- Validate dry-run first:

```bash
go run ./cmd/discord-bot slack-adventure-coda-backfill \
  --adventure-db ./examples/discord-bots/adventure/data/adventure.sqlite \
  --slack-state-db ./var/slack-adventure.sqlite
```

- Apply only after reviewing dry-run output:

```bash
go run ./cmd/discord-bot slack-adventure-coda-backfill \
  --adventure-db ./examples/discord-bots/adventure/data/adventure.sqlite \
  --slack-state-db ./var/slack-adventure.sqlite \
  --apply
```

### Technical details

The command requires `SLACK_BOT_TOKEN` or `--slack-bot-token` when `--apply` is used. Dry-run does not call Slack.

## Step 9: Generate Coda Storyboard Images

This step added an image-generation capability for completed adventures. When a story reaches its coda, the engine now asks the Go-owned OpenRouter module to generate one cohesive storyboard image from the entire adventure history.

The JavaScript layer still does not control provider details. Go owns the image model selection, defaulting to Gemini Flash Image, while JS only sends the storyboard prompt/story content and receives a generated image attachment or image URL.

### Prompt Context

**User prompt (verbatim):** "ok cool. Can we add a new capability which is to generate an image basd on the story for the coda? We may need to extend the LLM attachment."

**User follow-up (verbatim):** "Use gemini-2.5-flash-image, generate a storyboard from the entire story"

**Assistant interpretation:** Extend the OpenRouter host module with image generation and use it at final adventure coda time to create a storyboard from all scenes.

**Inferred user intent:** The user wants richer final adventure artifacts: a visual storyboard summarizing the whole run, generated by the backend-owned LLM/image provider path.

**Commit (code):** da9c47ff182f32bd1364811e7ae4e96f62af957b — "Generate adventure coda storyboards"

**Commit (code):** 0402dba1461200fa44300ca264f2a8ef3444b1f0 — "Use Gemini flash image model for storyboards"

### What I did

- Extended `adventure_llm` in `internal/jsdiscord/openrouter_module.go` with `generateImage(input)`.
- Added OpenRouter image request support using chat completions with image/text modalities.
- Defaulted image generation to `google/gemini-2.5-flash-image`, with override via `OPENROUTER_IMAGE_MODEL`.
- Parsed OpenRouter image responses from `message.images[0].image_url.url`.
- Added JS wrapper `llm.generateImage(...)`.
- Added adventure engine storyboard generation from the entire exported session/scenes.
- Final coda rendering now includes storyboard output:
  - data URLs are converted to base64 image file attachments,
  - non-data image URLs are rendered as a storyboard link.
- Extended normalized file payloads with `encoding: "base64"` support so generated images can become binary Discord files.
- Extended Slack response handling so image files are uploaded into the message thread via Slack `files.upload` instead of being inlined as binary text.

### Why

- The coda should be more than text; it should provide a visual memory of the completed adventure.
- The prompt asks for a storyboard from the full story, so the image prompt is built from all scene titles/narration snippets.
- Provider/model selection remains in Go to preserve the established boundary: JS requests capability, Go owns OpenRouter details.

### What worked

Validation passed:

```bash
go test ./internal/jsdiscord -run 'TestSlack|TestOpenRouter|TestAdventure' -count=1
go test ./... -count=1
```

### What didn't work

- No compile/test failures remained in this step. A design issue surfaced while implementing: Slack cannot display data URL images directly in Block Kit, so Slack image handling needed to upload generated image files rather than inline binary content.

### What I learned

- Generated image responses may come back as data URLs; the app needs to split MIME type and base64 content before attaching.
- Slack and Discord have different image paths: Discord can receive a file attachment through the existing response model, while Slack needs a follow-up file upload/thread upload for generated binary images.

### What was tricky to build

- Keeping JS provider-agnostic while still adding a new capability required extending the Go host module and adding only a thin JS wrapper.
- The coda renderer needed to return a plain response object with `files` and `components` instead of relying on `ui.message().file(...)`, because the builder's file method was not designed for base64 binary image content.
- Slack response payloads need internal `_files` stripped before calling Web API JSON endpoints, then uploaded separately after Slack returns/knows the message timestamp.

### What warrants a second pair of eyes

- Confirm the exact OpenRouter model slug `google/gemini-2.5-flash-image` is accepted in the deployed OpenRouter account. If not, set `OPENROUTER_IMAGE_MODEL` to the provider's current Gemini image slug.
- Review Slack `files.upload`; Slack has newer upload APIs, and this initial implementation uses the simpler multipart endpoint.
- Storyboard generation is synchronous at coda time and may add latency to final scene resolution.

### What should be done in the future

- Add a feature flag to disable coda image generation for low-credit/dev runs.
- Add tests with mocked OpenRouter image responses.
- Consider storing generated storyboard metadata in the adventure database so coda/history can re-render it later without regenerating.

### Code review instructions

- Review `generateImage`, `buildImageRequest`, and `generateImage` response parsing in `internal/jsdiscord/openrouter_module.go`.
- Review `generateStoryboard` in `examples/discord-bots/adventure/lib/engine.js`.
- Review final coda rendering in `examples/discord-bots/adventure/lib/render.js`.
- Review Slack file upload handling in `internal/jsdiscord/slack_backend.go`.
- Validate with:

```bash
go test ./... -count=1
```

### Technical details

Environment override:

```bash
export OPENROUTER_IMAGE_MODEL=google/gemini-2.5-flash-image
```

The storyboard prompt asks for a single wide image with 4-6 panels, no readable text/captions/UI, and uses all scene turns/titles/narration snippets from the completed session.
