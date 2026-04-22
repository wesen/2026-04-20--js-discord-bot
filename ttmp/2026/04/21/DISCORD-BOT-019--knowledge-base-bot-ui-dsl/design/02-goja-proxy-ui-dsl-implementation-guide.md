---
Title: Goja Proxy-Based UI DSL Implementation Guide
Ticket: DISCORD-BOT-019
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/jsdiscord/runtime.go
      Note: RuntimeState and Registrar — the pattern to follow for registering a new native module
    - Path: internal/jsdiscord/payload_components.go
      Note: Current component normalization — will be bypassed by Go-side builders
    - Path: internal/jsdiscord/payload_model.go
      Note: normalizePayload — the final consumer of builder output; needs a discordgo.MessageSend fast path
    - Path: internal/jsdiscord/bot_store.go
      Note: storeObject pattern — simple Go object → JS API example
    - Path: internal/jsdiscord/ui_showcase_runtime_test.go
      Note: Integration tests that will validate the Go-side builders end-to-end
    - Path: ttmp/2026/04/21/DISCORD-BOT-019--knowledge-base-bot-ui-dsl/scripts/01-proxy-poc.go
      Note: Proof of concept proving Goja Proxy works for builder pattern
    - Path: examples/discord-bots/ui-showcase/lib/ui/primitives.js
      Note: Current JS-side builders — will be replaced by require("ui") from Go
ExternalSources: []
Summary: Detailed implementation guide for a Go-native UI DSL module using Goja Proxy traps, replacing the JS-side builder approach with validated Go structs.
LastUpdated: 2026-04-22T00:00:00-04:00
WhatFor: Guide the implementation of require("ui") as a Go-native Goja module with Proxy-based builders.
WhenToUse: Use when implementing the Go-side UI DSL module, writing the builder structs, or integrating with the existing dispatch pipeline.
---

# Goja Proxy-Based UI DSL Implementation Guide

## Why Go-side builders?

The JS-side builders in `examples/discord-bots/ui-showcase/lib/ui/primitives.js` proved the DSL concept works, but exposed a fundamental problem: **JS chain objects leak through to the Go host**. The chain pattern (`select().placeholder("Pick").option("A", "a")`) returns a JS object with methods on it. When that object flows into `row()` → `message()` → Discord payload → Go's `normalizeComponents()`, the Go side sees `map[string]any` with no `"type"` key and crashes with `unsupported component type %!q(<nil>)`.

The error happens three layers deep in Go's payload normalization, far from the JS code that caused it. A Go-side DSL fixes this at the root:

1. **Builders are Go structs wrapped in Goja Proxies.** JS never sees raw data properties — only the chain methods the Go side exposes.
2. **Validation happens at construction time.** `ui.button("", "", "badstyle")` fails immediately with a clear message, not deep in `normalizeLeafComponent`.
3. **No `map[string]any` round-trip.** Go builders produce `discordgo.MessageComponent` trees directly.
4. **The existing `payload_components.go` stays** as a safety net for raw JS objects — no breaking change.

Proof of concept: `scripts/01-proxy-poc.go` confirms Goja's `vm.NewProxy(target, &ProxyTrapConfig{Get: ...})` works perfectly for builder chains.

## Architecture

### Module registration

Register `require("ui")` as a new native Goja module alongside `require("discord")`:

```
internal/jsdiscord/
  ui_module.go        — Registrar + Loader, registers the "ui" module
  ui_message.go       — MessageBuilder struct + Proxy
  ui_embed.go         — EmbedBuilder struct + Proxy
  ui_components.go    — ButtonBuilder, SelectBuilder, FormBuilder + Proxies
  ui_helpers.go       — row(), pager(), actions(), confirm(), ok(), error()
  ui_flow.go          — FlowHelper for stateful screens
```

The registration follows the same pattern as `runtime.go`:

```go
// ui_module.go
type UIRegistrar struct{}

func (r *UIRegistrar) ID() string { return "discord-ui-registrar" }

func (r *UIRegistrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
    reg.RegisterNativeModule("ui", UILoader)
    return nil
}

func UILoader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    _ = exports.Set("message", func(call goja.FunctionCall) goja.Value { return newMessageBuilder(vm) })
    _ = exports.Set("embed", func(call goja.FunctionCall) goja.Value { ... })
    _ = exports.Set("button", func(call goja.FunctionCall) goja.Value { ... })
    _ = exports.Set("select", func(call goja.FunctionCall) goja.Value { ... })
    _ = exports.Set("form", func(call goja.FunctionCall) goja.Value { ... })
    _ = exports.Set("flow", func(call goja.FunctionCall) goja.Value { ... })
    // helpers
    _ = exports.Set("row", ...)
    _ = exports.Set("pager", ...)
    _ = exports.Set("actions", ...)
    _ = exports.Set("confirm", ...)
    _ = exports.Set("ok", ...)
    _ = exports.Set("error", ...)
    _ = exports.Set("card", ...)
    _ = exports.Set("emptyResults", ...)
}
```

