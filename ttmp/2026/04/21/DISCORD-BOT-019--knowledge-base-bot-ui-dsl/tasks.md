# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Inspect the current knowledge-base bot UI surface
- [x] Write a detailed design brainstorm for a possible UI DSL
- [x] Write concrete example DSL sketches for multiple use cases
- [x] Add and maintain a diary while working
- [x] Decide on implementation approach — Goja Proxy-based Go-side builders

## Analysis goals

- [x] Map the current search/review/form UI architecture
- [x] Identify concrete duplication and awkwardness in the current interaction layer
- [x] Propose multiple DSL design directions with trade-offs
- [x] Recommend one preferred implementation path
- [x] Prove Goja Proxy works for builders (scripts/01-proxy-poc.go)

## JS prototype (superseded by Go approach)

- [x] Create `examples/discord-bots/ui-showcase/` example bot skeleton
- [x] Create JS-side `lib/ui/` with generic UI builder primitives
- [x] Wire up the showcase bot using the JS UI DSL
- [x] Write Go integration tests (ui_showcase_runtime_test.go)
- [x] Fix chain-object-leaking bug (auto-build in row())
- [x] Update README.md

## Go-side UI DSL module (current plan)

### Design and guide

- [x] Write Goja Proxy-based implementation guide (design/02-goja-proxy-ui-dsl-implementation-guide.md)

### Core module registration

- [ ] Create `internal/jsdiscord/ui_module.go` — UIRegistrar, UILoader, module exports table
- [ ] Wire UIRegistrar into the host alongside the existing discord Registrar
- [ ] Write unit test: `require("ui")` loads and exports all expected functions

### Builder: message

- [ ] Create `internal/jsdiscord/ui_message.go` — MessageBuilder struct + Proxy
- [ ] Methods: content(), ephemeral(), tts(), embed(), row(), file(), build()
- [ ] build() returns `map[string]any` compatible with existing `normalizePayload()`
- [ ] Write unit tests for MessageBuilder

### Builder: embed

- [ ] Create `internal/jsdiscord/ui_embed.go` — EmbedBuilder struct + Proxy
- [ ] Methods: title(), description(), color(), field(), fields(), footer(), author(), timestamp(), build()
- [ ] Validate: max 25 fields, color range, description length
- [ ] Write unit tests for EmbedBuilder

### Builder: card

- [ ] Add card() to `ui_helpers.go` — convenience embed builder with .meta() shortcut
- [ ] Write unit tests

### Builder: button

- [ ] Create `internal/jsdiscord/ui_components.go` — ButtonBuilder struct + Proxy
- [ ] Methods: disabled(), emoji(), url(), build()
- [ ] Validate: style must be one of primary/secondary/success/danger/link
- [ ] Write unit tests for ButtonBuilder

### Builder: select

- [ ] Add SelectBuilder to `ui_components.go` — SelectBuilder struct + Proxy
- [ ] Methods: placeholder(), option(), minValues(), maxValues(), disabled(), build()
- [ ] Validate: max 25 options, placeholder length
- [ ] Add typed select variants: userSelect, roleSelect, channelSelect, mentionableSelect
- [ ] Write unit tests for all select builders

### Builder: form

- [ ] Add FormBuilder to `ui_components.go` — FormBuilder struct + Proxy
- [ ] Methods: text(), textarea(), required(), value(), placeholder(), min(), max(), build()
- [ ] Validate: customId/title required, max 5 action rows
- [ ] Write unit tests for FormBuilder

### Resolve helpers

- [ ] Create `internal/jsdiscord/ui_errors.go` — methodOwner map, wrongParentError(), typeMismatchError(), unknownMethodError()
- [ ] Every builder's Get trap uses three branches: own methods, wrong-parent methods, truly unknown
- [ ] Write unit tests: wrong-parent error messages, type-mismatch errors, unknown-method errors

### Helper functions

- [ ] Create `internal/jsdiscord/ui_helpers.go` — row(), pager(), actions(), confirm(), ok(), error(), emptyResults()
- [ ] Each returns the correct Go type directly
- [ ] Write unit tests for helpers

### Flow helper

- [ ] Create `internal/jsdiscord/ui_flow.go` — FlowHelper struct + Proxy
- [ ] Methods: load(ctx), save(ctx, state), clear(ctx), id(suffix), componentIds(names), pagerIds()
- [ ] Uses ctx.store under the hood with namespace-based keys
- [ ] Write unit tests for FlowHelper

### Integration: update ui-showcase bot

- [ ] Replace `lib/ui/primitives.js` with `module.exports = require("ui")`
- [ ] Keep `lib/ui/screen.js` (flow state is domain logic)
- [ ] Verify all existing integration tests pass unchanged
- [ ] Remove `lib/ui/index.js` (no longer needed)

### Integration: validate with existing tests

- [ ] All 9 tests in `ui_showcase_runtime_test.go` pass against Go builders
- [ ] Knowledge-base bot still works (raw JS payloads still work through existing normalize path)
- [ ] Ping bot still works
- [ ] Run full test suite: `go test ./internal/jsdiscord/ -v`

### Cleanup

- [ ] Remove `lib/ui/primitives.js` and `lib/ui/index.js`
- [ ] Update ticket diary with implementation notes
- [ ] Final commit and changelog update
