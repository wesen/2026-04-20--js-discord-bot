---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: internal/jsdiscord/host_commands.go
      Note: Command serialization — where all 6 gaps live
    - Path: internal/jsdiscord/host_maps.go
      Note: Attachment resolution in optionMap() needed for Gap 1
    - Path: ttmp/2026/04/21/DISCORD-BOT-020--discord-interaction-types-demo-bot-with-user-commands-message-commands-and-subcommands/sources/discord-application-commands-docs.md
      Note: Official Discord API reference for all missing features
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---




# Discord JS Bot Framework — Missing UI Surface Types & Command Metadata

> **Audience:** Backend developer familiar with Go and the discordgo library.  
> **Goal:** Identify every Discord API feature our framework does not yet expose to JavaScript bot scripts, and provide a concrete implementation plan for each.

---

## Executive Summary

Our `discord-js-bot` framework (Go host + Goja JS runtime) exposes most of Discord's UI surface to JavaScript bots, but six features are missing. Five are **command metadata/option features** that Discord's API supports but our `host_commands.go` serializer ignores. One is an **option type** (`ATTACHMENT`) that our `optionTypeFromSpec()` switch statement simply lacks a case for.

This ticket tracks adding all six features so bot developers can:
- Accept file uploads as slash command options
- Restrict channel pickers to specific channel types
- Require certain Discord permissions to use a command
- Mark commands as age-restricted
- Control where commands appear (guild vs. DM vs. user install)
- Localize command names and descriptions for international users

---

## 1. How the Framework Serializes Commands

Before listing gaps, understand the serialization path. When a JS bot calls `command("foo", { options: {...} }, handler)`, the framework:

1. **JS side** (`bot.go`): `command()` stores a `commandDraft` with the raw JS spec.
2. **Describe phase** (`descriptor.go`): The draft is serialized to a plain map via `commandSnapshotsFromDrafts()`.
3. **Discord conversion** (`host_commands.go`): `applicationCommandFromSnapshot()` reads the map and builds a `discordgo.ApplicationCommand` struct.
4. **Sync phase** (`bot.go` `SyncCommands()`): The struct is sent to Discord via `ApplicationCommandBulkOverwrite()`.

**The gaps are all in step 3** — `host_commands.go` does not read certain fields from the JS spec map.

```
JS command("ping", {
  description: "...",
  options: { ... },        ← host_commands.go reads this
  nsfw: true,              ← NOT read today
  default_member_permissions: "8",  ← NOT read today
  contexts: [0, 1],        ← NOT read today
  name_localizations: {...}  ← NOT read today
})
        │
        ▼
  commandDraft (bot.go)
        │
        ▼
  commandSnapshotsFromDrafts() (descriptor.go)
        │
        ▼
  applicationCommandFromSnapshot() (host_commands.go)
        │
        ▼
  discordgo.ApplicationCommand
        │
        ▼
  Discord API
```

---

## 2. The Six Gaps

### Gap 1: ATTACHMENT Option Type (Slash Command File Upload)

**What it is:** Discord option type `11` (`ApplicationCommandOptionTypeAttachment`). Lets users upload a file directly as a command argument. The file appears as an `Attachment` object in the interaction payload.

**Why it matters:** Without this, bots cannot receive files via slash commands. Users must DM files or upload them to a channel and reference by URL. For an archive bot, this would let users upload a template for formatting, or a JSON config for export settings.

**Current code:** `internal/jsdiscord/host_commands.go:optionTypeFromSpec()`

```go
switch strings.ToLower(strings.TrimSpace(fmt.Sprint(mapping["type"]))) {
case "", "string":
    return discordgo.ApplicationCommandOptionString, nil
// ... 9 more cases ...
case "sub_command_group":
    return discordgo.ApplicationCommandOptionSubCommandGroup, nil
default:
    return discordgo.ApplicationCommandOptionString, fmt.Errorf("unsupported option type %q", mapping["type"])
}
```

**Missing:** No case for `"attachment"`.

**Implementation:**

```go
case "attachment":
    return discordgo.ApplicationCommandOptionAttachment, nil
```