Registration in the host happens through the existing engine registrar pipeline — the `UIRegistrar` is added next to the `discord` Registrar.

### The Proxy pattern

Each builder is a Go struct that owns its data. A Goja Proxy wraps it:

```go
type MessageBuilder struct {
    content    string
    ephemeral  bool
    embeds     []*discordgo.MessageEmbed
    components []discordgo.MessageComponent
    files      []*discordgo.File
    tts        bool
}

func newMessageBuilder(vm *goja.Runtime) goja.Value {
    target := vm.NewObject()
    b := &MessageBuilder{}

    proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
        Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
            switch property {
            case "content":
                return vm.ToValue(func(call goja.FunctionCall) goja.Value {
                    b.content = call.Argument(0).String()
                    return receiver // chain
                })
            case "ephemeral":
                return vm.ToValue(func(call goja.FunctionCall) goja.Value {
                    b.ephemeral = true
                    return receiver
                })
            case "embed":
                return vm.ToValue(func(call goja.FunctionCall) goja.Value {
                    // Accept either a goja.Proxy (embed builder) or raw object
                    emb := resolveEmbed(vm, call.Argument(0))
                    b.embeds = append(b.embeds, emb)
                    return receiver
                })
            case "row":
                return vm.ToValue(func(call goja.FunctionCall) goja.Value {
                    // Auto-build each component, wrap in ActionsRow
                    row := resolveRow(vm, call.Arguments)
                    b.components = append(b.components, row)
                    return receiver
                })
            case "file":
                return vm.ToValue(func(call goja.FunctionCall) goja.Value {
                    // ... add file
                    return receiver
                })
            case "build":
                return vm.ToValue(func(call goja.FunctionCall) goja.Value {
                    // Return a map[string]any that normalizePayload already understands.
                    // This is the bridge: Go structs → existing pipeline.
                    payload := buildMessagePayload(b)
                    return vm.ToValue(payload)
                })
            default:
                return goja.Undefined()
            }
        },
    })
    return vm.ToValue(proxy)
}
```

### No backward compat — builders are the only path

The `ui` module is strict: `row()`, `embed()`, `form()`, etc. only accept Go builder proxies. Passing a raw JS object produces a clear error like `"expected a ui.button() builder, got object"`. This eliminates the `resolveX` bridge layer entirely.

Existing bots that don't use `require("ui")` continue returning raw JS objects to the dispatch pipeline — that path (`normalizePayload()` → `map[string]any` normalization) is unchanged. The `ui` module is opt-in: use it for clean payloads, or return raw objects if you prefer. But you can't mix them.

### The `build()` bridge — typed fast path

When JS calls `.build()` on a message builder, the Go side returns a `*normalizedResponse` directly. A single 3-line addition to `normalizePayload()` handles it:

```go
func normalizePayload(payload any) (*normalizedResponse, error) {
    switch v := payload.(type) {
    case *normalizedResponse:
        return v, nil          // ← Go-side builder fast path
    case nil:
        return &normalizedResponse{}, nil
    case string:
        // ... existing cases unchanged
```

This skips all `map[string]any` parsing. No `payload_components.go` involvement for `ui` module users. The builder already holds `[]*discordgo.MessageEmbed` and `[]discordgo.MessageComponent` — the exact types the dispatch pipeline needs.

Bots that don't use `require("ui")` continue to hit the `map[string]any` case as before. No breakage.

## Builder specifications

### `ui.message()`

| Method | Args | Validates | Mutates |
|--------|------|-----------|---------|
| `.content(text)` | string | — | Sets content |
| `.ephemeral()` | — | — | Sets ephemeral=true |
| `.tts()` | — | — | Sets tts=true |
| `.embed(e)` | embed builder or object | Must be valid embed | Appends to embeds |
| `.row(...components)` | component builders | Max 5 rows, max 5 per row | Appends ActionsRow |
| `.file(name, content)` | string, string | name required | Appends File |
| `.build()` | — | — | Returns payload |

