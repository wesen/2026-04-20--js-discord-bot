---
Title: Implementation Diary
Ticket: DISCORD-BOT-FRAMEWORK
Status: active
Topics:
    - discord
    - goja
    - javascript
    - go
    - framework
    - embedding
    - glazed
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: design-doc/01-framework-extraction-design-and-implementation-guide.md
      Note: The main design document this diary accompanies
ExternalSources: []
Summary: Chronological diary for the framework extraction design ticket.
LastUpdated: 2026-04-22T10:00:00-04:00
WhatFor: Record the analysis steps, design decisions, and deliverables for making the discord-bot a reusable framework.
WhenToUse: Use when resuming or reviewing the DISCORD-BOT-FRAMEWORK work.
---

# Diary

## Goal

Record the analysis, design, and documentation work for extracting the JS Discord bot runtime into a reusable Go framework that other applications can embed — for example, a web server with a database that also exposes a Discord bot surface.

## Step 1: Create ticket and analyze current coupling

### What I did

- Created docmgr ticket `DISCORD-BOT-FRAMEWORK`.
- Read every Go file in the project to map coupling points between packages.
- Read all relevant docmgr diaries from DISCORD-BOT-002 through DISCORD-BOT-020, plus the UI-DSL design tickets and the PASS1–4 cleanup passes.
- Read the go-go-goja engine, modules, and runtimeowner packages to understand the module registration path.
- Started the design document.

### Key findings

The current codebase has five packages with tight but discoverable coupling:

1. **`internal/jsdiscord/`** — The core value. This package owns the goja runtime, the `defineBot` DSL, `BotHandle`, `DispatchRequest`, `DiscordOps`, payload normalization, and event routing. It is the main extraction target. It currently depends only on `discordgo`, `goja`, and `go-go-goja/engine`.

2. **`internal/bot/`** — The live host wiring. This package creates the `discordgo.Session`, registers all Discord event handlers, and routes them to `jsdiscord.Host.Dispatch*` methods. It depends on `config.Settings` for credentials and on `jsdiscord` for the host. This is the package that must become pluggable.

3. **`internal/botcli/`** — The CLI orchestration. Discovery, bot resolution, dynamic Glazed schema building, and the `bots run` execution path. This is application-level code, not framework code. Embedders would replace this with their own wiring.

4. **`internal/config/`** — A Glazed-backed `Settings` struct for Discord credentials. Simple and already fairly independent. The framework would expose its own config interface.

5. **`cmd/discord-bot/`** — The CLI entrypoint. Wires Glazed commands, help system, and the `botcli` commands together. This stays as the standalone application.

### What I learned

- The `jsdiscord.Host` struct is *almost* already a framework API. It takes a script path, creates a runtime, loads the bot, and exposes `Describe()`, `DispatchInteraction()`, `DispatchReady()`, etc. The main coupling is that `NewHost()` calls `engine.NewBuilder()` internally and that `buildDiscordOps()` reaches directly into the `discordgo.Session`.

- The `DiscordOps` struct is the key extensibility seam. It is already a bag of function pointers (`GuildFetch`, `ChannelSend`, etc.). Today these are populated by `build*Ops()` helpers that close over a `*discordgo.Session`. A framework would let embedders provide their own ops — or additional ops — without touching the host.

- The `DispatchRequest` struct carries everything a JS handler sees: args, config, discord ops, reply/followUp/edit/defer/showModal callbacks. This is the second key seam. Today the responder callbacks are built from `discordgo.Interaction` and `discordgo.Session`, but they could just as well be built from an HTTP handler or a test mock.

### Next step

Write the design document with three design options, full API references, pseudocode, and diagrams.

## Step 2: Write the design document

### What I did

- Wrote a comprehensive design document with:
  - Executive summary
  - Detailed problem statement with the five coupling points
  - Full current architecture walkthrough with package map, dependency graph, and file-by-file coupling analysis
  - Three design options with full pseudocode, diagrams, and trade-off analysis
  - Detailed implementation plan with file mapping table and code diffs
  - Complete API reference
  - Five concrete embedding examples
  - Open questions and alternatives considered

