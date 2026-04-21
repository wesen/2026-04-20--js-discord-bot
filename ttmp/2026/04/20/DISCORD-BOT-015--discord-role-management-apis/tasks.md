# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Add and maintain a diary as implementation progresses

## Planned implementation phases

### Phase 1 — role create/update core
- [ ] Add `ctx.discord.roles.create(guildID, payload)`
- [ ] Add `ctx.discord.roles.update(guildID, roleID, payload)`
- [ ] Add runtime tests and example commands
- [ ] Validate with focused and full Go test runs

### Phase 2 — destructive and ordering helpers
- [ ] Decide whether delete belongs here or in a follow-up
- [ ] Consider role reordering separately from field updates
- [ ] Update docs and diary/changelog
