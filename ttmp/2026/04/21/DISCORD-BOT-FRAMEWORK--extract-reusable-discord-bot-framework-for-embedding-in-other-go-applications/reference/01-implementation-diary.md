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

## Step 7: Add a custom runtime-module seam and advanced single-bot example

The next Track A task after the minimal embedding example was the more advanced case from the task list: one explicit bot plus a custom module/runtime extension. I still did not want to jump all the way to a public `botcli` package or to a huge host abstraction, so I chose the narrowest useful seam: let the public single-bot package pass custom runtime module registrars down into `jsdiscord.NewHost(...)`.

That keeps the single-bot story intact while proving a downstream app can extend the JS runtime with its own Go-native `require()` modules.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue with the next Track A implementation slice, not just planning.

**Inferred user intent:** The user wants the framework extraction to keep moving in code-sized, reviewable increments until the single-bot path is both useful and extensible.

### What I did

- Added `internal/jsdiscord/host_options.go` with host-option plumbing for runtime module registrars.
- Updated `internal/jsdiscord/host.go` so `NewHost(ctx, scriptPath, opts...)` now accepts optional host options and appends custom runtime registrars after the built-in Discord/UI registrars.
- Updated `internal/jsdiscord/descriptor.go` so `LoadBot(...)` can also forward host options.
- Updated `internal/bot/bot.go` so `NewWithScript(...)` can pass host options through to `jsdiscord.LoadBot(...)` while preserving existing call sites.
- Extended the public package in `pkg/framework/framework.go` with:
  - `WithRuntimeModuleRegistrars(...)`
- Added package tests in `pkg/framework/framework_test.go` covering:
  - failure when a bot script requires a missing custom module
  - success when the same script is constructed with a custom runtime registrar
- Added a second embedding example:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-custom-module/main.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-custom-module/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-custom-module/bot/index.js`
- Updated top-level discoverability docs:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/discord-bots/README.md`
- Ran:
  - `gofmt -w internal/jsdiscord/host.go internal/jsdiscord/host_options.go internal/jsdiscord/descriptor.go internal/bot/bot.go pkg/framework/framework.go pkg/framework/framework_test.go examples/framework-custom-module/main.go`
  - `go test ./pkg/framework ./examples/framework-single-bot ./examples/framework-custom-module ./internal/jsdiscord ./...`

### Why

The design document explicitly calls out custom runtime/module injection as a core requirement for embedders. Without at least one public hook here, `pkg/framework` would remain a thin wrapper around the fixed built-in runtime. This slice adds the first extension seam while still keeping the public API focused on the single-bot case.

### What worked

- The custom registrar path is enough to support a real downstream-style extension: `require("app")` from JavaScript, implemented in Go by the embedding application.
- The new tests prove both sides of the behavior:
  - the script fails without the registrar
  - the script loads successfully with the registrar
- The new example is self-contained and demonstrates the feature without mixing in repository discovery.

### What didn't work

- My first assertion for the missing-module failure expected the module name (`"app"`) in the error text. The actual Goja require error was `"Invalid module"`, so I had to relax the assertion to match the runtime's real error string.

### What I learned

- We do not need a full public `NewHost(...)` yet to unlock useful extensibility. A targeted `WithRuntimeModuleRegistrars(...)` option already opens a meaningful escape hatch.
- The internal architecture was already close to supporting this; the main issue was simply that `jsdiscord.NewHost()` had no option channel and the public package had no way to reach that layer.

### What was tricky to build

- The main tricky part was choosing where to put the seam. If the option lived only on `pkg/framework`, it would feel magical and hard to evolve. If it lived only on `internal/jsdiscord`, downstream embedders still could not reach it. The right answer for this slice was both: internal host options plus one public framework option that maps onto them.
- Another subtle point was preserving existing call sites. I handled that by making the new host-option parameter variadic in both `jsdiscord.NewHost(...)` and `bot.NewWithScript(...)`, so the old code paths continued to compile unchanged.

### What warrants a second pair of eyes

- Whether the next lower-level public seam should be `framework.NewHost(...)`, or whether we should first add one more focused option such as runtime initializers or additional require options.
- Whether `WithRuntimeModuleRegistrars(...)` should remain the preferred public extension hook, or whether a future `WithRuntimeFactory(...)` should supersede it once Track A grows lower-level APIs.

### What should be done in the future

- Decide whether Track A now needs a public `NewHost(...)` to support lower-level embedding beyond the current option surface.
- If not, Track A may be sufficiently established to start Track B and promote `internal/botcli` into an optional public package.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/jsdiscord/host_options.go` and `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/jsdiscord/host.go`.
- Then review `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework.go` and `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework_test.go`.
- Finally review the example app at `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-custom-module/`.
- Validate with:
  - `go test ./pkg/framework ./examples/framework-single-bot ./examples/framework-custom-module ./internal/jsdiscord ./...`
  - `GOWORK=off go run ./examples/framework-custom-module` (with Discord env vars set)

### Technical details

- Built-in runtime registrars still load first: the Discord registrar and UI registrar remain default behavior.
- `WithRuntimeModuleRegistrars(...)` appends additional `engine.RuntimeModuleRegistrar` values after the built-ins.
- The public option accepts `engine.RuntimeModuleRegistrar`, so downstream embedders can define native modules without importing any internal package.
- The custom example registers an `app` native module and uses it from a normal `defineBot(...)` script via `require("app")`.

## Step 8: Conclude the main merge and start Track B with a public bootstrap package

Before starting the next implementation slice I had to deal with the repository state: the branch was already in the middle of a `main` merge. The merge itself turned out to be straightforward — mostly the new show-space bot work from `main` — and the index was already in the "all conflicts fixed" state. I validated the combined branch, concluded the merge commit, and then immediately used the clean branch to start Track B with the smallest public extraction: a bootstrap package for repo-driven bot repository resolution.

### Prompt Context

**User prompt (verbatim):** "merge main, which shouldn't have too many changes affecting us.

Then continue"

**Assistant interpretation:** Finish the pending merge cleanly, make sure the branch is healthy, and then continue the framework extraction instead of stopping at the merge.

**Inferred user intent:** The user wants the branch rebased mentally onto current project reality before more framework work lands, and wants progress to continue immediately after the merge.

### What I did

- Inspected `git status` and confirmed the repository was already in an unfinished merge state with all conflicts resolved and the merged files staged.
- Ran:
  - `go test ./...`
- Concluded the merge with:
  - `git commit -m "Merge main into task/discord-bot-framework"`
- Started Track B by adding a new public package:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap.go`
- Exposed a first public repo-bootstrap API there:
  - `Repository`
  - `Bootstrap`
  - `BuildBootstrap(rawArgs, opts...)`
  - `WithWorkingDirectory(...)`
  - `WithEnvironmentVariable(...)`
  - `WithDefaultRepositories(...)`
  - `WithRepositoryFlagName(...)`