### What I found

The codebase is already well-structured for extraction. The main work is:

1. Parameterizing `NewHost()` to accept engine builder configuration
2. Parameterizing `baseDispatchRequest()` to accept ops customizers
3. Parameterizing `buildContext()` to accept context extenders
4. Growing `DiscordOps.Extensions` for custom namespace injection
5. Wrapping everything in a `Framework` convenience layer

### Next step

Upload the design document to reMarkable.

## Step 3: Add jsverbs integration and RuntimeFactory override

### What I did

- Read the entire loupedeck codebase to understand how it creates custom goja runtimes and integrates with jsverbs.
- Read the `go-go-goja/pkg/jsverbs` package in full — `model.go`, `scan.go`, `runtime.go`, `binding.go`, `command.go` — to understand the scan/build/invoke pipeline.
- Read loupedeck's `verbs/bootstrap.go` to understand repository discovery from CLI, env, config, and embedded FS.
- Read loupedeck's `verbs/command.go` to understand how verbs become Cobra commands with custom invokers.
- Read loupedeck's `runtime/js/runtime.go` and `registrar.go` to understand the custom runtime factory pattern.

### Key findings

The loupedeck pattern reveals a 4th coupling point I missed in the original analysis:

- **Runtime creation is hardcoded.** The discord-bot's `jsdiscord.NewHost()` calls `engine.NewBuilder()` internally with fixed configuration. Loupedeck needs to control this — it registers its own modules (ui, gfx, easing, anim, present, state, metrics) and its own environment object. The framework must let embedders replace the entire runtime creation process.

The jsverbs system provides three reusable pieces:
1. `jsverbs.ScanDir()` / `ScanFS()` — find annotated JS files
2. `Registry.CommandsWithInvoker()` — build Glazed commands with custom execution
3. `Registry.RequireLoader()` — provide a custom require loader for verb invocation

### What I added to the design

Three new concepts:
1. **RuntimeFactory** interface — lets embedders control how goja runtimes are created for both bot scripts and verb invocations
2. **Repository** + **RepositoryDiscovery** — unified repository model for both bot discovery and jsverb scanning
3. **VerbRegistry** + **RegisterVerbs()** — framework helper for scanning repos and registering jsverb CLI commands

Updated the comparison table and the file mapping.

### Next step

Re-upload the updated design document to reMarkable.

## Step 4: Start Track A by improving the standalone single-bot run path

This was the first deliberate implementation slice after splitting the framework work into two tracks: keep single-bot embedding first-class, and make repo-driven `botcli` workflows optional. I chose to start with the single-bot side so the simple path gets better immediately instead of being treated as legacy behavior while all the attention goes to dynamic repository loading.

The concrete change in this step was small but important: I added `--sync-on-start` to the standalone `discord-bot run` command so the explicit single-bot path keeps the same operational convenience that the new host-managed repo-discovered run commands already have.

### Prompt Context

**User prompt (verbatim):** "Alright, let's work on them task by task, commit as you go, keep a diary."

**Assistant interpretation:** Start executing the framework-extraction plan incrementally, keep each change focused, and record the work as it happens.

**Inferred user intent:** The user wants disciplined implementation progress with small reviewable slices, intermediate commits, and durable ticket documentation.

### What I did

