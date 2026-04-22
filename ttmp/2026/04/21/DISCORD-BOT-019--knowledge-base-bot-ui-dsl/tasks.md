# Tasks

## ✅ Completed: documentation and analysis

- [x] Create ticket workspace
- [x] Inspect the current knowledge-base bot UI surface
- [x] Write a detailed design brainstorm for a possible UI DSL
- [x] Write concrete example DSL sketches for multiple use cases
- [x] Add and maintain a diary while working
- [x] Decide on implementation approach — Goja Proxy-based Go-side builders
- [x] Map the current search/review/form UI architecture
- [x] Identify concrete duplication and awkwardness in the current interaction layer
- [x] Propose multiple DSL design directions with trade-offs
- [x] Recommend one preferred implementation path
- [x] Prove Goja Proxy works for builders (`scripts/01-proxy-poc.go`)
- [x] Write Goja Proxy implementation guide (`design/02-goja-proxy-ui-dsl-implementation-guide.md`)

## ✅ Completed: JS prototype (superseded by Go approach)

- [x] Create `examples/discord-bots/ui-showcase/` example bot skeleton
- [x] Create JS-side `lib/ui/` with generic UI builder primitives
- [x] Wire up the showcase bot using the JS UI DSL
- [x] Write Go integration tests (`ui_showcase_runtime_test.go`)
- [x] Fix chain-object-leaking bug (auto-build in row())
- [x] Update README.md

---

## Phase 1: Infrastructure — module registration and error system

Goal: `require("ui")` loads from Go, exports all functions. Error system in place. No builders yet — just the skeleton.

- [ ] **1.1** Create `internal/jsdiscord/ui_module.go`
  - `UIRegistrar` struct with `ID()` and `RegisterRuntimeModules()`
  - `UILoader(vm, moduleObj)` — sets all export names on `moduleObj.Get("exports")`
  - Export names: `message`, `embed`, `card`, `button`, `select`, `userSelect`, `roleSelect`, `channelSelect`, `mentionableSelect`, `form`, `flow`, `row`, `pager`, `actions`, `confirm`, `ok`, `error`, `emptyResults`
  - Each export starts as a stub that panics with `"not yet implemented"`

- [ ] **1.2** Wire `UIRegistrar` into `internal/jsdiscord/host.go`
  - Add `&UIRegistrar{}` to the `WithRuntimeModuleRegistrars()` call (line ~32)
  - Currently: `WithRuntimeModuleRegistrars(NewRegistrar(Config{}))`
  - Becomes: `WithRuntimeModuleRegistrars(NewRegistrar(Config{}), &UIRegistrar{})`

- [ ] **1.3** Add `*normalizedResponse` fast path to `internal/jsdiscord/payload_model.go`
  - Add `case *normalizedResponse: return v, nil` as first case in `normalizePayload()`

- [ ] **1.4** Create `internal/jsdiscord/ui_errors.go`
  - `methodOwner` map: every known DSL method → the builder it belongs to (e.g., `"ephemeral" → "ui.message()"`, `"footer" → "ui.embed()"`)
  - `wrongParentError(vm, builderName, method, owner)` — returns `vm.NewTypeError`
  - `unknownMethodError(vm, builderName, method, available)` — returns `vm.NewTypeError` listing available methods
  - `typeMismatchError(vm, builderName, method, expected, got)` — returns `vm.NewTypeError` with suggestion

- [ ] **1.5** Write `internal/jsdiscord/ui_module_test.go`
  - Test: `require("ui")` loads without error
  - Test: all expected export names exist and are functions
  - Test: calling a stub export panics with "not yet implemented"

- [ ] **1.6** Run full test suite — verify nothing breaks
  - `go test ./internal/jsdiscord/ -v`
  - Existing bots that don't use `require("ui")` should be unaffected

**Commit:** `feat(ui): add require("ui") module skeleton with error system`

---

## Phase 2: Core builders — message and embed

Goal: `ui.message()` and `ui.embed()` work end-to-end. A JS bot can build a message with an embed and return it through the dispatch pipeline.

- [ ] **2.1** Create `internal/jsdiscord/ui_embed.go`
  - `EmbedBuilder` struct: title, description, color, fields, footer, author, timestamp
  - `newEmbedBuilder(vm, title)` — returns Goja Proxy wrapping the struct
  - Proxy `Get` trap: three branches (own methods, wrong-parent via methodOwner, unknown)
  - Methods: `title()`, `description()`, `color()`, `field()`, `fields()`, `footer()`, `author()`, `timestamp()`, `build()`
  - `build()` returns `*discordgo.MessageEmbed`
  - Validation: max 25 fields, color 0–0xFFFFFF, description max 4096

