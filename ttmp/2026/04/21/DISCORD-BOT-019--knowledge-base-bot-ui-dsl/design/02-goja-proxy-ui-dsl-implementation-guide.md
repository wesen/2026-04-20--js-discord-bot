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

### The `resolveX` helpers

When `row()` receives arguments, they could be:
- A Goja Proxy wrapping a Go builder (the happy path)
- A plain `map[string]any` from raw JS (the backward-compat path)

The `resolveX` helpers detect which and handle both:

```go
func resolveComponent(vm *goja.Runtime, arg goja.Value) discordgo.MessageComponent {
    // Try to unwrap our Go builder
    if proxy, ok := arg.(*goja.Proxy); ok {
        if builder, ok := extractBuilder(proxy); ok {
            return builder.Build() // returns discordgo.MessageComponent
        }
    }
    // Fall back to normalizing raw JS objects
    mapping, _ := arg.Export().(map[string]any)
    comp, _ := normalizeComponent(mapping)
    return comp
}
```

This means **existing bots that return raw JS objects still work**. The Go-side builders are an additional fast path.

### The `build()` bridge

When JS calls `.build()` on a message builder, the Go side needs to produce something the existing `normalizePayload()` in `payload_model.go` can consume. There are two options:

**Option A: Return `map[string]any` (simplest, immediate compatibility)**

```go
func buildMessagePayload(b *MessageBuilder) map[string]any {
    payload := map[string]any{}
    if b.content != "" {
        payload["content"] = b.content
    }
    if b.ephemeral {
        payload["ephemeral"] = true
    }
    if len(b.embeds) > 0 {
        payload["embeds"] = b.embeds
    }
    if len(b.components) > 0 {
        payload["components"] = b.components
    }
    // ...
    return payload
}
```

The existing `normalizePayload()` already handles `map[string]any` with `[]*discordgo.MessageEmbed` and `[]discordgo.MessageComponent` values — it was written generically enough. This is the recommended first step.

**Option B: Return a typed `normalizedResponse` (future optimization)**

Later, add a fast path in `normalizePayload()`:

```go
func normalizePayload(payload any) (*normalizedResponse, error) {
    // Fast path for Go-side builders
    if nr, ok := payload.(*normalizedResponse); ok {
        return nr, nil
    }
    // ... existing map[string]any path
}
```

This skips all the map parsing entirely. Only worth doing once the builders prove stable.

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
  ui_message.go       — MessageBuilder, newMessageBuilder(vm)
  ui_embed.go         — EmbedBuilder, newEmbedBuilder(vm, title)
  ui_components.go    — ButtonBuilder, SelectBuilder, FormBuilder
  ui_selects.go       — UserSelectBuilder, RoleSelectBuilder, etc.
  ui_helpers.go       — row, pager, actions, confirm, ok, error, card, emptyResults
  ui_flow.go          — FlowHelper, newFlowHelper(vm, namespace, init)
  ui_resolve.go       — resolveComponent, resolveEmbed, resolveRow, extractBuilder
```

No changes to existing files during implementation. The `UIRegistrar` is wired into the host separately.

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

## Error messages design

Go-side builders can produce clear, actionable errors:

```
ui.button("", "Click", "primary")
→ panic: ui.button: customId must not be empty

ui.button("id", "Click", "badstyle")
→ panic: ui.button: unknown style "badstyle", use one of: primary, secondary, success, danger, link

ui.select("id").option("", "value")
→ panic: ui.select.option: label must not be empty

ui.message().embed("not a builder")
→ panic: ui.message.embed: expected an embed builder (from ui.embed()), got string

ui.form("", "Title")
→ panic: ui.form: customId must not be empty
```

Compare with the old error:
```
unsupported component type %!q(<nil>)
```

## Performance notes

The Proxy approach adds a function call indirection per property access. In practice this is negligible because:
1. Builder chains are short (5–15 method calls per screen)
2. The heavy work is Discord API calls, not JS→Go dispatch
3. The alternative (raw JS objects → Go normalization) also involves reflection

If profiling ever shows Proxy overhead, the fast path (`*normalizedResponse` in `normalizePayload`) eliminates the map entirely.
