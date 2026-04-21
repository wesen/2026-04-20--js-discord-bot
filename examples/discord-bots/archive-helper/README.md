# Archive Helper Bot

Download Discord channels and threads as Markdown archives.

## What it does

- **`/archive-channel`** — Archives messages from the current channel as a `.md` file attachment.
- **Apps → Archive Thread** — Right-click any message inside a thread to archive the entire thread.

## Commands

### `/archive-channel`

```
/archive-channel [limit: 500] [before_message_id: "..."]
```

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `limit` | integer | No | 500 | Max messages to archive |
| `before_message_id` | string | No | — | Stop archiving when this message ID is reached |

The bot fetches messages with pagination (100 per request), renders them as Markdown with YAML frontmatter, and delivers the file as a Discord attachment.

### Archive Thread (message context menu)

Right-click any message inside a thread → Apps → **Archive Thread**.

The bot infers the thread ID from the message, fetches all thread messages, and delivers the archive as a file attachment in the thread.

## Running the bot

```bash
go run ./cmd/discord-bot bots \
  --bot-repository ./examples/discord-bots \
  run archive-helper \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --sync-on-start
```

### Runtime config

| Flag | Default | Description |
|------|---------|-------------|
| `--default-limit` | 500 | Default max messages per archive request |

## Archive file format

Each archive is a Markdown file with YAML frontmatter:

```yaml
---
source: "discord"
server: "My Server"
server_id: "1234567890123456789"
channel: "general"
channel_id: "9876543210987654321"
archived_at: "2026-04-21T14:32:00.000Z"
message_count: 347
---
```

Messages are rendered chronologically (oldest first):

```markdown
**alice** *(2026-04-21 14:30 UTC)*:
Hey everyone, check out this design doc!

**bob** *(2026-04-21 14:31 UTC)*:
Looks great, thanks for sharing!
```

## File structure

```
archive-helper/
  index.js          # Bot definition, commands, events
  lib/
    fetcher.js      # Message pagination logic
    renderer.js     # Markdown rendering
  README.md         # This file
```

## Permissions needed

- `Read Messages/View Channels`
- `Read Message History`
- `Send Messages` (to deliver archive files)
- `Attach Files` (to deliver archive files)
- `Use Slash Commands`
- `Use Application Commands`
