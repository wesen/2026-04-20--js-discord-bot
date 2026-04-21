---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: pkg/doc/topics/discord-js-bot-api-reference.md
      Note: Updated with pagination example and runtime constraints
    - Path: pkg/doc/tutorials/building-and-running-discord-js-bots.md
      Note: Updated with runtime environment warning and framework overview
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---



# Discord Archive Helper Bot — Analysis, Design & Implementation Guide

> **Audience:** New intern with basic programming knowledge. No prior Discord API experience assumed.
> **Goal:** By the end of this document, you should understand what we are building, why, and exactly how to implement it.

---

## Executive Summary

We need a **helper bot** that connects to Discord, reads messages from channels and threads, and produces clean Markdown archives. The bot is implemented as a JavaScript module that runs inside a Go-hosted runtime. Think of it as a "print to PDF" button for Discord conversations — producing structured, searchable, offline-readable archives delivered as file attachments in Discord.

---

## 1. What Is Discord? A Primer for the Uninitiated

Discord is a chat platform organized around **servers** (also called "guilds"). Each server contains **channels**, which can be:

- **Text channels** — persistent streams of messages, like Slack channels.
- **Voice channels** — for audio/video calls (not relevant here).
- **Forum channels** — text channels where each "post" is a thread.
- **Threads** — temporary or permanent sub-conversations spun off from a parent channel.

### Key Concepts You Must Know

| Concept | Explanation | Why It Matters for Us |
|---------|-------------|----------------------|
| **Server / Guild** | A community space. Has a unique numeric ID. | We need the guild ID to know which server to read from. |
| **Channel** | A named chat room inside a guild. Has a unique numeric ID. | We fetch messages from channels. |
| **Thread** | A mini-channel attached to a parent message or forum post. Has its own ID. | Threads contain messages we must archive separately. |
| **Message** | A single chat post. Can contain text, images, embeds, reactions, replies. | This is the atomic unit we convert to Markdown. |
| **User / Member** | A Discord account. Messages are authored by users. | We need usernames for attribution in the archive. |
| **Bot** | A special kind of Discord user that connects via API rather than the website. | Our archive helper is a bot. |
| **Token** | A secret string that authenticates the bot. Like a password. | We need this to run the bot. **Never commit it to Git.** |
| **Intents** | Permissions the bot declares when connecting. Discord uses these to decide what events to send. | We need the `MessageContent` intent to read message text. |
| **Rate Limits** | Discord caps how many API calls we can make per second. | The host handles this automatically, but we still paginate carefully. |

---

## 2. How Our Discord Bot Framework Works

This is the most important section. **We do not use raw Node.js or the `discord.js` npm package.** Our bots run inside a custom framework with a specific shape.

### 2.1 The Big Picture: Go Host + JS Bot Script

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Go Host Process                                │
│  (handles Discord gateway, rate limits, command sync, event dispatch)       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌──────────────────┐      ┌──────────────────┐      ┌─────────────────┐   │
│   │  Go Discord      │─────▶│  Goja JS Runtime │─────▶│  Your Bot Script│   │
│   │  Gateway Client  │      │  (loads + runs)  │      │  (index.js)     │   │
│   └──────────────────┘      └──────────────────┘      └─────────────────┘   │
│                                                                             │
│        Discord events ──────▶ dispatched to JS handlers                     │
│        Slash commands ──────▶ routed to JS command handlers                 │
│        JS ctx.discord.* ────▶ outbound API calls via Go host                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

The Go host owns the process. It connects to Discord, syncs slash commands, receives gateway events, and exposes a JavaScript API. Your bot is just a JavaScript module that declares what it handles.

### 2.2 Bot Script Structure

Every bot lives at `examples/discord-bots/<bot-name>/index.js` and exports one bot definition:

```javascript
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "archive-helper",
    description: "Download channels and threads as Markdown archives",
    category: "utilities",
  })

  // Your handlers go here
})
```

The builder callback receives registration helpers. You destructure only the ones you need:

| Helper | Purpose |
|--------|---------|
| `configure(options)` | Set bot name, description, category, runtime config fields |
| `command(name, spec?, handler)` | Register a slash command |
| `event(name, handler)` | Register a gateway event handler |
| `component(customId, handler)` | Handle button/select-menu clicks |
| `modal(customId, handler)` | Handle modal submissions |
| `autocomplete(commandName, optionName, handler)` | Supply autocomplete suggestions |

### 2.3 How the Bot Is Run

You do **not** run the bot with `node`. You run it through the Go CLI:

```bash
go run ./cmd/discord-bot bots \
  --bot-repository ./examples/discord-bots \
  run archive-helper \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --sync-on-start
```

The `bots` subcommand:
1. Scans `--bot-repository` for bot scripts
2. Loads the one named `archive-helper`
3. Reads its metadata and command registrations
4. Syncs slash commands to Discord (if `--sync-on-start`)
5. Opens the Discord gateway connection
6. Dispatches events and commands to your JS handlers

### 2.4 Runtime Config: Passing Settings to Your Bot

If your bot needs settings at startup (like an output directory or file naming pattern), declare them in `configure({ run: { fields: ... }})`:

```javascript
configure({
  name: "archive-helper",
  description: "Download channels and threads as Markdown archives",
  run: {
    fields: {
      output_prefix: {
        type: "string",
        help: "Filename prefix for generated archives",
        default: "discord-archive",
      },
      include_threads: {
        type: "bool",
        help: "Whether to automatically include threads",
        default: true,
      },
    },
  },
})
```

Each field becomes:
- A key in `ctx.config` inside your handlers
- A CLI flag when running the bot (e.g., `--output-prefix`, `--include-threads`)

### 2.5 What Is NOT Available

Because the JS runs inside a Goja runtime (not Node.js):

- **No `fs` module** — you cannot write files to disk from JS.
- **No `process.env`** — secrets come through CLI flags, not environment variables.
- **No `npm install`** — the only modules available are what the host provides.
- **No `fetch()` or HTTP clients** — all Discord API access goes through `ctx.discord.*`.

**For file output**, you send the generated Markdown back to Discord as a file attachment using `ctx.discord.channels.send()` with a `files` payload.

---

## 3. Problem Statement

### The Pain Point

Manuel participates in many Discord servers with valuable discussions — technical debates, project planning, design decisions, troubleshooting sessions. Discord is a great live medium, but:

- **Search is weak** for long-running threads.
- **No bulk export** — Discord has no "download this thread as a file" feature.
- **Data loss risk** — servers can be deleted, threads auto-archive, accounts banned.
- **Offline reading** — sometimes you want to read on a plane or on a reMarkable device.

### What We Need

A bot that, given a channel or thread, produces a clean Markdown file containing:

- All messages in chronological order.
- Proper author attribution with timestamps.
- Formatted text (bold, italic, code blocks, links).
- Embeds rendered as Markdown callouts or tables.
- Attachment URLs preserved.
- Reactions summarized.
- Reply chains indicated.

The output is delivered as a Discord file attachment, which can then be saved locally.

---

## 4. Proposed Solution

### 4.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Discord Archive Helper Bot                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   User runs /archive-channel or /archive-thread in Discord                  │
│        │                                                                    │
│        ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Slash Command Handler  (JS)                                        │   │
│   │  • Parse args (channel/thread ID, message limit, options)           │   │
│   │  • Defer the interaction (acknowledge immediately)                  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│        │                                                                    │
│        ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Message Fetcher  (JS → ctx.discord.messages.list)                  │   │
│   │  • Paginate through messages (before cursor, limit 100)             │   │
│   │  • Loop until no more messages or limit reached                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│        │                                                                    │
│        ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Thread Discovery  (JS → ctx.discord.threads.fetch)                 │   │
│   │  • For channels with threads, enumerate and fetch each              │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│        │                                                                    │
│        ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Markdown Renderer  (JS, pure logic)                                │   │
│   │  • Convert Discord markup → Markdown                                │   │
│   │  • Format headers, embeds, attachments, reactions                   │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│        │                                                                    │
│        ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  File Delivery  (JS → ctx.discord.channels.send with files payload) │   │
│   │  • Send the generated Markdown as a file attachment                 │   │
│   │  • Edit the deferred interaction with a summary + download link     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Data Flow

