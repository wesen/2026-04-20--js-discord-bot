# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Relate core files and upload bundle to reMarkable

## Planned implementation phases

### Phase 1 — richer payload normalization
- [x] Add files/attachments to outbound payloads
- [x] Add reply-to/message-reference support
- [x] Add richer webhook edit/send normalization

### Phase 2 — host capability objects
- [x] Expose `ctx.discord.channels.send(...)`
- [x] Expose `ctx.discord.messages.edit(...)`, `delete(...)`, and `react(...)`
- [x] Decide whether these should live under `ctx.discord` or `discord.*`

### Phase 3 — channel and thread helpers
- [ ] Add thread creation helpers
- [ ] Add basic fetch helpers for channels/messages where practical
- [ ] Add examples and runtime tests