### `ui.embed(title)`

| Method | Args | Validates | Mutates |
|--------|------|-----------|---------|
| `.description(text)` | string | Max 4096 chars | Sets description |
| `.color(value)` | number | 0–0xFFFFFF | Sets color |
| `.field(name, value, inline?)` | string, string, bool | name/value required | Appends field (max 25) |
| `.footer(text)` | string | Max 2048 chars | Sets footer |
| `.author(name)` | string | — | Sets author |
| `.timestamp()` | — | — | Sets timestamp to now |
| `.build()` | — | — | Returns `*discordgo.MessageEmbed` |

### `ui.button(customId, label, style)`

| Method | Args | Validates | Mutates |
|--------|------|-----------|---------|
| `.disabled(flag?)` | bool | — | Sets disabled |
| `.emoji(name)` | string | — | Sets emoji |
| `.url(href)` | string | Must be URL | Sets url (changes to link style) |
| `.build()` | — | — | Returns `discordgo.Button` |

Style validation happens at construction: `"primary"`, `"secondary"`, `"success"`, `"danger"`, `"link"`. Unknown style → immediate panic with clear message.

### `ui.select(customId)`

| Method | Args | Validates | Mutates |
|--------|------|-----------|---------|
| `.placeholder(text)` | string | Max 150 chars | Sets placeholder |
| `.option(label, value, description?)` | string, string, string | label/value required, max 25 options | Appends option |
| `.minValues(n)` | number | 0–25 | Sets minValues |
| `.maxValues(n)` | number | 0–25 | Sets maxValues |
| `.disabled(flag?)` | bool | — | Sets disabled |
| `.build()` | — | — | Returns `discordgo.SelectMenu` |

Variant builders for typed selects:
- `ui.userSelect(customId)` — type=userSelect
- `ui.roleSelect(customId)` — type=roleSelect
- `ui.channelSelect(customId)` — type=channelSelect
- `ui.mentionableSelect(customId)` — type=mentionableSelect

### `ui.form(customId, title)`

| Method | Args | Validates | Mutates |
|--------|------|-----------|---------|
| `.text(id, label)` | string, string | id/label required | Starts text field |
| `.textarea(id, label)` | string, string | id/label required | Starts paragraph field |
| `.required(flag?)` | bool | — | Marks current field required |
| `.value(text)` | string | — | Sets current field default value |
| `.placeholder(text)` | string | — | Sets current field placeholder |
| `.min(n)` | number | — | Sets minLength |
| `.max(n)` | number | — | Sets maxLength |
| `.build()` | — | customId + title required | Returns modal payload map |

### `ui.flow(namespace, options)`

The flow helper manages per-user per-channel state using `ctx.store`. It does not produce Discord payloads — it's a state management tool.

| Method | Args | Returns |
|--------|------|---------|
| `.load(ctx)` | context | Current state object |
| `.save(ctx, state)` | context, object | Saved state |
| `.clear(ctx)` | context | — |
| `.id(suffix)` | string | `"namespace:suffix"` |
| `.componentIds(names)` | array of strings | `{ name: "namespace:name", ... }` |
| `.pagerIds()` | — | `{ previous: "namespace:previous", next: "namespace:next" }` |

Implementation: wraps `ctx.store.get(key)` / `ctx.store.set(key, value)` with a key derived from `guildId.channelId.userId`.

### Helper functions

These are simple functions (not builders) that return complete payloads or components:

- `ui.row(...components)` — wraps auto-built components in `ActionsRow`
- `ui.pager(prevId, nextId, {hasPrevious, hasNext})` — returns `ActionsRow` with Previous/Next buttons
- `ui.actions([{id, label, style}, ...])` — returns `ActionsRow` with buttons
- `ui.confirm(confirmId, cancelId, {title, body, ...})` — returns complete ephemeral message payload
- `ui.card(title)` — returns an embed builder with `.meta()` shortcut
- `ui.ok(content)` — returns `{content, ephemeral: true}`
- `ui.error(content)` — returns `{content: "⚠️ " + content, ephemeral: true}`
- `ui.emptyResults(query)` — returns ephemeral "no results" message

## Integration with existing pipeline

The key insight is that `normalizePayload()` in `payload_model.go` already handles typed Go values:

