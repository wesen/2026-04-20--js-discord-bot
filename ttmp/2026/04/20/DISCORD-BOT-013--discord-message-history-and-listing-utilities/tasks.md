# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Add and maintain a diary as implementation progresses

## Planned implementation phases

### Phase 1 — message history core
- [x] Add `ctx.discord.messages.list(channelID, payload?)`
- [x] Support a narrow payload shape with `before`, `after`, `around`, and `limit`
- [x] Add runtime tests for message listing helpers
- [x] Update the moderation example bot with message history commands
- [x] Validate with focused and full Go test runs

### Phase 2 — operator docs and caveats
- [x] Update reference docs with the implemented API surface and caveats
- [x] Update example README with permission/failure-mode notes
- [x] Refresh the diary and changelog after implementation
