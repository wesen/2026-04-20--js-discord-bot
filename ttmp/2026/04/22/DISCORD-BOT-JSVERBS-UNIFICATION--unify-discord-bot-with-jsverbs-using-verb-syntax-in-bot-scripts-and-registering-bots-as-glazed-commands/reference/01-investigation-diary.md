---
Title: Investigation diary
Ticket: DISCORD-BOT-JSVERBS-UNIFICATION
Status: active
Topics:
    - discord-bot
    - jsverbs
    - glazed
    - cli
    - bot-registration
    - command-discovery
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological log of the investigation into unifying discord-bot with jsverbs."
LastUpdated: 2026-04-22T18:00:00-04:00
WhatFor: "Track what was tried, what worked, what failed, and what to do next."
WhenToUse: "When resuming work on this ticket or when a new engineer needs to understand the investigation path."
---

# Investigation diary

## 2026-04-22 — Deep-dive into discord-bot architecture

### What was asked

Analyze how the discord-bot registers its JavaScript bot scripts, compare with jsverbs from go-go-goja, and determine if bot scripts can use `__verb__` syntax while remaining runnable as bots. Create a detailed document for an intern.

### What worked

1. **Explored the discord-bot JS API** by reading example scripts:
   - `examples/discord-bots/support/index.js` shows `defineBot`, `command`, `event`, `configure`
   - `examples/discord-bots/knowledge-base/index.js` shows `configure({ run: { fields: {...} } })`

2. **Traced the full Host lifecycle**:
   - `internal/jsdiscord/host.go:NewHost` → builds engine, loads script
   - `internal/jsdiscord/runtime.go:Loader` → registers `"discord"` module with `defineBot`
   - `internal/jsdiscord/runtime.go:defineBot` → creates `botDraft`, calls builder fn, returns bot object
   - `internal/jsdiscord/bot_compile.go:CompileBot` → extracts `describe`, `dispatchCommand`, etc.
   - `internal/jsdiscord/descriptor.go:descriptorFromDescribe` → parses `map[string]any` into `BotDescriptor`

3. **Understood the dispatch mechanism**:
   - `BotHandle.dispatchCommand` receives a `DispatchRequest` (rich context object)
   - Handler gets `ctx` with `ctx.args`, `ctx.discord`, `ctx.reply`, `ctx.edit`, `ctx.defer`
   - This is fundamentally different from jsverbs' one-shot function call

4. **Compared with jsverbs architecture**:
   - jsverbs: static scan (Tree-sitter), `__verb__` metadata, one-shot execution
   - discord-bot: runtime load (Goja execution), `defineBot` API, long-running event dispatch
   - Both use go-go-goja engine, but for completely different purposes

5. **Confirmed `__verb__` + `defineBot` coexistence is possible**:
   - Tree-sitter scans for `__verb__` calls at the AST level — it doesn't execute code
   - `defineBot` is a runtime API that executes when the script loads
   - A single file can have both; we just need no-op polyfills for `__verb__` in the Discord runtime

### What was tricky

- The `bots run` command uses `DisableFlagParsing: true` and manually parses ~200 lines of custom flag logic (`run_static_args.go`). This is a major anti-pattern when Glazed already provides all of this.
- The dynamic schema parsing in `run_dynamic_schema.go` creates a **throwaway** `cobra.Command` just to parse flags, then discards it. This is fragile and bypasses all of Glazed's help rendering.
- `bots list` and `bots help` print plain text instead of using Glazed's structured output pipeline.
- The Discord `ctx` object is much richer than jsverbs' parsed args — it includes Discord entity snapshots, API proxies, and response helpers. Unifying the handler signatures (Level C) is probably not worth the complexity.

### Commands run

```bash
# Explore discord-bot source
find /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli -name "*.go" | sort
find /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/jsdiscord -name "*.go" | sort
cat /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/discord-bots/support/index.js

# Search for RunSchema usage
rg -n "run:" /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/ -A 5

# Compare with jsverbs
cat /home/manuel/workspaces/2026-04-22/discord-bot-framework/go-go-goja/pkg/jsverbs/scan.go | head -50
cat /home/manuel/workspaces/2026-04-22/discord-bot-framework/go-go-goja/pkg/jsverbs/command.go | head -50
```

## 2026-04-22 — Implementation progress

### What changed

1. **Added jsverbs metadata polyfills to the Discord runtime** in `internal/jsdiscord/runtime.go`.
   - The runtime now defines no-op globals for `__package__`, `__section__`, `__verb__`, and `doc`.
   - This means bot scripts can contain jsverbs metadata without crashing when loaded through `require("discord")`.

