---
Title: js-discord-bot Code Quality Report
Ticket: CODEQUAL-2026-0421
Status: active
Topics:
    - code-quality
    - architecture
    - refactoring
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/bot/bot.go
      Note: 350-line session host with 11 repetitive event handlers
    - Path: internal/botcli/run_schema.go
      Note: 346-line manual flag parser that reimplements Cobra behavior
    - Path: internal/jsdiscord/bot.go
      Note: "1"
    - Path: internal/jsdiscord/host_dispatch.go
      Note: 593-line dispatch file with 18 repeated DispatchRequest constructions
    - Path: internal/jsdiscord/host_payloads.go
      Note: 736-line normalization layer for 12 Discord payload types
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---






# js-discord-bot Code Quality Report

**Ticket:** CODEQUAL-2026-0421  
**Date:** 2026-04-21  
**Scope:** Full codebase quality review — organization, clarity, API design, duplication, and maintainability  
**Target audience:** New engineering interns joining the project  
**Lines of Go examined:** ~8,800  
**Lines of JS examined:** ~5,300  

---

## 1. Executive Summary

This report is a structural and organizational health check of `js-discord-bot`, a Go-hosted Discord bot runtime that lets developers write bot logic in JavaScript. The project works, is well-tested, and has a coherent architecture. However, several areas have accumulated enough complexity that they will slow down onboarding and increase the risk of bugs:

1. **Three files are oversized** and violate the single-responsibility principle. `bot.go` is a 1,300-line monolith that mixes VM bridge logic, draft bookkeeping, and JavaScript object construction.
2. **Heavy repetition** exists in dispatch construction, ops builder functions, and test setup. Roughly 15–20% of the `jsdiscord` package is copy-paste structural code.
3. **Several APIs are confusing** for newcomers: the `DispatchRequest` struct has 25 fields, `DiscordOps` is a 25-field function-pointer bag, and `botDraft.finalize` dynamically creates 600+ lines of closures inside a closure.
4. **Deprecated directory duality** exists between `examples/bots` and `examples/discord-bots`.
5. **Nil-check noise** is pervasive. While defensive, it adds visual clutter and suggests types could be stronger.

None of these are emergencies. They are **gradual-friction** issues: each one makes the codebase a little harder to extend, test, and explain. The recommended path is a series of small, phased refactors rather than a big rewrite.

---

## 2. Project Orientation for New Interns

Before we talk about problems, you need to understand what this system is and how the pieces fit together. This section is prose-heavy because the goal is to build mental models, not just list files.

### 2.1 What is this project?

`js-discord-bot` is a **bridge** between two worlds:

- **Go** owns the hard parts: the Discord websocket connection, session lifecycle, command registration, and process management.
- **JavaScript** owns the bot behavior: slash commands, event handlers, buttons, modals, autocomplete, and outbound Discord calls.

The JavaScript does **not** run in Node.js. It runs inside an embedded ECMAScript engine called **goja**, which is a pure-Go JavaScript implementation. This means:

- No `npm install`, no `package.json`, no node_modules resolution (except a minimal `require` polyfill).
- The JS API is hand-crafted by the Go code and injected into the runtime as a native module named `"discord"`.
- JS errors bubble up into Go. Go errors are translated back into JS exceptions.

### 2.2 The runtime model (one diagram)

```text
Operator types a CLI command
        |
        v
+----------------------------+
|   cmd/discord-bot          |   Cobra CLI entrypoint
+----------------------------+
        |
        v
+----------------------------+
|   internal/botcli          |   "bots list/help/run" subcommand
|   (bot discovery & runner) |
+----------------------------+
        |
        v
+----------------------------+
|   internal/bot             |   Creates discordgo.Session
|   (Discord session host)   |   Wires event handlers
+----------------------------+
        |
        v
+----------------------------+
|   internal/jsdiscord       |   Embeds goja runtime
|   (JS bridge)              |   Loads bot script
|                            |   Forwards events to JS handlers
+----------------------------+
        |
        v
+----------------------------+
|   examples/discord-bots/   |   JavaScript bot implementations
|   <bot>/index.js           |   defineBot(...)
+----------------------------+
```

### 2.3 Key concepts you must know

**Concept 1: One bot per process**
The host is designed so that `discord-bot bots run <name>` selects exactly one JavaScript bot and runs it. You cannot load two bots into the same process. Composition happens inside the JS layer.

**Concept 2: The `defineBot` contract**
Every JS bot file does this:

```js
const { defineBot } = require("discord")
module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "my-bot", description: "..." })
  command("ping", { description: "..." }, async (ctx) => {
    return { content: "pong" }
  })
})
```

The object returned by `defineBot` is called a **BotHandle** in Go. It has methods like `DispatchCommand`, `DispatchEvent`, `DispatchComponent`, etc. The Go host calls these methods when Discord events arrive.

**Concept 3: Request-scoped context (`ctx`)**
When a handler runs, it receives a `ctx` object that contains:
- `ctx.args` — parsed command arguments
- `ctx.reply(...)` / `ctx.edit(...)` / `ctx.followUp(...)` — ways to respond
- `ctx.discord.*` — outbound Discord API calls (fetch guilds, send messages, etc.)
- `ctx.store.*` — a simple in-memory key/value store
- `ctx.log.*` — structured logging

This is intentional. There is no global Discord client. Every operation is tied to the current request/event.

**Concept 4: The `DispatchRequest` struct**
This is the universal envelope that Go uses to pass data into JavaScript. It contains everything a handler might need: the event name, parsed arguments, the interaction object, the user, the guild, the channel, response functions, and the DiscordOps object. We will critique this design later.

### 2.4 Directory map