- Wrote focused tests in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap_test.go`
- Switched `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go` to use the new public `pkg/botcli.BuildBootstrap(...)` helper instead of keeping that logic private in the command package.
- Ran:
  - `gofmt -w pkg/botcli/bootstrap.go pkg/botcli/bootstrap_test.go cmd/discord-bot/root.go`
  - `go test ./pkg/botcli ./cmd/discord-bot ./internal/botcli ./...`

### Why

Track B is about making the repo-driven `botcli` behavior embeddable. The smallest useful public piece is not full Cobra integration yet — it is the bootstrap/repository resolution logic, because dynamic command registration depends on knowing the repositories before Cobra parses subcommands. By making that logic public first, the next slices can build on a stable embeddable foundation.

### What worked

- The merge from `main` was indeed low-drama once I recognized the repo was already mid-merge; `go test ./...` passed before I concluded it.
- Extracting the bootstrap logic into `pkg/botcli` was straightforward because `cmd/discord-bot/root.go` already contained a compact precedence implementation.
- The public package tests now cover the essential precedence rules:
  - CLI repositories override env and default
  - env repositories apply when CLI is absent
  - default repositories apply when both CLI and env are absent
  - repeated CLI repositories are deduplicated
  - downstream apps can override the flag name and env var name

### What didn't work

- N/A

### What I learned

- The root bootstrap logic was already a clean candidate for public extraction. It was application-owned only because of file placement, not because of deep coupling.
- Starting Track B with bootstrap extraction feels much safer than jumping directly to public Cobra integration; it lets the public package own the most awkward pre-parse behavior first.

### What was tricky to build

- The main subtlety was choosing the shape of the public types. For this slice I used public aliases for `Bootstrap`, `Repository`, and `DiscoveredBot` to avoid unnecessary conversion churn while the package is still being promoted incrementally.
- Another subtle point was preserving the exact precedence semantics of the existing root command so the public helper stays behaviorally identical to the app's current bootstrap.

### What warrants a second pair of eyes

- Whether the public type aliases should remain aliases, or whether the next Track B slice should convert them into explicitly owned `pkg/botcli` structs before more API surface lands.
- Whether `BuildBootstrap(...)` should grow additional public hooks next (for example app name or default source labels), or whether the next slice should jump to `pkg/botcli.NewCommand(...)` and Cobra integration.

### What should be done in the future

- Continue Track B with a public embeddable Cobra integration entrypoint that consumes the new bootstrap package.
- Move the dynamic command registration behavior itself behind the public package after that.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap.go`.
- Review `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap_test.go` for precedence and customization coverage.
- Then inspect `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go` to verify the app now consumes the public helper.
- Validate with:
  - `go test ./pkg/botcli ./cmd/discord-bot ./internal/botcli ./...`

### Technical details

- `BuildBootstrap(...)` preserves the same precedence as the root command had before extraction:
  1. raw argv `--bot-repository`
  2. `DISCORD_BOT_REPOSITORIES`
  3. `examples/discord-bots` if present
- The helper also preserves raw-argv pre-scan semantics by parsing the repository flag directly from `rawArgs` before Cobra command construction.
- The public package supports downstream customization of:
  - working directory
  - env var name
  - repository flag name
  - default repository paths

## Step 9: Expose a public embeddable Cobra command entrypoint

After the public bootstrap package landed, the next natural Track B slice was to expose the repo-driven command tree itself. This is the point where a downstream application can actually mount the `bots` subtree into its own Cobra root without importing anything from `internal/botcli`.

I kept this slice intentionally narrow too: just make the existing command builder reachable through `pkg/botcli`, prove it embeds cleanly into an arbitrary root, and then switch the app itself to consume that public path.

### Prompt Context

**User prompt (verbatim):** implicit continuation after the previous slice

**Assistant interpretation:** Keep going task-by-task on Track B while the branch is clean and the previous slice is validated.

**Inferred user intent:** The user wants the optional public `botcli` package to become real in successive reviewable increments, not as one giant rewrite.

### What I did

- Added `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command.go` exposing:
  - `NewBotsCommand(bootstrap)`
  - `NewCommand(bootstrap...)`
- Added `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go` with downstream-style tests that:
  - mount the public `bots` command under an arbitrary Cobra root
  - execute a discovered `status` verb successfully
  - verify dynamic `knowledge-base run --help` wiring works through the public package
- Updated `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go` so the application now uses:
  - `pkg/botcli.BuildBootstrap(...)`
  - `pkg/botcli.NewCommand(...)`
  instead of consuming `internal/botcli` directly from the root command.
- Ran:
  - `gofmt -w pkg/botcli/command.go pkg/botcli/command_test.go cmd/discord-bot/root.go`
  - `go test ./pkg/botcli ./cmd/discord-bot ./internal/botcli ./...`

### Why

Bootstrap extraction alone was not enough for downstream embedders to benefit. They still needed a public way to mount the discovered bot command tree into an existing Cobra application. This slice closes that gap with the smallest possible API move: public wrappers over the existing command builder plus downstream-style tests.

### What worked

- The public command wrappers were enough to embed the `bots` subtree into a separate Cobra root cleanly.
- The downstream tests prove not just static command mounting but actual dynamic behavior:
  - discovered verbs
  - host-managed run help
  - full parser wiring through the public surface
- Switching `cmd/discord-bot/root.go` to the public package means the app itself now dogfoods the extracted Track B API.

### What didn't work

- N/A

### What I learned

- The boundary between `internal/botcli` and `pkg/botcli` can be promoted incrementally. There is no need to rewrite the command-building internals all at once; the public package can wrap and progressively absorb them.
- Having the main app consume the public wrappers immediately is useful because it turns the standalone binary into a live integration test for the extracted package.

### What was tricky to build

- The main subtlety was choosing the right level of public exposure. I did not expose new option types or a public root-command builder yet; I only made the already-working `bots` subtree embeddable. That keeps the surface area small while still satisfying the core downstream integration requirement.
- Another subtle point was test shape. The new tests had to look like real downstream usage rather than just calling the underlying internal package directly, so I mounted the public command under a fresh Cobra root and executed through that root.

### What warrants a second pair of eyes

- Whether the next Track B slice should add a public `NewRootCommand(rawArgs, opts...)`, or whether it is cleaner to keep the package focused on the `bots` subtree and let downstream apps own their own roots.
- Whether the public type aliases introduced in Step 8 should become fully owned `pkg/botcli` structs before more public command options are added.

### What should be done in the future

- Continue Track B with the behavior-specific internals that still only live in `internal/botcli`:
  - entrypoint-only scanning policy
  - explicit-verb-only scanning
  - host-managed run semantics
  - app-name/runtime-factory configurability where needed
- Add one docs/example slice showing a downstream app that combines `pkg/framework` and `pkg/botcli` together.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command.go`.
- Review `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go` for the downstream embedding shape.
- Then inspect `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go` to confirm the app now uses the public package end-to-end.
- Validate with:
  - `go test ./pkg/botcli ./cmd/discord-bot ./internal/botcli ./...`

### Technical details

- `pkg/botcli.NewBotsCommand(...)` currently delegates to the existing internal builder, giving downstream users the same behavior as the standalone app without duplicating the implementation yet.
- `pkg/botcli.NewCommand(...)` mirrors the current application convenience API while keeping `Bootstrap` as the main public input.
- The standalone `discord-bot` root now consumes the public package for both repository bootstrap and `bots` command registration, which keeps the extracted API exercised by the app itself.

## Step 10: Add `WithAppName(...)` to the public botcli package

The next small Track B step was about configurability rather than extraction breadth. The public `pkg/botcli` package could already resolve repositories and mount the `bots` subtree, but dynamic bot commands still assumed one hardcoded env prefix: `DISCORD_*`. For downstream embedders, that is too rigid. The right next slice was a narrow option that changes the Glazed app name used by the dynamic parser, because that is what controls the env prefix for bot credentials and other env-backed fields.

### Prompt Context

**User prompt (verbatim):** "ok, and after that it probably makes sense to test tings for real to validate before we move on"

**Assistant interpretation:** Implement `WithAppName(...)` next, then do real manual validation before continuing to further framework work.

**Inferred user intent:** The user wants the public `botcli` package to become truly downstream-usable, and wants a pause for reality-based validation before more abstraction layers are added.

### What I did

- Added internal command-option plumbing in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/options.go`
- Updated `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command.go` so the dynamic parser config is built from a configurable app name instead of a hardcoded `"discord"` string.
- Added a new internal regression test in `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command_test.go` proving that setting the app name to `wezen` makes the parser honor:
  - `WEZEN_BOT_TOKEN`
  - `WEZEN_APPLICATION_ID`