**JS usage:**

```javascript
command("import-config", {
  description: "Upload a config file",
  options: {
    config: {
      type: "attachment",
      description: "JSON config file",
      required: true,
    },
  },
}, async (ctx) => {
  const attachment = ctx.args.config
  // attachment = { id, filename, content_type, size, url }
  return { content: `Received ${attachment.filename} (${attachment.size} bytes)` }
})
```

**Handler payload shape:** The framework must also ensure `ctx.args` exposes attachment objects. Today `optionMap()` in `host_maps.go` only copies `option.Value`. For attachments, Discord sends an `attachment` object in the resolved data, not a simple value. This requires mapping `resolved.attachments` into `ctx.args`.

**Effort:** Small. One line in `optionTypeFromSpec()`, plus mapping resolved attachments in `host_maps.go`.

---

### Gap 2: `channel_types` Restriction for CHANNEL Options

**What it is:** When a slash command has a `CHANNEL` option, Discord shows a channel picker. By default it shows ALL channel types (text, voice, categories, forums, etc.). The `channel_types` field restricts the picker to specific types.

**Why it matters:** An archive bot's `/archive-channel` command should only accept text channels and forum channels, not voice channels or categories. Without this, users can pick invalid channels and the command fails at runtime.

**Discord API:**

```json
{
  "name": "channel",
  "type": 7,
  "channel_types": [0, 15]
}
```

Where `0` = `GuildText`, `15` = `GuildForum`.

**Current code:** `host_commands.go:optionSpecToDiscord()` reads `minLength`, `maxLength`, `minValue`, `maxValue`, but never `channel_types`.

**Implementation:**

In `optionSpecToDiscord()`, after setting `Type`, add:

```go
if optionType == discordgo.ApplicationCommandOptionChannel {
    if raw, ok := mapping["channel_types"]; ok {
        if types, ok := raw.([]any); ok {
            ret.ChannelTypes = make([]discordgo.ChannelType, 0, len(types))
            for _, t := range types {
                if n, ok := toInt(t); ok {
                    ret.ChannelTypes = append(ret.ChannelTypes, discordgo.ChannelType(n))
                }
            }
        }
    }
}
```

**JS usage:**

```javascript
command("archive-channel", {
  options: {
    target: {
      type: "channel",
      description: "Channel to archive",
      required: true,
      channel_types: [0, 15], // text and forum only
    },
  },
})
```

**Effort:** Small. ~10 lines in `host_commands.go`, plus doc update.

---

### Gap 3: `default_member_permissions`

**What it is:** A permission bitfield string that controls who can see/use a command. Setting `"0"` hides it from everyone except admins. Setting `"8"` (Administrator) restricts it. Setting `"1024"` (ManageMessages) restricts to moderators.

**Why it matters:** The archive bot should probably be restricted to users with `ReadMessageHistory` or `ManageMessages`. Without this, any server member can archive channels, potentially exposing sensitive discussions.

**Discord API:**

```json
{
  "name": "archive-channel",
  "default_member_permissions": "1024"
}
```

**Current code:** `applicationCommandFromSnapshot()` reads `name`, `description`, `options` but not `default_member_permissions`.

**Implementation:**

In `applicationCommandFromSnapshot()`, after setting `Description`, add:

```go
if raw, ok := snapshot["default_member_permissions"]; ok {
    if perms := strings.TrimSpace(fmt.Sprint(raw)); perms != "" {
        ret.DefaultMemberPermissions = &perms
    }
}
```

Note: `DefaultMemberPermissions` is a `*string` in discordgo (pointer, so nil = no restriction).

**JS usage:**

```javascript
configure({
  name: "archive-helper",
  // ...
})

command("archive-channel", {
  description: "Archive a channel",
  default_member_permissions: "1024", // ManageMessages
  options: { ... },
})
```

**Alternative design:** Put it in `configure()` so it applies to all commands:

```javascript
configure({
  name: "archive-helper",
  default_member_permissions: "1024",
})
```

This is simpler but less flexible. **Recommendation:** Support both. `configure()` sets a default; individual `command()` calls can override.