```text
cmd/discord-bot/               CLI entrypoint (main.go, root.go, commands.go)
internal/
  bot/                         Live Discord session wrapper (bot.go)
  botcli/                      Named bot repository discovery & runner
    bootstrap.go               Directory scanning, "does this look like a bot?"
    command.go                 Cobra subcommands: list, help, run
    run_schema.go              Runtime config parsing (pre-parser + Glazed schema)
    runtime.go                 The actual `Run()` loop
    model.go                   DiscoveredBot, Repository, Bootstrap types
  config/                      Config decoding and validation (config.go)
  jsdiscord/                   THE BIG PACKAGE — JS runtime bridge
    runtime.go                 Module registration, RuntimeState
    bot.go                     BotHandle, botDraft, defineBot API, dispatch
    host.go                    Host struct — loads script into goja
    host_dispatch.go           Forwards Discord events into JS handlers
    host_payloads.go           Normalizes JS return values into Discord structs
    host_responses.go          interactionResponder, channelResponder
    host_commands.go           Converts JS command specs into ApplicationCommand
    host_maps.go               Converts discordgo structs into map[string]any
    host_ops.go                Orchestrates ops builder calls
    host_ops_*.go              One file per ops category (messages, members, ...)
    host_ops_helpers.go        Normalization helpers for ops payloads
    host_logging.go            Structured logging helpers
    descriptor.go              BotDescriptor, InspectScript, parsing helpers
    store.go                   MemoryStore (in-memory KV for JS bots)
examples/
  discord-bots/                Real example bots (ping, knowledge-base, moderation, ...)
  bots/                        OLD/deprecated examples
  bots-dupe-a/, bots-dupe-b/   Test fixtures for duplicate-name detection
pkg/doc/                       Embedded help pages (Markdown → Cobra help)
testdata/                      Additional test fixtures
```

---

## 3. Architecture Deep Dive

This section explains the runtime flow with enough detail that you can trace a single Discord interaction from the websocket to the JS handler and back.

### 3.1 Entry points

There are two ways to start the bot:

**Path A: Named bot runner (recommended)**
```bash
discord-bot bots run ping --bot-repository ./examples/discord-bots
```
This flows through:
1. `cmd/discord-bot/root.go` → `botcli.NewCommand()`
2. `internal/botcli/command.go` → `newRunCommand()`
3. `internal/botcli/run_schema.go` → `preparseRunArgs()` (custom pre-parser)
4. `internal/botcli/bootstrap.go` → `DiscoverBots()` → `ResolveBot()`
5. `internal/botcli/runtime.go` → `runSelectedBots()`
6. `internal/bot/bot.go` → `NewWithScript()`

**Path B: Direct host command (legacy/advanced)**
```bash
discord-bot run --bot-script ./examples/discord-bots/ping/index.js
```
This flows through:
1. `cmd/discord-bot/commands.go` → `newRunCommand()` (Glazed command)
2. `cmd/discord-bot/commands.go` → `RunIntoGlazeProcessor()`
3. `internal/bot/bot.go` → `New()`

### 3.2 The Go-JavaScript bridge

The bridge lives in `internal/jsdiscord`. Here is the exact lifecycle:

**Step 1: Module registration**
When the goja runtime starts, `runtime.go` registers a native module:

```go
reg.RegisterNativeModule(state.ModuleName(), state.Loader)
```

**Step 2: Script loading**
`host.go` creates the runtime and calls `rt.Require.Require(absScript)`. The JS file executes.

**Step 3: `defineBot` execution**
Inside JS, `defineBot` is called. It invokes the builder function, which calls `command(...)`, `event(...)`, etc. Each call appends to a **draft** object (`botDraft`).

**Step 4: Finalization**
After the builder finishes, `draft.finalize(vm)` is called. This constructs a JS object with methods like `dispatchCommand`, `dispatchEvent`, etc. Each method is a **closure** that captures the draft arrays and looks up the correct handler.

**Step 5: Compilation**
`bot.go`'s `CompileBot()` extracts those methods from the JS object and wraps them in a `BotHandle`.

**Step 6: Dispatch**
When a Discord event arrives, `host_dispatch.go` builds a `DispatchRequest` and calls `h.handle.DispatchEvent()` (or `DispatchCommand`, etc.).

**Step 7: Promise resolution**
If the JS handler returns a Promise, `bot.go`'s `waitForPromise()` polls it every 5ms until it settles. This is a busy-wait loop inside the goja runtime.

### 3.3 Discord event flow

Let's trace a slash command `/ping` from the websocket:

```text
Discord Gateway
      |
      v
discordgo.Session (library)
      |
      v
internal/bot/bot.go
  handleInteractionCreate()
      |
      v
internal/jsdiscord/host_dispatch.go
  DispatchInteraction()
      |
      v
  switch interaction.Type:
    InteractionApplicationCommand
      |
      v
  switch data.CommandType:
    ChatInput Command
      |
      v
  check for subcommands:
    if data.Options[0].Type == SubCommand
      -> DispatchSubcommand(...)
    else
      -> DispatchCommand(...)
      |
      v
internal/jsdiscord/bot.go
  BotHandle.DispatchCommand()
      |
      v
  buildDispatchInput(vm, ctx, request)
      -> creates JS object with all request fields
      |
      v
  bot.dispatch() -> calls JS closure
      |
      v
JS handler runs:
  async (ctx) => { return { content: "pong" } }
      |
      v
JS return value exported to Go
      |
      v
settleValue() recursively resolves Promises
      |
      v
Result returned to host_dispatch.go
      |
      v
emitEventResult() calls responder.Reply()
      |
      v
interactionResponder.Reply()
  -> session.InteractionRespond()
```

### 3.4 Bot discovery and the CLI

The `botcli` package implements a small package manager for JS bots:

1. You specify one or more `--bot-repository` directories.
2. `bootstrap.go` walks each directory, skipping `node_modules` and hidden folders.
3. For every `.js` file that is either `index.js` or directly inside the repo root, it checks if the file contains `defineBot` and `require("discord")`.
4. If so, it loads the script into a temporary goja runtime and calls `describe()` to get metadata.
5. Bots are deduplicated by name. If two repos contain a bot with the same name, it's an error.
6. The `run` command uses a **two-phase parser**: first it manually extracts known flags (`--bot-token`, `--sync-on-start`, etc.), then it passes everything else to a Glazed schema parser for the bot's runtime config fields.

---

## 4. Code Quality Assessment

### 4.1 Methodology

We examined the codebase using these criteria:

| Criterion | How we measured |
|-----------|-----------------|
| File size | `wc -l` on every `.go` file |
| Package size | Total lines per package directory |
| Duplication | Visual inspection + grep for repeated struct literal shapes |
| API clarity | Could a new intern guess what a function does from its signature? |
| Single responsibility | Does one file/type/function do one thing? |
| Nil safety vs noise | Are nil checks defensive or symptomatic of weak types? |
| Test patterns | Are tests DRY? Do they use helpers consistently? |

### 4.2 Summary of findings

| Category | Count | Severity |
|----------|-------|----------|
| Oversized files (>300 lines) | 5 | Medium |
| Heavy repetition blocks | 6 | Medium |
| Confusing APIs | 4 | Medium-High |
| Deprecated/legacy paths | 2 | Low |
| Goroutine safety concerns | 2 | Medium |
| Nil-check noise | Pervasive | Low (but tedious) |

---

