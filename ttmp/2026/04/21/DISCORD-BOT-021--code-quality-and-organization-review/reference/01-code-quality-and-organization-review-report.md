---
Title: Code Quality and Organization Review Report
Ticket: DISCORD-BOT-021
Status: active
Topics:
    - backend
    - go
    - javascript
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go
      Note: Main runtime quality hot spot; currently mixes compiler, dispatch, context, capability binding, and settlement logic
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_dispatch.go
      Note: Repetitive event-to-dispatch bridging layer
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_payloads.go
      Note: Large normalization file where payload complexity accumulates
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go
      Note: Live Discord session wrapper with repetitive handler shells and dead fallback interaction code
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/run_schema.go
      Note: Custom pre-parser and dynamic config parser with high contract-drift risk
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js
      Note: Canonical large example showing alias duplication and heavy manual interaction wiring
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/store.js
      Note: Example store layer that mixes schema, seeding, normalization, search, and repository responsibilities
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js
      Note: Appears unreferenced and stale relative to the current knowledge-base bot entrypoint
ExternalSources: []
Summary: Detailed code quality review of the repo with concrete organizational hot spots, stale-code findings, and cleanup sketches.
LastUpdated: 2026-04-21T07:35:00-04:00
WhatFor: Capture a maintainability-focused review of the repo for future cleanup planning and intern onboarding.
WhenToUse: Use when planning refactors, architectural cleanup, file splits, or example-bot ergonomics work.
---

# Executive summary

This repository has a solid architectural core, but it now shows the classic symptoms of a fast-moving system that successfully shipped multiple slices in sequence:

- several files have become **accumulation points**,
- some internal APIs are more map-based and implicit than they need to be,
- some fallback or legacy artifacts no longer match the current single-bot architecture,
- and the largest example bot is starting to expose the friction in the current authoring API.

The highest-leverage cleanup targets are:

1. `internal/jsdiscord/bot.go`
2. `internal/jsdiscord/host_payloads.go`
3. `internal/jsdiscord/host_dispatch.go`
4. `internal/bot/bot.go`
5. `internal/botcli/run_schema.go`
6. `examples/discord-bots/knowledge-base/`
7. stale example/registration artifacts that no longer match the current runtime model

## Overall judgment

### What is strong
- The repo has a clear product idea: one Go host, one selected JS bot, request-scoped Discord operations.
- The host/descriptor/runtime split is conceptually good.
- Example bots are doing real work and therefore reveal real API quality issues instead of toy-only issues.

### What is weak
- Too much behavior is concentrated in a few files.
- The Go↔JS boundary is often expressed as `map[string]any` rather than typed internal contracts.
- Some APIs appear supported in one layer while being dead, stale, or no-op in another.
- The biggest example bot is starting to encode framework friction instead of just business logic.

# Review method

I used a code-quality-first review strategy rather than a bug hunt.

## Focus areas
- file/package size hot spots
- large or conceptually overloaded modules
- stale or deprecated artifacts
- confusing APIs and implicit contracts
- duplicated or repetitive code with maintainability cost
- canonical example quality as a proxy for API ergonomics

## Explicit exclusion
I did **not** focus this report on the colleague’s in-flight work around commands/subcommands/message commands. The report instead focuses on the structural seams that will matter regardless of how that feature slice lands.

# Architecture map

```text
cmd/discord-bot
    ↓
internal/config
    ↓
internal/bot
    ↓
internal/jsdiscord/host + runtime + bot compiler
    ↓
examples/discord-bots/<name>/index.js
```

More concretely:

```text
Cobra / Glazed CLI
  ├── direct host commands: run / validate-config / sync-commands
  └── repository runner: bots list / help / run

Live Discord session wrapper
  └── attaches Discordgo handlers and forwards events/interactions

JS host bridge
  ├── loads one JS script
  ├── exposes require("discord")
  ├── compiles the bot definition
  ├── builds ctx objects
  └── normalizes response payloads back into Discordgo structs

Example bot
  └── uses command/event/component/modal/autocomplete/configure
```

