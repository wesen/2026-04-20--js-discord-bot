---
Title: Discord Credentials and Setup
Ticket: DISCORD-BOT-001
Status: active
Topics:
    - backend
    - chat
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "What Discord credentials the bot needs, where to get them, and how to wire them into local development."
LastUpdated: 2026-04-20T10:04:43.000894693-04:00
WhatFor: "Provide a copy/paste-ready checklist for bot setup, credential acquisition, and invite URLs."
WhenToUse: "Use when creating the Discord application, configuring local development, or inviting the bot to a server."
---

# Discord Credentials and Setup

## Goal

Document exactly which Discord credentials the Go bot needs, where each one comes from, and how to wire them into a local development environment without leaking secrets.

## Context

A Discord bot usually needs a small set of values from the Discord Developer Portal. The minimal gateway-based setup uses a bot token and an application ID. During development, a guild ID is also very useful because guild-scoped slash commands appear faster than global ones.

If you later add a browser-based install flow or interactions endpoint, you may need extra values such as the public key or OAuth2 client secret. Those are optional for a simple gateway bot.

## Quick Reference

### Required credentials

| Value | Environment variable | Needed for | Where to get it |
| --- | --- | --- | --- |
| Bot token | `DISCORD_BOT_TOKEN` | Authenticating the bot session to the Discord gateway | Developer Portal → your application → **Bot** tab → **Reset Token** / **Copy** |
| Application ID | `DISCORD_APPLICATION_ID` | Slash command registration and invite URL generation | Developer Portal → your application → **General Information** |

### Recommended development credentials

| Value | Environment variable | Needed for | Where to get it |
| --- | --- | --- | --- |
| Guild ID | `DISCORD_GUILD_ID` | Fast command registration in a private test server | In Discord, enable **Developer Mode**, right-click the server, and choose **Copy Server ID** |
| Log level | `DISCORD_LOG_LEVEL` | Local debugging and troubleshooting | Chosen by you |

### Optional credentials

| Value | Environment variable | Needed for | Where to get it |
| --- | --- | --- | --- |
| Public key | `DISCORD_PUBLIC_KEY` | Only if you expose an interactions HTTP endpoint | Developer Portal → **General Information** |
| Client secret | `DISCORD_CLIENT_SECRET` | Only if you implement OAuth2 login or a web install flow | Developer Portal → **OAuth2** |

### Required Discord scopes for inviting the bot

Use these scopes in the OAuth2 URL:

- `bot`
- `applications.commands`

### Common permissions for a simple bot

Start small. Typical permissions for a basic utility bot are:

- Send Messages
- View Channels
- Read Message History
- Use Application Commands

Only add more permissions if the bot really needs them.

### Suggested environment file

```bash
DISCORD_BOT_TOKEN=your_bot_token_here
DISCORD_APPLICATION_ID=123456789012345678
DISCORD_GUILD_ID=123456789012345678
DISCORD_LOG_LEVEL=info
```

### Example invite URL

Replace the application ID and permissions bitfield:

```text
https://discord.com/api/oauth2/authorize?client_id=YOUR_APPLICATION_ID&permissions=YOUR_PERMISSION_BITFIELD&scope=bot%20applications.commands
```

## Usage Examples

### 1) Create the Discord application

1. Go to the Discord Developer Portal.
2. Create a new application.
3. Open the application’s **Bot** tab.
4. Add a bot user if one does not exist yet.
5. Reset the token and copy it once.
6. Save the token into a password manager or secret manager.

### 2) Prepare a local `.env` file

```bash
DISCORD_BOT_TOKEN=...
DISCORD_APPLICATION_ID=...
DISCORD_GUILD_ID=...
```

Load it locally only. Do not commit it.

### 3) Invite the bot to a private test guild

1. Enable Developer Mode in Discord if you need the server ID.
2. Generate an OAuth2 invite URL with scopes `bot` and `applications.commands`.
3. Invite the bot to your private server.
4. Give it the minimal permissions it needs.

### 4) Sync guild commands during development

Use a guild ID while iterating so slash commands show up quickly.

```text
discord-bot sync-commands --guild-id 123456789012345678
```

### 5) Run the bot

```text
discord-bot run
```

## Related

- `design-doc/01-implementation-and-architecture-guide.md`
- `reference/01-diary.md`