## 5. Detailed Findings

Each finding follows this template:
- **Problem** — what is wrong and why it matters
- **Where to look** — exact files and line ranges
- **Example snippet** — the code in question
- **Why it matters** — maintenance/runtime implications
- **Cleanup sketch** — pseudocode or structural layout

---

### 5.1 Large files and packages

#### 5.1.1 `internal/jsdiscord/bot.go` — 1,293 lines

**Problem:** This file contains at least five distinct responsibilities:
1. The `BotHandle` type and its `CompileBot()` constructor.
2. The `dispatch()` generic dispatcher and promise settlement logic.
3. The `botDraft` type and all draft registration methods (`command`, `event`, `component`, `modal`, `autocomplete`, `configure`).
4. The `finalize()` method, which dynamically builds a massive JS object with 7 closures.
5. Helper functions: `exportMap`, `cloneMap`, snapshot builders, finders, `buildDispatchInput`, `buildContext`, `storeObject`, `discordOpsObject`, `loggerObject`, `applyFields`.

**Where to look:** `internal/jsdiscord/bot.go`, lines 1–1293.

**Example snippet (the finalize method):**
```go
func (d *botDraft) finalize(vm *goja.Runtime) goja.Value {
    // ... builds a JS object with 7 closures ...
    _ = bot.Set("dispatchCommand", func(call goja.FunctionCall) goja.Value {
        // ~25 lines
    })
    _ = bot.Set("dispatchSubcommand", func(call goja.FunctionCall) goja.Value {
        // ~20 lines
    })
    // ... 5 more closures ...
    return bot
}
```

The `finalize` method alone is ~180 lines. Each closure repeats the same pattern: validate argument count, extract fields from the input object, look up the handler, build a context, call the handler, handle errors.

**Why it matters:**
- A new intern cannot reason about this file without loading all 1,300 lines into working memory.
- The file changes for many unrelated reasons: adding a new draft type, changing dispatch semantics, tweaking the JS context shape, or adding a new store method.
- Code review diffs are large and unfocused.

**Cleanup sketch:**
Split into subpackages or at least separate files:

```text
internal/jsdiscord/
  bot_handle.go          // BotHandle, CompileBot, dispatch, settleValue
  bot_draft.go           // botDraft and registration methods
  bot_finalizer.go       // finalize() and the 7 dispatch closures
  bot_context.go         // buildDispatchInput, buildContext, storeObject, loggerObject
  discord_ops.go         // discordOpsObject, DiscordOps struct
```

The `finalize()` closures can be extracted into standalone functions:

```go
func makeDispatchCommand(vm *goja.Runtime, commands []*commandDraft, ...) goja.Callable {
    return func(call goja.FunctionCall) goja.Value {
        // implementation
    }
}
```

---

#### 5.1.2 `internal/jsdiscord/host_payloads.go` — 736 lines

**Problem:** This file is a pure normalization layer. It takes `any` (usually `map[string]any` from JS) and converts it into discordgo structs: `InteractionResponseData`, `WebhookParams`, `WebhookEdit`, `MessageSend`, `MessageEdit`, `ApplicationCommandOptionChoice`, `MessageEmbed`, `MessageComponent`, etc.

It does one thing, but it does it for **12 different target types**.

**Where to look:** `internal/jsdiscord/host_payloads.go`, lines 1–736.

**Example snippet:**
```go
func normalizeResponsePayload(payload any) (*discordgo.InteractionResponseData, error) { ... }
func normalizeModalPayload(payload any) (*discordgo.InteractionResponseData, error) { ... }
func normalizeWebhookParams(payload any) (*discordgo.WebhookParams, error) { ... }
func normalizeWebhookEdit(payload any) (*discordgo.WebhookEdit, error) { ... }
func normalizeMessageSend(payload any) (*discordgo.MessageSend, error) { ... }
func normalizeChannelMessageEdit(...) (*discordgo.MessageEdit, error) { ... }
func normalizePayload(payload any) (*normalizedResponse, error) { ... }
// plus embeds, components, files, mentions, choices, text inputs, ...
```

**Why it matters:**
- The file is not complex, but it is **long**. Finding the right helper requires scrolling.
- Adding a new payload type means appending to an already-large file.
- There is no shared contract or interface. Each normalizer is independent.

**Cleanup sketch:**
Group by target domain:

```text
internal/jsdiscord/payload/
  response.go      // normalizeResponsePayload, normalizeModalPayload
  webhook.go       // normalizeWebhookParams, normalizeWebhookEdit
  message.go       // normalizeMessageSend, normalizeChannelMessageEdit
  embed.go         // normalizeEmbed, normalizeEmbedArray
  component.go     // normalizeComponent, normalizeLeafComponent, normalizeSelectMenu, normalizeTextInput
  choice.go        // normalizeAutocompleteChoices, normalizeAutocompleteChoice
  file.go          // normalizeFiles
  mention.go       // normalizeAllowedMentions
  reference.go     // normalizeMessageReference
```

Each file would be 50–120 lines. The package-level `normalizedResponse` struct can stay as a shared type in `payload/common.go`.

---

#### 5.1.3 `internal/jsdiscord/host_dispatch.go` — 593 lines

**Problem:** This file contains 10+ dispatch methods (`DispatchReady`, `DispatchGuildCreate`, `DispatchGuildMemberAdd`, `DispatchMessageCreate`, `DispatchInteraction`, etc.) and 3 helpers. Every event dispatch repeats the same `DispatchRequest` construction pattern.

**Where to look:** `internal/jsdiscord/host_dispatch.go`, lines 1–593.

**Example snippet (typical event dispatch):**
```go
func (h *Host) DispatchGuildMemberAdd(ctx context.Context, session *discordgo.Session, member *discordgo.GuildMemberAdd) error {
    if h == nil || h.handle == nil || member == nil || member.Member == nil {
        return nil
    }
    _, err := h.handle.DispatchEvent(ctx, DispatchRequest{
        Name:     "guildMemberAdd",
        Member:   memberMap(member.Member),
        User:     userMap(member.User),
        Guild:    guildMap(member.GuildID),
        Me:       currentUserMap(session),
        Metadata: map[string]any{"scriptPath": h.scriptPath},
        Config:   cloneMap(h.runtimeConfig),
        Command:  map[string]any{"event": "guildMemberAdd"},
        Discord:  buildDiscordOps(h.scriptPath, session),
    })
    return err
}
```

This exact shape appears 10 times with only the field names changing.