- Picked the first concrete Track A task: improve the current standalone single-bot operator flow.
- Updated `cmd/discord-bot/commands.go` so `newRunCommand()` now exposes `--sync-on-start` in the standalone `run` command help.
- Wired `runCommand.RunIntoGlazeProcessor(...)` to call `bot.SyncCommands()` before `bot.Open()` when `--sync-on-start` is set.
- Added a root-level regression test in `cmd/discord-bot/root_test.go` to verify `discord-bot run --help` now shows `--sync-on-start`.
- Ran:
  - `gofmt -w cmd/discord-bot/commands.go cmd/discord-bot/root_test.go`
  - `go test ./cmd/discord-bot ./internal/botcli -run 'StandaloneRunHelpShowsSyncOnStart|RootLevelBotRepositoryFlagRegistersKnowledgeBaseRunVerb|LegacyRunSyntaxWorksForUiShowcase' -v`
  - `go test ./...`
- Updated the framework ticket `tasks.md` and `changelog.md` to record this first implementation slice.

### Why

The framework split only makes sense if the simple path remains simple and well-supported. Adding `--sync-on-start` to the standalone run path is a concrete way to keep the single-bot experience first-class rather than letting the new repo-driven command tree become the only polished workflow.

### What worked

- The standalone `run` command now advertises `--sync-on-start` just like the host-managed dynamic run commands.
- The implementation was straightforward because `internal/bot/bot.go` already had `SyncCommands()` as a first-class method.
- `go test ./...` passed after the change.

### What didn't work

- N/A

### What I learned

- The standalone and dynamic run paths had drifted a little: the dynamic repo-driven run path already had `--sync-on-start`, but the explicit single-bot path did not.
- This kind of drift is exactly why the framework-extraction work should keep Track A visible: otherwise the simple path slowly becomes the less-capable path.

### What was tricky to build

- The code change itself was not hard, but the tricky part was choosing the right first slice. It would have been easy to jump directly into public-`botcli` extraction because that is the bigger architectural change, but that would have left the single-bot story under-specified. I handled that by taking a small user-facing improvement on the single-bot side first and recording it explicitly in the ticket tasks.

### What warrants a second pair of eyes

- Whether `--sync-on-start` should remain a command-local flag on `run` only, or whether the future extracted single-bot framework API should also expose a similarly named option directly.
- Whether the standalone `run` path should eventually share more code with the host-managed `botRunCommand` path so these feature differences cannot drift again.

### What should be done in the future

- Continue Track A with the actual public single-bot framework extraction (`framework.New(...)` / `NewHost(...)`).
- After that, move to Track B and promote `internal/botcli` into an optional public package.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/commands.go` and inspect `newRunCommand()` plus `runCommand.RunIntoGlazeProcessor(...)`.
- Then review `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root_test.go` for the new help regression coverage.
- Validate with:
  - `go test ./...`
  - `go run ./cmd/discord-bot run --help`

### Technical details

- The implementation uses the existing `bot.SyncCommands()` method from `internal/bot/bot.go`.
- The new flag is parsed from the default Glazed section using `vals.GetField(schema.DefaultSlug, "sync-on-start")`.
- The runtime order is now:
  1. parse config
  2. create bot
  3. optionally sync commands
  4. open gateway session
  5. block until shutdown

## Step 5: Create the first public single-bot framework package

After improving the standalone single-bot CLI path, the next Track A slice was to expose that simplicity as a public package instead of leaving it trapped behind application-only commands. I kept this slice intentionally small: one package, one constructor, explicit script selection, explicit or env-backed credentials, runtime config injection, and no repository scanning at all.

This is the first real code step toward the “single explicit bot remains the easy path” design. It does not solve the full extraction yet, but it proves that downstream code can import a stable public package and create one bot instance without touching `internal/botcli`.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue with the next focused framework-extraction task, keep it reviewable, and record it in the ticket diary.

**Inferred user intent:** The user wants the design split turned into real code incrementally, with single-bot simplicity preserved as a first-class use case.

### What I did

- Added a new public package:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework.go`
- Implemented an initial public API:
  - `framework.New(opts...)`
  - `framework.WithScript(...)`
  - `framework.WithCredentials(...)`
  - `framework.WithCredentialsFromEnv()`
  - `framework.WithRuntimeConfig(...)`
  - `framework.WithSyncOnStart(...)`
