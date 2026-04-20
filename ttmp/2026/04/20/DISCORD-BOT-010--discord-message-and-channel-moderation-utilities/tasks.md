# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Add and maintain a diary as implementation progresses

## Planned implementation phases

### Phase 1 — message inspection and pinning
- [x] Add `ctx.discord.messages.fetch(channelID, messageID)`
- [x] Add `ctx.discord.messages.pin(channelID, messageID)`
- [x] Add `ctx.discord.messages.unpin(channelID, messageID)`
- [x] Add `ctx.discord.messages.listPinned(channelID)`
- [x] Add normalized message snapshot helpers for fetched/pinned messages
- [x] Add runtime tests for fetch/pin/unpin/listPinned
- [x] Update the moderation example bot with pin/fetch commands
- [x] Validate with focused and full Go test runs

### Phase 2 — message bulk deletion
- [x] Add `ctx.discord.messages.bulkDelete(channelID, messageIDs)`
- [x] Decide the accepted input forms for message ID lists (`[]string`, `[]any`, object payload, or all three)
- [x] Add structured logging around bulk delete requests
- [x] Add runtime tests for bulk deletion
- [x] Update the moderation example bot with a bulk-delete command
- [x] Validate with focused and full Go test runs

### Phase 3 — channel utilities
- [ ] Add `ctx.discord.channels.fetch(channelID)`
- [ ] Add `ctx.discord.channels.setTopic(channelID, topic)`
- [ ] Add `ctx.discord.channels.setSlowmode(channelID, seconds)`
- [ ] Add normalized channel snapshot helpers
- [ ] Add runtime tests for fetch/topic/slowmode helpers
- [ ] Update the moderation example bot with channel utility commands
- [ ] Validate with focused and full Go test runs

### Phase 4 — docs and operator guidance
- [ ] Update reference docs with the implemented API surface and caveats
- [ ] Update example README with permissions/failure-mode notes
- [ ] Add a small playbook for debugging message/channel moderation flows