# Findings

## 1. `internal/jsdiscord/bot.go` is the main “god file” in the live runtime

### Problem
`internal/jsdiscord/bot.go` has become the single biggest concentration point in the live runtime. It currently mixes at least five distinct responsibilities: bot-definition DSL building, callable handle compilation, request dispatch, promise settlement, and JS context/capability binding.

### Where to look
- `internal/jsdiscord/bot.go:66` — `type DiscordOps`
- `internal/jsdiscord/bot.go:97` — `type DispatchRequest`
- `internal/jsdiscord/bot.go:135` — `CompileBot(...)`
- `internal/jsdiscord/bot.go:238` — `(*BotHandle).dispatch(...)`
- `internal/jsdiscord/bot.go:395` — `(*botDraft).command(...)`
- `internal/jsdiscord/bot.go:564` — `(*botDraft).configure(...)`
- `internal/jsdiscord/bot.go:575` — `(*botDraft).finalize(...)`
- `internal/jsdiscord/bot.go:891` — `buildDispatchInput(...)`
- `internal/jsdiscord/bot.go:942` — `buildContext(...)`
- `internal/jsdiscord/bot.go:1027` — `discordOpsObject(...)`

### Example snippet
```go
func buildContext(vm *goja.Runtime, store *MemoryStore, input *goja.Object, kind, name string, metadata map[string]any) *goja.Object {
    ctx := vm.NewObject()
    setObjectField(vm, ctx, "args", input.Get("args"))
    setObjectField(vm, ctx, "options", input.Get("args"))
    setObjectField(vm, ctx, "values", input.Get("values"))
    ...
    _ = ctx.Set("store", storeObject(vm, store))
    _ = ctx.Set("log", loggerObject(vm, kind, name, metadata))
    ...
}
```

and later in the same file:

```go
func discordOpsObject(vm *goja.Runtime, ctx context.Context, ops *DiscordOps) *goja.Object {
    root := vm.NewObject()
    guilds := vm.NewObject()
    roles := vm.NewObject()
    threads := vm.NewObject()
    channels := vm.NewObject()
    messages := vm.NewObject()
    members := vm.NewObject()
    ...
}
```

### Why it matters
This file is not just large in a mechanical sense. It is large in a **conceptual** sense.

That matters because:

- a reviewer has to context-switch between “compile a bot,” “wait for promises,” and “bind Discord ops” in one scroll,
- adding one new runtime surface increases collision risk with unrelated concerns,
- and newcomers cannot easily tell which code defines the public JS API and which code is just transport glue.

### Cleanup sketch
Keep the package the same for now, but split the file by responsibility.

```text
internal/jsdiscord/
  bot_compile.go        // CompileBot, botDraft registration/finalize
  bot_dispatch.go       // BotHandle dispatch + settleValue + promise waiting
  bot_context.go        // DispatchRequest, buildDispatchInput, buildContext
  bot_store.go          // storeObject
  bot_ops.go            // DiscordOps + discordOpsObject
  bot_logging.go        // loggerObject + applyFields
```

Pseudocode for the shape:

```go
type DispatchRequest struct { ... }           // bot_context.go

type BotHandle struct { ... }                 // bot_compile.go / bot_dispatch.go

func CompileBot(...) (*BotHandle, error)      // bot_compile.go
func (h *BotHandle) dispatch(...)             // bot_dispatch.go
func buildContext(...) *goja.Object           // bot_context.go
func discordOpsObject(...) *goja.Object       // bot_ops.go
```

This is a low-risk structural cleanup because it preserves behavior while making ownership clearer.

## 2. The runtime boundary uses too much `map[string]any` instead of typed internal contracts

### Problem
The Go↔JS boundary understandably uses plain objects, but the internals continue to rely on `map[string]any` longer than necessary. That makes the code harder to refactor, harder to autocomplete, and easier to drift.