**Why it matters:**
- Adding a new event means copy-pasting ~15 lines and hoping you didn't forget a field.
- The `DispatchInteraction` method is 250+ lines of nested `switch` statements. It is the most complex function in the entire codebase.

**Cleanup sketch:**
Introduce a builder for `DispatchRequest`:

```go
type DispatchBuilder struct {
    host *Host
    name string
}

func (b *DispatchBuilder) WithEvent(name string) *DispatchBuilder { ... }
func (b *DispatchBuilder) WithMember(m *discordgo.Member) *DispatchBuilder { ... }
func (b *DispatchBuilder) WithUser(u *discordgo.User) *DispatchBuilder { ... }
// ... etc ...

func (b *DispatchBuilder) Build() DispatchRequest { ... }
```

Then `DispatchGuildMemberAdd` becomes:

```go
func (h *Host) DispatchGuildMemberAdd(ctx context.Context, s *discordgo.Session, m *discordgo.GuildMemberAdd) error {
    if !h.canDispatch() || m == nil || m.Member == nil { return nil }
    req := NewDispatchBuilder(h).WithEvent("guildMemberAdd").
        WithMember(m.Member).WithUser(m.User).WithGuild(m.GuildID).WithSession(s).Build()
    _, err := h.handle.DispatchEvent(ctx, req)
    return err
}
```

---

#### 5.1.4 `internal/jsdiscord/descriptor.go` — 397 lines

**Problem:** Parses the JS `describe()` output into `BotDescriptor`. Contains 7 `parse*Descriptors` functions that are nearly identical.

**Where to look:** `internal/jsdiscord/descriptor.go`, lines 1–397.

**Example snippet:**
```go
func parseComponentDescriptors(raw any) []ComponentDescriptor { ... }
func parseModalDescriptors(raw any) []ModalDescriptor { ... }
```

These two functions differ only in:
- The return type
- The field name they extract (`customId`)
- The struct they append to

**Why it matters:** Adding a new descriptor type (e.g., `ContextMenuDescriptor`) means copying another 15-line function.

**Cleanup sketch:**
Use a generic helper:

```go
func parseDescriptors[T any](raw any, extract func(map[string]any) (T, bool)) []T { ... }
```

Then:
```go
components := parseDescriptors(desc["components"], func(m map[string]any) (ComponentDescriptor, bool) {
    id := mapString(m, "customId")
    return ComponentDescriptor{CustomID: id}, id != ""
})
```

---

#### 5.1.5 `internal/bot/bot.go` — 350 lines

**Problem:** Wires 11 Discord event handlers into `discordgo.Session`, then implements each handler. Every handler has the same shape:

```go
func (b *Bot) handleX(session *discordgo.Session, x *discordgo.X) {
    if x == nil { return }
    if x.Author != nil && x.Author.Bot { return }  // sometimes
    if b.jsHost != nil {
        if err := b.jsHost.DispatchX(...); err != nil {
            log.Error().Err(err).Msg("failed to dispatch X")
        }
    }
}
```

**Where to look:** `internal/bot/bot.go`, lines 169–320.

**Why it matters:**
- 11 methods * ~8 lines each = ~88 lines of structural duplication.
- `handleInteractionCreate` also contains a fallback hardcoded `ping`/`echo` command handler that is dead code once a JS bot is loaded.

**Cleanup sketch:**
Use a generic handler wrapper:

```go
type discordHandler[T any] func(*discordgo.Session, T)

func dispatchEvent[T any](b *Bot, s *discordgo.Session, evt T, dispatchFn func(context.Context, *discordgo.Session, T) error) {
    if b.jsHost == nil { return }
    if err := dispatchFn(context.Background(), s, evt); err != nil {
        log.Error().Err(err).Msg("failed to dispatch event")
    }
}
```

---

#### 5.1.6 The Go↔JS boundary uses too much `map[string]any` instead of typed internal contracts

**Problem:** The Go↔JS boundary understandably uses plain objects, but the *interior* of the system continues to rely on `map[string]any` longer than necessary. That makes the code harder to refactor, harder to autocomplete, and easier to drift.

**Where to look:**
- `internal/jsdiscord/bot.go:97` — `DispatchRequest` fields are map-heavy
- `internal/jsdiscord/descriptor.go:115` — `descriptorFromDescribe(...)`
- `internal/jsdiscord/descriptor.go:153` through `344` — repeated `parse*Descriptor` helpers from raw maps
- `internal/jsdiscord/host_maps.go` — many functions produce anonymous map shapes used transitively elsewhere

**Example snippet:**
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

**Why it matters:**
`map[string]any` is good at the *edge* of the system, but weaker as the *interior* representation. The current approach increases:
- typo risk in keys,
- drift between maps produced in `host_maps.go` and maps expected in JS-facing docs/tests,
- reviewer difficulty when tracing shape changes,
- and the amount of ad hoc conversion code in parsing layers.

**Cleanup sketch:**
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

---

### 5.2 Repetitive / duplicated code

#### 5.2.1 `DispatchRequest` construction in `host_dispatch.go`

**Problem:** There are 18 call sites that construct a `DispatchRequest`. Each one repeats the same base fields:

```go
Metadata: map[string]any{"scriptPath": h.scriptPath},
Config:   cloneMap(h.runtimeConfig),
Discord:  buildDiscordOps(h.scriptPath, session),
Me:       currentUserMap(session),
```

**Where to look:** `internal/jsdiscord/host_dispatch.go`, lines 16–520.

**Count:** 18 repetitions of the metadata/config/discord/me quartet.

**Cleanup sketch:**
Add a constructor on `Host`:

```go
func (h *Host) baseDispatchRequest(session *discordgo.Session) DispatchRequest {
    return DispatchRequest{
        Metadata: map[string]any{"scriptPath": h.scriptPath},
        Config:   cloneMap(h.runtimeConfig),
        Discord:  buildDiscordOps(h.scriptPath, session),
        Me:       currentUserMap(session),
    }
}
```

---

#### 5.2.2 Host ops builder functions (`host_ops_*.go`)

**Problem:** There are 6 files (`host_ops_channels.go`, `host_ops_guilds.go`, `host_ops_members.go`, `host_ops_messages.go`, `host_ops_roles.go`, `host_ops_threads.go`) that follow the exact same mechanical pattern:

```go
func buildXOps(ops *DiscordOps, scriptPath string, session *discordgo.Session) {
    if ops == nil || session == nil { return }
    ops.X = func(ctx context.Context, ...) (...) {
        _ = ctx
        // validate inputs
        // call discordgo
        // log lifecycle
        // return
    }
}
```