```
User types in Discord:
  /archive-channel limit: 500

        │
        ▼
  ┌─────────────────────────────────────┐
  │ Go host receives interaction        │
  │ routes to JS command handler        │
  └─────────────────────────────────────┘
        │
        ▼
  ┌─────────────────────────────────────┐
  │ JS handler calls ctx.defer()        │
  │ "Archiving channel..."              │
  └─────────────────────────────────────┘
        │
        ▼
  ┌─────────────────────────────────────┐
  │ Resolve channel from ctx.channel.id │
  │ (the channel where cmd was run)     │
  └─────────────────────────────────────┘
        │
        ▼
  ┌─────────────────────────────────────┐
  │ Fetch messages via                  │
  │ ctx.discord.messages.list()         │
  │ with pagination loop                │
  └─────────────────────────────────────┘
        │
        ▼
  ┌─────────────────────────────────────┐
  │ For each batch:                     │
  │   • Append to message array         │
  │   • Use last message ID as "before" │
  │   • Repeat until empty or limit     │
  └─────────────────────────────────────┘
        │
        ▼
  ┌─────────────────────────────────────┐
  │ Reverse to chronological order      │
  │ (oldest first)                      │
  └─────────────────────────────────────┘
        │
        ▼
  ┌─────────────────────────────────────┐
  │ Render to Markdown string           │
  └─────────────────────────────────────┘
        │
        ▼
  ┌─────────────────────────────────────┐
  │ Send file via                       │
  │ ctx.discord.channels.send()         │
  │ with files: [{name, content}]       │
  └─────────────────────────────────────┘
        │
        ▼
  ┌─────────────────────────────────────┐
  │ Edit deferred reply with summary:   │
  │ "Archived 347 messages."            │
  └─────────────────────────────────────┘
```

---

## 5. Deep Dive: Framework APIs for Archiving

### 5.1 Message Fetching (Pagination)

The framework provides `ctx.discord.messages.list(channelId, payload?)`. It returns up to 100 messages per call. To get all messages, you paginate backwards using the `before` parameter:

```
ctx.discord.messages.list(channelId, { limit: 100 })
  → returns messages [950..1049] (newest first)

ctx.discord.messages.list(channelId, { limit: 100, before: "950" })
  → returns messages [850..949]

...repeat until response is empty or limit reached...
```

**Pseudocode — Message Fetching Loop:**

```javascript
async function fetchAllMessages(ctx, channelId, maxMessages) {
  const allMessages = []
  let lastMessageId = null
  const limit = 100 // Discord max per request

  while (true) {
    const options = { limit: limit }
    if (lastMessageId) {
      options.before = lastMessageId
    }

    const batch = await ctx.discord.messages.list(channelId, options)

    if (!batch || batch.length === 0) {
      break
    }

    allMessages.push(...batch)
    lastMessageId = batch[batch.length - 1].id

    // Respect user-defined max
    if (maxMessages && allMessages.length >= maxMessages) {
      allMessages.splice(maxMessages)
      break
    }
  }

  // Messages arrive newest-first; reverse for chronological order
  return allMessages.reverse()
}
```

**Important rules from the framework:**
- You can use at most one anchor: `before`, `after`, or `around`.
- `limit` defaults to 25 and is capped at 100.
- The host handles rate limiting automatically.
- Returns plain JavaScript objects (not class instances).

### 5.2 Thread Discovery

For archiving threads, you need to discover them first. The current framework API for threads:

```javascript
// Fetch a thread snapshot by ID
const thread = await ctx.discord.threads.fetch(threadId)
// thread shape: { id, guildID, parentID, name, type, archived, locked, ... }

// Start a thread (not needed for archive, but good to know)
const started = await ctx.discord.threads.start(channelId, {
  name: "...",
  type: "public",
})
```

**Note:** The current framework does not expose a "list all threads in channel" API. For our archive bot, we have two approaches:

1. **Thread-as-target:** The user provides a thread ID directly (`/archive-thread thread_id: ...`).
2. **Event-based capture:** Use `event("messageCreate", ...)` to capture messages passively, building an archive over time in SQLite via `require("database")`.

For the initial implementation, we focus on **explicit slash commands** targeting a specific channel or thread.

### 5.3 Channel Info

```javascript
const channel = await ctx.discord.channels.fetch(channelId)
// channel shape: { id, guildID, name, type, topic, position, rateLimitPerUser }
```

---

## 6. Message Object Structure

Here is what a message looks like when returned from `ctx.discord.messages.list()` or `ctx.discord.messages.fetch()`:

```js
{
  id: "1234567890123456789",
  content: "Hey everyone, check out this design doc!",
  guildID: "9876543210987654321",
  channelID: "111222333444555666",
  author: {
    id: "444555666777888999",
    username: "alice",
    discriminator: "0",
    bot: false,
  },
  // Note: the current framework returns a simplified shape.
  // Full embeds, attachments, reactions may be available depending
  // on the host implementation version. Check ctx.discord.messages.fetch
  // for the richest single-message payload.
}
```

**Current normalized message shape (what you can rely on):**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Message snowflake ID |
| `content` | string | Message text content |
| `guildID` | string | Parent guild ID |
| `channelID` | string | Parent channel ID |
| `author.id` | string | Author user ID |
| `author.username` | string | Author display name |
| `author.discriminator` | string | Legacy discriminator (often "0") |
| `author.bot` | boolean | Whether the author is a bot |

**Note:** The framework's `messages.list` returns a simplified shape. For richer data (embeds, attachments, reactions), you may need to call `ctx.discord.messages.fetch(channelId, messageId)` for individual messages, or the host may enrich the list payload in future versions. Design your renderer defensively — check for field presence before using it.

---

## 7. Markdown Rendering in Detail

### 7.1 Discord Markup → Markdown

Discord uses a simplified markup language. We convert it to standard Markdown:

| Discord Syntax | Markdown Equivalent | Notes |
|----------------|---------------------|-------|
| `**bold**` | `**bold**` | Same. |
| `*italic*` or `_italic_` | `*italic*` | Same. |
| `__underline__` | `<u>underline</u>` | No native MD underline; use HTML or omit. |
| `~~strikethrough~~` | `~~strikethrough~~` | Same. |
| `` `code` `` | `` `code` `` | Same. |
| ```` ```lang\ncode\n``` ```` | Same | Same. |
| `> quote` | `> quote` | Same. |
| `>>> multi-line quote` | `> line 1\n> line 2` | Convert to multiple `>` lines. |
| `<#channelId>` | `#channel-name` | Resolve ID to name if possible. |
| `<@userId>` | `@username` | Resolve ID to username. |
| `<@&roleId>` | `@role-name` | Resolve ID to role name. |
| `<:emoji:123>` | `:emoji:` | Or use the CDN URL for custom emoji. |
| `https://...` | `<https://...>` or bare URL | Same. |

**Pseudocode — Content Converter:**

```javascript
function discordToMarkdown(content, ctx) {
  // Mentions: <@123> → @username
  content = content.replace(/<@(\d+)>/g, (match, userId) => {
    return "@" + userId // Ideally resolved; fallback to raw ID
  })

  // Channel mentions: <#123> → #channel-name
  content = content.replace(/<#(\d+)>/g, (match, channelId) => {
    return "#" + channelId // Ideally resolved
  })

  // Role mentions: <@&123> → @role-name
  content = content.replace(/<@&(\d+)>/g, (match, roleId) => {
    return "@" + roleId
  })

  // Custom emoji: <:name:id> → :name:
  content = content.replace(/<:(\w+):(\d+)>/g, (match, name, id) => {
    return `:${name}:`
  })

  // Animated emoji: <a:name:id>
  content = content.replace(/<a:(\w+):(\d+)>/g, (match, name, id) => {
    return `:${name}:`
  })

  // Multi-line quotes: >>> \n... → > line1\n> line2
  content = content.replace(/^>>>(\s*)\n?((?:.|\n)*)/gm, (match, space, quote) => {
    return quote.split("\n").map(line => "> " + line).join("\n")
  })

  return content
}
```