### Where to look
- `internal/jsdiscord/bot.go:97` — `DispatchRequest` fields are map-heavy
- `internal/jsdiscord/descriptor.go:115` — `descriptorFromDescribe(...)`
- `internal/jsdiscord/descriptor.go:153` through `344` — repeated `parse*Descriptor` helpers from raw maps
- `internal/jsdiscord/host_maps.go` — many functions produce anonymous map shapes used transitively elsewhere

### Example snippet
```go
type DispatchRequest struct {
    Name        string
    Args        map[string]any
    Values      any
    Command     map[string]any
    Interaction map[string]any
    Message     map[string]any
    Before      map[string]any
    User        map[string]any
    Guild       map[string]any
    Channel     map[string]any
    ...
}
```

### Why it matters
`map[string]any` is good at the *edge* of the system, but weaker as the *interior* representation.

The current approach increases:

- typo risk in keys,
- drift between maps produced in `host_maps.go` and maps expected in JS-facing docs/tests,
- reviewer difficulty when tracing shape changes,
- and the amount of ad hoc conversion code in parsing layers.

### Cleanup sketch
Do not remove plain-object export at the JS boundary. Instead, introduce typed internal structs and convert to maps at the last step.

```go
type DispatchEnvelope struct {
    Name        string
    Args        map[string]any
    Interaction InteractionSnapshot
    Message     MessageSnapshot
    User        UserSnapshot
    Guild       GuildSnapshot
    Channel     ChannelSnapshot
    ...
}

func (e DispatchEnvelope) ToJSMap() map[string]any { ... }
```

Likewise for descriptor parsing:

```go
type rawDescribeSnapshot struct {
    Metadata      map[string]any
    Commands      []map[string]any
    Events        []map[string]any
    Components    []map[string]any
    ...
}
```

This keeps the JS authoring API flexible without forcing the internal Go code to stay loosely typed forever.

## 3. `internal/jsdiscord/host_payloads.go` is a second major accumulation point

### Problem
`internal/jsdiscord/host_payloads.go` is the payload-normalization kitchen sink. It handles interaction responses, modal payloads, webhook edits, message sends, embeds, components, files, mentions, and more.

### Where to look
- `internal/jsdiscord/host_payloads.go:11` — `normalizedResponse`
- `internal/jsdiscord/host_payloads.go:22` — `normalizeResponsePayload(...)`
- `internal/jsdiscord/host_payloads.go:41` — `normalizeModalPayload(...)`
- `internal/jsdiscord/host_payloads.go:111` — `normalizeWebhookParams(...)`
- `internal/jsdiscord/host_payloads.go:167` — `normalizeChannelMessageEdit(...)`
- `internal/jsdiscord/host_payloads.go:184` — `normalizePayload(...)`
- `internal/jsdiscord/host_payloads.go:249` — `normalizeEmbedArray(...)`
- `internal/jsdiscord/host_payloads.go:364` — `normalizeComponents(...)`
- `internal/jsdiscord/host_payloads.go:595` — `normalizeAllowedMentions(...)`
- `internal/jsdiscord/host_payloads.go:625` — `normalizeFiles(...)`

### Example snippet
```go
func normalizePayload(payload any) (*normalizedResponse, error) {
    switch v := payload.(type) {
    case nil:
        return &normalizedResponse{}, nil
    case string:
        return &normalizedResponse{Content: v}, nil
    case map[string]any:
        ret := &normalizedResponse{}
        if content, ok := v["content"]; ok {
            ret.Content = fmt.Sprint(content)
        }
        ...
```

### Why it matters
This file is one of the places most likely to get harder every time a new Discord feature lands.

That matters because:

- new response kinds naturally pile into this layer,
- the normalization rules are hard to scan and hard to test locally in isolation,
- and a large “switch over all possible payload shapes” style discourages small, composable additions.

### Cleanup sketch
Split by payload concern and make the shared intermediate shape explicit.

