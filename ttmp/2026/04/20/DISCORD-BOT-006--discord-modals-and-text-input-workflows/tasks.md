# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Relate core files and upload bundle to reMarkable

## Planned implementation phases

### Phase 1 — JS runtime contract
- [x] Add `modal(customId, handler)` to the JS builder API
- [x] Add `ctx.showModal(payload)` to interaction contexts that can open modals
- [x] Add modal descriptors to `describe()` output

### Phase 2 — host modal responses
- [x] Normalize modal payloads into `discordgo.InteractionResponseModal`
- [x] Normalize text input components and action rows for modal bodies
- [x] Reject invalid modal payloads with clear errors

### Phase 3 — modal submit handling
- [x] Route `InteractionModalSubmit` into JavaScript
- [x] Expose `ctx.modal` and `ctx.values`
- [x] Add tests for command/component -> showModal -> modal submit flows

### Phase 4 — examples and docs
- [x] Update the example bot to demonstrate a modal workflow
- [ ] Document custom ID stability and text-input style conventions
