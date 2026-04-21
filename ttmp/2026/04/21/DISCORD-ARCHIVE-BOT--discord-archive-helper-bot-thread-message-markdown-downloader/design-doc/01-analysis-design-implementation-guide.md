---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/21/DISCORD-ARCHIVE-BOT--discord-archive-helper-bot-thread-message-markdown-downloader/design-doc/01-analysis-design-implementation-guide.md
      Note: Primary design document for the archive bot
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

We need a **helper bot** that connects to Discord, reads messages from channels and threads, and saves them as clean Markdown files on disk. Think of it as a "print to PDF" button, but for Discord conversations — producing structured, searchable, offline-readable archives. The bot runs from the command line, takes a channel or thread URL, fetches all messages (handling pagination automatically), renders them into Markdown with proper formatting for embeds, attachments, reactions, and timestamps, and writes them to a local folder tree.

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
| **Rate Limits** | Discord caps how many API calls we can make per second. | We must handle 429 errors and back off. |

### The Discord API Landscape

Discord offers two ways to interact programmatically:

1. **REST API** — request/response, like a normal web API. Used for fetching historical messages, channels, guilds.
2. **Gateway (WebSocket)** — persistent connection for real-time events (new messages, user joins, etc.).

For an archive bot, we primarily use the **REST API** because we care about historical data, not real-time events. However, some Discord libraries (like `discord.js`) wrap both and make our lives easier.

---

## 2. Problem Statement

### The Pain Point

Manuel participates in many Discord servers with valuable discussions — technical debates, project planning, design decisions, troubleshooting sessions. Discord is a great live medium, but:

- **Search is weak** for long-running threads.
- **No bulk export** — Discord has no "download this thread as a file" feature.
- **Data loss risk** — servers can be deleted, threads auto-archive, accounts banned.
- **Offline reading** — sometimes you want to read on a plane or on a reMarkable device.

### What We Need

A tool that, given a Discord channel or thread URL, produces a clean Markdown file (or folder of files) containing:

- All messages in chronological order.
- Proper author attribution with timestamps.
- Formatted text (bold, italic, code blocks, links).
- Embeds rendered as Markdown callouts or tables.
- Attachment URLs preserved (or files downloaded).
- Reactions summarized.
- Reply chains indicated.

---

## 3. Proposed Solution

### 3.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Discord Archive Helper Bot                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌──────────────┐     ┌──────────────┐     ┌──────────────────────────┐   │
│   │   CLI Input  │────▶│   Fetcher    │────▶│   Markdown Renderer      │   │
│   │   (yargs)    │     │  (discord.js)│     │  (custom + marked)       │   │
│   └──────────────┘     └──────────────┘     └──────────────────────────┘   │
│          │                    │                        │                    │
│          │                    │                        │                    │
│          ▼                    ▼                        ▼                    │
│   ┌──────────────┐     ┌──────────────┐     ┌──────────────────────────┐   │
│   │ Config (.env)│     │  Rate Limiter│     │   File Writer            │   │
│   │  DISCORD_    │     │  (built-in)  │     │  (fs + path)             │   │
│   │  TOKEN, etc. │     │              │     │                          │   │
│   └──────────────┘     └──────────────┘     └──────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Discord Servers                                │
│                         (via HTTPS + WebSocket)                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Data Flow