Each function closure repeats:
- `strings.TrimSpace` on every string argument
- `fmt.Errorf("... requires ...")` validation
- `logLifecycleDebug(...)` with a map containing `script`, `action`, and IDs

**Where to look:** All `internal/jsdiscord/host_ops_*.go` files.

**Count:** ~25 closure definitions across 6 files, each 8–15 lines.

**Cleanup sketch:**
Extract a helper for the common validation/logging pattern:

```go
func validatedCall1[T any](scriptPath, action, id1 string, fn func() (T, error)) (T, error) {
    id1 = strings.TrimSpace(id1)
    if id1 == "" { var zero T; return zero, fmt.Errorf("%s requires ID", action) }
    result, err := fn()
    if err == nil {
        logLifecycleDebug(action+" from javascript", map[string]any{
            "script": scriptPath, "action": action, "id": id1,
        })
    }
    return result, err
}
```

Then `ops.GuildFetch` becomes:

```go
ops.GuildFetch = func(ctx context.Context, guildID string) (map[string]any, error) {
    return validatedCall1(scriptPath, "discord.guilds.fetch", guildID, func() (map[string]any, error) {
        guild, err := session.Guild(guildID)
        if err != nil { return nil, err }
        return guildSnapshotMap(guild), nil
    })
}
```

---

#### 5.2.3 Map conversion helpers (`host_maps.go`)

**Problem:** Every conversion function has the same guard:

```go
func userMap(user *discordgo.User) map[string]any {
    if user == nil { return map[string]any{} }
    return map[string]any{"id": user.ID, "username": user.Username, ...}
}
```

There are ~20 such functions. They are simple, but they could be generated or templated.

**Where to look:** `internal/jsdiscord/host_maps.go`.

**Cleanup sketch:**
Keep as-is for now, but consider a code-generation approach if the Discord API surface grows. A small Go program that reads discordgo struct tags and emits `*Map` functions would eliminate manual maintenance.

---

#### 5.2.4 Test organization

**Problem:** The runtime tests are valuable, but they are very concentrated. `internal/jsdiscord/runtime_test.go` alone is over 1,200 lines, and it acts as a catch-all for many unrelated behaviors.

**Where to look:**
- `internal/jsdiscord/runtime_test.go` — mixes command snapshots, async settlement, event dispatch, Discord ops, autocomplete, thread utilities, moderation APIs
- `internal/jsdiscord/knowledge_base_runtime_test.go` — integration test with its own inline helpers

**Cleanup sketch:**

**Step 1 — extract shared helpers:**

```go
package testutil

func LoadTestBot(t *testing.T, script string) *BotHandle { ... }
func WriteBotScript(t *testing.T, source string) string { ... }
```

**Step 2 — split by behavior family:**

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

---

### 5.3 Confusing code and poor APIs

#### 5.3.1 `bot.go`: The monolithic `botDraft.finalize`

**Problem:** `finalize` dynamically constructs a JS object by creating 7 closures inline. Each closure captures the entire draft state (commands, subcommands, events, etc.) by value via `append([]*T(nil), draft.field...)`. This is both memory-inefficient and hard to follow.

**Where to look:** `internal/jsdiscord/bot.go`, lines 380–570.

**Example snippet:**
```go
_ = bot.Set("dispatchCommand", func(call goja.FunctionCall) goja.Value {
    if len(call.Arguments) != 1 {
        panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchCommand expects one input object")))
    }
    input := objectFromValue(vm, call.Arguments[0])
    name := strings.TrimSpace(input.Get("name").String())
    if name == "" {
        panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchCommand input name is empty")))
    }
    command := findCommand(commands, name)
    if command == nil {
        command = findCommand(userCommands, name)
    }
    if command == nil {
        command = findCommand(messageCommands, name)
    }
    if command == nil {
        panic(vm.NewGoError(fmt.Errorf("discord bot %q has no command named %q", moduleName, name)))
    }
    ctx := buildContext(vm, store, input, "command", name, metadata)
    result, err := command.handler(goja.Undefined(), ctx)
    if err != nil {
        panic(vm.NewGoError(err))
    }
    return result
})
```

**Why it matters:**
- The closure captures `commands`, `userCommands`, `messageCommands`, `store`, `metadata`, and `moduleName`.
- The fallback chain `commands -> userCommands -> messageCommands` is implicit and surprising.
- Errors are reported via `panic(vm.NewGoError(...))` inside a closure, which is idiomatic for goja but noisy.

**Cleanup sketch:**
Replace inline closures with named functions that take the needed state as arguments:

```go
func makeDispatchCommand(
    vm *goja.Runtime,
    moduleName string,
    commands, userCommands, messageCommands []*commandDraft,
    store *MemoryStore,
    metadata map[string]any,
) goja.Callable {
    return func(call goja.FunctionCall) goja.Value {
        input := requireObjectArg(vm, call, "dispatchCommand")
        name := requireStringField(vm, input, "name")
        cmd := findCommandInAll(commands, userCommands, messageCommands, name)
        if cmd == nil {
            panic(vm.NewGoError(fmt.Errorf("no command %q", name)))
        }
        ctx := buildContext(vm, store, input, "command", name, metadata)
        result, err := cmd.handler(goja.Undefined(), ctx)
        if err != nil {
            panic(vm.NewGoError(err))
        }
        return result
    }
}
```

Then `finalize` becomes:
```go
_ = bot.Set("dispatchCommand", makeDispatchCommand(vm, moduleName, commands, userCommands, messageCommands, store, metadata))
```

---

#### 5.3.2 `host_dispatch.go`: Nesting depth

**Problem:** `DispatchInteraction` is a 250-line method with 4 levels of `switch` nesting:

```go
switch interaction.Type {                      // level 1
    case InteractionApplicationCommand:
        switch data.CommandType {              // level 2
            case UserApplicationCommand:       // leaf
            case MessageApplicationCommand:    // leaf
            default:                           // Chat input
                if len(data.Options) > 0 && data.Options[0].Type == SubCommand {
                    // subcommand path        // level 3 (implicit)
                } else {
                    // top-level command      // level 3 (implicit)
                }
        }
    case InteractionMessageComponent:           // level 1
    case InteractionModalSubmit:                // level 1
    case InteractionApplicationCommandAutocomplete: // level 1
}
```

**Where to look:** `internal/jsdiscord/host_dispatch.go`, lines 226–530.

**Why it matters:**
- Each leaf repeats the same `DispatchRequest` construction and result handling.
- The method is too long to fit on one screen. Reviewers miss details.

