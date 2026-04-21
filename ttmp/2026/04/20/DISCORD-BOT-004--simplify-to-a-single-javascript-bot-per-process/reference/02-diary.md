---
Title: Diary
Ticket: DISCORD-BOT-004
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
    - Path: examples/discord-bots/README.md
      Note: Updated operator-facing example commands and startup-config guidance
    - Path: examples/discord-bots/knowledge-base/index.js
      Note: 'Example bot demonstrating configure({ run: ... }) and ctx.config'
    - Path: internal/botcli/command.go
      Note: Single-bot run command now resolves help and runtime config dynamically
    - Path: internal/botcli/command_test.go
      Note: Coverage for dynamic runtime config parsing and single-bot argument enforcement
    - Path: internal/botcli/run_schema.go
      Note: Two-stage bot run parsing and dynamic Glazed/Cobra runtime field parsing
    - Path: internal/botcli/runtime.go
      Note: Runtime config propagation and parsed-values output
    - Path: internal/jsdiscord/bot.go
      Note: Added ctx.config to the runtime dispatch/context contract
    - Path: internal/jsdiscord/descriptor.go
      Note: Typed parsing for bot run-schema metadata and normalized runtime field names
    - Path: internal/jsdiscord/host.go
      Note: The host implementation was later split into smaller same-package files while keeping behavior intact
    - Path: internal/jsdiscord/host_dispatch.go
      Note: Event and interaction dispatch entrypoints extracted from the former monolithic host.go
    - Path: internal/jsdiscord/host_ops_messages.go
      Note: Request-scoped Discord message operations now live in a dedicated file
    - Path: internal/jsdiscord/host_ops_channels.go
      Note: Request-scoped Discord channel operations now live in a dedicated file
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Request-scoped Discord member operations now live in a dedicated file
    - Path: internal/jsdiscord/host_payloads.go
      Note: Response and payload normalization helpers now live in a dedicated file
    - Path: internal/jsdiscord/runtime_test.go
      Note: Coverage for ctx.config visibility and runtime host operations
    - Path: examples/discord-bots/ping/index.js
      Note: Example ping bot now uses the timer module for deferred search behavior
ExternalSources: []
Summary: Chronological implementation diary for the single-bot runtime simplification and startup-config work.
LastUpdated: 2026-04-20T21:45:00-04:00
WhatFor: Capture implementation steps, exact validation commands, failures, and review guidance for DISCORD-BOT-004.
WhenToUse: Use when reviewing or resuming the DISCORD-BOT-004 implementation work.
---


# Diary

## Goal

Track the implementation work for simplifying the live runtime back to one selected JavaScript bot per process, then layering bot-level startup config on top of that cleaner model.

## Step 1: Simplify the live runner back to one selected bot

This step converted the live bot runner back to the model captured in the ticket design doc: one selected bot implementation, one JS runtime, one live Discord session. The main goal was to stop extending the multi-bot runtime path after we had already decided that composition belongs inside JavaScript rather than in the host.

The code changes were intentionally narrow. I changed the runtime entrypoints and bot CLI contract first, validated them with focused Go tests and direct `go run` checks, then updated the ticket bookkeeping so the documentation reflects the implemented cleanup.

### Prompt Context

**User prompt (verbatim):** "ok, let's do the cleanup DISCORD-BOT-004, we had a colleague already remove the support for the example bot, now all the bots are handled through `bots run`."

**Assistant interpretation:** Implement the first real code slice of DISCORD-BOT-004 by removing the live multi-bot runner behavior and aligning the runtime with the single-bot-per-process decision.

**Inferred user intent:** Reduce architectural complexity before adding more runtime features, and align the code with the intended operator-facing model.

