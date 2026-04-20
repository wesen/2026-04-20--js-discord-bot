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
- [x] Add `reactionAdd` session handling in `internal/bot/bot.go`
- [x] Add `reactionRemove` session handling in `internal/bot/bot.go`
- [x] Add `Host.DispatchReactionAdd(...)`
- [x] Add `Host.DispatchReactionRemove(...)`
- [x] Normalize reaction payloads (emoji, user, message/channel/guild IDs, member where available)
- [x] Extend JS dispatch context with `ctx.reaction`
- [x] Add runtime tests for reaction events
- [x] Update one example bot to demonstrate reaction event usage
- [x] Validate with focused and full Go test runs

### Phase 1C — guild member events
- [x] Add `guildMemberAdd` session handling in `internal/bot/bot.go`
- [x] Add `guildMemberUpdate` session handling in `internal/bot/bot.go`
- [x] Add `guildMemberRemove` session handling in `internal/bot/bot.go`
- [x] Add `Host.DispatchGuildMemberAdd(...)`
- [x] Add `Host.DispatchGuildMemberUpdate(...)`
- [x] Add `Host.DispatchGuildMemberRemove(...)`
- [x] Normalize member payloads (user, nick, roles, joinedAt, pending, guild ID)
- [x] Extend JS dispatch context with `ctx.member`
- [x] Add runtime tests for guild member events
- [x] Update one example bot to demonstrate guild member event usage
- [x] Validate with focused and full Go test runs

### Phase 2 — moderation/admin host APIs
- [x] Add `ctx.discord.members.addRole(...)`
- [x] Add `ctx.discord.members.removeRole(...)`
- [x] Add `ctx.discord.members.timeout(...)`
- [ ] Decide whether `kick`, `ban`, and `unban` belong in the first moderation slice
- [x] Add explicit structured logging around destructive/admin Discord operations
- [x] Add runtime tests for moderation/admin host methods
- [x] Document permissions, error surfaces, and safety expectations

### Phase 3 — examples and operator guidance
- [x] Add one moderation-oriented example bot that uses the new event and admin surfaces
- [x] Document privileged intent requirements (`GuildMembers`, reaction intents where needed)
- [x] Document permission expectations and common failure modes
- [x] Add a small operator/debug playbook for tracing event and moderation flows
