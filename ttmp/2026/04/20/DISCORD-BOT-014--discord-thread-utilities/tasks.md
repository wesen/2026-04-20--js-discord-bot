# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Add and maintain a diary as implementation progresses

## Planned implementation phases

### Phase 1 — thread lookup and membership core
- [ ] Add `ctx.discord.threads.fetch(threadID)`
- [ ] Add `ctx.discord.threads.join(threadID)`
- [ ] Add `ctx.discord.threads.leave(threadID)`
- [ ] Add runtime tests and example commands
- [ ] Validate with focused and full Go test runs

### Phase 2 — thread creation and lifecycle helpers
- [ ] Add a small thread start helper
- [ ] Decide whether archive/lock behavior belongs in this ticket or a follow-up
- [ ] Update docs and diary/changelog