### What I did
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go` to use one `*jsdiscord.Host` in the live path instead of `*jsdiscord.MultiHost`.
- Replaced `NewWithScripts(...)` usage with `NewWithScript(...)` in the live bot runner.
- Changed `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command.go` so `discord-bot bots run` accepts exactly one bot selector.
- Simplified `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/runtime.go` so the run request carries one bot instead of a slice of bots.
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command_test.go` to assert single-bot behavior and reject multiple bot selectors.
- Validated with:
  - `GOWORK=off go test ./internal/botcli ./internal/bot ./cmd/discord-bot ./internal/jsdiscord`
  - `GOWORK=off go test ./...`
  - `GOWORK=off go run ./cmd/discord-bot bots run knowledge-base --bot-repository ./examples/discord-bots --bot-token test-token --application-id test-app --print-parsed-values`
  - `GOWORK=off go run ./cmd/discord-bot bots run knowledge-base support --bot-repository ./examples/discord-bots --bot-token test-token --application-id test-app`

### Why
- The multi-bot host path had become architectural debt once the design decision changed.
- Keeping the live path single-bot makes the upcoming runtime-config / startup-flag work substantially simpler.

### What worked
- The bot CLI now rejects multi-bot selection cleanly.
- Parsed-values output now reflects one selected bot.
- The live runtime path no longer depends on `MultiHost`.

### What didn't work
- N/A.

### What I learned
- The cleanup touched fewer files than expected because descriptor/discovery work stayed useful. The real complexity lived at the host-runner boundary, not in discovery.

### What was tricky to build
- The main sharp edge was keeping the live runtime path single-bot without deleting discovery or experimental multi-host code prematurely. The right compromise was to remove `MultiHost` from the main runtime path while leaving the file in place until the rest of the cleanup ticket is finished.

### What warrants a second pair of eyes
- Whether `internal/jsdiscord/multihost.go` should be deleted now or retained until all startup-config work lands.
- Whether direct `discord-bot run` / `sync-commands` should gain a clearer operator message when no `--bot-script` is provided.

### What should be done in the future
- Implement bot-level startup config on top of the simplified single-bot runner.

### Code review instructions
- Start with `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go`.
- Then review `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command.go` and `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/runtime.go`.
- Validate with the exact commands listed above.

### Technical details
- The implemented public UX is now again:
  - `discord-bot bots list`
  - `discord-bot bots help <bot>`
  - `discord-bot bots run <bot>`

## Step 2: Start the bot-level run-schema and diary workflow

After the single-bot cleanup landed, I moved directly to the next intended slice: dynamic bot startup config. The goal is to let a selected bot describe startup fields under `configure({ run: ... })`, parse them after bot selection, and inject the resolved values into JavaScript as `ctx.config`.

At the same time, the user asked me to keep a diary while working, so this step also adds the ticket-local diary document and starts recording the implementation chronology before more code changes pile up.

### Prompt Context

**User prompt (verbatim):** "keep a diary as you work"

**Assistant interpretation:** Add and maintain a real implementation diary inside the DISCORD-BOT-004 ticket while continuing the active coding work.

**Inferred user intent:** Preserve implementation context, decisions, and validation steps as the ticket evolves so the work remains reviewable and resumable.

### What I did
- Added `/home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/20/DISCORD-BOT-004--simplify-to-a-single-javascript-bot-per-process/reference/02-diary.md`.
- Re-read the current single-bot runtime files and Glazed parsing APIs to prepare the run-schema implementation.
- Began extending `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor.go` so the descriptor can expose typed run-schema metadata from `configure({ run: ... })`.

### Why
- The upcoming dynamic startup-config work spans runtime metadata, parsing, and JS context injection. That is exactly the kind of work that benefits from an explicit continuation diary.

### What worked
- The ticket now has a dedicated diary doc.
- The next implementation seam is clear: descriptor parsing first, then dynamic parser wiring, then `ctx.config` injection.

### What didn't work
- N/A so far in this step.

### What I learned
- The Glazed/Cobra stack is usable here without converting the whole `bots run` command into a static Glazed command upfront. A two-stage approach is plausible: resolve the bot first, then build a dynamic parser from its run schema.

### What was tricky to build
- The main design challenge is preserving the “same medium” goal for bot startup fields even though `bots run` is currently a hand-written Cobra command. The likely solution is a small static pre-parse for bot selection plus a dynamic Glazed parser for bot-defined runtime fields.