```text
internal/jsdiscord/
  payload_model.go          // normalizedResponse + small shared types
  payload_message.go        // normalizePayload, normalizeMessageSend, edits
  payload_embeds.go         // normalizeEmbed, fields, authors, images
  payload_components.go     // rows, buttons, selects, modals, text inputs
  payload_files.go          // file/path/buffer normalization
  payload_mentions.go       // allowedMentions + references
  payload_autocomplete.go   // autocomplete choices
```

Also tighten the test layout to match the split:

```text
internal/jsdiscord/
  payload_message_test.go
  payload_components_test.go
  payload_files_test.go
```

## 4. `internal/jsdiscord/host_dispatch.go` repeats the same event-bridging pattern many times

### Problem
The host dispatch layer repeatedly assembles near-identical `DispatchRequest` objects for each event type. The fields differ, but the shape of the code is largely the same.

### Where to look
- `internal/jsdiscord/host_dispatch.go:11` — `DispatchReady(...)`
- `internal/jsdiscord/host_dispatch.go:30` — `DispatchGuildCreate(...)`
- `internal/jsdiscord/host_dispatch.go:102` — `DispatchMessageCreate(...)`
- `internal/jsdiscord/host_dispatch.go:190` — `DispatchReactionAdd(...)`
- `internal/jsdiscord/host_dispatch.go:221` — `DispatchReactionRemove(...)`
- `internal/jsdiscord/host_dispatch.go:251` — `DispatchInteraction(...)`

### Example snippet
```go
result, err := h.handle.DispatchEvent(ctx, DispatchRequest{
    Name:     "messageCreate",
    Message:  messageMap(message.Message),
    User:     userMap(message.Author),
    Guild:    guildMap(message.GuildID),
    Channel:  channelMap(message.ChannelID),
    Me:       currentUserMap(session),
    Metadata: map[string]any{"scriptPath": h.scriptPath},
    Config:   cloneMap(h.runtimeConfig),
    Discord:  buildDiscordOps(h.scriptPath, session),
    Reply:    responder.Reply,
    FollowUp: responder.FollowUp,
    Edit:     responder.Edit,
    Defer:    responder.Defer,
})
```

### Why it matters
The repetition itself is not catastrophic, but it creates three maintenance costs:

- adding a new shared field means editing many dispatch functions,
- event-specific differences are hidden inside large repeated blocks,
- and the same responder/config/metadata plumbing distracts from the actual event meaning.

### Cleanup sketch
Extract shared envelope builders.

```go
func (h *Host) baseRequest(session *discordgo.Session) DispatchRequest {
    return DispatchRequest{
        Me:       currentUserMap(session),
        Metadata: map[string]any{"scriptPath": h.scriptPath},
        Config:   cloneMap(h.runtimeConfig),
        Discord:  buildDiscordOps(h.scriptPath, session),
    }
}

func (h *Host) eventRequest(session *discordgo.Session, name string) DispatchRequest {
    req := h.baseRequest(session)
    req.Name = name
    req.Command = map[string]any{"event": name}
    return req
}
```

For responder-backed events:

```go
func (h *Host) withChannelResponder(req DispatchRequest, r *channelResponder) DispatchRequest {
    req.Reply = r.Reply
    req.FollowUp = r.FollowUp
    req.Edit = r.Edit
    req.Defer = r.Defer
    return req
}
```

This would make each event handler shorter and more semantic.

## 5. `internal/bot/bot.go` still carries repetitive handler shells and a dead fallback interaction path

### Problem
`internal/bot/bot.go` repeats similar guard-and-forward logic for each Discordgo handler, and `handleInteractionCreate(...)` still contains a fallback ping/echo implementation that appears unreachable under the current architecture.

### Where to look
- `internal/bot/bot.go:28` — `NewWithScript(...)`
- `internal/bot/bot.go:102` — `SyncCommands()`
- `internal/bot/bot.go:169` — `handleReady(...)`
- `internal/bot/bot.go:183` — `handleGuildCreate(...)`
- `internal/bot/bot.go:191` — `handleGuildMemberAdd(...)`
- `internal/bot/bot.go:233` — `handleMessageCreate(...)`
- `internal/bot/bot.go:294` — `handleInteractionCreate(...)`