```go
// From payload_model.go:
case map[string]any:
    ret := &normalizedResponse{}
    // ...
    embeds, err := normalizeEmbeds(v)        // handles []*discordgo.MessageEmbed
    components, err := normalizeComponents(v["components"]) // handles []discordgo.MessageComponent
```

And `normalizeComponents()` already handles typed values:

```go
case []discordgo.MessageComponent:
    return v, nil
```

So when a Go builder's `.build()` returns a `map[string]any` containing `[]discordgo.MessageComponent` and `[]*discordgo.MessageEmbed`, the existing pipeline **already works**. No changes needed in `payload_model.go` or `payload_components.go`.

The only change is in `normalizePayload()` — add a case for the builder's output type:

```go
// In normalizePayload, the map[string]any case already works.
// Optionally, add a fast path later:
case *normalizedResponse:
    return v, nil
```

## File plan

```
internal/jsdiscord/
  ui_module.go        — UIRegistrar, UILoader, export table
  ui_message.go       — MessageBuilder struct + Proxy, build() returns *normalizedResponse
  ui_embed.go         — EmbedBuilder struct + Proxy
  ui_components.go    — ButtonBuilder, SelectBuilder, FormBuilder + Proxies
  ui_selects.go       — UserSelectBuilder, RoleSelectBuilder, etc.
  ui_helpers.go       — row, pager, actions, confirm, ok, error, card, emptyResults
  ui_flow.go          — FlowHelper for stateful screens
  ui_errors.go        — methodOwner map, wrongParentError, typeMismatchError
```

Changes to existing files:
- `payload_model.go` — add 1 case (`*normalizedResponse`) to `normalizePayload()`
- Host registration — add `UIRegistrar` alongside existing `Registrar`

## Testing strategy

### Unit tests (Go only, no JS)

```go
func TestMessageBuilder(t *testing.T) {
    vm := goja.New()
    msg := newMessageBuilder(vm)
    // ... call methods, verify build() output
}
```

### Integration tests (through the dispatch pipeline)

The existing `ui_showcase_runtime_test.go` tests are exactly right — they call `DispatchCommand` and verify the response shape. When the Go-side builders replace the JS ones, these same tests should pass unchanged. That's the migration validation.

### Proxy behavior tests

```go
func TestProxyBuilderChainReturnsSelf(t *testing.T) {
    // Verify that sel.placeholder("x") returns the proxy (for chaining)
    // Verify that sel.build() returns the raw target (not the proxy)
    // Verify that sel.unknownProp returns undefined
}
```

## Migration path from JS to Go builders

### Phase 1: Add Go module alongside JS

1. Implement `require("ui")` Go module with all builders
2. The JS `lib/ui/primitives.js` becomes `module.exports = require("ui")`
3. `screen.js` (flow helper) stays in JS — it's domain logic, not Discord infrastructure
4. All existing tests continue to pass

### Phase 2: Validate and remove JS fallback

1. Add comprehensive Go unit tests for each builder
2. Run the existing integration tests against the Go builders
3. Remove `lib/ui/primitives.js` and `lib/ui/index.js`
4. The `ui-showcase` bot's `require("./lib/ui")` now resolves to `require("ui")` from Go

### What stays in JS

- `lib/demo-store.js` — domain data, no Discord API involvement
- `lib/ui/screen.js` — flow state management using `ctx.store`, pure JS logic
- `index.js` — bot wiring, command handlers, domain composition

### What gets removed from JS

- `lib/ui/primitives.js` — replaced by Go native module
- `lib/ui/index.js` — no longer needed

## Error messages and wrong-parent detection

The Go-side builders track their own type identity. Each Proxy's `Get` trap knows what struct it wraps, so it can distinguish between "unknown method" and "method that belongs to a different builder".

### The method-owner map

`ui_errors.go` holds a global lookup:

```go
var methodOwner = map[string]string{
    // MessageBuilder
    "content": "ui.message()",
    "ephemeral": "ui.message()",
    "tts":      "ui.message()",
    "embed":    "ui.message()",
    "row":      "ui.message()",
    "file":     "ui.message()",

    // EmbedBuilder
    "description": "ui.embed()",
    "color":      "ui.embed()",
    "field":      "ui.embed()",
    "fields":     "ui.embed()",
    "footer":     "ui.embed()",
    "author":     "ui.embed()",
    "timestamp":  "ui.embed()",

    // ButtonBuilder
    "disabled": "ui.button()",
    "emoji":    "ui.button()",

    // SelectBuilder
    "placeholder": "ui.select()",
    "option":      "ui.select()",
    "options":     "ui.select()",
    "minValues":   "ui.select()",
    "maxValues":   "ui.select()",

    // FormBuilder
    "text":     "ui.form()",
    "textarea": "ui.form()",
    "required": "ui.form()",
    "value":    "ui.form()",
    "min":      "ui.form()",
    "max":      "ui.form()",
}
```