- [ ] **2.2** Write `internal/jsdiscord/ui_embed_test.go`
  - Test: chain returns self for each method
  - Test: `build()` returns `*discordgo.MessageEmbed`
  - Test: fields accumulate correctly
  - Test: validation (too many fields, bad color)
  - Test: wrong-parent error (e.g., `.ephemeral()` → "you probably meant ui.message()")
  - Test: unknown method error lists available methods
  - Test: raw JS object passed to `.field()` → type error

- [ ] **2.3** Create `internal/jsdiscord/ui_message.go`
  - `MessageBuilder` struct: content, ephemeral, tts, embeds, components, files
  - `newMessageBuilder(vm)` — returns Goja Proxy
  - Proxy `Get` trap: three branches (own, wrong-parent, unknown)
  - Methods: `content()`, `ephemeral()`, `tts()`, `embed()`, `row()`, `file()`, `build()`
  - `.embed(e)` — accepts only `*discordgo.MessageEmbed` or embed builder Proxy; rejects raw objects
  - `.row(...components)` — accepts only button/select builder Proxies; auto-builds each, wraps in `discordgo.ActionsRow`; rejects raw objects
  - `.build()` returns `*normalizedResponse` (typed fast path)
  - Validation: max 5 component rows, max 5 components per row, max 10 embeds

- [ ] **2.4** Write `internal/jsdiscord/ui_message_test.go`
  - Test: full chain `.content("hi").ephemeral().embed(e).build()` produces `*normalizedResponse`
  - Test: `build()` output passes through `normalizePayload()` fast path
  - Test: multiple embeds accumulate
  - Test: multiple rows accumulate
  - Test: raw JS object in `.embed()` → type error with suggestion
  - Test: raw JS object in `.row()` → type error with suggestion
  - Test: wrong-parent errors for embed methods (`.field()` on message → "you probably meant ui.embed()")
  - Test: max rows validation

- [ ] **2.5** Wire `message` and `embed` exports in `ui_module.go`
  - Replace stubs with actual `newMessageBuilder(vm)` and `newEmbedBuilder(vm, title)` calls

- [ ] **2.6** Integration test: return a ui.message from a bot command
  - Write a small Go test that loads a JS script using `require("ui")`
  - Script returns `ui.message().content("hello").ephemeral().build()` from a command
  - Verify the dispatch pipeline produces the correct response

**Commit:** `feat(ui): implement message and embed builders with Proxy traps`

---

## Phase 3: Component builders — button and select

Goal: Buttons and selects work inside `ui.message().row(...)`. All select menu types supported.

- [ ] **3.1** Create `internal/jsdiscord/ui_components.go`
  - `ButtonBuilder` struct: customId, label, style, disabled, emoji, url
  - `newButtonBuilder(vm, customId, label, style)` — returns Goja Proxy
  - Style validation at construction: primary/secondary/success/danger/link
  - Methods: `disabled()`, `emoji()`, `url()`, `build()`
  - `build()` returns `discordgo.Button`

- [ ] **3.2** Write button tests in `internal/jsdiscord/ui_components_test.go`
  - Test: construction validates style
  - Test: `disabled(true)` sets Disabled
  - Test: `build()` returns `discordgo.Button`
  - Test: wrong-parent error (`.placeholder()` on button → "you probably meant ui.select()")
  - Test: empty customId → error

- [ ] **3.3** Add `SelectBuilder` to `internal/jsdiscord/ui_components.go`
  - `SelectBuilder` struct: customId, placeholder, options, minValues, maxValues, disabled, menuType
  - `newSelectBuilder(vm, customId)` — returns Goja Proxy, menuType = StringSelectMenu
  - Methods: `placeholder()`, `option()`, `minValues()`, `maxValues()`, `disabled()`, `build()`
  - `build()` returns `discordgo.SelectMenu`
  - Validation: max 25 options, label/value required, placeholder max 150

- [ ] **3.4** Write select tests in `internal/jsdiscord/ui_components_test.go`
  - Test: options accumulate
  - Test: max 25 options validation
  - Test: `build()` returns correct `discordgo.SelectMenu`
  - Test: wrong-parent error (`.ephemeral()` on select → "you probably meant ui.message()")