- Added the public option surface in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go`
- Updated `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command.go` so public command builders accept command options, including `WithAppName(...)`.
- Added a public-package test in `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go` proving the public API accepts the option and still embeds correctly into a downstream Cobra root.
- Updated `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go` to dogfood the option explicitly with `WithAppName("discord")`.
- Ran:
  - `gofmt -w internal/botcli/options.go internal/botcli/command.go internal/botcli/command_test.go pkg/botcli/options.go pkg/botcli/command.go pkg/botcli/command_test.go cmd/discord-bot/root.go`
  - `go test ./internal/botcli ./pkg/botcli ./cmd/discord-bot ./...`

### Why

The `botcli` package is supposed to be embeddable by other apps, and other apps should not be forced to adopt the `DISCORD_*` env prefix convention. In Glazed, the env prefix comes from the parser app name. So `WithAppName(...)` is the smallest meaningful public option that turns a hardcoded standalone-app assumption into a downstream-friendly behavior.

### What worked

- The parser-level regression test confirms the behavior change is real, not just cosmetic: when the app name is `wezen`, the env middleware reads `WEZEN_BOT_TOKEN` and `WEZEN_APPLICATION_ID`.
- The public package can now surface the option cleanly without exposing the parser internals.
- The standalone root command now exercises the public option path directly.

### What didn't work

- N/A

### What I learned

- This was a good example of why the Track B work benefits from extracting behavior one seam at a time. The command tree was already public, but one small hardcoded parser detail still made it less reusable than it looked.
- The actual env-prefix behavior lives lower than I first wanted to touch; the right fix was not a public-only wrapper, but internal option plumbing plus a public wrapper over that plumbing.

### What was tricky to build

- The main subtlety was the API shape. I wanted a public option without exploding the surface area, so I kept it to `WithAppName(...)` and threaded that one value down to the existing parser config builder.
- Another subtle point was choosing the right regression test. Executing `--help` does not exercise env middleware, so I used the same parser-level route as the existing env test and added a non-default-prefix variant.

### What warrants a second pair of eyes

- Whether `WithAppName(...)` should remain the only public parser-related option for a while, or whether future slices should expose a broader parser/runtime config object.
- Whether we should eventually update `internal/config.Validate()` so its missing-env-variable error messages can reflect non-default prefixes when the app name is customized.

### What should be done in the future

- Perform real manual validation before adding more Track B surface area.
- After validation, continue with the next public-configurability seam only if the current extracted API feels solid in practice.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/options.go` and `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command.go`.
- Review the new env-prefix regression coverage in `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command_test.go`.
- Then review the public surface in `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go` and `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command.go`.
- Validate with:
  - `go test ./internal/botcli ./pkg/botcli ./cmd/discord-bot ./...`

### Technical details

- Glazed derives the env prefix from `strings.ToUpper(AppName)`, so `WithAppName("wezen")` makes the parser look for `WEZEN_*` variables.
- The public package keeps the default app name at `discord`, so standalone behavior remains unchanged unless a downstream app opts into a custom value.
- The public command wrappers now accept command options and forward them into the internal builder, which keeps the implementation centralized while making the behavior configurable.

## Step 11: Do real manual validation before moving on

After the `WithAppName(...)` slice was implemented and covered by unit tests, I paused to validate the behavior in real command executions rather than immediately continuing Track B. The goal was not to achieve a successful Discord connection with fake credentials, but to prove that env-backed credentials were actually being accepted in both the standalone app and a downstream-style embedding scenario.

### What I did

Ran these manual validations from the repository root:

1. Standalone app command inventory still works after the recent `pkg/botcli` extraction:

```bash
GOWORK=off go run ./cmd/discord-bot --bot-repository ./examples/discord-bots bots list --output json
```

2. Standalone app dynamic run command still accepts the default `DISCORD_*` env prefix and gets far enough to load the JS bot before failing authentication with the fake token:

```bash
DISCORD_BOT_TOKEN=token-from-env \
DISCORD_APPLICATION_ID=app-from-env \
GOWORK=off timeout 15s \
  go run ./cmd/discord-bot --bot-repository ./examples/discord-bots bots ui-showcase run
```

3. Downstream-style app using the new public `pkg/botcli.WithAppName("wezen")` accepts the custom `WEZEN_*` env prefix and also gets far enough to load the JS bot before failing authentication with the fake token:

```bash
WEZEN_BOT_TOKEN=token-from-custom-env \
WEZEN_APPLICATION_ID=app-from-custom-env \
GOWORK=off timeout 15s \
  go run ./tmp-manual-botcli-app-XXXX.go bots ui-showcase run
```

(The downstream app was a tiny temporary Cobra root that mounted `pkg/botcli.NewCommand(bootstrap, pkg/botcli.WithAppName("wezen"))`.)

### What worked

- `bots list --output json` still worked through the current branch after the merge and the public package extraction.
- The standalone app command with `DISCORD_*` env vars did **not** fail with `missing required environment variables`; instead it loaded the JavaScript bot implementation and then failed at Discord authentication, which is the expected behavior with a fake token.
- The downstream-style app with `WithAppName("wezen")` behaved the same way using `WEZEN_*` env vars: it loaded the bot and only failed later at Discord authentication.

### Why this matters

This is the most useful kind of validation at this stage. It proves the env-prefix behavior is not just passing parser-level unit tests; it is effective in actual command execution paths for both:
- the standalone `discord-bot` app, and
- a downstream embedding app using the public `pkg/botcli` package.

### What should be done next

Only after this checkpoint does it make sense to continue Track B. The next slice should probably target a deeper public-behavior seam such as runtime-factory/configurability or moving more scanning/host-managed-run ownership into `pkg/botcli`.

## Step 12: Add runtime-module customization to the public botcli package

With the bootstrap layer, public Cobra entrypoint, and `WithAppName(...)` in place, the biggest remaining asymmetry between Track A and Track B was runtime extensibility. The single-bot public package could already inject custom runtime modules, but the repo-driven public `pkg/botcli` path still assumed the fixed built-in runtime. This slice closes that gap by making custom runtime module registrars flow through the repo-driven command system too.

The key realization here was that repo-driven runtime customization is not just about the final `run` command. It affects three different places:
1. bot discovery/inspection (`InspectScript`) because top-level `require("app")` must work while discovering bots,
2. ordinary jsverbs invocation (`status`, etc.), and
3. host-managed `run` construction for Discord bot scripts.

### Prompt Context

**User prompt (verbatim):** "do the whole slice"

**Assistant interpretation:** Implement the complete Track B runtime-customization slice rather than stopping at partial plumbing.

**Inferred user intent:** The user wants one coherent chunk of progress: public API, internal plumbing, tests, and validation together.

### What I did

- Added shared test helpers for custom-module bot repositories in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/test_helpers_test.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/test_helpers_test.go`
- Extended internal command options in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/options.go`
  with:
  - `WithRuntimeModuleRegistrars(...)`
  - conversion into `jsdiscord.HostOption`
- Updated discovery/inspection plumbing:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/bootstrap.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/jsdiscord/descriptor.go`
  so bot discovery can inspect scripts that require custom modules.
- Updated command helpers:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/list_command.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/help_command.go`
  to use the customized discovery path.
- Updated ordinary jsverbs execution in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/jsverbs_invoker.go`
  so normal discovered verbs get the extra runtime registrars too.
- Updated the host-managed run path in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/bot_run_command.go`
  so the actual bot runtime receives the same host options.