### Example snippet
```go
func (b *Bot) handleInteractionCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
    if b.jsHost != nil {
        if err := b.jsHost.DispatchInteraction(context.Background(), session, interaction); err != nil {
            log.Error().Err(err).Msg("failed to dispatch interaction to javascript bot")
        }
        return
    }

    if interaction.Type != discordgo.InteractionApplicationCommand {
        return
    }
    ... ping/echo fallback ...
}
```

### Why it matters
`NewWithScript(...)` requires a non-empty script and loads a host before returning success, so the `b.jsHost == nil` fallback is no longer the architectural norm. Keeping the old path around tells the reader that the system still supports an older mode when it effectively does not.

The repetitive handler shells also make it harder to see which handlers are conceptually unique versus just transport wrappers.

### Cleanup sketch
First, remove the stale fallback branch after confirming no tests or operators rely on it.

Second, introduce small forwarding helpers.

```go
func (b *Bot) forward(name string, fn func() error) {
    if b.jsHost == nil {
        return
    }
    if err := fn(); err != nil {
        log.Error().Err(err).Str("event", name).Msg("failed to dispatch event to javascript bot")
    }
}
```

Then handlers become:

```go
func (b *Bot) handleGuildCreate(session *discordgo.Session, guild *discordgo.GuildCreate) {
    b.forward("guildCreate", func() error {
        return b.jsHost.DispatchGuildCreate(context.Background(), session, guild)
    })
}
```

That would not radically reduce LOC, but it would reduce boilerplate and clarify intent.

## 6. `internal/botcli/run_schema.go` is a clarity hot spot because it manually re-implements parsing around Cobra/Glazed

### Problem
The bot runner’s startup-config path currently relies on a custom pre-parser that manually peels apart arguments before dynamic Glazed parsing runs. The result works, but it is hard for a newcomer to know which parser is authoritative.

### Where to look
- `internal/botcli/command.go:126` — `newRunCommand()`
- `internal/botcli/run_schema.go:45` — `preparseRunArgs(...)`
- `internal/botcli/run_schema.go:198` — `buildRunSchema(...)`
- `internal/botcli/run_schema.go:263` — `parseRuntimeConfigArgs(...)`
- `internal/botcli/run_schema.go:312` — `printRunSchema(...)`

### Example snippet
```go
cmd := &cobra.Command{
    Use:                "run <bot>",
    DisableFlagParsing: true,
    RunE: func(cmd *cobra.Command, args []string) error {
        parsed, err := preparseRunArgs(args, defaultPreParsedRunArgs())
        if err != nil {
            return err
        }
        ...
    },
}
```

### Why it matters
This design has a real cognitive cost:

- Cobra declares flags, but does not actually parse them in the normal way for this command.
- Some flags appear in help because they are declared on the Cobra command, but their values are consumed by separate code paths.
- The code is more fragile because static runner flags and dynamic bot flags are managed by hand.

This is not a correctness complaint. It is a **maintainability and contract-clarity** complaint.

### Cleanup sketch
Keep the two-phase model if necessary, but make it explicit in the code structure.

Suggested split:

```text
internal/botcli/
  run_static_args.go     // preParsedRunArgs + static flag parsing only
  run_dynamic_schema.go  // buildRunSchema + parseRuntimeConfigArgs
  run_help.go            // printRunSchema + selector-aware help rendering
```

Also rename the static phase to signal intent:

```go
func parseStaticRunnerArgs(...) (StaticRunnerArgs, error)
```

That makes the architecture more honest than the current “Cobra command with flags, but manual parse owns reality” shape.

## 7. `internal/jsdiscord/runtime.go` still exposes lifecycle surfaces that appear unused or half-retained