**Cleanup sketch:**
Extract each top-level branch into a private method:

```go
func (h *Host) dispatchApplicationCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) error { ... }
func (h *Host) dispatchMessageComponent(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error { ... }
func (h *Host) dispatchModalSubmit(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error { ... }
func (h *Host) dispatchAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error { ... }
```

---

#### 5.3.3 The `DiscordOps` struct

**Problem:** `DiscordOps` is a struct with 25 function-pointer fields. It is used as a **bag of callbacks** that gets passed through `DispatchRequest` into `buildContext`, then into `discordOpsObject`, which creates a JS object with nested namespaces.

**Where to look:** `internal/jsdiscord/bot.go`, lines 68–98 (`DiscordOps` definition) and lines 850–1050 (`discordOpsObject`).

**Why it matters:**
- Adding a new Discord API surface means editing 4 places: the struct, `buildDiscordOps`, `discordOpsObject` (both the `ops == nil` branch and the real branch).
- The nil-branch in `discordOpsObject` is 30 lines of no-op stubs. This is defensive but tedious.

**Cleanup sketch:**
Consider a registry pattern:

```go
type DiscordOp func(ctx context.Context, args ...any) (any, error)

type DiscordOpRegistry map[string]DiscordOp

func buildDiscordOps(scriptPath string, session *discordgo.Session) DiscordOpRegistry {
    return DiscordOpRegistry{
        "guilds.fetch":   makeGuildFetch(scriptPath, session),
        "roles.list":     makeRoleList(scriptPath, session),
        // ...
    }
}
```

Then `discordOpsObject` becomes a single loop over the registry keys, creating JS functions dynamically. The nil case is just an empty registry.

---

#### 5.3.4 The `DispatchRequest` struct

**Problem:** `DispatchRequest` has 25 fields. Most dispatchers only use a subset, but the struct forces every call site to know about all of them.

**Where to look:** `internal/jsdiscord/bot.go`, lines 100–136.

```go
type DispatchRequest struct {
    Name        string
    RootName    string
    SubName     string
    Args        map[string]any
    Values      any
    Command     map[string]any
    Interaction map[string]any
    Message     map[string]any
    Before      map[string]any
    User        map[string]any
    Guild       map[string]any
    Channel     map[string]any
    Member      map[string]any
    Reaction    map[string]any
    Me          map[string]any
    Metadata    map[string]any
    Config      map[string]any
    Component   map[string]any
    Modal       map[string]any
    Focused     map[string]any
    Discord     *DiscordOps
    Reply       func(context.Context, any) error
    FollowUp    func(context.Context, any) error
    Edit        func(context.Context, any) error
    Defer       func(context.Context, any) error
    ShowModal   func(context.Context, any) error
}
```

**Why it matters:**
- It is a "God struct." It knows too much.
- Call sites copy fields they don't need because the struct is flat.
- Adding a new context field (e.g., `Thread`) means editing this struct and every constructor.

**Cleanup sketch:**
Group fields into sub-structs:

```go
type DispatchRequest struct {
    Identity   DispatchIdentity     // Name, RootName, SubName
    Payload    DispatchPayload      // Args, Values, Command, Interaction, Message, Before, Component, Modal, Focused
    Context    DispatchContext      // User, Guild, Channel, Member, Reaction, Me
    Runtime    DispatchRuntime      // Metadata, Config
    Discord    *DiscordOps
    Responders Responders           // Reply, FollowUp, Edit, Defer, ShowModal
}
```

This makes call sites clearer:
```go
req := DispatchRequest{
    Identity: DispatchIdentity{Name: data.Name},
    Payload:  DispatchPayload{Args: args, Command: cmdMap},
    Context:  DispatchContext{User: interactionUserMap(i), Guild: guildMap(i.GuildID)},
    Runtime:  h.runtimeDispatch(),
    Discord:  buildDiscordOps(h.scriptPath, session),
    Responders: Responders{Reply: responder.Reply, ...},
}
```

---

### 5.4 Deprecated or legacy patterns

#### 5.4.1 Manual flag parsing in `botcli/run_schema.go`

**Problem:** The `bots run` command disables normal flag parsing and then implements its own flag parser (`preparseRunArgs`). It manually handles `--bot-token`, `--sync-on-start`, etc., then passes unknown flags into a dynamic Glazed schema parser.

**Where to look:** `internal/botcli/run_schema.go`, lines 1–346; `internal/botcli/command.go`, lines 80–140.

**Why it matters:**
- The CLI framework already knows how to parse flags. Reimplementing this is error-prone.
- The manual parser does not support shorthand flags.
- It is 200+ lines of string-splitting logic.
- Some flags appear in help because they are declared on the command, but their values are consumed by separate code paths, making it hard to know which parser is authoritative.

**Cleanup sketch — low-risk (recommended first):**
Keep the two-phase model if necessary, but make it explicit in the code structure:

```text
internal/botcli/
  run_static_args.go     // static flag parsing only
  run_dynamic_schema.go  // buildRunSchema + parseRuntimeConfigArgs
  run_help.go            // printRunSchema + selector-aware help rendering
```

Also rename the static phase to signal intent:

```go
func parseStaticRunnerArgs(...) (StaticRunnerArgs, error)
```

**Cleanup sketch — larger change (discuss before doing):**
Replace the custom pre-parser entirely. Define all known flags on the command normally. Collect remaining args after `--` as dynamic args:

```bash
discord-bot bots run ping --bot-repository ./examples -- --db-path ./data.sqlite
```

If the `--` separator is unacceptable, use `cmd.FParseErrWhitelist.UnknownFlags = true` and post-process unknown flags.

---

#### 5.4.2 Stale example artifacts

**Problem:** There are at least two stale example surfaces that no longer match the current single-bot `defineBot(...)` runtime model:

1. `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js` appears unreferenced.
2. `examples/bots/` still documents an old `__verb__`-based example repository even though current bot discovery explicitly looks for `defineBot` + `require("discord")`.

**Where to look:**
- `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js`
- `examples/bots/README.md` and `examples/bots/discord.js`
- `internal/botcli/bootstrap.go:184` — `looksLikeBotScript(...)`:

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

**Why it matters:**
This is exactly the kind of stale code that confuses a new intern:
- one directory says "this is the local example repository,"
- but the current discovery code will never treat its scripts as live bot implementations,
- and the knowledge-base bot has an older alternate registration file that is no longer wired in.

**Cleanup sketch:**
- Delete `examples/bots/` and `examples/bots-dupe-a/`, `examples/bots-dupe-b/`. Move anything valuable into `examples/discord-bots/`.
- Delete or archive `register-knowledge-bot.js`.
- Update `README.md`.