- Exposed the public option in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go`
- Added regression tests in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command_test.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go`
- Ran:
  - `gofmt -w ...`
  - `go test ./internal/botcli ./pkg/botcli ./pkg/framework ./cmd/discord-bot ./internal/jsdiscord ./...`
- Performed a real downstream-style manual validation by creating a tiny temporary Cobra app that mounted `pkg/botcli.NewCommand(...)` with `WithRuntimeModuleRegistrars(...)`, pointed it at a temp bot repository whose script calls `require("app")`, and executed:
  - `bots custom-module-bot status --output json`

### What worked

- Discovery now succeeds for bot scripts that require a custom module, as long as the registrar is provided.
- Ordinary jsverbs execution (`status`) now also sees the custom runtime module.
- The public package test and the manual downstream app both showed the same success case: the custom-module bot was discovered and the `status` verb returned structured data from `require("app")`.
- The manual run produced JSON proving the custom module was actually active:
  - `{"active": true, "module": "app", "description": "manual custom module works"}`

### What didn't work

- There were no code-level failures in the slice itself, but one manual-validation shell command initially failed because I passed temporary-file environment variables incorrectly in a one-off shell/python snippet. I reran the validation with a simpler shell-only setup and it worked.

### Why this mattered

Before this change, `pkg/botcli` looked public but still could not support one of the most important embedding requirements: top-level custom `require()` modules in repo-driven bot scripts. After this slice, the public repo-driven path is much closer to parity with the single-bot public path.

### What I learned

- Runtime customization for the repo-driven path is more cross-cutting than for the single-bot path, because discovery itself is runtime-backed.
- The right abstraction boundary here is not just a public option on `pkg/botcli`; it is consistent propagation through discovery, verb invocation, and run construction.

### What was tricky to build

- The subtle part was remembering that `InspectScript()` loads the script, so custom-module support had to be threaded into discovery before any command tree could even exist.
- Another subtlety was keeping internal and public options aligned without duplicating too much logic. I handled that the same way as the `WithAppName(...)` slice: public options convert into internal command options, and the implementation remains centralized internally.

### What warrants a second pair of eyes

- Whether the next runtime-related public seam should be `WithRuntimeFactory(...)` rather than just more registrar-based hooks.
- Whether we should start moving more of the scanning/registration internals physically into `pkg/botcli` now that the behavioral seams are becoming public and stable.

### What should be done in the future

- Decide whether to add a broader runtime factory hook next, or move to the next public-behavior extraction slice (entrypoint-only scan policy / host-managed run ownership).
- Add a durable downstream example app combining `pkg/framework` and `pkg/botcli` once the remaining Track B seams settle.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/options.go` and `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command.go`.
- Then review the three runtime touchpoints:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/bootstrap.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/jsverbs_invoker.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/bot_run_command.go`
- Review the public option surface in `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go`.
- Validate with:
  - `go test ./internal/botcli ./pkg/botcli ./pkg/framework ./cmd/discord-bot ./internal/jsdiscord ./...`

### Technical details

- `WithRuntimeModuleRegistrars(...)` is now available on both the internal and public botcli command builders.
- Discovery uses `jsdiscord.InspectScript(..., hostOpts...)`, so top-level custom-module imports work while building the command tree.
- Ordinary jsverbs invocation builds a runtime with the built-in Discord registrar plus any custom registrars.
- Host-managed `run` forwards the same host options into `bot.NewWithScript(...)`, which keeps the final live runtime consistent with discovery and jsverb invocation.

## Step 13: Add a durable downstream example app that combines both public layers

After the recent Track B work, the public API surface was finally broad enough to justify a durable combined example instead of more isolated feature demos. The missing piece was a concrete downstream app that uses both extracted packages in one process: the simple explicit-bot path for one built-in bot, and the optional repo-driven `botcli` path for discovered multi-bot workflows.

This slice implements exactly that and uses it to make the intended public split explicit in the repo docs.

### Prompt Context

**User prompt (verbatim):** "ok, continue."

**Assistant interpretation:** Take the next sensible implementation slice now that the merge is complete.

**Inferred user intent:** The user wants the framework extraction to keep advancing in practical increments, with examples that make the public package story legible and testable.

### What I did

- Added a new combined downstream example app:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/main.go`
- Added its README:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md`
- Added a built-in explicit bot script used only by that app:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/builtin-bot/index.js`
- The app now does both:
  - mounts repo-driven bots under `bots` using `pkg/botcli`
  - exposes `run-builtin` that starts one explicit built-in bot using `pkg/framework`
- Updated docs in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/discord-bots/README.md`
- Ran:
  - `gofmt -w examples/framework-combined/main.go`
  - `go test ./examples/framework-combined ./pkg/framework ./pkg/botcli ./cmd/discord-bot ./...`
  - `go run ./examples/framework-combined bots list --output json`

### What worked

- The combined example compiles cleanly as part of the normal repo test run.
- The repo-driven `bots` subtree worked in a real run through the combined app.
- The docs are now much clearer about the intended split:
  - `pkg/framework` = simple explicit built-in bot path
  - `pkg/botcli` = optional repo-driven multi-bot path

### Why this mattered

Until this slice, the public split existed in code and in my reasoning, but not yet as a stable example someone could point at and imitate. The new combined example is the first concrete downstream-shaped app that shows how both public layers fit together without relying on the standalone `cmd/discord-bot` binary.

### What I learned

- The combined app does not need to be fancy to be useful. A single `run-builtin` command plus a mounted `bots` subtree is enough to make the architecture legible.
- This example is also a good integration checkpoint: if it keeps compiling and `bots list` keeps working, then the public package split remains coherent.

### What was tricky to build

- The main judgment call was how much app behavior to include. I deliberately kept it small: one explicit built-in bot script and the public `bots` subtree. Adding more feature flags or extra runtime hooks here would have made it a demo of everything instead of a clear architecture example.
- Another small choice was how to resolve repositories. I used `pkg/botcli.BuildBootstrap(rawArgs, WithDefaultRepositories("examples/discord-bots"))` so the example mirrors the real pre-scan pattern while staying simple.

### What warrants a second pair of eyes

- Whether the combined example should eventually grow one more variant that also uses `WithRuntimeModuleRegistrars(...)` on the `pkg/botcli` side, or whether that would muddy the core conceptual split.
- Whether the top-level README now has enough information to treat the integration/positioning task as done. I marked it done because the split is explicitly stated and backed by a concrete example, but that is worth sanity-checking.

### What should be done in the future