2. **Added bot-repo jsverbs scanning** in `internal/botcli/jsverbs_scan.go`.
   - Important correction: scanning the entire repo root caused helper libraries under bot directories to be treated as verbs.
   - The wrapper now reuses `discoverScriptCandidates()` and scans **only bot entrypoint scripts**.
   - After scanning with `jsverbs.ScanSources`, it patches each `FileSpec.AbsPath` so host-managed commands know the real script file path.

3. **Added `botRunCommand` as a `cmds.BareCommand`** in `internal/botcli/bot_run_command.go`.
   - It decodes Discord credentials using `appconfig.FromValues`.
   - It builds a runtime config map from all parsed CLI values.
   - It launches the bot with `bot.NewWithScript(cfg, scriptPath, runtimeConfig)` and blocks on `<-ctx.Done()`.
   - This is the host-managed implementation of `__verb__("run")`.

4. **Added the field-name bridge** in `internal/botcli/field_name.go`.
   - `runtimeFieldInternalName()` converts kebab-case CLI flag names to snake_case JS config keys.
   - Example: `--db-path` → `ctx.config.db_path`.
   - Matching the jsdiscord naming behavior exactly mattered; the first implementation over-inserted underscores for consecutive capitals (`APIKey`), which the tests caught.

5. **Rewrote the `bots` command tree** in `internal/botcli/command.go`.
   - `bots list` is now a Glazed command (`listBotsCommand`).
   - `bots help <bot>` is now a Glazed command (`helpBotsCommand`).
   - Discovered jsverbs are registered under the command tree via `glazed_cli.AddCommandsToRootCommand`, which preserves parent commands inferred from filenames (e.g. `demo-bot status`, `demo-bot run`).
   - For standard verbs in bot scripts, the code now uses a custom `botVerbInvoker` so the runtime includes the Discord registrar. This fixed the key issue where `require("discord")` failed for non-run jsverbs like `status`.
   - The old manual parser files were removed: `run_static_args.go`, `run_dynamic_schema.go`, `run_help.go`.

6. **Updated root wiring** in `cmd/discord-bot/root.go`.
   - The root command now builds a default bootstrap from `DISCORD_BOT_REPOSITORIES`.
   - If the env var is unset, it falls back to `examples/discord-bots` for local/dev usage.
   - If the env var is set, the examples directory is **not** automatically appended.

### What worked

- `go test ./...` now passes across the whole repo.
- End-to-end manual checks succeeded:
  - `DISCORD_BOT_REPOSITORIES=internal/botcli/testdata/scanner-repo go run ./cmd/discord-bot bots list --output json`
  - `DISCORD_BOT_REPOSITORIES=internal/botcli/testdata/scanner-repo go run ./cmd/discord-bot bots help demo --output json`
  - `DISCORD_BOT_REPOSITORIES=internal/botcli/testdata/scanner-repo go run ./cmd/discord-bot bots demo-bot status --output json`
  - `DISCORD_BOT_REPOSITORIES=internal/botcli/testdata/scanner-repo go run ./cmd/discord-bot bots demo-bot run --help`

### What was tricky

- **Capturing Glazed output in tests**: `root.SetOut()` does not capture Glazed structured output because Glazed writes to `os.Stdout` directly. Tests had to temporarily redirect `os.Stdout`/`os.Stderr` with `os.Pipe()`.
- **Standard jsverbs in bot scripts initially failed**: `status` in `demo-bot.js` failed with `Invalid module` because the standard jsverbs runtime did not register the Discord module. The fix was a custom `botVerbInvoker` that builds an engine runtime with both `registry.RequireLoader()` and `jsdiscord.NewRegistrar(...)`.
- **Scanning too much source**: scanning whole bot repos inferred verbs from helper libraries (e.g. `knowledge-base/lib/reactions.js`), which then failed binding because of destructured parameters. Restricting scanning to `discoverScriptCandidates()` fixed this.
- **Preserving command parents**: directly calling `BuildCobraCommandFromCommand` and `root.AddCommand` flattened discovered verbs. Switching to `AddCommandsToRootCommand` preserved inferred parent commands (`demo-bot status`, `demo-bot run`).

### What to do next

1. Add a richer example bot script that uses the unified `__verb__("run")` + `defineBot` pattern in `examples/discord-bots/`.
2. Decide whether to migrate existing example bots from `configure({ run: ... })` to `__verb__("run", { fields: ... })`.
3. Add higher-level docs for the new usage pattern in `examples/discord-bots/README.md`.
4. Consider whether root-level `--bot-repository` flags should be added later, or whether the environment-variable bootstrap is sufficient.
