# Show Space Discord Bot — Operator Runbook

## Purpose

This runbook explains how to start, configure, and maintain the show-space bot in the venue Discord server.

## 1) Start the bot

### Phase 1 run (seeded JSON only)

Use this when you want to exercise the announcement workflow without a database:

```bash
GOWORK=off go run ./cmd/discord-bot bots run show-space \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --upcoming-shows-channel-id "$UPCOMING_SHOWS_CHANNEL_ID" \
  --staff-channel-id "$STAFF_CHANNEL_ID" \
  --admin-role-id "$ADMIN_ROLE_ID" \
  --booker-role-id "$BOOKER_ROLE_ID" \
  --sync-on-start
```

Add `--debug` if you want the debug-only role lookup command to be available while you are wiring the server.

### Phase 2 run (SQLite-backed persistence)

Use this when you want show records to persist and be manageable by ID:

```bash
GOWORK=off go run ./cmd/discord-bot bots run show-space \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --upcoming-shows-channel-id "$UPCOMING_SHOWS_CHANNEL_ID" \
  --staff-channel-id "$STAFF_CHANNEL_ID" \
  --admin-role-id "$ADMIN_ROLE_ID" \
  --booker-role-id "$BOOKER_ROLE_ID" \
  --db-path ./examples/discord-bots/show-space/data/shows.sqlite \
  --sync-on-start
```

## 2) Required configuration

The bot expects these runtime fields:

- `upcomingShowsChannelId` — public show announcements channel
- `announcementsChannelId` — optional general announcement channel
- `staffChannelId` — private staff log/summary channel
- `adminRoleId` — role ID that grants full access
- `bookerRoleId` — role ID that grants show-management access
- `timeZone` — optional IANA timezone for date display
- `dbPath` — optional SQLite path for phase-2 persistence
- `seedFromJson` — optional bool that controls whether the DB seeds from `shows.json` when empty
- `debug` — optional bool that enables debug-only helper commands like `/debug-roles`

## 3) Seed or migrate initial shows

### Phase 1

The bot reads its initial upcoming-show list from `examples/discord-bots/show-space/shows.json`.

### Phase 2

When `dbPath` is configured and the database is empty, the bot seeds the SQLite store from `shows.json` automatically.

If you want to start over, delete the SQLite file and restart the bot.

## 4) Common operator actions

### Post a new show

Use `/announce` for a quick post-and-pin flow, or `/add-show` when you want the bot to store the show in the DB and return an ID.

### Look up a show

Use `/show <id>` to inspect the current record.

### Check role IDs

Enable `--debug` and run `/debug-roles` to list the current guild role IDs and names.

### Cancel a show

Use `/cancel-show <id>` to mark the show cancelled, unpin the original message, and post a cancellation notice.

### Archive a finished show

Use `/archive-show <id>` to mark the show archived and unpin the original announcement.

### Clean up old pins

Use `/unpin-old` for a manual pin sweep in `#upcoming-shows`.

Use `/archive-expired` when you want the bot to archive expired shows and post a quiet staff summary.

## 5) Permissions checklist

Make sure the bot can:

- view the target guild
- view/send messages in `#upcoming-shows`
- pin and unpin messages in `#upcoming-shows`
- send messages in `#staff` if you want archive summaries
- read member roles so `@admin` and `@booker` gating works

## 6) Troubleshooting

### `/announce` says it cannot find the message to pin

Check that the bot can send messages in the configured show channel and that no other bot is posting identical embed titles at the same time.

### `/upcoming` is empty

Confirm the bot is running with the right `shows.json` data or with the DB path pointing at the correct SQLite file.

### `@booker` cannot use write commands

Confirm the configured role IDs match the actual Discord role IDs and that the member has the expected role.

### Archived shows never appear in `/past-shows`

Confirm the show was actually archived, not just cancelled. `past-shows` shows archived or past entries.