- Continue Track B with the remaining behavior-ownership extraction work (scan policy / host-managed run semantics / potentially runtime factory).
- Keep the combined example compiling and use it as a regression checkpoint for future public API changes.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/main.go`.
- Then review the docs in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
- Validate with:
  - `go test ./examples/framework-combined ./pkg/framework ./pkg/botcli ./cmd/discord-bot ./...`
  - `go run ./examples/framework-combined bots list --output json`

### Technical details

- `run-builtin` uses `pkg/framework` directly and does not depend on repository discovery.
- The `bots` subtree is mounted through `pkg/botcli.NewCommand(...)` and uses `pkg/botcli.BuildBootstrap(...)` with a default repository rooted at `examples/discord-bots`.
- The built-in bot script receives a small runtime config map (`mode`, `source`) so it is easy to confirm the explicit built-in path is really the `pkg/framework` path.

## Step 14: Complete slices A/B/C for public botcli ownership and runtime factory support

At this point the public `pkg/botcli` package was still half-wrapper, half-owner. The next major goal was to make it actually own the remaining high-value behavior that users of the public package care about: scanning policy, host-managed run semantics, and a broader runtime factory hook.

The user explicitly asked to add tasks and do the whole A/B/C set, so I treated this as one coherent slice instead of three separate mini-slices.

### Prompt Context

**User prompt (verbatim):** "add tasks and do A B C"

**Assistant interpretation:** Update the task list with the specific A/B/C breakdown and then implement all three slices rather than deferring part of the plan.

**Inferred user intent:** The user wants the remaining public-ownership work to become concrete, both in planning docs and in code, and wants the public `pkg/botcli` package to become substantially more self-owned in one go.

### What I did

#### Planning
- Updated the framework task list to add explicit A/B/C sub-slices for:
  - public scan-policy ownership
  - public host-managed run ownership
  - public runtime-factory support

#### Slice A — public scan-policy ownership
- Added `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/discover.go` with public ownership of:
  - real bot entrypoint discovery only
  - `IncludePublicFunctions: false` scanning
  - `AbsPath` preservation for scanned files
  - public `ResolveBot(...)`
- This moved the practical scan policy used by the public command builder out of the internal package.

#### Slice B — public host-managed run ownership
- Added `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/run_description.go`
- Added `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/runtime_helpers.go`
- Added `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go`
- Reworked `pkg/botcli.NewBotsCommand(...)` / `NewCommand(...)` so the public package now owns:
  - explicit `__verb__("run")` as a `BareCommand`
  - synthetic `run` fallback when a bot has no explicit run metadata
  - both `bots <bot> run` and `bots run <bot>` shapes
  - public list/help command wiring
- Kept the standalone app on the public package path, so `cmd/discord-bot` continues to dogfood the extracted behavior.

#### Slice C — broader runtime-factory hook
- Expanded `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go`
- Added `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/runtime_factory.go`
- Added `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/invoker.go`
- Exposed:
  - `WithRuntimeFactory(...)`
  - `RuntimeFactory`
  - `RuntimeFactoryFunc`
  - `HostOptionsProvider`
- The factory now affects ordinary jsverbs runtime creation, and if it implements `HostOptionsProvider`, it can also contribute host options used for discovery and host-managed bot runs.

#### Tests
- Strengthened public package tests in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/test_helpers_test.go`
- Added public regressions for:
  - no leaked helper verbs from bot libraries
  - both run command shapes
  - custom runtime-module registrars
  - custom runtime-factory behavior with discovery/run host-option parity
- Ran:
  - `go test ./pkg/botcli ./cmd/discord-bot ./internal/botcli ./...`

#### Real validation
- Validated both run shapes through the combined downstream example:
  - `go run ./examples/framework-combined bots ui-showcase run --help`
  - `go run ./examples/framework-combined bots run ui-showcase --help`
- Validated a temporary downstream app using `WithRuntimeFactory(...)` against a temp repo whose bot script requires `require("app")`.
- That manual run returned structured JSON successfully:
  - `{ "active": true, "module": "app", "description": "manual factory works" }`

### What worked

- The public package now owns much more of the behavior that had previously lived only in the internal package.
- The missing public regressions are now in place for helper-leak prevention and both run command shapes.
- The runtime factory hook works in a real downstream-style app, not just in tests.

### What didn't work

- The initial pass on the public tests had a purely test-local import collision in `pkg/botcli/command_test.go` (`require` package name collision). Fixing the import aliases resolved it immediately.

### Why this mattered

This slice turns `pkg/botcli` from a mostly public façade into a package that genuinely owns its important public behavior. That reduces the gap between “public API” and “internal implementation with public wrappers,” especially for the scan/run behaviors downstream embedders actually depend on.

### What I learned

- The scan policy and host-managed run semantics fit naturally together in the public package because they are both part of how the public command tree comes into existence.
- A runtime factory becomes more useful when it can influence both ordinary jsverb invocation and host-backed discovery/run behavior; otherwise it feels incomplete.

### What was tricky to build

- The biggest subtlety was making the runtime-factory hook broad enough to be useful without over-designing it. I solved that by separating two concerns:
  - `RuntimeFactory` for ordinary verb runtime creation
  - optional `HostOptionsProvider` so the same factory can influence discovery and host-managed runs
- Another subtle point was keeping the public types stable while moving more real behavior into the public package. Using the existing public aliases for `Bootstrap`, `Repository`, and `DiscoveredBot` let me do that incrementally.

### What warrants a second pair of eyes

- Whether the current `RuntimeFactory` / `HostOptionsProvider` split is the right long-term public API, or whether a future refactor should fold those into a single richer factory type.
- Whether the remaining internal package should eventually be further reduced now that the public command builder owns most of the behavior downstream users care about.

### What should be done in the future

- Decide whether the next step is to keep extracting more implementation out of `internal/botcli`, or to pause and let the now-broader public API settle.
- Potentially add one more durable example using `WithRuntimeFactory(...)` if that becomes a core recommended pattern rather than an advanced escape hatch.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/discover.go` and `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go`.
- Then review `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go`, `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/runtime_factory.go`, and `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/invoker.go`.
- Finally review the strengthened public regressions in `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go`.

### Technical details

- Public scan policy now discovers only real bot entrypoints and always scans with `IncludePublicFunctions: false`.
- Public command construction now synthesizes host-managed run commands itself instead of relying on the internal command builder.
- `WithRuntimeFactory(...)` customizes ordinary jsverbs runtime creation; if the factory also implements `HostOptionsProvider`, its host options are applied to discovery and host-managed bot runs too.

## Step 15: Stabilization/docs pass plus review deliverables

After the A/B/C public-ownership slice, the user explicitly asked to pause and assess the result rather than immediately pushing deeper into implementation. That was the right move. The code now has enough public surface area that the key question is no longer “can we extract this?” but “is the design good enough, and where is the cleanup burden now concentrated?”

So this step was intentionally documentation-heavy: first tighten the public docs, then write two review documents from different angles — one for cleanup/code quality and one for design judgment.

### Prompt Context

**User prompt (verbatim):** "1. 

for 2.  make a detailed code review report in the ticket. focus on deprecation / unclear code / bloated files and packages / backwards compatibility (unnecessary) and legacy wrappers (clean cut)

then for 3. make a review document (since you already did the code review at this oint) that allows me to make a judgment and see where we are at in terms of design.

Upload 2 and 3 to remarkable once written"

**Assistant interpretation:** Do a stabilization/docs pass first, then write a detailed code-review cleanup report and a separate design-state review document, and publish those review docs to reMarkable.

**Inferred user intent:** The user wants an evaluation checkpoint before more code work: one document to show what should be cleaned up, and one document to show whether the overall framework extraction is on the right design track.

### What I did

- Performed a stabilization/docs pass by updating:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md`
- The doc updates now explicitly call out the public `pkg/botcli` option surface:
  - `WithAppName(...)`
  - `WithRuntimeModuleRegistrars(...)`
  - `WithRuntimeFactory(...)`
- Gathered file inventory and evidence with commands such as:
  - `find pkg/botcli internal/botcli examples/framework-combined -maxdepth 3 -type f | sort`
  - `wc -l pkg/botcli/*.go internal/botcli/*.go | sort -nr | head -n 40`
  - `nl -ba <file> | sed -n '<range>'`
- Wrote the detailed code-review report at:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/reference/02-public-botcli-code-review-cleanup-report.md`
- Wrote the design judgment document at:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/design-doc/02-framework-extraction-design-review-and-decision-guide.md`
- Ran:
  - `docmgr doctor --ticket DISCORD-BOT-FRAMEWORK --stale-after 30`