**Effort:** Small. ~5 lines in `host_commands.go`, plus propagate through the descriptor.

---

### Gap 4: `nsfw` Flag

**What it is:** Marks a command as age-restricted. Discord will hide it in channels that are not marked NSFW, and warn users before execution.

**Why it matters:** Less critical for an archive bot, but useful for moderation tools or any bot with adult-oriented features.

**Discord API:**

```json
{
  "name": "spoilers",
  "nsfw": true
}
```

**Current code:** Not read from spec.

**Implementation:**

```go
if raw, ok := snapshot["nsfw"]; ok {
    if nsfw, ok := raw.(bool); ok {
        ret.NSFW = nsfw
    }
}
```

**JS usage:**

```javascript
command("spoilers", {
  description: "Show spoiler content",
  nsfw: true,
})
```

**Effort:** Trivial. 3 lines.

---

### Gap 5: `contexts` / `integration_types`

**What it is:** Controls WHERE a command can be used:
- `integration_types`: Where the app is installed (`GUILD_INSTALL` = server, `USER_INSTALL` = user)
- `contexts`: Where the interaction happens (`GUILD` = server channel, `BOT_DM` = DM with bot, `PRIVATE_CHANNEL` = group DM)

**Why it matters:** The archive bot probably only makes sense in guilds (not DMs). Without this, the command appears everywhere the app is installed, which is confusing.

**Discord API:**

```json
{
  "name": "archive-channel",
  "contexts": [0],
  "integration_types": [0]
}
```

**Current code:** Not read from spec.

**Implementation:**

```go
// contexts
if raw, ok := snapshot["contexts"]; ok {
    if arr, ok := raw.([]any); ok {
        ret.Contexts = make([]discordgo.InteractionContextType, 0, len(arr))
        for _, item := range arr {
            if n, ok := toInt(item); ok {
                ret.Contexts = append(ret.Contexts, discordgo.InteractionContextType(n))
            }
        }
    }
}

// integration_types
if raw, ok := snapshot["integration_types"]; ok {
    if arr, ok := raw.([]any); ok {
        ret.IntegrationTypes = make([]discordgo.ApplicationIntegrationType, 0, len(arr))
        for _, item := range arr {
            if n, ok := toInt(item); ok {
                ret.IntegrationTypes = append(ret.IntegrationTypes, discordgo.ApplicationIntegrationType(n))
            }
        }
    }
}
```

**JS usage:**

```javascript
configure({
  name: "archive-helper",
  contexts: [0], // GUILD only
  integration_types: [0], // GUILD_INSTALL only
})
```

**Effort:** Small. ~20 lines. Requires checking discordgo struct field names (may be `InteractionContextTypes` or similar).

---

### Gap 6: `name_localizations` / `description_localizations`

**What it is:** Multi-language support for command names and descriptions. Discord shows the localized version based on the user's client language.

**Why it matters:** International servers. A German user sees `/archiv-kanal` instead of `/archive-channel`.

**Discord API:**

```json
{
  "name": "archive-channel",
  "name_localizations": {
    "de": "archiv-kanal",
    "fr": "archiver-canal"
  },
  "description_localizations": {
    "de": "Archiviert Nachrichten aus dem aktuellen Kanal",
    "fr": "Archive les messages du canal actuel"
  }
}
```

**Current code:** Not read from spec.

**Implementation:**

```go
if raw, ok := snapshot["name_localizations"]; ok {
    if locs, ok := raw.(map[string]any); ok {
        ret.NameLocalizations = make(map[discordgo.Locale]string)
        for k, v := range locs {
            ret.NameLocalizations[discordgo.Locale(k)] = fmt.Sprint(v)
        }
    }
}

if raw, ok := snapshot["description_localizations"]; ok {
    if locs, ok := raw.(map[string]any); ok {
        ret.DescriptionLocalizations = make(map[discordgo.Locale]string)
        for k, v := range locs {
            ret.DescriptionLocalizations[discordgo.Locale(k)] = fmt.Sprint(v)
        }
    }
}
```