### Problem
The runtime registrar still contains global runtime-state registration and lookup helpers whose public shape suggests they matter broadly, but they do not appear to be used outside `runtime.go` itself.

### Where to look
- `internal/jsdiscord/runtime.go:14` — `RuntimeStateContextKey`
- `internal/jsdiscord/runtime.go:16` — `runtimeStateByVM`
- `internal/jsdiscord/runtime.go:74` — `RegisterRuntimeState(...)`
- `internal/jsdiscord/runtime.go:81` — `UnregisterRuntimeState(...)`
- `internal/jsdiscord/runtime.go:88` — `LookupRuntimeState(...)`

### Example snippet
```go
const RuntimeStateContextKey = "discord.runtime"

var runtimeStateByVM sync.Map

func LookupRuntimeState(vm *goja.Runtime) *RuntimeState {
    if vm == nil {
        return nil
    }
    value, ok := runtimeStateByVM.Load(vm)
    if !ok {
        return nil
    }
    state, _ := value.(*RuntimeState)
    return state
}
```

### Why it matters
Unused lifecycle surfaces are not dangerous only because of runtime cost. They are also dangerous because they imply an API contract that future maintainers may preserve unnecessarily.

A newcomer reading this file may assume:
- some other subsystem relies on `LookupRuntimeState`,
- the context key is part of a broader contract,
- or VM-level state lookup is a supported extension seam.

If none of that is true anymore, the code should say so by becoming smaller.

### Cleanup sketch
Choose one of two options and document it explicitly.

#### Option A — delete the unused surfaces
If lookup is no longer needed:
- remove `LookupRuntimeState`
- remove the unused context key if nothing reads it
- collapse the implementation to the minimum runtime registration required

#### Option B — keep them, but write a comment explaining the intended future extension seam
If the team wants to preserve the hooks intentionally, add a comment like:

```go
// RuntimeStateContextKey and VM registration are retained as future extension seams
// for runtime-level inspectors. They are intentionally unused today.
```

Right now the code reads like an API in search of consumers.

## 8. The repo still contains stale or deprecated example artifacts that conflict with the current architecture

### Problem
There are at least two stale example surfaces that no longer match the current single-bot `defineBot(...)` runtime model:

1. `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js` appears unreferenced.
2. `examples/bots/` still documents an old `__verb__`-based example repository even though current bot discovery explicitly looks for `defineBot` + `require("discord")`.

### Where to look
- `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js`
- `examples/discord-bots/knowledge-base/index.js`
- `internal/botcli/bootstrap.go:142` — `discoverScriptCandidates(...)`
- `internal/botcli/bootstrap.go:184` — `looksLikeBotScript(...)`
- `examples/bots/README.md`
- `examples/bots/discord.js`

### Example snippet
From discovery:

```go
func looksLikeBotScript(path string) (bool, error) {
    ...
    if !strings.Contains(text, "defineBot") {
        return false, nil
    }
    if strings.Contains(text, `require("discord")`) || strings.Contains(text, `require('discord')`) {
        return true, nil
    }
    return false, nil
}
```

From the legacy example:

```js
function greet(name, excited) { ... }
__verb__("greet", { ... })
```

### Why it matters
This is exactly the kind of stale code that confuses a new intern:

- one directory says “this is the local example repository for `discord-bot bots ...`,”
- but the current discovery code will never treat its scripts as live bot implementations,
- and the knowledge-base bot has an older alternate registration file that is no longer wired in.

This does not break production, but it absolutely harms readability and trust in the repo.

### Cleanup sketch
Make a clean decision and reflect it in the tree.

#### For `register-knowledge-bot.js`
- either delete it,
- or move it to `archive/` or `sources/` in a ticket workspace if it is historically useful.

#### For `examples/bots/`
Choose one:
- delete it from this repo,
- move it to a historical/legacy examples area with explicit warnings,
- or update its README to say it is a legacy jsverbs artifact not used by the current `defineBot` runner.