- Doctor initially warned about missing topic vocabulary (`framework`, `embedding`), so I fixed that with:
  - `docmgr vocab add --category topics --slug framework --description 'Reusable framework APIs and embedding-oriented package design'`
  - `docmgr vocab add --category topics --slug embedding --description 'Embedding one package or runtime into another host application'`
- Reran doctor successfully after the vocabulary fix.
- Prepared the two review docs for reMarkable upload.

### What worked

- The two review docs ended up complementary instead of redundant:
  - the code-review report focuses on duplication, deprecation candidates, unclear code, bloated files, and legacy wrappers,
  - the design review focuses on whether the public split is coherent and what judgment to make about the current design state.
- `docmgr doctor` passed cleanly after the vocabulary fix.
- The stabilization/docs pass was enough to make the new public option surface easier to evaluate without growing the code surface further.

### What didn't work

- `docmgr doctor` initially failed with unknown topic vocabulary on the ticket index for `framework` and `embedding`. That was a metadata problem, not a content problem, and adding the two missing slugs resolved it.

### What I learned

- The architecture is now strong enough that the highest-value review question is not “should we keep extracting?” but “should we now optimize for cleanup instead of more features?”
- Writing the code-review report and the design review back-to-back made the repo’s current state much clearer: the design has become coherent, but the duplication/transition burden is now the biggest risk.

### What was tricky to build

- The hardest part was keeping the two review docs distinct. It would have been easy for the design review to become just another list of code smells. I avoided that by making the code-review document issue-centric and file-backed, and making the design review explicitly judgment-centric.
- Another subtle point was constraining the stabilization/docs pass. The new public runtime-factory hook could easily have triggered another feature pass, but the user asked for evaluation, not more implementation, so I kept the docs focused on clarifying the existing public API.

### What warrants a second pair of eyes

- Whether the cleanup recommendation in the code-review report is too aggressive, especially around making a cleaner cut against `internal/botcli` duplication.
- Whether the design review strikes the right balance between “the design is ready enough to trust” and “the codebase still needs a cleanup/stabilization phase before calling the extraction finished.”

### What should be done in the future

- Use the two new review docs to decide whether the next work should be:
  - stabilization/cleanup only,
  - one more narrow public API refinement, or
  - a longer pause before further extraction.
- If implementation continues, bias toward deletion/reduction instead of adding more feature surface.

### Code review instructions

- Read the code-review report first:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/reference/02-public-botcli-code-review-cleanup-report.md`
- Then read the design judgment guide:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/design-doc/02-framework-extraction-design-review-and-decision-guide.md`
- If you want the shortest code path after reading those docs, inspect:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/discover.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/bootstrap.go`

### Technical details

- The docs pass now explicitly explains the three main public `pkg/botcli` customization hooks.
- The code-review report is cleanup-biased and evidence-first.
- The design review is decision-biased and intended to help a human reviewer decide where the extraction stands and what kind of work should happen next.

## Step 16: Remove the first public aliases and wrapper constructors

The user’s follow-up made the direction explicit: stop treating backwards compatibility as something to preserve and make a clean cut instead. I started with the public `pkg/botcli` surface because it was the easiest place to reduce ambiguity without changing the core runtime model.

This slice intentionally removed only the first layer of wrapper behavior: the public package now owns its model types directly, and first-party callers now use the canonical `NewBotsCommand(...)` constructor rather than the panic-based `NewCommand(...)` helper. That makes the next compatibility-removal slices cleaner because the app and examples already speak the public API more directly.

### Prompt Context

**User prompt (verbatim):** "yes, add tasks and do it task by task, committing at appropriate intervals, keeping a diary.

We don't need backwards compatibility, so we can remove all aliases and backwawrds wrappers. do them all."

**Assistant interpretation:** Add explicit cleanup tasks to the framework ticket, then implement the compatibility-removal work in focused slices with real commits and diary updates after each slice.

**Inferred user intent:** The user wants a deliberate cleanup pass that follows through on the review recommendations instead of keeping transitional wrappers around for convenience.

**Commit (code):** d6b09d6 — "Make public botcli own its API types and constructor"

### What I did

- Added a new cleanup section to:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/tasks.md`
- Added a new public model file:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/model.go`
- Moved public ownership of:
  - `BotRepositoryFlag`
  - `Bootstrap`
  - `Repository`
  - `DiscoveredBot`
  into `pkg/botcli` directly instead of aliasing `internal/botcli`
- Updated:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap.go`
  so it no longer imports or aliases the internal package
- Removed the panic-based public wrapper `NewCommand(...)` from:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go`
- Switched first-party callers/tests to the canonical constructor:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/main.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go`
- Also removed explicit `WithAppName("discord")` usage from first-party roots because the default already applies.
- Ran:
  - `gofmt -w pkg/botcli/model.go pkg/botcli/bootstrap.go pkg/botcli/commands_impl.go pkg/botcli/command_test.go cmd/discord-bot/root.go examples/framework-combined/main.go`
  - `go test ./pkg/botcli ./cmd/discord-bot ./examples/framework-combined ./...`

### What worked

- The public package detached cleanly from the internal model types without forcing other runtime changes.
- Removing `NewCommand(...)` was straightforward once the callers/tests were switched over.
- The explicit default `WithAppName("discord")` plumbing really was just noise in first-party code; removing it did not affect behavior.

### What didn't work

- N/A in this slice. The code changes compiled and tested cleanly on the first pass.

### What I learned

- The public alias removal is a good first step because it improves ownership clarity immediately without forcing the larger deletion of `internal/botcli` in the same commit.
- Once the panic wrapper is gone, the remaining public surface already feels more honest: constructor errors are now handled where commands are mounted.

### What was tricky to build

- The subtle part was keeping the slice narrow enough. It would have been easy to remove the legacy `bots run <bot>` shape at the same time, but that would have mixed public API cleanup with CLI-behavior cleanup and made it harder to review. I kept this step focused on ownership and wrapper removal only, leaving the command-shape cut for the next slice.

### What warrants a second pair of eyes

- Whether any downstream-facing docs outside the active framework ticket still present `NewCommand(...)` as the public embedding API and should now be updated or explicitly treated as historical.
- Whether we want to keep the placeholder `pkg/botcli/command.go` file around once the constructor surface now lives fully in `commands_impl.go`.

### What should be done in the future

- Next remove the legacy `bots run <bot>` compatibility path and its compatibility-specific helper code/tests/docs.
- After that, delete the duplicated `internal/botcli` implementation and move any still-needed fixtures into the public package.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/model.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go`
- Then verify the first-party callers:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/main.go`
- Validate with:
  - `go test ./pkg/botcli ./cmd/discord-bot ./examples/framework-combined ./...`

### Technical details

- `pkg/botcli` now owns its public model types directly instead of exporting aliases to `internal/botcli`.
- `NewBotsCommand(...)` is now the only public constructor; the panic-based wrapper is gone.
- First-party code now relies on the default app name implicitly instead of redundantly passing `WithAppName("discord")`.

## Step 17: Remove the legacy `bots run <bot>` compatibility path

With the public aliases gone, the next clean cut was CLI behavior itself. The user explicitly said backwards compatibility is not needed, so I removed the old nested compatibility path instead of keeping two equivalent ways to start a named bot.

I changed the public command builder, removed the compatibility-only helper, updated the regression to assert the canonical form only, and then swept the operator-facing docs and error text so the repo now consistently points readers at `bots <bot> run`.

### Prompt Context

**User prompt (verbatim):** (same as Step 16)

**Assistant interpretation:** Continue the cleanup by removing behavioral compatibility, not just public API wrappers.

**Inferred user intent:** The user wants one clear command shape rather than carrying compatibility syntax indefinitely.

**Commit (code):** 303d987 — "Drop legacy bots run compatibility path"

### What I did