### 7.2 Message Header Format

Each message in the archive should have a consistent header:

```markdown
**alice** *(2024-03-15 14:32 UTC)*:
```

For replies, prepend a quote block (if reply data is available):

```markdown
> **bob** *(2024-03-15 14:28 UTC)*: We need a new design system...

**alice** *(2024-03-15 14:32 UTC)*: Hey everyone, check out this design doc!
```

### 7.3 File Output Format

The generated Markdown is a single string delivered as a file attachment. It should include YAML frontmatter for metadata:

```yaml
---
source: discord
server: "My Server Name"
server_id: "1234567890123456789"
channel: "design-discussion"
channel_id: "9876543210987654321"
thread: "color-palette-review"
thread_id: "111222333444555666"
archived_at: "2026-04-21T10:00:00Z"
message_count: 147
bot_version: "1.0.0"
---
```

---

## 8. Implementation Plan

### Phase 0: Create the Bot File

Create `examples/discord-bots/archive-helper/index.js`:

```javascript
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "archive-helper",
    description: "Download channels and threads as Markdown archives",
    category: "utilities",
    run: {
      fields: {
        default_limit: {
          type: "integer",
          help: "Default maximum messages to archive per request",
          default: 500,
        },
      },
    },
  })

  // Handlers defined in phases below
})
```

### Phase 1: The `/archive-channel` Command

```javascript
command("archive-channel", {
  description: "Archive messages from the current channel as Markdown",
  options: {
    limit: {
      type: "integer",
      description: "Maximum messages to archive (default: 500)",
      required: false,
    },
  },
}, async (ctx) => {
  const channelId = ctx.channel && ctx.channel.id
  if (!channelId) {
    return { content: "This command must be run in a channel.", ephemeral: true }
  }

  await ctx.defer({ ephemeral: true })

  const maxMessages = ctx.args.limit || ctx.config.default_limit || 500

  // Fetch channel info for metadata
  const channel = await ctx.discord.channels.fetch(channelId)

  // Fetch all messages with pagination
  const messages = await fetchAllMessages(ctx, channelId, maxMessages)

  // Render to Markdown
  const markdown = renderArchive(channel, null, messages)

  // Send as file attachment
  await ctx.discord.channels.send(channelId, {
    content: `📄 Archive of #${channel.name}: ${messages.length} messages`,
    files: [
      {
        name: `${sanitize(channel.name)}--archive.md`,
        content: markdown,
      },
    ],
  })

  // Update the deferred reply
  await ctx.edit({
    content: `Archived ${messages.length} messages from #${channel.name}.`,
    ephemeral: true,
  })
})
```

### Phase 2: The `/archive-thread` Command

```javascript
command("archive-thread", {
  description: "Archive messages from a thread as Markdown",
  options: {
    thread_id: {
      type: "string",
      description: "Thread ID to archive",
      required: true,
    },
    limit: {
      type: "integer",
      description: "Maximum messages to archive (default: 500)",
      required: false,
    },
  },
}, async (ctx) => {
  const threadId = ctx.args.thread_id
  const maxMessages = ctx.args.limit || ctx.config.default_limit || 500

  await ctx.defer({ ephemeral: true })

  // Fetch thread info
  const thread = await ctx.discord.threads.fetch(threadId)

  // Fetch parent channel for metadata
  const channel = await ctx.discord.channels.fetch(thread.parentID)

  // Fetch all messages from the thread
  const messages = await fetchAllMessages(ctx, threadId, maxMessages)

  // Render to Markdown
  const markdown = renderArchive(channel, thread, messages)

  // Send as file attachment to the current channel
  const currentChannelId = ctx.channel && ctx.channel.id
  await ctx.discord.channels.send(currentChannelId, {
    content: `📄 Archive of thread "${thread.name}": ${messages.length} messages`,
    files: [
      {
        name: `${sanitize(channel.name)}--${sanitize(thread.name)}--archive.md`,
        content: markdown,
      },
    ],
  })

  await ctx.edit({
    content: `Archived ${messages.length} messages from thread "${thread.name}".`,
    ephemeral: true,
  })
})
```

### Phase 3: Helper Functions

```javascript
// Fetch all messages with pagination
async function fetchAllMessages(ctx, channelId, maxMessages) {
  const allMessages = []
  let lastMessageId = null
  const pageSize = 100

  while (true) {
    const options = { limit: pageSize }
    if (lastMessageId) {
      options.before = lastMessageId
    }

    const batch = await ctx.discord.messages.list(channelId, options)
    if (!batch || batch.length === 0) {
      break
    }

    allMessages.push(...batch)
    lastMessageId = batch[batch.length - 1].id

    if (maxMessages && allMessages.length >= maxMessages) {
      allMessages.splice(maxMessages)
      break
    }
  }

  return allMessages.reverse()
}

