# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Relate core files and upload bundle to reMarkable

## Planned implementation phases

### Phase 1 — inbound event expansion
- [ ] Add `messageUpdate` and `messageDelete`
- [ ] Add `reactionAdd` and `reactionRemove`
- [ ] Add `guildMemberAdd`, `guildMemberUpdate`, and `guildMemberRemove`
- [ ] Add thread/channel lifecycle events where practical

### Phase 2 — moderation/admin host APIs
- [ ] Add role assignment helpers
- [ ] Add timeout helpers
- [ ] Add kick/ban helpers if desired
- [ ] Add careful audit/logging guidance for destructive operations

### Phase 3 — examples and operator guidance
- [ ] Add one moderation-oriented example bot
- [ ] Document permission expectations and failure modes