- Removed the legacy alias-registration loop from:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go`
- Removed the compatibility-only helper from:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/run_description.go`
- Updated the public regression in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go`
  so it asserts the canonical `bots <bot> run` path and verifies that the old `bots run <bot>` invocation no longer surfaces run help
- Updated operator-facing docs and help text in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/doc/topics/discord-js-bot-api-reference.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/doc/tutorials/building-and-running-discord-js-bots.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/bot/bot.go`
- Tried to run:
  - `gofmt -w pkg/botcli/commands_impl.go pkg/botcli/run_description.go pkg/botcli/command_test.go README.md pkg/doc/topics/discord-js-bot-api-reference.md pkg/doc/tutorials/building-and-running-discord-js-bots.md internal/bot/bot.go`
- That failed immediately because I accidentally passed Markdown files to `gofmt`.
- Reran correctly with:
  - `gofmt -w pkg/botcli/commands_impl.go pkg/botcli/run_description.go pkg/botcli/command_test.go internal/bot/bot.go`
  - `go test ./pkg/botcli ./cmd/discord-bot ./examples/framework-combined ./...`

### What worked

- The public command builder no longer registers the old compatibility path.
- The canonical path still exposes the expected bot run flags/help.
- The old invocation no longer produces the run-help surface, which is enough to stop advertising it as a supported path.

### What didn't work

- I mistakenly ran `gofmt` on Markdown files. The exact command was:
  - `gofmt -w pkg/botcli/commands_impl.go pkg/botcli/run_description.go pkg/botcli/command_test.go README.md pkg/doc/topics/discord-js-bot-api-reference.md pkg/doc/tutorials/building-and-running-discord-js-bots.md internal/bot/bot.go`
- The relevant errors were:
  - `README.md:1:1: illegal character U+0023 '#'`
  - `pkg/doc/topics/discord-js-bot-api-reference.md:1:1: expected 'package', found '--'`
  - `pkg/doc/tutorials/building-and-running-discord-js-bots.md:1:1: expected 'package', found '--'`
- Re-running `gofmt` on only Go files resolved the issue immediately.

### What I learned

- The old `bots run <bot>` path was not failing hard in the help-oriented regression; instead Cobra/help handling ended up showing the broader `bots` help. That means the right regression is not “expect an error” but “the old invocation must no longer expose run-help content.”
- Once the docs are updated in the same slice, the clean cut feels much more coherent because the repository stops teaching two command shapes at once.

### What was tricky to build

- The subtle part was the regression behavior. I expected `bots run ui-showcase --help` to error, but it returned root help instead. The symptom was a failed `require.Error(...)` assertion in the new test. I adjusted the test to verify the important behavioral invariant instead: the old path no longer exposes run flags or the old compatibility usage string.

### What warrants a second pair of eyes

- Whether any additional public docs outside the top-level README and embedded help pages still present `bots run <bot>` as current guidance and should be cleaned up in a later docs sweep.
- Whether we want to add one explicit CLI regression at the standalone root level for the old path, or keep the package-level regression as sufficient coverage.

### What should be done in the future

- Delete the duplicated `internal/botcli` implementation now that the public path owns the behavior we still want to keep.
- Move any test fixtures that still live under `internal/botcli` into `pkg/botcli` before deleting the package.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/run_description.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go`
- Then inspect the operator-facing docs:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/doc/topics/discord-js-bot-api-reference.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/doc/tutorials/building-and-running-discord-js-bots.md`
- Validate with:
  - `go test ./pkg/botcli ./cmd/discord-bot ./examples/framework-combined ./...`
  - `go run ./examples/framework-combined bots ui-showcase run --help`
  - `go run ./examples/framework-combined bots run ui-showcase --help`

### Technical details

- `pkg/botcli` now exposes only the canonical `bots <bot> run` path.
- The compatibility-specific run-description helper is gone.
- The docs and missing-script error text now consistently point users at `bots <bot> run`.

## Step 18: Delete the duplicated `internal/botcli` package

This was the last major cleanup slice from the review. After the first two cuts, the standalone app and the public examples were already running entirely through `pkg/botcli`, so keeping `internal/botcli` around no longer bought anything except duplication risk.

I moved the one shared scanner fixture into the public package, switched the remaining tests/docs to that location, removed the entire internal package, and then validated the repository on the public-path-only tree. That is the clean-cut outcome the earlier review recommended.

### Prompt Context

**User prompt (verbatim):** (same as Step 16)

**Assistant interpretation:** Finish the cleanup by removing the duplicated internal botcli implementation instead of leaving a shadow package behind.

**Inferred user intent:** The user wants the review recommendations fully applied, not partially applied behind compatibility shims.

**Commit (code):** dbe3501 — "Delete duplicated internal botcli package"

### What I did

- Added the shared scanner fixture to:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/testdata/scanner-repo/demo-bot.js`
- Updated fixture lookups in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root_test.go`
- Updated public docs/examples that still referenced the internal package path:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md`
- Deleted the entire duplicated implementation under:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/`
- Ran:
  - `gofmt -w pkg/botcli/command_test.go cmd/discord-bot/root_test.go`
  - `go test ./...`
- Performed manual validation:
  - `go run ./cmd/discord-bot --bot-repository ./pkg/botcli/testdata/scanner-repo bots demo-bot status --output json`
  - `go run ./examples/framework-combined bots ui-showcase run --help`
  - `go run ./examples/framework-combined bots run ui-showcase --help`
- Ran:
  - `docmgr doctor --ticket DISCORD-BOT-FRAMEWORK --stale-after 30`
- Doctor initially warned that the ticket index still referenced deleted `internal/botcli/runtime.go`, so I replaced that related-file entry in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/index.md`
  with a current public file reference, then reran doctor successfully.

### What worked

- The repository no longer depends on `internal/botcli` anywhere in live code.
- The fixture move was enough to keep the downstream-style and standalone tests working without the old internal package.
- `go test ./...` passed cleanly after the deletion.
- The manual checks confirmed the intended behavior split:
  - canonical `bots <bot> run` still exposes run help,
  - legacy `bots run <bot>` now only falls back to generic `bots` help and is no longer a supported path.

### What didn't work

- `docmgr doctor` found one stale related-file entry after the deletion:
  - `missing_related_file — related file not found: internal/botcli/runtime.go`
- Fixing the ticket index and rerunning doctor resolved it.

### What I learned

- Once the public package genuinely owns the behavior, deleting the internal duplicate is less risky than continuously keeping the two implementations synchronized.
- The remaining cleanup after package deletion is mostly documentation/bookkeeping, not code architecture.

### What was tricky to build

- The tricky part was not the deletion itself but making sure fixture ownership moved first. If I had deleted `internal/botcli` before relocating `scanner-repo`, both the public package tests and the standalone root tests would have lost a convenient stable fixture. Moving the fixture into `pkg/botcli/testdata` first kept the final cut simple.
- Another small edge was ticket metadata: deleting a package can leave docmgr references stale even when the code is fully correct. Running doctor immediately after the deletion caught that before the bookkeeping drifted.

### What warrants a second pair of eyes

- Whether there are any broader historical design docs we still want to update to mention that `internal/botcli` is now gone rather than merely “to be reduced later.”
- Whether we want an additional public example specifically for `pkg/botcli/testdata`-style fixture usage, or whether the current tests/docs are enough.

### What should be done in the future

- N/A for the specific cleanup requested here. The major alias/wrapper/compatibility cuts are done.

### Code review instructions

- Start with the deletion-oriented diff around:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/testdata/scanner-repo/demo-bot.js`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root_test.go`
- Then confirm the package is actually gone:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/`
- Validate with:
  - `go test ./...`
  - `go run ./cmd/discord-bot --bot-repository ./pkg/botcli/testdata/scanner-repo bots demo-bot status --output json`
  - `go run ./examples/framework-combined bots ui-showcase run --help`