### How each builder uses it

Every Proxy `Get` trap has three branches:

```go
Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
    // 1. Methods that belong to THIS builder → return the chain function
    switch property {
    case "content", "ephemeral", "embed", "row", "build":
        return /* the method */
    }

    // 2. Methods that belong to a DIFFERENT builder → "wrong parent" error
    if owner, ok := methodOwner[property]; ok {
        panic(vm.NewTypeError(
            "ui.message: .%s() is not available here. "+
            "You probably meant to call this on %s.",
            property, owner))
    }

    // 3. Truly unknown → generic error
    panic(vm.NewTypeError(
        "ui.message: unknown method .%s(). "+
        "Available: content, ephemeral, embed, row, file, build.",
        property))
}
```

### Examples

**Wrong parent — method exists but on the wrong builder:**
```
ui.button("id", "Click", "primary").ephemeral()
→ Error: ui.button: .ephemeral() is not available here.
         You probably meant to call this on ui.message().
```

```
ui.button("id", "Click", "primary").footer("text")
→ Error: ui.button: .footer() is not available here.
         You probably meant to call this on ui.embed().
```

**Bad arguments at construction:**
```
ui.button("", "Click", "primary")
→ Error: ui.button: customId must not be empty

ui.button("id", "Click", "badstyle")
→ Error: ui.button: unknown style "badstyle", use one of: primary, secondary, success, danger, link
```

**Type mismatch — raw JS object where a builder is expected:**
```
ui.message().embed({title: "Hello"})
→ Error: ui.message.embed: expected a ui.embed() builder, got object.
         Use ui.embed("Hello").description("...") to create an embed.

ui.message().row({type: "button", label: "Click"})
→ Error: ui.message.row: expected a ui.button() or ui.select() builder, got object.
         Use ui.button("id", "Label", "primary") to create a button.
```

**Truly unknown method:**
```
ui.message().unknownThing()
→ Error: ui.message: unknown method .unknownThing().
         Available: content, ephemeral, embed, row, file, build.
```

### Raw JS objects are errors in ui methods

The `ui` module does **not** accept raw `map[string]any` or plain JS objects as component/embed arguments. This is intentional:

- `ui.message().embed(<raw object>)` → type error, "expected a ui.embed() builder"
- `ui.message().row(<raw object>)` → type error, "expected a ui.button() or ui.select() builder"
- `ui.row(<raw object>)` → type error

Bots that don't use `require("ui")` continue to return raw JS objects through the existing `normalizePayload()` → `map[string]any` path. That pipeline is unchanged and still works. The `ui` module is opt-in: use it for validated, type-safe payloads, or return raw objects if you prefer — but you can't mix them.

In the future, heuristics could be added to detect common raw-object shapes and suggest the equivalent builder call:
```
ui.message().embed({title: "Hello", color: 0xFF0000})
→ Error: ui.message.embed: expected a ui.embed() builder, got object.
         Did you mean: ui.embed("Hello").color(0xFF0000)?
```

This is a future enhancement, not a day-one requirement.

### Compare with the old errors

**Before (JS-side builders):**
```
unsupported component type %!q(<nil>)
```
Three layers deep in Go normalization, no indication what JS code caused it.

**Before (raw JS typo):**
```
ui.message().ephemral()
→ TypeError: ui.message(...).ephemral is not a function
```
Silent `undefined`, crashes later at an unrelated location.

**After (Go-side builders):**
```
ui.message().ephemral()
→ Error: ui.message: unknown method .ephemral().
         Available: content, ephemeral, embed, row, file, build.
         Did you mean: ephemeral()?
```

## Performance notes

The Proxy approach adds a function call indirection per property access. In practice this is negligible because:
1. Builder chains are short (5–15 method calls per screen)
2. The heavy work is Discord API calls, not JS→Go dispatch
3. The alternative (raw JS objects → Go normalization) also involves reflection

If profiling ever shows Proxy overhead, the fast path (`*normalizedResponse` in `normalizePayload`) eliminates the map entirely.