```
User runs:
  node archive-bot.js --channel 123456789 --output ./archives/

        │
        ▼
  ┌─────────────┐
  │ Parse args  │ ──▶ Validate channel ID format, check output dir exists
  └─────────────┘
        │
        ▼
  ┌─────────────┐
  │  Login      │ ──▶ Connect to Discord Gateway with bot token + intents
  └─────────────┘
        │
        ▼
  ┌─────────────┐
  │  Resolve    │ ──▶ Fetch channel object from Discord API
  │  Channel    │     (confirms it exists and bot can see it)
  └─────────────┘
        │
        ▼
  ┌─────────────┐
  │  Fetch      │ ──▶ GET /channels/{id}/messages with pagination
  │  Messages   │     (loop until no more messages)
  └─────────────┘
        │
        ▼
  ┌─────────────┐
  │  Discover   │ ──▶ Check for active threads in channel
  │  Threads    │     (forum channels: every post is a thread)
  └─────────────┘
        │
        ▼
  ┌─────────────┐
  │  For each   │ ──▶ Recursively fetch messages from each thread
  │  thread...  │
  └─────────────┘
        │
        ▼
  ┌─────────────┐
  │  Render     │ ──▶ Convert message objects to Markdown strings
  │  Markdown   │
  └─────────────┘
        │
        ▼
  ┌─────────────┐
  │  Write      │ ──▶ Save to disk: YYYY/MM/DD/channel-name--thread-name.md
  │  Files      │
  └─────────────┘
        │
        ▼
  ┌─────────────┐
  │  Report     │ ──▶ Print summary: N messages, M threads, saved to path
  │  Summary    │
  └─────────────┘
```

---

## 4. Deep Dive: Discord API for Archiving

### 4.1 Authentication

Every request to Discord must include an `Authorization` header:

```
Authorization: Bot YOUR_BOT_TOKEN_HERE
```

**Critical security rule:** The token is a secret. Store it in a `.env` file, load it via `dotenv`, and **never** commit `.env` to version control. Add `.env` to `.gitignore` immediately.

### 4.2 Required Intents

When connecting a bot, you must declare which events you want. For archiving, you need:

| Intent | Purpose |
|--------|---------|
| `Guilds` | Access to guild/channel structure. |
| `GuildMessages` | Access to messages in guild channels. |
| `MessageContent` | **Essential.** Without this, message text is empty. |
| `GuildMessageReactions` | Access to reaction counts (optional but nice). |

### 4.3 Fetching Messages (Pagination)

Discord's `GET /channels/{channel.id}/messages` endpoint returns **up to 100 messages per request**. To get all messages, you paginate backwards using the `before` parameter:

```
GET /channels/123456789/messages?limit=100
  → returns messages [950..1049] (newest first)

GET /channels/123456789/messages?limit=100&before=950
  → returns messages [850..949]

...repeat until response is empty...
```

**Pseudocode — Message Fetching Loop:**

```
function fetchAllMessages(channelId):
    allMessages = []
    lastMessageId = null

    loop:
        if lastMessageId is null:
            url = "/channels/{channelId}/messages?limit=100"
        else:
            url = "/channels/{channelId}/messages?limit=100&before={lastMessageId}"

        batch = discordAPI.get(url)

        if batch is empty:
            break

        allMessages.addAll(batch)
        lastMessageId = batch.last().id

        // Rate limit: Discord allows ~5 requests per second per channel
        sleep(200ms)

    // Messages arrive newest-first; reverse for chronological order
    return reverse(allMessages)
```

### 4.4 Thread Discovery

#### Regular Text Channels with Threads

A text channel may have "active" and "archived" threads. Use:

```
GET /channels/{channel.id}/threads/archived/public
GET /channels/{channel.id}/threads/archived/private
GET /guilds/{guild.id}/threads/active
```

#### Forum Channels

In a forum channel, **every post is a thread**. The channel itself contains no messages — only thread listings. Fetch the threads, then fetch messages from each thread.

### 4.5 Rate Limiting

Discord returns rate-limit headers on every response:

```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 3
X-RateLimit-Reset-After: 1.250
X-RateLimit-Bucket: abc123...
```

If you hit the limit, Discord returns `429 Too Many Requests`. You **must** back off and retry. The `discord.js` library handles this automatically, which is why we use it.

---

## 5. Message Object Structure

Here is what a Discord message looks like when returned from the API. Understanding every field is critical because we must render each into Markdown.