- [ ] **3.5** Create `internal/jsdiscord/ui_selects.go`
  - `newUserSelectBuilder(vm, customId)` — menuType = UserSelectMenu
  - `newRoleSelectBuilder(vm, customId)` — menuType = RoleSelectMenu
  - `newChannelSelectBuilder(vm, customId)` — menuType = ChannelSelectMenu
  - `newMentionableSelectBuilder(vm, customId)` — menuType = MentionableSelectMenu
  - Each returns a Proxy with only `placeholder()`, `minValues()`, `maxValues()`, `disabled()`, `build()`

- [ ] **3.6** Write typed select tests
  - Test each variant produces the correct menu type
  - Test that `.option()` is not available on typed selects (wrong-parent or not applicable)

- [ ] **3.7** Wire `button`, `select`, `userSelect`, `roleSelect`, `channelSelect`, `mentionableSelect` exports in `ui_module.go`

- [ ] **3.8** Integration test: message with buttons and selects
  - Test that `ui.message().row(ui.button(...), ui.select(...)).build()` produces a response with correct components

**Commit:** `feat(ui): implement button, select, and typed select builders`

---

## Phase 4: Form builder and helper functions

Goal: Modal forms work. Helper functions (row, pager, actions, confirm, ok, error, card, emptyResults) all work.

- [ ] **4.1** Add `FormBuilder` to `internal/jsdiscord/ui_components.go`
  - `FormBuilder` struct: customId, title, fields (current field tracking)
  - `newFormBuilder(vm, customId, title)` — returns Goja Proxy
  - Methods: `text()`, `textarea()`, `required()`, `value()`, `placeholder()`, `min()`, `max()`, `build()`
  - `build()` returns `map[string]any` compatible with existing `normalizeModalPayload()`
  - Validation: customId/title required, max 5 fields (Discord modal limit)

- [ ] **4.2** Write form tests in `internal/jsdiscord/ui_components_test.go`
  - Test: chain accumulates fields
  - Test: `build()` produces correct modal payload shape
  - Test: max 5 fields validation
  - Test: `required()`, `min()`, `max()` set field properties
  - Test: wrong-parent error (`.ephemeral()` on form → "you probably meant ui.message()")

- [ ] **4.3** Create `internal/jsdiscord/ui_helpers.go`
  - `rowFunc(vm, ...components)` — accepts only builder Proxies, auto-builds each, returns `discordgo.ActionsRow`
  - `pagerFunc(vm, prevId, nextId, controls)` — returns `discordgo.ActionsRow` with Previous/Next buttons
  - `actionsFunc(vm, definitions)` — accepts array of `{id, label, style}`, returns `discordgo.ActionsRow`
  - `confirmFunc(vm, confirmId, cancelId, options)` — returns `*normalizedResponse` (ephemeral message with confirm/cancel buttons)
  - `okFunc(vm, content)` — returns `*normalizedResponse` (ephemeral, simple content)
  - `errorFunc(vm, content)` — returns `*normalizedResponse` (ephemeral, "⚠️ " prefix)
  - `emptyResultsFunc(vm, query)` — returns `*normalizedResponse` (ephemeral, "No results found")
  - `cardFunc(vm, title)` — returns an embed builder Proxy with extra `.meta(name, value, inline)` method

- [ ] **4.4** Write `internal/jsdiscord/ui_helpers_test.go`
  - Test each helper produces the correct Go type
  - Test `row()` rejects raw JS objects
  - Test `confirm()` produces ephemeral message with two buttons
  - Test `card()` returns an embed builder that also has `.meta()`

- [ ] **4.5** Wire `form`, `row`, `pager`, `actions`, `confirm`, `ok`, `error`, `card`, `emptyResults` exports in `ui_module.go`

**Commit:** `feat(ui): implement form builder and helper functions`

---

## Phase 5: Flow helper

Goal: `ui.flow()` works for managing per-user per-channel state across interactions.

- [ ] **5.1** Create `internal/jsdiscord/ui_flow.go`
  - `FlowHelper` struct: namespace, init state
  - `newFlowHelper(vm, namespace, init)` — returns Goja Proxy
  - `stateKey(ctx)` — derives key from `guild.id.channel.id.user.id`
  - Methods: `load(ctx)`, `save(ctx, state)`, `clear(ctx)`, `id(suffix)`, `componentIds(names)`, `pagerIds()`
  - `load()` reads from `ctx.store.get(key)` and merges with init defaults
  - `save()` writes to `ctx.store.set(key, state)`
  - `id(suffix)` returns `"namespace:suffix"`

