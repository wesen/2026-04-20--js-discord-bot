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

### What to do next

1. **Phase 1**: Convert `bots list` to a `cmds.GlazeCommand` — straightforward, high value
2. **Phase 2**: Convert `bots run` to use `BuildCobraCommandFromCommand` with a generated schema per-bot
3. **Phase 3**: Add `__verb__`/`__section__`/`__package__` polyfills to `internal/jsdiscord/runtime.go`
4. **Phase 4**: Add a `jsverbs.ScanDir` pass over bot repositories to discover CLI verbs in bot scripts
5. **Write tests** for each phase