```json
{
  "id": "1234567890123456789",
  "channel_id": "9876543210987654321",
  "author": {
    "id": "111222333444555666",
    "username": "alice",
    "global_name": "Alice",
    "bot": false
  },
  "content": "Hey everyone, check out this design doc!",
  "timestamp": "2024-03-15T14:32:01.123000+00:00",
  "edited_timestamp": null,
  "tts": false,
  "mention_everyone": false,
  "mentions": [],
  "mention_roles": [],
  "attachments": [
    {
      "id": "999888777666555444",
      "filename": "design.png",
      "size": 204800,
      "url": "https://cdn.discordapp.com/attachments/.../design.png",
      "content_type": "image/png",
      "width": 1200,
      "height": 800
    }
  ],
  "embeds": [
    {
      "title": "Design System v2",
      "description": "Updated color palette and typography...",
      "url": "https://figma.com/file/...",
      "color": 3447003,
      "fields": [
        { "name": "Status", "value": "In Review", "inline": true },
        { "name": "Owner", "value": "Alice", "inline": true }
      ]
    }
  ],
  "reactions": [
    { "emoji": { "name": "👍" }, "count": 5, "me": false },
    { "emoji": { "name": "🎉" }, "count": 2, "me": false }
  ],
  "pinned": false,
  "type": 0,
  "flags": 0,
  "referenced_message": {
    "id": "111222333444555666",
    "author": { "username": "bob" },
    "content": "We need a new design system..."
  }
}
```

### 5.1 Rendering Map: API Field → Markdown

| API Field | Markdown Output | Example |
|-----------|-----------------|---------|
| `author.username` | Bold username prefix | **`alice:`** Hey everyone... |
| `timestamp` | Small timestamp in parentheses | **`alice`** *(2024-03-15 14:32)*: |
| `content` | Main message body (Discord → MD conversion) | Hey everyone, check out... |
| `attachments[]` | Image embed or link | `![design.png](URL)` or `[design.png](URL)` |
| `embeds[]` | Blockquote / callout with title, fields, description | See section 7.3 |
| `reactions[]` | Inline summary: 👍×5 🎉×2 | `👍 5  🎉 2` |
| `referenced_message` | Quote block indicating reply | `> **bob:** We need a new...` |
| `edited_timestamp` | "(edited)" suffix | ...design doc! *(edited)* |

---

## 6. Markdown Rendering in Detail

### 6.1 Discord Markup → Markdown

Discord uses a simplified markup language. We must convert it to standard Markdown:

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
| `<#channelId>` | `[#channel-name](discord://discord.com/channels/...)` | Resolve ID to name if possible. |
| `<@userId>` | `@username` | Resolve ID to username. |
| `<@&roleId>` | `@role-name` | Resolve ID to role name. |
| `<:emoji:123>` | `:emoji:` | Or use the CDN URL for custom emoji. |
| `https://...` | `<https://...>` or bare URL | Same. |

**Pseudocode — Content Converter:**

```
function discordToMarkdown(content, guildContext):
    // Mentions: <@123> → @username
    content = replaceRegex(content, /<@(\d+)>/g, (match, userId) => {
        user = guildContext.resolveUser(userId)
        return "@" + (user?.username || "unknown-user")
    })

    // Channel mentions: <#123> → #channel-name
    content = replaceRegex(content, /<#(\d+)>/g, (match, channelId) => {
        channel = guildContext.resolveChannel(channelId)
        return "#" + (channel?.name || "unknown-channel")
    })

    // Role mentions: <@&123> → @role-name
    content = replaceRegex(content, /<@&(\d+)>/g, (match, roleId) => {
        role = guildContext.resolveRole(roleId)
        return "@" + (role?.name || "unknown-role")
    })

    // Custom emoji: <:name:id> → :name: or image link
    content = replaceRegex(content, /<:(\w+):(\d+)>/g, (match, name, id) => {
        return `![:${name}:](https://cdn.discordapp.com/emojis/${id}.png)`
    })

    // Animated emoji: <a:name:id>
    content = replaceRegex(content, /<a:(\w+):(\d+)>/g, (match, name, id) => {
        return `![:${name}:](https://cdn.discordapp.com/emojis/${id}.gif)`
    })

    // Multi-line quotes: >>> \n... → > line1\n> line2
    content = replaceRegex(content, /^>>>\s*\n((?:.|\n)*)/gm, (match, quote) => {
        return quote.split("\n").map(line => "> " + line).join("\n")
    })

    return content