- [ ] **5.2** Write `internal/jsdiscord/ui_flow_test.go`
  - Test: `load()` returns init state when no state stored
  - Test: `save()` persists, `load()` retrieves
  - Test: `clear()` removes state
  - Test: `id("select")` returns `"myflow:select"`
  - Test: `componentIds(["a", "b"])` returns correct map
  - Test: `pagerIds()` returns previous/next IDs

- [ ] **5.3** Wire `flow` export in `ui_module.go`

**Commit:** `feat(ui): implement flow helper for stateful screens`

---

## Phase 6: Migrate ui-showcase bot

Goal: The ui-showcase bot uses `require("ui")` from Go instead of JS-side builders. All existing integration tests pass unchanged.

- [ ] **6.1** Create `examples/discord-bots/ui-showcase/lib/ui/primitives.js` replacement
  - Replace with `module.exports = require("ui")`
  - This makes all existing `require("./lib/ui")` calls in the bot resolve to the Go module

- [ ] **6.2** Update `examples/discord-bots/ui-showcase/lib/ui/screen.js`
  - Verify `ui.flow` from Go works the same as the JS `flow()` function
  - If API differs, adjust `screen.js` to match the Go export
  - Keep `screen.js` as JS (domain logic, uses `ctx.store`)

- [ ] **6.3** Remove `examples/discord-bots/ui-showcase/lib/ui/index.js`
  - No longer needed — `screen.js` can be required directly or the bot can `require("ui")` + `require("./lib/ui/screen")`

- [ ] **6.4** Run all 9 integration tests in `ui_showcase_runtime_test.go`
  - `go test ./internal/jsdiscord/ -run TestUIShowcase -v -count=1`
  - All tests must pass without modifications

- [ ] **6.5** Fix any API mismatches
  - If the Go builder API differs from the JS one (e.g., method names, argument order), update `index.js` bot code to match
  - Prefer changing the bot code, not the Go API, unless the Go API is clearly wrong

**Commit:** `refactor(ui-showcase): migrate to Go-side ui module`

---

## Phase 7: Validate and harden

Goal: Full test suite green. Error messages verified. Knowledge-base and other bots unaffected.

- [ ] **7.1** Run full test suite
  - `go test ./internal/jsdiscord/ -v -count=1`
  - All tests green — including knowledge-base, ping, poker, etc.

- [ ] **7.2** Verify non-ui bots unaffected
  - Knowledge-base bot: `go test ./internal/jsdiscord/ -run TestKnowledgeBase -v`
  - The bot doesn't use `require("ui")`, so it should still return raw JS objects through the existing `normalizePayload()` path

- [ ] **7.3** Write comprehensive error message tests
  - Test every wrong-parent scenario (button.ephemeral, select.footer, form.content, etc.)
  - Test every type-mismatch scenario (raw object to embed, raw object to row)
  - Test unknown method on each builder
  - Test argument validation on each builder constructor

- [ ] **7.4** Test edge cases
  - Empty builder: `ui.message().build()` with no methods called
  - Maximum limits: 10 embeds, 5 rows, 25 fields, 25 options
  - Unicode in labels, descriptions, content
  - Very long strings (near Discord limits)

- [ ] **7.5** Run `go vet ./internal/jsdiscord/`
- [ ] **7.6** Run `golangci-lint run ./internal/jsdiscord/`

**Commit:** `test(ui): comprehensive error and edge case tests`

---

## Phase 8: Cleanup and documentation

Goal: Remove JS-side builder code. Update all docs. Final state.

- [ ] **8.1** Delete `examples/discord-bots/ui-showcase/lib/ui/primitives.js`
- [ ] **8.2** Delete `examples/discord-bots/ui-showcase/lib/ui/index.js`
- [ ] **8.3** Update `examples/discord-bots/README.md` — note that ui-showcase now uses Go-side `require("ui")`
- [ ] **8.4** Update ticket diary with implementation notes and commit log
- [ ] **8.5** Update ticket changelog
- [ ] **8.6** Final commit
- [ ] **8.7** Upload final ticket docs to reMarkable

**Commit:** `chore(ui): remove JS-side builders, update docs`
