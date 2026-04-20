# Example Discord bot implementations

This repository exercises the named bot runner model.

## Bots

- `ping/` — Discord JS API showcase with buttons, modals, autocomplete, and outbound operations
- `knowledge-base/` — relative `require()` helper, search/article commands, message event
- `support/` — deferred/edit/follow-up interaction flow, embeds, buttons, guild event
- `moderation/` — embeds, components, ephemeral responses, message lifecycle, reaction, guild-member events, and member moderation host APIs
- `poker/` — video poker hand management, Hold'em action advice, buttons, and modals
- `announcements.js` — root-level bot script to exercise direct file discovery

## Example commands

```bash
GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots
GOWORK=off go run ./cmd/discord-bot bots help ping --bot-repository ./examples/discord-bots
GOWORK=off go run ./cmd/discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots
GOWORK=off go run ./cmd/discord-bot bots help poker --bot-repository ./examples/discord-bots
GOWORK=off go run ./cmd/discord-bot bots run ping --bot-repository ./examples/discord-bots --bot-token "$DISCORD_BOT_TOKEN" --application-id "$DISCORD_APPLICATION_ID" --guild-id "$DISCORD_GUILD_ID" --sync-on-start
GOWORK=off go run ./cmd/discord-bot bots run poker --bot-repository ./examples/discord-bots --bot-token "$DISCORD_BOT_TOKEN" --application-id "$DISCORD_APPLICATION_ID" --guild-id "$DISCORD_GUILD_ID" --sync-on-start
GOWORK=off go run ./cmd/discord-bot bots run knowledge-base --bot-repository ./examples/discord-bots --bot-token "$DISCORD_BOT_TOKEN" --application-id "$DISCORD_APPLICATION_ID" --guild-id "$DISCORD_GUILD_ID" --index-path ./docs/local-index --sync-on-start
```

## Runtime notes

- Use `/ping` for the JS showcase bot with buttons, modals, autocomplete, outbound operations, and a deferred `/search` demo.
- `/search` shows a private "Searching..." state, waits about 2 seconds, then edits in the results.
- `knowledge-base` now demonstrates bot startup config via `configure({ run: ... })`; for example `index_path` becomes the CLI flag `--index-path` and is exposed in JavaScript as `ctx.config.index_path`.
- Use `/poker-help` in Discord to see the command list and examples.
- `/poker-help` includes quick-action buttons and modal entry points for rank/action examples.
- `!kb`, `!support`, `!modping`, `!poker`, and `!pingjs` message triggers exercise each bot's own `messageCreate` handling.
- `moderation` now also logs message edit/delete lifecycle events, reaction add/remove events, and guild member join/update/remove events to demonstrate the early DISCORD-BOT-009 event-expansion slices.
- `moderation` also now includes host-backed `mod-add-role`, `mod-timeout`, `mod-kick`, `mod-ban`, and `mod-unban` commands that demonstrate `ctx.discord.members.*` operations using explicit Discord IDs.

## Moderation / event prerequisites

- Event-heavy moderation flows depend on gateway intents including:
  - `GuildMessages`
  - `GuildMessageReactions`
  - `GuildMembers`
  - `MessageContent`
- Moderation commands must be run in a guild context.
- `mod-add-role`, `mod-timeout`, `mod-kick`, and `mod-ban` require the bot to have the corresponding Discord permissions and sufficient role hierarchy over the target member/role.
- `mod-unban` requires unban permissions in the guild.
- The current `timeout(...)` slice supports `durationSeconds`, `until`, and `clear: true`; it does not yet send an audit-log reason.
- The current `ban(...)` slice supports `reason` and `deleteMessageDays`.
