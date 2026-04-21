# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Add and maintain a diary as implementation progresses

## Planned implementation phases

### Phase 1 — thread lookup and membership core
- [x] Add `ctx.discord.threads.fetch(threadID)`
- [x] Add `ctx.discord.threads.join(threadID)`
- [x] Add `ctx.discord.threads.leave(threadID)`
- [x] Add runtime tests and example commands
- [x] Validate with focused and full Go test runs

### Phase 2 — thread creation and lifecycle helpers
- [x] Add a small thread start helper
- [x] Decide whether archive/lock behavior belongs in this ticket or a follow-up
- [x] Update docs and diary/changelog