### Technical details

- The scanner fixture now lives under `pkg/botcli/testdata`, so the public package owns both the implementation and the fixture used to validate it.
- `internal/botcli` has been removed completely.
- `docmgr doctor` passes again after replacing the stale index reference to deleted internal botcli code.

## Step 19: Split the remaining botcli command "god file"

After the larger clean cut, the biggest remaining code-organization smell in `pkg/botcli` was still obvious: the command builder file had become the place where everything happened at once. That was no longer a design problem, but it was still a readability and maintenance problem.

So this slice was deliberately structural rather than behavioral. I split the old command builder into focused files, kept the public constructor in one place, and removed the placeholder file that used to exist only as a comment shell.

### Prompt Context

**User prompt (verbatim):** "do the cleanup pass on botcli. do it all, including the stronger when to use this.

Add missing tasks, then do task by task as usual"

**Assistant interpretation:** Add the remaining botcli cleanup work to the framework ticket, then execute it in focused slices with commits and diary updates after each slice.

**Inferred user intent:** The user wants the post-clean-cut botcli package to be not just functionally correct but also cleanly structured and clearly documented.

**Commit (code):** eee42c6 — "Split botcli command builder into focused files"

### What I did

- Added follow-up cleanup tasks to:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/tasks.md`
- Replaced the old oversized command file with focused command files:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_root.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_list.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_help.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_run.go`
- Removed:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command.go`
- Moved the shared bool flag helper into:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/runtime_helpers.go`
- Ran:
  - `gofmt -w pkg/botcli/command_root.go pkg/botcli/command_list.go pkg/botcli/command_help.go pkg/botcli/command_run.go pkg/botcli/runtime_helpers.go`
  - `go test ./pkg/botcli ./cmd/discord-bot ./examples/framework-combined ./...`

### What worked

- The split was behavior-preserving; tests passed without needing command-shape or runtime changes.
- `NewBotsCommand(...)` is easier to navigate now because the constructor and orchestration live in one file and the command implementations live elsewhere.
- Removing the placeholder `command.go` file makes the package structure feel much less half-refactored.

### What didn't work

- N/A in this slice. The structural split compiled and tested cleanly on the first pass.

### What I learned

- Once the public package is the only implementation left, file-structure cleanup becomes much more worthwhile because there is no second package mirroring the same concepts anymore.
- The placeholder-file smell mattered mostly for human navigation, but fixing it makes the package feel more intentional immediately.

### What was tricky to build

- The main subtlety was deciding how far to split. I wanted to reduce the “god file” pressure without exploding the package into too many tiny files. The compromise was to split by command concern — root, list, help, run — and keep the command-registration orchestration with the public constructor in `command_root.go`.

### What warrants a second pair of eyes

- Whether `command_root.go` should stay as the orchestration center or whether a later pass should further separate static command registration from discovered-command registration.
- Whether the current file naming is the clearest long-term layout, or whether a future `doc.go` / package-comment file should become the first file readers hit.

### What should be done in the future

- Finish the second remaining cleanup task by adding stronger guidance for when `WithRuntimeFactory(...)` is really needed versus `WithRuntimeModuleRegistrars(...)`.
- Then validate the cleaned package layout plus public docs/examples together.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_root.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_list.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_help.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_run.go`
- Validate with:
  - `go test ./pkg/botcli ./cmd/discord-bot ./examples/framework-combined ./...`

### Technical details

- `NewBotsCommand(...)` still owns the public constructor, but the individual command implementations no longer share a single oversized file.
- The placeholder `pkg/botcli/command.go` file is gone.
- Shared bool/runtime-config extraction now lives together in `runtime_helpers.go`.

## Step 20: Add stronger botcli runtime-customization guidance and validate the cleaned package

The last part of the requested botcli cleanup was not structural but communicative: make it much clearer when a downstream embedder should stop at simple hooks and when they should reach for the most powerful hook. That especially applied to `WithRuntimeFactory(...)`, which had become real and useful but still looked more mysterious than it needed to.

I handled that in three layers at once: package-level docs for `go doc`, stronger code comments on the relevant interfaces/options, and human-facing README/example guidance that gives a concrete decision ladder instead of just listing hooks.

### Prompt Context

**User prompt (verbatim):** (same as Step 19)

**Assistant interpretation:** Finish the botcli cleanup by making the advanced runtime hook understandable and by validating the final cleaned package/documentation state.

**Inferred user intent:** The user wants the package to be easy to judge and use correctly, not just internally tidy.

### What I did

- Added a new package doc at:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/doc.go`
- Strengthened the public comments in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/runtime_factory.go`
- Updated human-facing docs in:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md`
- The guidance now explicitly says:
  - use `WithAppName(...)` when only env-prefix behavior changes
  - use `WithRuntimeModuleRegistrars(...)` when scripts simply need extra native modules and default runtime construction is fine
  - use `WithRuntimeFactory(...)` only when runtime creation itself must change
  - implement `HostOptionsProvider` when the same customization should also affect discovery and host-managed bot runs
- Ran:
  - `gofmt -w pkg/botcli/doc.go pkg/botcli/options.go pkg/botcli/runtime_factory.go`
  - `go test ./pkg/botcli ./cmd/discord-bot ./examples/framework-combined ./...`
  - `go doc ./pkg/botcli | head -n 40`
  - `go run ./examples/framework-combined bots ui-showcase run --help`

### What worked

- The package overview renders cleanly through `go doc`, which is exactly the right proof that the new package-level guidance is discoverable to embedders.
- The decision ladder now appears in both code-facing and human-facing places, so readers do not have to reverse-engineer the intended order of hooks from the implementation.
- The cleaned package layout still validates after the doc pass.

### What didn't work

- N/A in this slice. The docs/comments pass validated cleanly without follow-up fixes.

### What I learned

- The strongest way to explain a powerful hook is usually by giving readers permission not to use it. Once the docs explicitly say “use module registrars first; use a runtime factory only when runtime creation itself must change,” the public surface looks much less over-engineered.
- `go doc` is a useful validation tool for package cleanup work because it confirms that the first thing embedders will read is actually coherent.

### What was tricky to build

- The tricky part was making the guidance specific enough to be useful without over-prescribing every future downstream use case. I solved that by framing the guidance as a "smallest hook first" decision ladder and giving concrete examples of what counts as “runtime creation itself must change” — module roots, require behavior, builder/runtime setup, and lifecycle details.

### What warrants a second pair of eyes

- Whether the new package doc is the right level of detail for `go doc`, or whether future downstream users would benefit from one dedicated advanced example of `RuntimeFactoryFunc` as well.
- Whether the README wording now strikes the right balance between encouraging the simple path and still advertising the advanced hook as intentionally supported.

### What should be done in the future

- N/A for the requested cleanup pass. The remaining package-level structural and guidance issues identified in the review have been addressed.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/doc.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md`
  - `/home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md`
- Validate with:
  - `go test ./pkg/botcli ./cmd/discord-bot ./examples/framework-combined ./...`
  - `go doc ./pkg/botcli | head -n 40`
  - `go run ./examples/framework-combined bots ui-showcase run --help`

### Technical details

- The public guidance now uses a deliberate “smallest hook first” model.
- `WithRuntimeFactory(...)` is explicitly documented as the advanced escape hatch for changing runtime creation itself.
- `HostOptionsProvider` is now documented as the bridge that keeps advanced runtime customization consistent across ordinary jsverb execution, discovery, and host-managed bot runs.