- Added a public `framework.Bot` wrapper with:
  - `Open()`
  - `Run(ctx)`
  - `SyncCommands()`
  - `Close()`
- Kept the package deliberately simple by wrapping `internal/bot.NewWithScript(...)` rather than trying to solve runtime factories or transport abstraction in the same slice.
- Added tests in `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework_test.go` for:
  - missing script error
  - env-backed credentials
  - explicit credentials + runtime config
- Ran:
  - `gofmt -w pkg/framework/framework.go pkg/framework/framework_test.go`
  - `go test ./pkg/framework ./cmd/discord-bot ./internal/botcli -run 'TestNew|StandaloneRunHelpShowsSyncOnStart|LegacyRunSyntaxWorksForUiShowcase' -v`
  - `go test ./...`
- Updated the framework ticket tasks and changelog to record the new public package.

### Why

The design says embedders should have a first-class single-bot path that does not require repository discovery, jsverbs scanning, or `botcli`. The quickest way to make that real was to expose a thin public wrapper around the existing single-bot runtime behavior.

### What worked

- The new package can construct one explicit bot using an existing example script with no repository bootstrap logic.
- `WithCredentialsFromEnv()` works against the existing `DISCORD_BOT_TOKEN` / `DISCORD_APPLICATION_ID` environment variables.
- Runtime config injection is exposed in the public package without forcing users through the repo-driven command path.
- The package-level tests passed, and the full repo test suite stayed green.

### What didn't work

- N/A

### What I learned

- The current `internal/bot.NewWithScript(...)` path is already a good foundation for the public single-bot API. The biggest missing piece was not deep runtime surgery — it was a public package boundary and a cleaner option surface.
- This confirms the split is viable: Track A can move independently from Track B for a while.

### What was tricky to build

- The main tricky part was resisting scope creep. It would have been tempting to add runtime-factory hooks, custom transports, or a public `NewHost(...)` in the same change because the design document discusses them heavily. I avoided that by defining the smallest package that still proves the single-bot story externally: `New`, options, and a thin wrapper around the existing runtime.
- Another subtle point was making sure the tests exercised the public path without opening a real Discord session. The fix was to use `framework.New(...)` against an existing example script and stop at construction time; `discordgo.New(...)` and bot loading are enough to validate the package shape without needing network activity.

### What warrants a second pair of eyes

- The package name and path choice: `pkg/framework` is plausible, but we may still want to rename it to something more explicit if the repo later grows multiple public packages.
- Whether `Open()` should always honor `SyncOnStart`, or whether sync-on-start should remain exclusively part of `Run(ctx)` and higher-level startup helpers.
- Whether the next slice should add a public `NewHost(...)` now, or wait until the runtime-factory work is ready.

### What should be done in the future