---

### 5.5 Runtime and maintenance concerns

#### 5.5.1 Goroutine safety

**Problem:** `interactionResponder` and `channelResponder` use `sync.Mutex` to protect their `acked`/`replied` flags. This is correct. However, `BotHandle.dispatch()` calls into the goja runtime, which is **not goroutine-safe**. The code uses `runtimeowner.Runner.Call()` to serialize access, which is good, but this is an external dependency (`go-go-goja`) and the safety contract is not documented in this repo.

**Where to look:** `internal/jsdiscord/bot.go`, lines 175–200 (`dispatch` method).

**Cleanup sketch:**
Add a comment at the top of `bot.go`:

```go
// BotHandle dispatch methods are NOT safe for concurrent use.
// All calls must be serialized by the runtime owner (runtimebridge.Lookup).
// See go-go-goja/pkg/runtimeowner for details.
```

---

#### 5.5.2 Error handling

**Problem:** Many functions return `nil` on nil input rather than an error. This is defensive but can hide bugs.

**Example:**
```go
func (h *Host) DispatchReady(...) error {
    if h == nil || h.handle == nil || ready == nil { return nil }
    ...
}
```

**Where to look:** `host_dispatch.go`, every dispatch method.

**Why it matters:**
- Calling `DispatchReady` on a nil Host is almost certainly a programmer error. Silently returning nil makes it harder to find.
- In Go, it is more idiomatic to let the caller ensure the receiver is non-nil, or to return an error if it is.

**Cleanup sketch:**
For new code, prefer early error returns:

```go
func (h *Host) DispatchReady(...) error {
    if h == nil { return fmt.Errorf("host is nil") }
    if h.handle == nil { return fmt.Errorf("host handle is nil") }
    if ready == nil { return fmt.Errorf("ready event is nil") }
    ...
}
```

This is a behavioral change, so it should be done gradually and with test updates.

---

#### 5.5.3 Promise polling busy-wait

**Problem:** `waitForPromise` polls the promise state every 5ms in a tight loop:

```go
for {
    select {
    case <-ctx.Done(): return nil, ctx.Err()
    default:
    }
    ret, err := owner.Call(ctx, "...", func(...) { ... })
    ...
    switch snapshot.State {
    case goja.PromiseStatePending:
        time.Sleep(5 * time.Millisecond)
    ...
    }
}
```

**Where to look:** `internal/jsdiscord/bot.go`, lines 245–280.

**Why it matters:**
- 5ms is arbitrary. For a bot that handles many concurrent interactions, this creates unnecessary CPU load.
- goja may expose a notification mechanism (e.g., channels or callbacks) that would allow event-driven resolution.

**Cleanup sketch:**
Document the tradeoff. If goja does not support event-driven promises, consider increasing the sleep to 10–20ms and adding a `sync.Cond` or channel-based notification if the goja runtime ever supports it.

---

#### 5.5.4 `internal/jsdiscord/runtime.go` exposes unused lifecycle surfaces

**Problem:** The runtime registrar still contains global runtime-state registration and lookup helpers whose public shape suggests they matter broadly, but they do not appear to be used outside `runtime.go` itself.

**Where to look:**
- `internal/jsdiscord/runtime.go:14` — `RuntimeStateContextKey`
- `internal/jsdiscord/runtime.go:16` — `runtimeStateByVM`
- `internal/jsdiscord/runtime.go:74` — `RegisterRuntimeState(...)`
- `internal/jsdiscord/runtime.go:81` — `UnregisterRuntimeState(...)`
- `internal/jsdiscord/runtime.go:88` — `LookupRuntimeState(...)`

**Example snippet:**
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

**Why it matters:**
Unused lifecycle surfaces imply an API contract that future maintainers may preserve unnecessarily. A newcomer reading this file may assume:
- some other subsystem relies on `LookupRuntimeState`,
- the context key is part of a broader contract,
- or VM-level state lookup is a supported extension seam.

If none of that is true anymore, the code should say so by becoming smaller.

**Cleanup sketch:**
Choose one of two options and document it explicitly.

**Option A — delete the unused surfaces:**
If lookup is no longer needed:
- remove `LookupRuntimeState`
- remove the unused context key if nothing reads it
- collapse the implementation to the minimum runtime registration required

**Option B — keep them, but write a comment:**
```go
// RuntimeStateContextKey and VM registration are retained as future extension seams
// for runtime-level inspectors. They are intentionally unused today.
```

Right now the code reads like an API in search of consumers.

---

## 6. Recommendations

These recommendations incorporate findings from both this review and a parallel review by a colleague. The priorities are ordered by risk and by what unblocks later work.

### Pass 1 — stale code and dead branches (lowest risk)

1. **Delete or archive stale artifacts:**
   - `examples/bots/` and `examples/bots-dupe-a/`, `examples/bots-dupe-b/`
   - `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js`
   - One commit. ~30 minutes.

2. **Remove dead fallback interaction code** from `internal/bot/bot.go` (`handleInteractionCreate` ping/echo branch).
   - One commit. ~15 minutes.

3. **Add a comment or delete unused surfaces** in `internal/jsdiscord/runtime.go` (`LookupRuntimeState`, context key, sync.Map).
   - One commit. ~15 minutes.

### Pass 2 — file-size cleanup without semantic changes (low risk)

4. **Split `internal/jsdiscord/bot.go`** by responsibility:
   ```text
   bot_compile.go    // CompileBot, botDraft registration/finalize
   bot_dispatch.go   // BotHandle dispatch + settleValue + promise waiting
   bot_context.go    // DispatchRequest, buildDispatchInput, buildContext
   bot_store.go      // storeObject
   bot_ops.go        // DiscordOps + discordOpsObject
   bot_logging.go    // loggerObject + applyFields
   ```
   - Pure move refactor. ~2 hours.

5. **Split `internal/jsdiscord/host_payloads.go`** by payload concern:
   ```text
   payload_model.go      // normalizedResponse + shared types
   payload_message.go    // message send/edit normalization
   payload_embeds.go     // embed normalization
   payload_components.go // component normalization
   payload_files.go      // file normalization
   payload_mentions.go   // allowedMentions
   ```
   - Pure move refactor. ~2 hours.

6. **Split large test files** by behavior family:
   ```text
   runtime_descriptor_test.go
   runtime_dispatch_test.go
   runtime_events_test.go
   runtime_payloads_test.go
   runtime_ops_messages_test.go
   runtime_ops_members_test.go
   runtime_ops_threads_test.go
   runtime_knowledge_base_test.go
   ```
   - Quality-of-life improvement. ~1 hour.