### What warrants a second pair of eyes
- The exact metadata shape to support first: whether to support both `run.fields` and `run.sections` immediately or stage them.
- The exact field-naming convention for JavaScript keys versus CLI flag names.

### What should be done in the future
- Finish the descriptor typing for `run` metadata.
- Build and validate the dynamic parser.
- Inject resolved values as `ctx.config` in the runtime.

### Code review instructions
- Start with `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor.go` once the run-schema changes are complete.
- Then inspect the upcoming `internal/botcli` dynamic parser wiring.

### Technical details
- Target JS authoring shape:

```js
configure({
  name: "knowledge-base",
  run: {
    fields: {
      index_path: {
        type: "string",
        help: "Path to the docs index",
      },
      read_only: {
        type: "bool",
        help: "Disable writes",
      }
    }
  }
})
```

- Target runtime use:

```js
ctx.config.index_path
ctx.config.read_only
```

## Step 3: Implement dynamic run-schema parsing and inject `ctx.config`

This step finished the first end-to-end version of bot startup config for the single-bot runner. The most important architectural result is that bot-defined runtime fields now travel through a real Glazed/Cobra parser after bot selection, instead of being bolted on as one-off manual flags.

I kept the implementation intentionally staged. First I taught the descriptor layer to parse `configure({ run: ... })` metadata into typed run-schema descriptors. Then I added a small two-stage run flow in `botcli`: a light pre-parse for static runner flags and bot selection, followed by a dynamic Glazed/Cobra parser for the selected bot’s runtime fields. Finally, I plumbed the parsed runtime config into the live host so JavaScript handlers can read it through `ctx.config`.

### Prompt Context

**User prompt (verbatim):** "ok, continue"

**Assistant interpretation:** Continue DISCORD-BOT-004 after the single-bot cleanup and implement the next architectural slice rather than stopping at the runtime simplification.

**Inferred user intent:** Finish the next meaningful chunk of the single-bot design, especially the startup-config architecture that becomes possible after removing multi-bot execution.

### What I did
- Extended `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor.go` with:
  - `RunSchemaDescriptor`
  - `RunSectionDescriptor`
  - `RunFieldDescriptor`
  - parsing of `metadata.run`
  - normalization from JS field names to Glazed internal names
