# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Write detailed architecture / implementation guide
- [x] Write API reference and planning notes
- [x] Relate core files and upload bundle to reMarkable

## Planned implementation phases

### Phase 1A — message lifecycle events
- [x] Add `messageUpdate` session handling in `internal/bot/bot.go`
- [x] Add `messageDelete` session handling in `internal/bot/bot.go`
- [x] Add `Host.DispatchMessageUpdate(...)`
- [x] Add `Host.DispatchMessageDelete(...)`
- [x] Normalize partial/update/delete-safe message payloads in `internal/jsdiscord/host.go`
- [x] Extend JS dispatch context so message lifecycle handlers can inspect the updated/deleted message payload
- [x] Add runtime tests for `messageUpdate` and `messageDelete`
- [x] Update one example bot to demonstrate message lifecycle events
- [x] Validate with focused and full Go test runs

### Phase 1B — reaction events
- [ ] Add `reactionAdd` session handling in `internal/bot/bot.go`
- [ ] Add `reactionRemove` session handling in `internal/bot/bot.go`
- [ ] Add `Host.DispatchReactionAdd(...)`
- [ ] Add `Host.DispatchReactionRemove(...)`
- [ ] Normalize reaction payloads (emoji, user, message/channel/guild IDs, member where available)
- [ ] Extend JS dispatch context with `ctx.reaction`
- [ ] Add runtime tests for reaction events
- [ ] Update one example bot to demonstrate reaction event usage
- [ ] Validate with focused and full Go test runs

### Phase 1C — guild member events
- [ ] Add `guildMemberAdd` session handling in `internal/bot/bot.go`
- [ ] Add `guildMemberUpdate` session handling in `internal/bot/bot.go`
- [ ] Add `guildMemberRemove` session handling in `internal/bot/bot.go`
- [ ] Add `Host.DispatchGuildMemberAdd(...)`
- [ ] Add `Host.DispatchGuildMemberUpdate(...)`
- [ ] Add `Host.DispatchGuildMemberRemove(...)`
- [ ] Normalize member payloads (user, nick, roles, joinedAt, pending, guild ID)
- [ ] Extend JS dispatch context with `ctx.member`
- [ ] Add runtime tests for guild member events
- [ ] Update one example bot to demonstrate guild member event usage
- [ ] Validate with focused and full Go test runs

### Phase 2 — moderation/admin host APIs
- [ ] Add `ctx.discord.members.addRole(...)`
- [ ] Add `ctx.discord.members.removeRole(...)`
- [ ] Add `ctx.discord.members.timeout(...)`
- [ ] Decide whether `kick`, `ban`, and `unban` belong in the first moderation slice
- [ ] Add explicit structured logging around destructive/admin Discord operations
- [ ] Add runtime tests for moderation/admin host methods
- [ ] Document permissions, error surfaces, and safety expectations

### Phase 3 — examples and operator guidance
- [ ] Add one moderation-oriented example bot that uses the new event and admin surfaces
- [ ] Document privileged intent requirements (`GuildMembers`, reaction intents where needed)
- [ ] Document permission expectations and common failure modes
- [ ] Add a small operator/debug playbook for tracing event and moderation flows
