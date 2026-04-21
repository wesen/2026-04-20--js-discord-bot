# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Add and maintain a diary as implementation progresses

## Planned implementation phases

### Phase 1 — guild and role lookup core
- [x] Add `ctx.discord.guilds.fetch(guildID)`
- [x] Add `ctx.discord.roles.list(guildID)`
- [x] Add `ctx.discord.roles.fetch(guildID, roleID)`
- [x] Add normalized guild and role snapshot helpers
- [x] Add runtime tests for guild/role lookup helpers
- [x] Update the moderation example bot with guild/role inspection commands
- [x] Validate with focused and full Go test runs

### Phase 2 — operator docs and caveats
- [x] Update reference docs with the implemented API surface and caveats
- [x] Update example README with permission/failure-mode notes
- [x] Refresh the diary and changelog after implementation