// Render a full archive
function renderArchive(channel, thread, messages) {
  const lines = []

  // Frontmatter
  lines.push("---")
  lines.push(`source: "discord"`)
  lines.push(`channel: "${escapeYaml(channel.name || "unknown")}"`)
  lines.push(`channel_id: "${channel.id}"`)
  if (thread) {
    lines.push(`thread: "${escapeYaml(thread.name || "unknown")}"`)
    lines.push(`thread_id: "${thread.id}"`)
  }
  lines.push(`archived_at: "${new Date().toISOString()}"`)
  lines.push(`message_count: ${messages.length}`)
  lines.push("---")
  lines.push("")

  // Messages
  for (const message of messages) {
    lines.push(renderMessage(message))
    lines.push("")
  }

  return lines.join("\n")
}

// Render a single message
function renderMessage(message) {
  const author = message.author && message.author.username || "unknown"
  const timestamp = message.timestamp || new Date().toISOString()
  const formattedTime = String(timestamp).replace("T", " ").slice(0, 19) + " UTC"

  const header = `**${author}** *(${formattedTime})*:`
  const body = discordToMarkdown(String(message.content || ""))

  return [header, body].join("\n")
}

// Convert Discord markup to Markdown
function discordToMarkdown(content) {
  // Mentions
  content = content.replace(/<@(\d+)>/g, "@user-$1")
  content = content.replace(/<#(\d+)>/g, "#channel-$1")
  content = content.replace(/<@&(\d+)>/g, "@role-$1")
  content = content.replace(/<:(\w+):(\d+)>/g, ":$1:")
  content = content.replace(/<a:(\w+):(\d+)>/g, ":$1:")

  // Multi-line quotes
  content = content.replace(/^>>>(\s*)\n?((?:.|\n)*)/gm, (m, s, quote) => {
    return quote.split("\n").map(l => "> " + l).join("\n")
  })

  return content
}

// Sanitize a name for use in filenames
function sanitize(name) {
  return String(name || "unknown")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 60)
}