- Extended `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go` so dispatch requests can carry `Config`, and the JS context now exposes `ctx.config`.
- Extended `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go` so the live host stores runtime config and injects it into every command/event/interaction dispatch.
- Added `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/run_schema.go` with:
  - static pre-parsing for `bots run`
  - dynamic Glazed/Cobra schema generation from bot run metadata
  - runtime config extraction from parsed Glazed values
  - help rendering for bot run fields
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command.go` to use the new two-stage parse flow.
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/runtime.go` to pass `RuntimeConfig` through to the live bot host and include it in parsed-values output.
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/command_test.go` with single-bot runtime-config parsing coverage.
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go` to prove `ctx.config` is visible in JavaScript.
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/multihost_test.go` so descriptor inspection proves run-schema parsing works.
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js` to demonstrate `configure({ run: ... })` plus `ctx.config` use.
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md` to show the single-bot run model and example runtime flags.
- Ran:
  - `gofmt -w internal/jsdiscord/descriptor.go internal/jsdiscord/bot.go internal/jsdiscord/host.go internal/jsdiscord/runtime_test.go internal/jsdiscord/multihost_test.go internal/bot/bot.go internal/botcli/command.go internal/botcli/runtime.go internal/botcli/run_schema.go internal/botcli/command_test.go`
  - `GOWORK=off go test ./internal/botcli ./internal/jsdiscord ./internal/bot ./cmd/discord-bot`
  - `GOWORK=off go test ./...`
  - `GOWORK=off go run ./cmd/discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots`
  - `GOWORK=off go run ./cmd/discord-bot bots run knowledge-base --bot-repository ./examples/discord-bots --bot-token test-token --application-id test-app --index-path ./docs/local-index --print-parsed-values`

### Why
- The single-bot cleanup only fully pays off if the host can now parse bot-specific runtime config without multi-bot flag collisions.
- Reusing the Glazed/Cobra parser path for runtime fields preserves env/config/default behavior in the same parsing medium instead of inventing a second incompatible flag system.

### What worked
- `bots help knowledge-base` now prints a `Run config:` block.
- `bots run knowledge-base --index-path ... --print-parsed-values` now shows both the selected bot and the resolved runtime config.
- JavaScript handlers can read startup config from `ctx.config`.
- The two-stage parse model works cleanly for the current single-bot runner.

### What didn't work
- I accidentally tried to run `gofmt` on non-Go files while formatting the step:
  - `examples/discord-bots/knowledge-base/index.js:1:1: expected 'package', found 'const'`
  - `examples/discord-bots/README.md:1:1: illegal character U+0023 '#'`
- An early version of the single-bot argument test failed because the custom pre-parser treated a second bot selector as a dynamic argument instead of a hard error. The resulting failed command tried to open Discord and produced:
  - `open discord session: websocket: close 4004: Authentication failed.`
- An early help-path version also had two operator-facing issues:
  - `bots run --help` returned `bot selector is required`
  - help text showed env-backed defaults for sensitive flags because the Cobra flag defaults were still populated from environment variables
- I fixed those by allowing selector-less help in the pre-parser and by removing env-backed defaults from the displayed Cobra flag definitions.

### What I learned
- A two-stage parser is enough here. We do not need to rewrite the whole `bots run` command as a static Glazed command to get bot-specific Glazed parsing.
- Using descriptor-level normalized internal field names lets us keep cleaner CLI flags like `--index-path` while still surfacing a predictable JS config key.

### What was tricky to build
- The hardest part was preserving the “same medium” goal for startup flags while still resolving the selected bot first. Cobra wants to know all flags before execution, but bot-defined fields only exist after discovery. The solution was to split parsing into two phases:
  1. pre-parse the small fixed set of runner flags and the selected bot name
  2. build a dynamic Glazed/Cobra parser for the bot’s run schema and parse the remaining args there
- Another subtle issue was that `NewCobraParserFromSections(...)` mutates the schema by adding command-settings sections. I had to make sure runtime-config extraction iterated only the real run-schema sections rather than accidentally pulling in command-settings fields.

### What warrants a second pair of eyes
- The field-name normalization policy from JS names to CLI/internal names.
- Whether the first implementation should already support more advanced run metadata such as grouped sections in help output, richer field types, or choices.
- Whether `--print-parsed-values` should also expose the raw run-schema itself, not only the resolved runtime values.

### What should be done in the future
- Extend run-schema support beyond the current initial field set if needed.
- Decide whether to support richer authoring metadata such as explicit CLI flag names, choices, or validation constraints.
- Consider whether the plain `discord-bot run` and `sync-commands` Glazed commands should eventually share more of this dynamic bot-run schema machinery.

### Code review instructions
- Start with `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/run_schema.go`.
- Then inspect `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor.go` and `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go`.
- Finally review `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go` to confirm runtime-config injection into the live host path.
- Validate with the exact commands listed above.

### Technical details
- Current example bot authoring shape:

```js
configure({
  name: "knowledge-base",
  run: {
    fields: {
      index_path: {
        type: "string",
        help: "Optional path label for the active docs index",
        default: "builtin-docs"
      },
      read_only: {
        type: "bool",
        help: "Disable write operations for future knowledge-base mutations",
        default: true
      }
    }
  }
})
```

- Current operator flow:

```bash
GOWORK=off go run ./cmd/discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots
GOWORK=off go run ./cmd/discord-bot bots run knowledge-base \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --index-path ./docs/local-index \
  --print-parsed-values