### Pass 3 — API clarity improvements (medium risk)

7. **Introduce typed internal envelopes** around the Go↔JS boundary (`DispatchEnvelope`, `InteractionSnapshot`, `MessageSnapshot`, etc.).
   - Converts `map[string]any` to typed structs internally; maps produced only at the JS boundary.
   - ~3 hours.

8. **Refactor `host_dispatch.go`** around base request builders:
   - `baseRequest()`, `eventRequest()`, `withChannelResponder()` helpers.
   - Eliminates 18 repetitions. ~1 hour.

9. **Make `internal/botcli` parsing phases explicit** in file and type names:
   - `preparseRunArgs` → `parseStaticRunnerArgs`
   - Split `run_schema.go` into `run_static_args.go`, `run_dynamic_schema.go`, `run_help.go`.
   - ~1 hour.

10. **Refactor `DiscordOps`** into a registry pattern (`map[string]DiscordOp`).
    - Reduces 4-place edit burden for new API surfaces. ~3 hours.

### Pass 4 — larger changes (discuss before doing)

11. **Replace manual flag parsing** with native CLI parsing.
    - May require `--` separator for dynamic args. ~3 hours.

12. **Code-generate map converters** (`host_maps.go`) from discordgo structs.
    - Requires a small codegen tool. ~1 day.

13. **Add structured error types** instead of `fmt.Errorf` everywhere.
    - Would enable better error classification in JS. ~1 day.

14. **Evaluate event-driven promise resolution** instead of polling.
    - Depends on goja capabilities.

### 6.1 Low-risk refactors (do these first)

1. **Delete deprecated examples** (`examples/bots/`, `examples/bots-dupe-*`). One commit. ~30 minutes.
2. **Extract test helpers** into `internal/jsdiscord/testutil/`. One commit. ~1 hour.
3. **Add `baseDispatchRequest` helper** to eliminate 18 repetitions in `host_dispatch.go`. One commit. ~30 minutes.
4. **Extract `DispatchInteraction` branches** into private methods. One commit. ~1 hour.
5. **Extract `makeDispatchCommand` etc.** from `finalize`. One commit. ~1 hour.

### 6.2 Medium-risk reorganizations (do after low-risk)

1. **Split `bot.go`** into 4–5 files by responsibility. This is a pure move refactor; no logic changes. ~2 hours.
2. **Split `host_payloads.go`** into a `payload/` subpackage. ~2 hours.
3. **Introduce `DispatchBuilder`** or `DispatchRequest` sub-structs. This changes many call sites but is mechanical. ~3 hours.
4. **Refactor `DiscordOps`** into a registry pattern. Reduces the 4-place edit burden for new API surfaces. ~3 hours.

### 6.3 Larger architectural changes (discuss before doing)

1. **Replace manual flag parsing in `botcli`** with Cobra-native parsing. This changes the CLI UX slightly (may require `--` separator). Discuss with users first.
2. **Code-generate map converters** (`host_maps.go`) from discordgo structs. Requires a small codegen tool. High payoff if the API surface grows.
3. **Add structured error types** instead of `fmt.Errorf` everywhere. Would enable better error classification in JS.
4. **Evaluate event-driven promise resolution** instead of polling. Depends on goja capabilities.

---

## 7. Appendix: File Reference Index

| File | Lines | Responsibility | Quality note |
|------|-------|----------------|--------------|
| `internal/jsdiscord/bot.go` | 1,293 | BotHandle, draft, finalize, context builders | **Too large**; split into 4–5 files |
| `internal/jsdiscord/runtime_test.go` | 1,205 | Unit tests for JS bridge | Move helpers to testutil |
| `internal/jsdiscord/host_payloads.go` | 736 | Payload normalization | **Too large**; split into `payload/` subpackage |
| `internal/jsdiscord/host_dispatch.go` | 593 | Event dispatch to JS | **Too large**; extract methods, builder |
| `internal/jsdiscord/descriptor.go` | 397 | BotDescriptor parsing | Use generic `parseDescriptors` helper |
| `internal/jsdiscord/host_maps.go` | 369 | discordgo → map conversion | Consider codegen |
| `internal/bot/bot.go` | 350 | Discord session host | Extract generic handler wrapper |
| `internal/botcli/run_schema.go` | 346 | Runtime config pre-parser | Replace with Cobra native parsing |
| `internal/jsdiscord/host_ops_helpers.go` | 303 | Ops normalization helpers | Fine; keep as-is |
| `internal/jsdiscord/knowledge_base_runtime_test.go` | 298 | Integration test for KB bot | Use shared testutil |
| `internal/jsdiscord/host_commands.go` | 243 | ApplicationCommand building | Fine; could use generics for sorting |
| `internal/botcli/command_test.go` | 242 | BotCLI tests | Fine |
| `cmd/discord-bot/commands.go` | 219 | Direct host commands | Fine |
| `internal/jsdiscord/host_responses.go` | 217 | Response emitters | Fine |
| `internal/botcli/command.go` | 215 | Bots subcommands | Fine |
| `internal/botcli/bootstrap.go` | 201 | Bot discovery | Fine; well-structured |
| `internal/jsdiscord/host_ops_messages.go` | 163 | Message ops closures | Repetitive pattern; use helper |
| `internal/jsdiscord/host_ops_members.go` | 163 | Member ops closures | Repetitive pattern; use helper |
| `internal/jsdiscord/store.go` | 148 | In-memory KV store | Fine |
| `internal/jsdiscord/runtime.go` | 147 | Module registration | **Unused surfaces**; document or delete |
| `internal/botcli/runtime.go` | 108 | Run loop | Fine |
| `internal/botcli/model.go` | 99 | Bot model types | Fine |
| `internal/jsdiscord/host.go` | 95 | Host loader | Fine |
| `internal/jsdiscord/host_ops_threads.go` | 87 | Thread ops closures | Repetitive pattern; use helper |
| `internal/jsdiscord/host_logging.go` | 87 | Logging helpers | Fine |
| `internal/jsdiscord/host_ops_channels.go` | 74 | Channel ops closures | Repetitive pattern; use helper |
| `internal/jsdiscord/descriptor_test.go` | 70 | Descriptor tests | Fine |
| `cmd/discord-bot/root.go` | 65 | Root command setup | Fine |
| `internal/config/config.go` | 61 | Config types | Fine |

---

*End of report.*