function escapeYaml(value) {
  return String(value || "").replace(/"/g, '\\"')
}
```

### Phase 4: Running the Bot

```bash
# List available bots
go run ./cmd/discord-bot bots \
  --bot-repository ./examples/discord-bots list

# Get help for our bot
go run ./cmd/discord-bot bots \
  --bot-repository ./examples/discord-bots help archive-helper

# Run it (with sync)
go run ./cmd/discord-bot bots \
  --bot-repository ./examples/discord-bots \
  run archive-helper \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --sync-on-start

# Run with custom default limit
go run ./cmd/discord-bot bots \
  --bot-repository ./examples/discord-bots \
  run archive-helper \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --default-limit 1000 \
  --sync-on-start
```

### Phase 5: Testing

1. **Start the bot** with `--sync-on-start` in a test server.
2. **Run `/archive-channel`** in a text channel with a few messages.
3. **Verify** the bot responds with an ephemeral "Archiving..." message, then delivers a `.md` file.
4. **Open the file** and check: chronological order, correct usernames, timestamps, content.
5. **Test `/archive-thread`** with a thread ID.
6. **Test the `limit` option** — try `limit: 5` and verify only 5 messages are included.
7. **Test edge cases:** empty channel, channel with only bot messages, very long messages.

---

## 9. File Delivery Strategy

Since the JS runtime has **no `fs` module**, we deliver archives through Discord itself:

### Option A: Channel Message with File Attachment (Recommended)

```javascript
await ctx.discord.channels.send(channelId, {
  content: "Here is your archive:",
  files: [
    {
      name: "general-chat--archive.md",
      content: markdownString,
    },
  ],
})
```

**Pros:** Simple, works today, user can download the file from Discord.
**Cons:** File is public to the channel, limited by Discord's file size limits (25MB for bots).

### Option B: Ephemeral Reply with File (Not Directly Supported)

Discord ephemeral messages cannot have file attachments in the same way. Use Option A for the file, then send an ephemeral summary to the command invoker.

### Option C: SQLite Storage via `require("database")`

For incremental archiving, you could store messages in SQLite and export later:

```javascript
const database = require("database")

event("messageCreate", async (ctx) => {
  // Store every message in SQLite
  database.exec(
    "INSERT INTO archive_messages (id, channel_id, content, author, timestamp) VALUES (?, ?, ?, ?, ?)",
    ctx.message.id,
    ctx.message.channelID,
    ctx.message.content,
    ctx.message.author.username,
    new Date().toISOString(),
  )
})
```

**Pros:** Passive capture, no manual command needed.
**Cons:** Requires database setup, more complex, not the immediate goal.

**For the initial implementation, use Option A.**

---

## 10. Technology Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| Host runtime | Go + Goja | The framework the whole repo uses. Handles Discord connectivity. |
| Bot language | JavaScript (ES5-ish) | What the Goja runtime executes. |
| Bot DSL | `require("discord")` | Framework-provided API for commands, events, config. |
| Discord API | `ctx.discord.*` | Framework wrapper around Discord REST API. |
| File output | `ctx.discord.channels.send()` with `files` payload | No `fs` in JS runtime; deliver via Discord. |
| Optional storage | `require("database")` | SQLite module for incremental capture (future). |
| Timing | `require("timer")` | `sleep()` for demos or rate-limit pacing. |

---

## 11. Error Handling Strategy

| Scenario | Behavior |
|----------|----------|
| Command run outside a channel | Return ephemeral: "This command must be run in a channel." |
| Invalid thread ID | Catch error from `ctx.discord.threads.fetch()`, return ephemeral error. |
| Channel/thread has no messages | Return ephemeral: "No messages found." |
| Rate limited | `ctx.discord.messages.list()` handles this internally; we just paginate. |
| Message content empty (bot messages, embed-only) | Render as "(empty)" or skip. |
| File too large for Discord | Split into multiple files or truncate. |

---

## 12. Security & Privacy Considerations

- **Token secrecy:** The bot token is passed as a CLI flag (`--bot-token`), not hardcoded. Never commit tokens.
- **Data sensitivity:** Archives contain conversations. The bot should only operate in channels it has permission to read.
- **Bot permissions:** Request only `Read Messages/View Channels` and `Read Message History`. Do not request `Send Messages` unless needed for delivery.
- **User data:** Archives may contain personal data. Store downloaded files securely. Do not redistribute without consent.

---

## 13. API Reference Cheat Sheet

### Framework Registration Helpers

| Helper | Signature | Purpose |
|--------|-----------|---------|
| `defineBot(builderFn)` | `builderFn({ command, event, configure, ... })` | Entrypoint for every bot |
| `configure(options)` | `options = { name, description, category, run? }` | Set metadata and runtime config |
| `command(name, spec?, handler)` | `spec = { description, options }` | Register a slash command |
| `event(name, handler)` | `name = "ready" \| "messageCreate" \| ...` | Register event handler |

### Framework Context Fields (`ctx`)

| Field | Purpose |
|-------|---------|
| `ctx.args` / `ctx.options` | Parsed command option values |
| `ctx.config` | Runtime config values from `configure({ run: ... })` |
| `ctx.channel.id` | ID of the channel where the interaction occurred |
| `ctx.guild.id` | ID of the guild where the interaction occurred |
| `ctx.user.id` | ID of the user who invoked the command |
| `ctx.me` | Bot's own user record `{ id, username }` |
| `ctx.defer({ ephemeral? })` | Acknowledge interaction for slow work |
| `ctx.edit(payload)` | Edit the deferred/initial response |
| `ctx.followUp(payload)` | Send an additional follow-up message |
| `ctx.reply(payload)` | Send a response (for events or non-deferred) |
| `ctx.log.info/debug/warn/error(msg, fields)` | Structured logging |
| `ctx.store.get/set/delete/keys/namespace` | Per-runtime in-memory key/value store |

### Discord Operations (`ctx.discord.*`)

#### Messages

```javascript
// List messages (paginated)
const messages = await ctx.discord.messages.list(channelId, {
  before: "msg-id",  // optional
  after: "msg-id",   // optional
  around: "msg-id",  // optional
  limit: 100,        // max 100
})

// Fetch one message
const message = await ctx.discord.messages.fetch(channelId, messageId)
```

#### Channels

```javascript
const channel = await ctx.discord.channels.fetch(channelId)
await ctx.discord.channels.send(channelId, {
  content: "...",
  files: [{ name: "file.md", content: "..." }],
})
```

#### Threads

```javascript
const thread = await ctx.discord.threads.fetch(threadId)
await ctx.discord.threads.join(threadId)
await ctx.discord.threads.leave(threadId)
const started = await ctx.discord.threads.start(channelId, {
  name: "...",
  type: "public",
})
```

### Response Payload Shape

```javascript
// Simple text
return { content: "Hello", ephemeral: true }

// With embeds and components
return {
  content: "Hello",
  embeds: [{ title: "...", description: "...", color: 0x5865F2 }],
  components: [
    {
      type: "actionRow",
      components: [
        { type: "button", style: "primary", label: "Click", customId: "my:button" },
      ],
    },
  ],
  ephemeral: true,
}

// File attachment via channel send
await ctx.discord.channels.send(channelId, {
  content: "Here is a file",
  files: [
    { name: "report.md", content: "# Report\n\nHello world." },
  ],
})
```

---

## 14. Open Questions

1. **Should we support incremental updates?** (e.g., "archive only new messages since last run") — This would require persistent storage (SQLite) and is a future enhancement.
2. **How should we handle message edits?** — Show original, latest, or both? Currently we capture the state at archive time.
3. **Should we render embeds and attachments?** — The current `messages.list` API returns a simplified shape. We may need `messages.fetch` for full richness.
4. **What about forum channels?** — Every post is a thread. We may need a `/archive-forum` command that loops through all threads.
5. **Should we add a passive capture mode?** — Using `event("messageCreate", ...)` to store all messages in SQLite automatically.

---

## 15. Glossary

| Term | Definition |
|------|-----------|
| **Guild** | Discord's internal name for a "server" — a community space. |
| **Channel** | A chat room or category within a guild. |
| **Thread** | A temporary or permanent sub-channel attached to a message. |
| **Embed** | A rich card with title, description, fields, images — used by bots and link previews. |
| **Intent** | A permission declaration that tells Discord which events the bot needs. |
| **Token** | A secret string used to authenticate API requests. |
| **Snowflake** | Discord's ID format — a large integer encoding timestamp + worker ID + sequence. |
| **Goja** | The Go JavaScript engine used by our framework (not Node.js). |
| **Native Module** | A JS module provided by the Go host, like `require("discord")` or `require("database")`. |
| **Defer** | Acknowledging a Discord interaction immediately, then editing the response later. |
| **Ephemeral** | A message visible only to the user who triggered it. |

---

*Document version: 2.0*
*Ticket: DISCORD-ARCHIVE-BOT*
*Created: 2026-04-21*
*Updated: 2026-04-21 (corrected for discord-js-bot framework)*