```

- Current JS runtime access:

```js
ctx.config.index_path
ctx.config.read_only
```

## Step 4: Debug the ping bot `/search` failure and improve runtime diagnostics

The next interruption was a real operator-reported bug while running the example `ping` bot. The `/search architecture` slash command reached the JavaScript runtime, deferred successfully, and then failed with the unhelpful host-side error `promise rejected: map[]`. That was enough to know a promise was failing inside JavaScript, but not enough to tell why.

I treated this as both an example-bot bug and a runtime-debuggability gap. The immediate example problem was that the ping bot implemented `sleep(...)` using `setTimeout`, but our runtime does not expose browser-style global timers. The platform already has a real timer module, so the example should have been written against `require("timer")` instead. The second problem was that rejected JavaScript `Error` objects were being stringified through `Export()`, which collapses many useful error objects into `map[]`.

### Prompt Context

**User prompt (verbatim):** "when running the ping bot and usig the /search architecture command, I get: ... failed to dispatch interaction to javascript bot error=\"promise rejected: map[]\" feel free to add more debug logging as well so that we can have an easier time tracking things"

**Assistant interpretation:** Reproduce and fix the ping example failure, and also make the host/runtime logging and error messages more actionable for future debugging.

**Inferred user intent:** The example bot should actually work, and when it does not, the logs should point to the real JavaScript failure rather than hiding it behind opaque exported-object formatting.

### What I did
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js` to replace the local `setTimeout`-based helper with:
  - `const { sleep } = require("timer")`
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go` so promise rejection reporting now snapshots both:
  - the exported settled value
  - a VM-side string/stack rendering of the rejected value
- This means rejected JavaScript errors now surface as real messages instead of `map[]` whenever possible.
- Updated `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go` to add debug-level dispatch logs and to wrap returned interaction errors with more context such as command/custom ID and script path.
- Added tests in `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime_test.go` covering:
  - clearer JavaScript error propagation from rejected promises
  - working timer-module sleep in command handlers
- Validated with:
  - `gofmt -w internal/jsdiscord/bot.go internal/jsdiscord/host.go internal/jsdiscord/runtime_test.go`
  - `GOWORK=off go test ./internal/jsdiscord ./internal/bot ./cmd/discord-bot ./internal/botcli`
  - `GOWORK=off go test ./...`
  - `GOWORK=off go run ./cmd/discord-bot bots help ping --bot-repository ./examples/discord-bots`
  - `GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots`

### Why
- Example bots are executable documentation. If the showcase bot breaks on a normal command path, it undermines trust in both the runtime and the docs.
- Runtime error messages should preserve JavaScript exception text and stack information whenever we can get it safely.
- Interaction dispatch logs need enough context to distinguish “which script / which command / which custom ID failed” without attaching a debugger first.

### What worked
- The ping bot now loads with the proper timer dependency instead of relying on a missing global `setTimeout`.
- Promise rejection errors now include real JavaScript details instead of `map[]`.
- The runtime test suite now covers both timer-backed async command behavior and clearer rejection reporting.
- `bots help ping` still works and confirms the example bot remains discoverable after the change.

### What didn't work
- The original error reporting path was too lossy for JavaScript `Error` objects. Exporting them directly was not sufficient for diagnostics because the resulting Go-side value often looked like an empty map.

### What I learned
- The timer capability already existed in the platform; the bug was in the example’s assumptions, not in async settlement itself.
- For JavaScript exceptions, VM-side stringification is much more useful than Go-side exported-object formatting.

### What was tricky to build
- The subtle part was improving rejection reporting without keeping raw `goja.Value` objects around longer than necessary. The fix snapshots both an exported value and a VM-rendered text description while still inside the owner-thread call.

### What warrants a second pair of eyes
- Whether we want even more structured interaction logging at info level for selected operations, or if debug-level dispatch traces plus richer wrapped errors are sufficient.
- Whether more example bots still assume browser globals that are not part of this runtime contract.

### What should be done in the future
- Consider a small “runtime environment assumptions” section in the example-bot README so authors know to use provided modules such as `timer` rather than browser globals.
- Add a smoke-test playbook for the example Discord bot repository once more live commands accumulate.

### Code review instructions
- Review `/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js` first to confirm the concrete bug fix.
- Then inspect `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go` around promise settlement and rejection formatting.
- Finally inspect `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go` for the new debug-level interaction dispatch logs and contextual error wrapping.
- Validate with the exact commands listed above.

### Technical details
- Old broken example pattern:

```js
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))
```

- Fixed example pattern:

```js
const { sleep } = require("timer")
```

- Intended improvement in future failure mode:
  - before: `promise rejected: map[]`
  - after: a real JavaScript error message such as `promise rejected: ReferenceError: setTimeout is not defined`

## Step 5: Delete the obsolete multi-host layer and add lifecycle logging

After fixing the ping bot bug, the next requested cleanup was to finish the architectural job and delete the leftover multi-host runtime code rather than just leaving it unused. At the same time, I added richer lifecycle debug logs so command and interaction handling are easier to trace during live runs.

This step had two goals that fit together well. First, remove code that represented a runtime shape the project has explicitly backed away from. Second, make the surviving single-bot host path more observable so future debugging focuses on the actual runtime contract instead of stale architectural branches.

### Prompt Context

**User prompt (verbatim):** "1. + 3. 5. 5a: delete 5b: explain"

**Assistant interpretation:** Commit the current work, add the requested logging pass, and complete the DISCORD-BOT-004 cleanup by deleting the obsolete multi-host runtime layer and documenting why.

**Inferred user intent:** Stop carrying dead-end architecture, make debugging easier in the path that remains, and preserve a written explanation for future maintainers.

### What I did
- Deleted:
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/multihost.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/multihost_test.go`
- Added `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor_test.go` so the useful descriptor/fallback-name coverage survives after removing the multi-host test file.
- Extended `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go` with richer debug lifecycle logging for:
  - interaction dispatch start
  - defer
  - reply
  - follow-up
  - edit
  - modal show
  - autocomplete reply
  - request-scoped Discord operations such as channel send, message edit, message delete, and reaction add
