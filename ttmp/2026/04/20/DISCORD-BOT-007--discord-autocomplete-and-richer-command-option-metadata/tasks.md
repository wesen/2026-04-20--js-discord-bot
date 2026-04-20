# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Relate core files and upload bundle to reMarkable

## Planned implementation phases

### Phase 1 — runtime autocomplete core
- [x] Add `autocomplete(commandName, optionName, handler)` to the JS builder API
- [x] Add autocomplete descriptors to `describe()` output
- [x] Route `InteractionApplicationCommandAutocomplete` into JavaScript

### Phase 2 — richer command option metadata
- [x] Support `autocomplete: true` in command option specs
- [x] Support static `choices` in option specs
- [x] Support `minLength`, `maxLength`, `minValue`, and `maxValue`
- [ ] Support channel type restrictions where practical

### Phase 3 — context and normalization
- [x] Expose `ctx.focused` and normalized option values to autocomplete handlers
- [x] Normalize autocomplete handler results into Discord choices
- [x] Add tests for focused-option dispatch and response payloads

### Phase 4 — examples and docs
- [x] Update example bots with one realistic autocomplete flow
- [x] Document when to prefer static `choices` versus dynamic `autocomplete`