**JS usage:**

```javascript
command("archive-channel", {
  description: "Archive messages from the current channel",
  name_localizations: {
    de: "archiv-kanal",
    fr: "archiver-canal",
  },
  description_localizations: {
    de: "Archiviert Nachrichten aus dem aktuellen Kanal",
    fr: "Archive les messages du canal actuel",
  },
  options: { ... },
})
```

**Also for options:** Each option should support `name_localizations` and `description_localizations` too. This is trickier because `optionSpecToDiscord()` already handles many fields. Add localization mapping there.

**Effort:** Medium. ~30 lines for commands, ~20 more for options. Requires checking discordgo type for `Locale`.

---

## 3. Bonus Gap: `attachment` in `ctx.args` Mapping

When a user uploads a file via an `ATTACHMENT` option, Discord does not send a simple value. It sends:

```json
{
  "data": {
    "options": [{"type": 11, "name": "config", "value": "attachment-id-123"}],
    "resolved": {
      "attachments": {
        "attachment-id-123": {
          "id": "attachment-id-123",
          "filename": "config.json",
          "size": 1024,
          "url": "https://cdn.discordapp.com/...",
          "proxy_url": "https://media.discordapp.net/...",
          "content_type": "application/json"
        }
      }
    }
  }
}
```

Our `optionMap()` in `host_maps.go` only copies `option.Value` (the attachment ID string). We need to resolve the ID to the full attachment object.

**Implementation:** In `host_dispatch.go` or wherever `ctx.args` is built, check if any option is type `Attachment`, look it up in `interaction.Data.Resolved.Attachments`, and replace the string ID with the full attachment map.

**Effort:** Small but requires understanding the dispatch flow.

---

## 4. Implementation Order (Recommended)

| Phase | Feature | Effort | Why first |
|-------|---------|--------|-----------|
| 1 | `ATTACHMENT` type | Small | Enables new UX (file uploads); one-liner + resolved mapping |
| 2 | `channel_types` | Small | Improves UX for all channel options; prevents invalid selections |
| 3 | `nsfw` | Trivial | One boolean field; good warm-up |
| 4 | `default_member_permissions` | Small | Security-critical; restricts command visibility |
| 5 | `contexts` / `integration_types` | Small | Prevents command pollution in wrong contexts |
| 6 | `name_localizations` / `description_localizations` | Medium | Nice-to-have; more complex due to nested option support |

---

## 5. Files to Modify

| File | What to change |
|------|---------------|
| `internal/jsdiscord/host_commands.go` | Add `attachment` case; read `channel_types`, `nsfw`, `default_member_permissions`, `contexts`, `integration_types`, localizations |
| `internal/jsdiscord/host_maps.go` | Resolve attachment IDs to full attachment objects in `optionMap()` |
| `internal/jsdiscord/bot.go` | Optionally: propagate `default_member_permissions` from `configure()` metadata |
| `pkg/doc/topics/discord-js-bot-api-reference.md` | Document all new features |
| `examples/discord-bots/` | Add example bot demonstrating new features |

---

## 6. Testing Strategy

For each feature:

1. **Write a minimal bot** that uses the new feature.
2. **Run with `--sync-on-start`** and verify the command appears correctly in Discord.
3. **Check Discord client behavior:**
   - `channel_types`: Does the channel picker filter correctly?
   - `default_member_permissions`: Does the command hide for users without the permission?
   - `nsfw`: Does Discord show an age-gate?
   - `contexts`: Does the command disappear from DMs?
   - `attachment`: Can you upload a file and does the bot receive it?
   - `localizations`: Change Discord client language, does the command name change?
4. **Run existing bots** (`ping`, `moderation`, `knowledge-base`) to ensure no regressions.

---

## 7. Related Tickets

- **DISCORD-ARCHIVE-BOT** — The archive bot that originally surfaced these gaps
- **DISCORD-BOT-020** — Interaction types demo bot (has context menu commands)

---

*Document version: 1.0*
*Ticket: DISCORD-BOT-FRAMEWORK-GAPS*
*Created: 2026-04-21*