- Updated the DISCORD-BOT-004 docs to explain why the multi-host code was deleted rather than merely left unused.
- Validated with:
  - `gofmt -w internal/jsdiscord/host.go internal/jsdiscord/descriptor_test.go`
  - `GOWORK=off go test ./internal/jsdiscord ./internal/bot ./cmd/discord-bot ./internal/botcli`
  - `GOWORK=off go test ./...`

### Why
- Leaving `multihost.go` in the tree kept signaling that host-level multi-bot composition might still be part of the intended architecture.
- The actual useful pieces from that earlier work had already been preserved elsewhere:
  - repository discovery
  - descriptor inspection
  - single-bot selection
- The code that remained was mostly dead architectural surface area.
- The logging additions matter because deleting dead code is only half the cleanup; the remaining live path should also be easier to reason about and debug.

### What worked
- The runtime still passes focused and full test runs after deleting the multi-host files.
- Descriptor coverage was preserved in a more appropriate dedicated test file.
- The host now emits richer debug logs around interaction lifecycle operations and Discord host actions.

### What didn't work
- N/A in this step.

### What I learned
- The actual reusable value from the old multi-host slice was never the live composition object itself. It was the descriptor/discovery groundwork around it.
- Once the architecture decision is firm, deleting obsolete runtime code is clearer than keeping it around “just in case.”

### What was tricky to build
- The subtle part was keeping the useful tests while removing the now-obsolete runtime layer. The solution was to move descriptor-oriented coverage into `descriptor_test.go` instead of letting it remain coupled to `multihost_test.go`.

### What warrants a second pair of eyes
- Whether any older planning tickets outside DISCORD-BOT-004 still overstate the relevance of the deleted multi-host runtime layer and should be cleaned up later.
- Whether some of the new lifecycle logs should eventually be promoted from debug to info for selected operator workflows.

### What should be done in the future
- Refresh any stale planning docs that still talk as if the live host uses multi-host routing.
- Decide whether to add one explicit operator-facing playbook for running with `--log-level debug` when tracing command lifecycle issues.

### Code review instructions
- Review `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go` for the lifecycle logging additions.
- Review `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor_test.go` to confirm the preserved descriptor coverage.
- Confirm that `internal/jsdiscord/multihost.go` and `internal/jsdiscord/multihost_test.go` are intentionally gone.