A clear repo is better than a repo with “possibly useful old stuff” in operator-facing paths.

## 9. The knowledge-base example is the clearest sign that the JS authoring API needs ergonomic cleanup

### Problem
The `knowledge-base` example is now large enough that it exposes repetitive command aliases, heavy manual component wiring, and an overly broad store module. Because this bot is also one of the canonical examples, that organizational strain matters more than it would in a private app.

### Where to look
- `examples/discord-bots/knowledge-base/index.js:113` — `command("ask", ...)`
- `examples/discord-bots/knowledge-base/index.js:132` — `command("kb-search", ...)`
- `examples/discord-bots/knowledge-base/index.js:151` — `command("article", ...)`
- `examples/discord-bots/knowledge-base/index.js:173` — `command("kb-article", ...)`
- `examples/discord-bots/knowledge-base/index.js:195` — `command("review", ...)`
- `examples/discord-bots/knowledge-base/index.js:218` — `command("kb-review", ...)`
- `examples/discord-bots/knowledge-base/index.js:241` — `command("recent", ...)`
- `examples/discord-bots/knowledge-base/index.js:255` — `command("kb-recent", ...)`
- `examples/discord-bots/knowledge-base/index.js:338` through `461` — review/search component handlers
- `examples/discord-bots/knowledge-base/index.js:499` — `buildTeachModal()`
- `examples/discord-bots/knowledge-base/lib/store.js:42` — `createKnowledgeStore()`
- `examples/discord-bots/knowledge-base/lib/store.js:65` — `ensureSchema()`
- `examples/discord-bots/knowledge-base/lib/store.js:335` — `search(...)`
- `examples/discord-bots/knowledge-base/lib/search.js:130` — `searchView(...)`
- `examples/discord-bots/knowledge-base/lib/search.js:159` — `buildSearchMessage(...)`
- `examples/discord-bots/knowledge-base/lib/review.js:130` — `buildEntryModal(...)`
- `examples/discord-bots/knowledge-base/lib/review.js:361` — `buildQueueMessage(...)`

### Example snippet
Alias duplication:

```js
command("ask", { ... }, async (ctx) => {
  ...
  return search.buildSearchMessage(search.searchView(ctx, store))
})

command("kb-search", { ... }, async (ctx) => {
  ...
  return search.buildSearchMessage(search.searchView(ctx, store))
})
```

Store accumulation:

```js
function createKnowledgeStore() {
  let configuredPath = ""
  let initialized = false

  function ensure(config) { ... }
  function ensureSchema() { ... }
  function ensureSeedData(config) { ... }
  function saveCandidate(config, candidate) { ... }
  function saveManualEntry(config, payload) { ... }
  function insertEntry(entry) { ... }
  function updateEntry(config, identifier, patch) { ... }
  function setStatus(config, identifier, status, reviewedBy, reviewNote) { ... }
  function listByStatus(config, status, limit) { ... }
  function search(config, query, limit) { ... }
}
```

### Why it matters
This example is doing too much in a few places:

- `index.js` is a registration file, a UI flow coordinator, and a small helper module all at once.
- `store.js` is schema migration, seeding, repository, mapper, search, and version history all at once.
- the repeated alias commands and interaction handlers make the example read more like framework plumbing than bot logic.

For a canonical example, that is a code quality smell because examples teach habits.

### Cleanup sketch
#### For `index.js`
Use the moderation bot’s “register by concern” pattern.

```text
examples/discord-bots/knowledge-base/
  index.js
  lib/
    register-capture.js
    register-manual-entry.js
    register-search.js
    register-review.js
    register-autocomplete.js
    register-reaction-promotions.js
```

Then `index.js` becomes composition only:

```js
module.exports = defineBot((api) => {
  configureBot(api)
  registerCapture(api, deps)
  registerManualEntry(api, deps)
  registerSearch(api, deps)
  registerReview(api, deps)
  registerAutocomplete(api, deps)
  registerReactionPromotions(api, deps)
})
```