- Continue Track A by deciding whether to expose a public `NewHost(...)` in addition to `framework.New(...)`.
- Add minimal public examples showing how a downstream app imports `pkg/framework` and runs one explicit bot.
- After Track A is comfortably established, start Track B and promote `internal/botcli` into an optional public package.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework.go` to review the public API surface and how it maps onto `internal/bot.NewWithScript(...)`.
- Review `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework_test.go` for the no-network package tests.
- Validate with:
  - `go test ./pkg/framework`
  - `go test ./...`

### Technical details

- `framework.New(...)` validates script presence before constructing the internal bot.
- `WithCredentialsFromEnv()` reads the same env vars as the standalone CLI:
  - `DISCORD_BOT_TOKEN`
  - `DISCORD_APPLICATION_ID`
  - `DISCORD_GUILD_ID`
  - `DISCORD_PUBLIC_KEY`
  - `DISCORD_CLIENT_ID`
  - `DISCORD_CLIENT_SECRET`
- `WithRuntimeConfig(...)` clones the provided map so callers do not share mutable config state with the package internals.
- The public package stays intentionally unaware of repositories, bootstraps, jsverbs scanning, and `botcli` command registration.

## Step 6: Add the first public single-bot embedding example

Once `pkg/framework` existed, the next risk was that it would remain invisible or feel hypothetical. I wanted the Track A story to be copy-pasteable, not just described by an API. So this slice adds a minimal example application plus a few doc links that make the new public single-bot path easy to discover.

### Prompt Context

**User prompt (verbatim):** "continue"

**Assistant interpretation:** Take the next focused Track A task and keep progressing with small commits.

**Inferred user intent:** The user wants the framework split turned into a concrete sequence of implementation slices, with docs and examples arriving early enough that the new API is usable by humans, not just tests.

### What I did

- Added a new example app at:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-single-bot/main.go`
- Added usage notes at:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-single-bot/README.md`
- Wired the example to the existing `unified-demo` bot script so the example demonstrates both explicit single-script startup and `ctx.config` runtime injection.
- Updated `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/discord-bots/README.md` to point readers at the new public embedding path.
- Updated `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md` to mention `pkg/framework` and the new example.
- Ran:
  - `gofmt -w examples/framework-single-bot/main.go`
  - `go test ./examples/framework-single-bot ./pkg/framework ./...`
- Updated the framework ticket tasks and changelog to mark the minimal embedded app example as done.

### Why

The public package only becomes genuinely useful when there is a clear answer to "what does embedding look like in real code?" A minimal example app answers that question quickly and keeps the single-bot path feeling first-class rather than theoretical.

### What worked

- The example compiles cleanly as part of `go test ./...`.
- The example shows the right shape for downstream users: env-backed credentials, one explicit script, runtime config, and context-driven shutdown.
- Reusing `examples/discord-bots/unified-demo/index.js` means the example demonstrates `ctx.config` behavior without needing another duplicate bot script.

### What didn't work

- N/A

### What I learned

- The most useful example for this slice is not a brand-new bot; it is a small Go wrapper around an existing example bot script. That keeps the docs focused on the embedding surface instead of duplicating bot behavior.
- The Track A story is now much easier to explain: standalone CLI, public package, and example app all point to the same single-bot runtime model.

### What was tricky to build

- The main judgment call was where to put the example. I placed it under `examples/framework-single-bot/` instead of inside `examples/discord-bots/` because it is an embedding application, not another discovered bot implementation.
- Another subtle point was choosing a signal handling pattern that is simple and portable. I used `signal.NotifyContext(context.Background(), os.Interrupt)` so the example remains small and easy to adapt.

### What warrants a second pair of eyes

- Whether we also want to surface this example from the embedded CLI docs (`pkg/doc/tutorials/...`) soon, or keep this slice limited to README-level discoverability.
- Whether the next Track A example should be the more advanced “one explicit bot + custom module/runtime” case from the task list, or whether we should first expose lower-level public hooks such as `NewHost(...)`.

### What should be done in the future

- Add the second Track A example: one explicit bot plus custom module/runtime hooks.
- Potentially add a lower-level public host API if that example needs more extension seams than `pkg/framework` currently exposes.
- Then move on to Track B and the optional public `botcli` package.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-single-bot/main.go`.
- Then review `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-single-bot/README.md`.
- Check the discoverability updates in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/discord-bots/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
- Validate with:
  - `go test ./examples/framework-single-bot ./pkg/framework ./...`
  - `GOWORK=off go run ./examples/framework-single-bot` (with Discord env vars set)

### Technical details

- The example intentionally uses `framework.WithCredentialsFromEnv()` so it matches the standalone CLI's credential-loading behavior.
- The script path is explicit: `examples/discord-bots/unified-demo/index.js`.
- Runtime config is passed through `framework.WithRuntimeConfig(...)` so `/unified-ping` can prove `ctx.config.db_path` and `ctx.config.api_key` are present.
- Shutdown is driven by a normal Go context derived from `signal.NotifyContext(...)`, which is the expected embedding pattern for downstream applications too.