### Technical details
- New debug logs now cover actions such as:
  - `defer`
  - `reply`
  - `followUp`
  - `edit`
  - `showModal`
  - `autocomplete.reply`
  - `discord.channels.send`
  - `discord.messages.edit`
  - `discord.messages.delete`
  - `discord.messages.react`
- The deletion rationale is now also recorded in `reference/01-single-bot-runner-reference-and-migration-notes.md`.

## Step 6: Split the monolithic host implementation into concern-specific files

After several more feature slices landed, `internal/jsdiscord/host.go` had become the main growth hotspot in the runtime. It was carrying constructor/setup code, event dispatch, interaction response helpers, request-scoped Discord operations, payload normalization, command-shape conversion, logging helpers, and object-mapping helpers all in one file. At that point, continuing to add new features there would have made every ticket slower to review and riskier to edit.

I treated this as an enabling refactor rather than as cosmetic cleanup. The goal was not to redesign the runtime or introduce new packages. The goal was to keep the current `jsdiscord` package intact while splitting the host implementation into smaller files with clearer landing zones for future work.

### Prompt Context

**User prompt (verbatim):** "ok, split files"

**Assistant interpretation:** Perform a same-package structural refactor of the host implementation so future Discord JS feature work no longer accumulates in one giant file.

**Inferred user intent:** Reduce future feature friction now that the host file has become large enough to slow down ongoing development.

### What I did
- Replaced the monolithic `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go` layout with a smaller same-package file set:
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_dispatch.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_logging.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_responses.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_channels.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_messages.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_members.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_ops_helpers.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_payloads.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_commands.go`
  - `/home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_maps.go`
- Kept the package boundary unchanged. Everything still lives under `internal/jsdiscord`.
- Split request-scoped Discord ops by concern so future features have obvious homes:
  - channel ops
  - message ops
  - member ops
- Ran:
  - `GOWORK=off go test ./internal/jsdiscord ./internal/bot ./cmd/discord-bot`
  - `GOWORK=off go test ./...`
- Committed the refactor as:
  - `6c1b87b` — `Split Discord JS host implementation by concern`

### Why
- The main development pain had become file size and navigability, not package architecture.
- A same-package split reduces future feature friction without forcing premature export/package-boundary decisions.
- Grouping Discord ops by concern gives the next feature tickets clearer insertion points.

### What worked
- Focused and full test suites still passed after the split.
- The working tree was clean after the refactor commit.
- The new file layout keeps behavior stable while making the host easier to review and extend.

### What didn't work
- My first automatic extraction pass left `host_payloads.go` missing the `normalizedResponse` type definition, which caused build failures:
  - `internal/jsdiscord/host_payloads.go:173:38: undefined: normalizedResponse`
  - `internal/jsdiscord/host_payloads.go:176:11: undefined: normalizedResponse`
  - `internal/jsdiscord/host_payloads.go:178:11: undefined: normalizedResponse`
  - `internal/jsdiscord/host_payloads.go:180:11: undefined: normalizedResponse`
  - `internal/jsdiscord/host_payloads.go:217:11: undefined: normalizedResponse`
- I fixed that by restoring the type declaration into `host_payloads.go` and rerunning the tests.

### What I learned
- The file split was justified once the host had become the default landing zone for almost every Discord runtime change.
- Same-package structural refactors are a good intermediate move before any deeper package decisions, because they keep the behavior stable while making the codebase easier to evolve.

### What was tricky to build
- The trickiest part was keeping the extraction mechanical enough to preserve behavior, while still ending with files grouped by actual concern rather than arbitrary line ranges.
- The Discord ops builder was especially worth splitting intentionally, because that is the seam future feature tickets will keep touching.

### What warrants a second pair of eyes
- Whether the current file boundaries are the most natural long-term grouping after the next few tickets land.
- Whether one or two additional helper files should eventually be carved out if payload normalization continues to grow.

### What should be done in the future
- Continue feature work against the new `host_*.go` layout rather than letting one file become the dumping ground again.
- Reassess package boundaries only after a few more feature slices make the stable seams more obvious.