#### For `store.js`
Split by responsibility:

```text
lib/
  store/
    index.js          // createKnowledgeStore facade
    schema.js         // ensureSchema
    seed.js           // seed entries
    repo.js           // CRUD/list/search entry persistence
    mapping.js        // row ↔ entry normalization
    search.js         // search query assembly / scoring helpers
```

#### For aliases and screen flows
Introduce small local helpers rather than repeating twin command bodies.

```js
registerAliasCommand(["ask", "kb-search"], makeSearchCommand(deps))
registerAliasCommand(["article", "kb-article"], makeArticleCommand(deps))
registerAliasCommand(["review", "kb-review"], makeReviewCommand(deps))
```

This is one of the clearest places where API ergonomics and code organization intersect.

## 10. Tests are organized more by history than by behavior

### Problem
The runtime tests are valuable, but they are very concentrated. `internal/jsdiscord/runtime_test.go` alone is over 1200 lines, and it acts as a catch-all for many unrelated behaviors.

### Where to look
- `internal/jsdiscord/runtime_test.go`
- `internal/jsdiscord/knowledge_base_runtime_test.go`
- `internal/botcli/command_test.go`

### Example
The runtime test file mixes multiple concerns:
- command snapshots
- async settlement
- event dispatch
- Discord ops
- autocomplete
- thread utilities
- moderation APIs

### Why it matters
Large integration-style test files are not inherently bad, but they become harder to navigate when a reviewer wants to answer a narrow question like:

- “Where are message ops tested?”
- “Where is reaction dispatch tested?”
- “Where do we prove autocomplete normalization?”

### Cleanup sketch
Split tests by behavior family.

```text
internal/jsdiscord/
  runtime_descriptor_test.go
  runtime_dispatch_test.go
  runtime_events_test.go
  runtime_payloads_test.go
  runtime_ops_messages_test.go
  runtime_ops_members_test.go
  runtime_ops_threads_test.go
  runtime_knowledge_base_test.go
```

This is a quality-of-life improvement for maintainers and reviewers, not a functional change.

# Prioritized cleanup plan

## Low-risk, high-value cleanups first

1. Delete or archive stale artifacts:
   - `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js`
   - legacy `examples/bots/` or at least clearly relabel it as legacy
2. Remove dead fallback interaction code from `internal/bot/bot.go`
3. Split `internal/jsdiscord/bot.go` by responsibility without changing behavior
4. Split `internal/jsdiscord/host_payloads.go` by payload concern
5. Split large test files by behavior family

## Medium-risk structural improvements

6. Introduce typed internal envelopes/snapshots around the Go↔JS boundary
7. Refactor `host_dispatch.go` around base request builders and responder helpers
8. Make `internal/botcli`’s two parsing phases explicit in file and type names

## Larger API/ergonomics cleanup

9. Reorganize the knowledge-base example by feature registration modules
10. Split the knowledge-base store into schema/seed/repo/mapping/search pieces
11. Add small alias/screen helper layers for the knowledge-base example

# Recommended review sequence for a second engineer

If someone else reviews or implements the cleanup, this is the order I would recommend.

## Pass 1 — stale code and dead branches
- `internal/bot/bot.go`
- `internal/jsdiscord/runtime.go`
- `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js`
- `examples/bots/`

## Pass 2 — file-size cleanup without semantic changes
- split `internal/jsdiscord/bot.go`
- split `internal/jsdiscord/host_payloads.go`
- split large tests

## Pass 3 — API clarity improvements
- typed envelopes/snapshots
- clearer `internal/botcli` parsing phases
- knowledge-base example ergonomics

# Final conclusion

This repo does not look like a project in trouble. It looks like a project that has succeeded in adding capability quickly and now needs a maintainability pass.

The most important insight is this:

> The architecture is mostly right, but the implementation has several accumulation points that now hide the architecture.

That is good news, because it means the right response is **cleanup and clarification**, not a rewrite.
