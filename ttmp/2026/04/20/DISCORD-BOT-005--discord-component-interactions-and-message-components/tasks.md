# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Relate core files and upload bundle to reMarkable

## Planned implementation phases

### Phase 1 — builder and descriptor model
- [x] Add `component(customId, handler)` to the JS builder API
- [x] Add component descriptors to `describe()` output
- [x] Extend `BotDescriptor` parsing for component metadata

### Phase 2 — host dispatch
- [x] Handle `InteractionMessageComponent` in `internal/jsdiscord/host.go`
- [x] Support button custom IDs and select-menu values in the dispatch request
- [x] Expose `ctx.component` and `ctx.values` to component handlers

### Phase 3 — outbound component payloads
- [x] Expand component normalization beyond buttons to include select menus
- [x] Support string select menus, user/role/channel/mentionable select menus, and menu options/defaults where practical
- [x] Add tests for outgoing payload normalization and inbound dispatch

### Phase 4 — examples and CLI visibility
- [x] Update example JS bots to demonstrate component handlers
- [ ] Surface components in `bots help <bot>` output if useful
- [ ] Document patterns for stable custom IDs and modular handler registration
