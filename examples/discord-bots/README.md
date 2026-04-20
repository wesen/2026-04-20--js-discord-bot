# Example Discord bot implementations

This repository exercises the named bot runner model.

## Bots

- `knowledge-base/` — relative `require()` helper, search/article commands, message event
- `support/` — deferred/edit/follow-up interaction flow, embeds, buttons, guild event
- `moderation/` — embeds, components, ephemeral responses, message event
- `announcements.js` — root-level bot script to exercise direct file discovery

## Example commands

```bash
GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots
GOWORK=off go run ./cmd/discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots
GOWORK=off go run ./cmd/discord-bot bots run knowledge-base --bot-repository ./examples/discord-bots --bot-token "$DISCORD_BOT_TOKEN" --application-id "$DISCORD_APPLICATION_ID" --guild-id "$DISCORD_GUILD_ID"
GOWORK=off go run ./cmd/discord-bot bots run knowledge-base support moderation --bot-repository ./examples/discord-bots --bot-token "$DISCORD_BOT_TOKEN" --application-id "$DISCORD_APPLICATION_ID" --guild-id "$DISCORD_GUILD_ID" --sync-on-start
```

## Runtime notes

- Slash-command names are intentionally unique across bots.
- `!kb`, `!support`, and `!modping` message triggers exercise `messageCreate` fan-out.