```

### 6.2 Message Header Format

Each message in the archive should have a consistent header:

```markdown
**alice** *(2024-03-15 14:32 UTC)*:
```

For replies, prepend a quote block:

```markdown
> **bob** *(2024-03-15 14:28 UTC)*: We need a new design system...

**alice** *(2024-03-15 14:32 UTC)*: Hey everyone, check out this design doc!
```

### 6.3 Embed Rendering

Embeds are rich cards that bots and link previews generate. They have:
- `title` + `url`
- `description`
- `color` (integer RGB)
- `fields[]` (name/value pairs, possibly inline)
- `image`, `thumbnail`, `video`, `footer`

**Markdown representation (callout style):**

```markdown
> **🔗 Design System v2**
> Updated color palette and typography...
>
> | Status | Owner |
> |--------|-------|
> | In Review | Alice |
>
> [Open in Figma](https://figma.com/file/...)
```

**Pseudocode — Embed Renderer:**

```
function renderEmbed(embed):
    lines = []

    if embed.title:
        if embed.url:
            lines.push("> **🔗 [" + embed.title + "](" + embed.url + ")**")
        else:
            lines.push("> **🔗 " + embed.title + "**")

    if embed.description:
        lines.push("> " + embed.description.replace("\n", "\n> "))

    if embed.fields and embed.fields.length > 0:
        // Group inline fields side-by-side, non-inline fields full-width
        tableRows = []
        currentRow = []
        for field in embed.fields:
            if field.inline and currentRow.length < 3:
                currentRow.push(field)
            else:
                if currentRow.length > 0:
                    tableRows.push(currentRow)
                currentRow = [field]
        if currentRow.length > 0:
            tableRows.push(currentRow)

        for row in tableRows:
            header = "| " + row.map(f => f.name).join(" | ") + " |"
            separator = "|" + row.map(() => " --- ").join("|") + "|"
            values = "| " + row.map(f => f.value).join(" | ") + " |"
            lines.push("> " + header)
            lines.push("> " + separator)
            lines.push("> " + values)
            lines.push("> ")

    if embed.image:
        lines.push("> ![embed image](" + embed.image.url + ")")

    if embed.footer:
        lines.push("> *" + embed.footer.text + "*")

    return lines.join("\n")
```

### 6.4 Attachment Rendering

```
function renderAttachment(attachment):
    if attachment.content_type starts with "image/":
        return "![" + attachment.filename + "](" + attachment.url + ")"
    else:
        return "[📎 " + attachment.filename + "](" + attachment.url + ")"
```

### 6.5 Reaction Summary

```
function renderReactions(reactions):
    if reactions.length == 0:
        return ""
    parts = reactions.map(r => r.emoji.name + " " + r.count)
    return "*Reactions: " + parts.join("  ") + "*"
```

---

## 7. File Output Structure

### 7.1 Directory Layout

```
archives/
└── 2026/
    └── 04/
        └── 21/
            ├── general-chat--2026-04-21.md
            ├── design-discussion--2026-04-21.md
            └── design-discussion/
                ├── thread--color-palette-review--2026-04-21.md
                ├── thread--typography-choices--2026-04-21.md
                └── thread--mobile-layout--2026-04-21.md
```

### 7.2 Filename Convention

```
{sanitized-channel-name}--{YYYY-MM-DD}.md
{sanitized-channel-name}/{thread|post}--{sanitized-thread-name}--{YYYY-MM-DD}.md
```

Sanitization rules:
- Lowercase.
- Replace spaces and special chars with `-`.
- Collapse multiple `-` into one.
- Max 80 characters.

### 7.3 Frontmatter

Each archive file should include YAML frontmatter for metadata:

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
participant_count: 8
bot_version: "1.0.0"
---
```

---

## 8. Technology Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| Language | Node.js (JavaScript/TypeScript) | Discord ecosystem is JS-heavy; excellent libraries exist. |
| Discord Library | `discord.js` v14 | Most mature, well-documented, handles pagination and rate limits. |
| CLI Framework | `yargs` or `commander` | Standard for Node CLI tools; good help generation. |
| Environment | `dotenv` | Load secrets from `.env` file. |
| Markdown Conversion | Custom + `marked` (optional) | Custom for Discord→MD; `marked` if we need parsing. |
| File I/O | Native `fs/promises` | No extra dependency needed. |
| Testing | `vitest` or `jest` | Unit tests for rendering functions. |
| Formatting | `prettier` | Consistent code style. |

---

## 9. Implementation Plan (Step by Step)

### Phase 0: Setup (Day 1)

1. **Create a Discord Application**
   - Go to https://discord.com/developers/applications
   - Click "New Application", name it "Archive Helper"
   - Navigate to "Bot" tab, click "Add Bot"
   - Under "Privileged Gateway Intents", enable **Message Content Intent**
   - Copy the **Token** (you will only see it once; save it in `.env`)

2. **Invite the Bot to Your Server**
   - OAuth2 → URL Generator
   - Scopes: `bot`
   - Permissions: `Read Messages/View Channels`, `Read Message History`
   - Copy the generated URL, open in browser, select your server

3. **Project Scaffold**
   ```
   discord-archive-bot/
   ├── .env                  # secrets (gitignored)
   ├── .gitignore            # node_modules, .env, archives/
   ├── package.json
   ├── src/
   │   ├── index.js          # CLI entry point
   │   ├── client.js         # Discord client setup
   │   ├── fetcher.js        # Message/thread fetching logic
   │   ├── renderer.js       # Markdown rendering
   │   ├── writer.js         # File output
   │   └── utils.js          # Sanitization, date formatting
   ├── tests/
   │   └── renderer.test.js
   └── README.md
   ```

### Phase 1: Core Connection (Day 1–2)

Implement `src/client.js`:

```javascript
// src/client.js
import { Client, GatewayIntentBits } from 'discord.js';
import dotenv from 'dotenv';

dotenv.config();

export function createClient() {
  const client = new Client({
    intents: [
      GatewayIntentBits.Guilds,
      GatewayIntentBits.GuildMessages,
      GatewayIntentBits.MessageContent,
    ],
  });
  return client;
}

export async function login(client) {
  const token = process.env.DISCORD_BOT_TOKEN;
  if (!token) {
    throw new Error('DISCORD_BOT_TOKEN not found in environment');
  }
  await client.login(token);
  console.log(`Logged in as ${client.user.tag}`);
}
```

### Phase 2: Fetching Messages (Day 2–3)

Implement `src/fetcher.js`:

```javascript
// src/fetcher.js
export async function fetchAllMessages(channel) {
  const messages = [];
  let lastId = null;

  while (true) {
    const options = { limit: 100 };
    if (lastId) options.before = lastId;

    const batch = await channel.messages.fetch(options);
    if (batch.size === 0) break;

    messages.push(...batch.values());
    lastId = batch.last().id;

    // Rate limit safety
    await sleep(200);
  }

  // Reverse to chronological order (oldest first)
  return messages.reverse();
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
```

Implement thread discovery:

```javascript
export async function fetchThreads(channel) {
  const threads = [];

  // Active threads in guild
  const active = await channel.guild.channels.fetchActiveThreads();
  for (const thread of active.threads.values()) {
    if (thread.parentId === channel.id) {
      threads.push(thread);
    }
  }

  // Archived threads
  let archived = await channel.threads.fetchArchived({ fetchAll: true });
  threads.push(...archived.threads.values());

  return threads;
}
```

### Phase 3: Rendering (Day 3–4)

Implement `src/renderer.js` with the conversion logic from section 6.

Key functions:
- `renderMessage(message)` → string
- `renderEmbed(embed)` → string
- `renderAttachment(attachment)` → string
- `discordToMarkdown(content, guild)` → string
- `renderReactions(reactions)` → string

### Phase 4: File Writing (Day 4)

Implement `src/writer.js`:

```javascript
// src/writer.js
import fs from 'fs/promises';
import path from 'path';

export async function writeArchive(outputDir, channel, thread, renderedMarkdown, meta) {
  const dateStr = new Date().toISOString().split('T')[0];
  const channelDir = path.join(outputDir, sanitize(channel.name));
  await fs.mkdir(channelDir, { recursive: true });

  let fileName;
  if (thread) {
    fileName = `thread--${sanitize(thread.name)}--${dateStr}.md`;
  } else {
    fileName = `${sanitize(channel.name)}--${dateStr}.md`;
  }

  const filePath = path.join(channelDir, fileName);
  const frontmatter = buildFrontmatter(meta);
  await fs.writeFile(filePath, frontmatter + '\n' + renderedMarkdown, 'utf-8');

  return filePath;
}

function sanitize(name) {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 80);
}

function buildFrontmatter(meta) {
  return '---\n' + Object.entries(meta)
    .map(([k, v]) => `${k}: ${JSON.stringify(v)}`)
    .join('\n') + '\n---\n';
}
```

### Phase 5: CLI Assembly (Day 4–5)

Wire everything in `src/index.js` using `yargs`:

```javascript
// src/index.js
import yargs from 'yargs';
import { hideBin } from 'yargs/helpers';
import { createClient, login } from './client.js';
import { fetchAllMessages, fetchThreads } from './fetcher.js';
import { renderChannel } from './renderer.js';
import { writeArchive } from './writer.js';

const argv = yargs(hideBin(process.argv))
  .option('channel', { type: 'string', demandOption: true, describe: 'Channel ID or URL' })
  .option('output', { type: 'string', default: './archives', describe: 'Output directory' })
  .option('threads', { type: 'boolean', default: true, describe: 'Include threads' })
  .option('download-attachments', { type: 'boolean', default: false, describe: 'Download attachment files' })
  .help()
  .argv;

async function main() {
  const client = createClient();
  await login(client);

  // Resolve channel from ID or URL
  const channelId = extractChannelId(argv.channel);
  const channel = await client.channels.fetch(channelId);

  if (!channel || !channel.isTextBased()) {
    throw new Error('Channel not found or not a text channel');
  }

  console.log(`Archiving #${channel.name}...`);

  // Fetch and render main channel messages
  const messages = await fetchAllMessages(channel);
  const mainMarkdown = renderChannel(channel, messages);
  await writeArchive(argv.output, channel, null, mainMarkdown, {
    source: 'discord',
    server: channel.guild.name,
    server_id: channel.guild.id,
    channel: channel.name,
    channel_id: channel.id,
    archived_at: new Date().toISOString(),
    message_count: messages.length,
  });

  // Fetch and render threads
  if (argv.threads) {
    const threads = await fetchThreads(channel);
    for (const thread of threads) {
      console.log(`  Archiving thread: ${thread.name}...`);
      const threadMessages = await fetchAllMessages(thread);
      const threadMarkdown = renderChannel(thread, threadMessages);
      await writeArchive(argv.output, channel, thread, threadMarkdown, {
        source: 'discord',
        server: channel.guild.name,
        server_id: channel.guild.id,
        channel: channel.name,
        channel_id: channel.id,
        thread: thread.name,
        thread_id: thread.id,
        archived_at: new Date().toISOString(),
        message_count: threadMessages.length,
      });
    }
  }

  console.log('Done.');
  await client.destroy();
  process.exit(0);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
```

### Phase 6: Testing & Hardening (Day 5–6)

- Write unit tests for `renderer.js` using sample message objects.
- Test with a real Discord channel (use a private test server first).
- Verify handling of:
  - Empty channels.
  - Channels with 10,000+ messages (rate limit behavior).
  - Channels with many threads.
  - Messages with complex embeds, multiple attachments, reactions.
  - Deleted messages (should be skipped gracefully).
  - Missing permissions (should produce clear error).

### Phase 7: Documentation (Day 6)

- Complete `README.md` with setup instructions, usage examples, and environment variables.
- Document the archive file format for future consumers.

---

## 10. Error Handling Strategy

| Scenario | Behavior |
|----------|----------|
| Invalid token | Clear error: "DISCORD_BOT_TOKEN invalid or missing" |
| Channel not found | Error: "Channel {id} not found or bot lacks access" |
| Missing Message Content Intent | Warning: "Messages will be empty; enable Message Content Intent in Developer Portal" |
| Rate limited | `discord.js` auto-retries; we add a 200ms sleep between batches. |
| Partial failure (one thread errors) | Log error, continue with remaining threads. |
| Disk full | Error with path; do not truncate files. |
| Network interruption | Retry up to 3 times with exponential backoff. |

---

## 11. Security & Privacy Considerations

- **Token secrecy:** The bot token grants access to everything the bot can see. Treat it like a password. Use `.env`, never log it, never commit it.
- **Data sensitivity:** Archives contain conversations. Store them in a secure location. Do not upload to public repositories.
- **Bot permissions:** Request only the minimum permissions needed (`Read Messages/View Channels`, `Read Message History`). Do not request `Send Messages`, `Manage Messages`, or admin privileges.
- **User data:** Be aware that Discord's Terms of Service and GDPR apply to user data. Do not redistribute archives without consent.

---

## 12. API Reference Cheat Sheet

### Discord.js v14 Key Classes

| Class | Purpose | Key Methods |
|-------|---------|-------------|
| `Client` | Bot connection | `.login(token)`, `.channels.fetch(id)`, `.destroy()` |
| `TextChannel` | Text channel | `.messages.fetch(options)`, `.threads.fetchArchived()` |
| `ThreadChannel` | Thread | Same as TextChannel (inherits) |
| `Message` | Single message | `.content`, `.author`, `.attachments`, `.embeds`, `.createdAt` |
| `MessageManager` | Channel's messages | `.fetch({ limit, before })` |
| `GuildChannelManager` | Guild channels | `.fetchActiveThreads()` |

### REST Endpoints (if not using discord.js)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/channels/{id}` | Get channel info |
| GET | `/channels/{id}/messages` | Get messages (paginated) |
| GET | `/channels/{id}/threads/archived/public` | Get archived public threads |
| GET | `/guilds/{id}/threads/active` | Get active threads |
| GET | `/channels/{id}/messages/{message.id}` | Get single message |

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DISCORD_BOT_TOKEN` | Yes | Bot authentication token |
| `DISCORD_GUILD_ID` | No | Default guild ID (optional) |
| `ARCHIVE_OUTPUT_DIR` | No | Default output directory (default: `./archives`) |

---

## 13. Open Questions

1. Should we download attachment files locally, or only save URLs?
2. Should we support incremental updates ("archive only new messages since last run")?
3. How should we handle message edits — show original, latest, or both?
4. Should we output HTML in addition to Markdown?
5. What is the target reMarkable reading experience? (This affects line length, image sizing, etc.)

---

## 14. Glossary

| Term | Definition |
|------|-----------|
| **Guild** | Discord's internal name for a "server" — a community space. |
| **Channel** | A chat room or category within a guild. |
| **Thread** | A temporary or permanent sub-channel attached to a message. |
| **Embed** | A rich card with title, description, fields, images — used by bots and link previews. |
| **Intent** | A permission declaration that tells Discord which events the bot needs. |
| **Token** | A secret string used to authenticate API requests. |
| **Snowflake** | Discord's ID format — a large integer encoding timestamp + worker ID + sequence. |
| **Rate Limit** | A cap on API requests per time window. |
| **CDN** | Content Delivery Network — where Discord hosts attachments and emoji images. |

---

*Document version: 1.0*
*Ticket: DISCORD-ARCHIVE-BOT*
*Created: 2026-04-21*
