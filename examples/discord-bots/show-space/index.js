const { defineBot } = require("discord")
const { createShowCatalog, loadSeedShows, normalizeShow } = require("./lib/shows")
const { createShowStore } = require("./lib/store")
const { localDateISO } = require("./lib/dates")
const { canAdminOnly, canManageShows, permissionDenied } = require("./lib/permissions")
const {
  showAnnouncementPayload,
  upcomingShowsText,
  showDetailPayload,
  pastShowsText,
  cancellationNotice,
  parseShowTitle,
} = require("./lib/render")

const catalog = createShowCatalog(loadSeedShows(), {
  timeZone: "America/New_York",
})
const store = createShowStore()

function trimText(value) {
  return String(value || "").trim()
}

function configValue(ctx, key) {
  return trimText((ctx && ctx.config && ctx.config[key]) || "")
}

function configChannelId(ctx, key) {
  return configValue(ctx, key)
}

function debugEnabled(ctx) {
  const value = ctx && ctx.config ? ctx.config.debug : false
  if (typeof value === "boolean") {
    return value
  }
  const text = String(value || "").trim().toLowerCase()
  return ["1", "true", "yes", "on"].includes(text)
}

function commandError(error) {
  return { content: `❌ ${String(error || "Something went wrong.")}`, ephemeral: true }
}

function hasDatabase(ctx) {
  return store.ensure(ctx.config)
}

function repoListUpcoming(ctx, limit) {
  if (hasDatabase(ctx)) {
    return store.listUpcoming(ctx.config, limit)
  }
  return catalog.listUpcoming(limit)
}

function repoListPast(ctx, limit) {
  if (hasDatabase(ctx)) {
    return store.listPast(ctx.config, limit)
  }
  return catalog.listPast(limit)
}

function repoGetShow(ctx, id) {
  if (hasDatabase(ctx)) {
    return store.getShow(ctx.config, id)
  }
  return catalog.getShow(id)
}

function repoCreateShow(ctx, rawShow) {
  if (hasDatabase(ctx)) {
    return store.createShow(ctx.config, rawShow)
  }
  return catalog.addShow(rawShow)
}

function repoAttachDiscordMessage(ctx, id, channelId, messageId) {
  if (hasDatabase(ctx)) {
    return store.attachDiscordMessage(ctx.config, id, channelId, messageId)
  }
  return catalog.attachDiscordMessage(id, channelId, messageId)
}

function repoCancelShow(ctx, id) {
  if (hasDatabase(ctx)) {
    return store.cancelShow(ctx.config, id)
  }
  return catalog.cancelShow(id)
}

function repoArchiveShow(ctx, id) {
  if (hasDatabase(ctx)) {
    return store.archiveShow(ctx.config, id)
  }
  return catalog.archiveShow(id)
}

function repoArchiveByDiscordMessage(ctx, channelId, messageId) {
  if (hasDatabase(ctx)) {
    return store.archiveByDiscordMessage(ctx.config, channelId, messageId)
  }
  return catalog.archiveByDiscordMessage(channelId, messageId)
}

function repoArchiveExpiredShows(ctx, referenceDate) {
  if (hasDatabase(ctx)) {
    return store.archiveExpiredShows(ctx.config, referenceDate)
  }
  return catalog.markExpiredArchived(referenceDate)
}

function buildAnnouncementInput(ctx, args) {
  return normalizeShow({
    artist: args.artist,
    date: args.date,
    doors_time: args.doors_time,
    age_restriction: args.age_restriction || args.age,
    price: args.price,
    notes: args.notes,
    source: "announcement",
  }, {
    timeZone: configValue(ctx, "timeZone"),
    referenceDate: new Date(),
  })
}

function findAnnouncementMessage(messages, show) {
  const wantedTitle = `🎵 ${show.artist} — ${show.displayDate}`
  const list = Array.isArray(messages) ? messages : []
  return list.find((message) => {
    if (!message) {
      return false
    }
    const embeds = Array.isArray(message.embeds) ? message.embeds : []
    return embeds.some((embed) => String((embed && embed.title) || "").trim() === wantedTitle)
  }) || null
}

function parsePinnedShowMessage(message) {
  const embeds = message && Array.isArray(message.embeds) ? message.embeds : []
  const embed = embeds[0] || {}
  const parsed = parseShowTitle(embed.title)
  if (!parsed.dateText) {
    return null
  }
  const parsedDate = new Date(parsed.dateText)
  if (Number.isNaN(parsedDate.getTime())) {
    return null
  }
  return {
    id: message.id,
    artist: parsed.artist,
    dateText: parsed.dateText,
    dateISO: localDateISO(parsedDate),
  }
}

async function postAnnouncement(ctx, rawShow) {
  const channelId = configChannelId(ctx, "upcomingShowsChannelId")
  if (!channelId) {
    return { ok: false, error: "upcomingShowsChannelId is not configured." }
  }

  const saved = repoCreateShow(ctx, rawShow)
  if (!saved.ok) {
    return { ok: false, error: saved.error }
  }

  await ctx.discord.channels.send(channelId, showAnnouncementPayload(saved.show))
  const recent = await ctx.discord.messages.list(channelId, { limit: 5 })
  const matched = findAnnouncementMessage(recent, saved.show)
  if (!matched) {
    return { ok: false, error: "Posted the announcement, but could not find the message to pin." }
  }
  await ctx.discord.messages.pin(channelId, matched.id)

  const attached = repoAttachDiscordMessage(ctx, saved.show.id, channelId, matched.id)
  if (!attached.ok) {
    return { ok: false, error: attached.error }
  }

  return { ok: true, show: attached.show || saved.show, messageId: matched.id }
}

async function archiveExpiredShows(ctx, options) {
  const channelId = configChannelId(ctx, "upcomingShowsChannelId")
  const staffChannelId = configChannelId(ctx, "staffChannelId")
  const referenceDate = new Date()
  const expired = repoArchiveExpiredShows(ctx, referenceDate)
  let unpinned = 0

  for (const show of Array.isArray(expired) ? expired : []) {
    const showChannelId = configValue({ config: { upcomingShowsChannelId: channelId } }, "upcomingShowsChannelId") || trimText(show.discordChannelId)
    const messageId = trimText(show.discordMessageId)
    if (showChannelId && messageId) {
      try {
        await ctx.discord.messages.unpin(showChannelId, messageId)
        unpinned += 1
      } catch (err) {
        // Keep going: the show is still archived in the store.
      }
    }
  }

  if (options && options.logStaff && staffChannelId && expired.length > 0) {
    try {
      await ctx.discord.channels.send(staffChannelId, {
        content: `📦 Auto-archived ${expired.length} past show(s).`,
      })
    } catch (err) {
      // Ignore staff-log failures so the maintenance flow still finishes.
    }
  }

  return { archived: expired.length, unpinned }
}

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "show-space",
    description: "Venue operations bot for upcoming shows and pinned announcements",
    category: "venues",
    run: {
      fields: {
        upcomingShowsChannelId: { type: "string", help: "Public channel for show announcements and pins", default: "" },
        announcementsChannelId: { type: "string", help: "Public channel for general announcements", default: "" },
        staffChannelId: { type: "string", help: "Private staff channel for summaries", default: "" },
        adminRoleId: { type: "string", help: "Discord role ID for admins", default: "" },
        bookerRoleId: { type: "string", help: "Discord role ID for bookers", default: "" },
        timeZone: { type: "string", help: "IANA timezone for display formatting", default: "America/New_York" },
        dbPath: { type: "string", help: "SQLite database path for phase-2 persistence", default: "" },
        seedFromJson: { type: "bool", help: "Seed the database from shows.json when empty", default: true },
        debug: { type: "bool", help: "Enable debug-only helper commands like role lookup", default: false },
      },
    },
  })

  event("ready", async (ctx) => {
    const showCount = hasDatabase(ctx) ? store.countShows(ctx.config) : catalog.listAll().length
    ctx.log.info("show-space bot ready", {
      user: ctx.me && ctx.me.username,
      script: ctx.metadata && ctx.metadata.scriptPath,
      shows: showCount,
      dbEnabled: hasDatabase(ctx),
    })
  })

  command("upcoming", {
    description: "Show upcoming shows",
  }, async (ctx) => {
    const shows = repoListUpcoming(ctx, 25)
    return {
      content: upcomingShowsText(shows),
      ephemeral: true,
    }
  })

  command("debug-roles", {
    description: "List guild role IDs for debugging (requires --debug)",
  }, async (ctx) => {
    if (!debugEnabled(ctx)) {
      return {
        content: "Debug mode is disabled. Re-run the bot with --debug to use this command.",
        ephemeral: true,
      }
    }
    const guildId = ctx.guild && ctx.guild.id
    if (!guildId) {
      return commandError("This command requires a guild context.")
    }
    const roles = await ctx.discord.roles.list(guildId)
    const guildName = trimText((ctx.guild && ctx.guild.name) || guildId)
    const lines = Array.isArray(roles) && roles.length > 0
      ? roles.map((role) => `• ${trimText(role && role.name) || "(unnamed)"} — ${trimText(role && role.id) || "(no id)"}`).join("\n")
      : "No roles found."
    return {
      content: `Guild roles for ${guildName}:\n\n${lines}`,
      ephemeral: true,
    }
  })

  command("announce", {
    description: "Post and pin a show announcement in #upcoming-shows",
    options: {
      artist: { type: "string", description: "Artist or band name", required: true },
      date: { type: "string", description: "Show date", required: true },
      doors_time: { type: "string", description: "Doors time", required: true },
      age_restriction: { type: "string", description: "Age restriction", required: true },
      price: { type: "string", description: "Ticket price", required: true },
      notes: { type: "string", description: "Optional notes", required: false },
    },
  }, async (ctx) => {
    if (!canManageShows(ctx)) {
      return permissionDenied()
    }
    const normalized = buildAnnouncementInput(ctx, ctx.args)
    if (!normalized.ok) {
      return commandError(normalized.error)
    }
    const posted = await postAnnouncement(ctx, normalized.show)
    if (!posted.ok) {
      return commandError(posted.error)
    }
    return {
      content: "✅ Posted and pinned in #upcoming-shows",
      ephemeral: true,
    }
  })

  command("add-show", {
    description: "Save a show, post the announcement, and store the Discord message ID",
    options: {
      artist: { type: "string", description: "Artist or band name", required: true },
      date: { type: "string", description: "Show date", required: true },
      doors_time: { type: "string", description: "Doors time", required: true },
      age: { type: "string", description: "Age restriction", required: true },
      price: { type: "string", description: "Ticket price", required: true },
      notes: { type: "string", description: "Optional notes", required: false },
    },
  }, async (ctx) => {
    if (!canManageShows(ctx)) {
      return permissionDenied()
    }
    const normalized = buildAnnouncementInput(ctx, {
      artist: ctx.args.artist,
      date: ctx.args.date,
      doors_time: ctx.args.doors_time,
      age_restriction: ctx.args.age,
      price: ctx.args.price,
      notes: ctx.args.notes,
    })
    if (!normalized.ok) {
      return commandError(normalized.error)
    }
    const posted = await postAnnouncement(ctx, normalized.show)
    if (!posted.ok) {
      return commandError(posted.error)
    }
    return {
      content: `✅ Show added — ID #${String(posted.show.id)}. Posted and pinned.`,
      ephemeral: true,
    }
  })

  command("show", {
    description: "Return the full details for one show",
    options: {
      id: { type: "string", description: "Show ID", required: true },
    },
  }, async (ctx) => {
    const show = repoGetShow(ctx, ctx.args.id)
    if (!show) {
      return commandError(`No show found for ID ${JSON.stringify(String(ctx.args.id || ""))}.`)
    }
    return {
      ...showDetailPayload(show),
      ephemeral: true,
    }
  })

  command("cancel-show", {
    description: "Cancel a show, unpin the original announcement, and post a cancellation notice",
    options: {
      id: { type: "string", description: "Show ID", required: true },
    },
  }, async (ctx) => {
    if (!canManageShows(ctx)) {
      return permissionDenied()
    }
    const show = repoGetShow(ctx, ctx.args.id)
    if (!show) {
      return commandError(`No show found for ID ${JSON.stringify(String(ctx.args.id || ""))}.`)
    }
    const cancelled = repoCancelShow(ctx, ctx.args.id)
    if (!cancelled.ok) {
      return commandError(cancelled.error)
    }
    if (trimText(show.discordChannelId) && trimText(show.discordMessageId)) {
      try {
        await ctx.discord.messages.unpin(show.discordChannelId, show.discordMessageId)
      } catch (err) {
        // Keep the cancellation flow moving even if the pin is already gone.
      }
    }
    const announceChannelId = configChannelId(ctx, "upcomingShowsChannelId") || trimText(show.discordChannelId)
    if (announceChannelId) {
      try {
        await ctx.discord.channels.send(announceChannelId, cancellationNotice(cancelled.show || show))
      } catch (err) {
        // The record is still cancelled even if the cancellation notice fails.
      }
    }
    return {
      content: `✅ Show #${String((cancelled.show && cancelled.show.id) || ctx.args.id)} cancelled and unpinned.`,
      ephemeral: true,
    }
  })

  command("archive-show", {
    description: "Archive a completed show and unpin its announcement",
    options: {
      id: { type: "string", description: "Show ID", required: true },
    },
  }, async (ctx) => {
    if (!canAdminOnly(ctx)) {
      return permissionDenied()
    }
    const show = repoGetShow(ctx, ctx.args.id)
    if (!show) {
      return commandError(`No show found for ID ${JSON.stringify(String(ctx.args.id || ""))}.`)
    }
    const archived = repoArchiveShow(ctx, ctx.args.id)
    if (!archived.ok) {
      return commandError(archived.error)
    }
    if (trimText(show.discordChannelId) && trimText(show.discordMessageId)) {
      try {
        await ctx.discord.messages.unpin(show.discordChannelId, show.discordMessageId)
      } catch (err) {
        // Ignore already-unpinned messages.
      }
    }
    return {
      content: `✅ Show #${String((archived.show && archived.show.id) || ctx.args.id)} archived and unpinned.`,
      ephemeral: true,
    }
  })

  command("past-shows", {
    description: "Return recently archived shows",
  }, async (ctx) => {
    const shows = repoListPast(ctx, 5)
    return {
      content: pastShowsText(shows),
      ephemeral: true,
    }
  })

  command("unpin-old", {
    description: "Unpin expired show announcements from #upcoming-shows",
  }, async (ctx) => {
    if (!canAdminOnly(ctx)) {
      return permissionDenied()
    }
    const channelId = configChannelId(ctx, "upcomingShowsChannelId")
    if (!channelId) {
      return commandError("upcomingShowsChannelId is not configured.")
    }
    const pinned = await ctx.discord.messages.listPinned(channelId)
    const nowISO = localDateISO(new Date())
    let removed = 0
    for (const message of Array.isArray(pinned) ? pinned : []) {
      const parsed = parsePinnedShowMessage(message)
      if (!parsed || !parsed.dateISO || parsed.dateISO >= nowISO) {
        continue
      }
      try {
        await ctx.discord.messages.unpin(channelId, message.id)
        removed += 1
        repoArchiveByDiscordMessage(ctx, channelId, message.id)
      } catch (err) {
        // Keep going if one pin is already gone.
      }
    }
    return {
      content: `Removed ${removed} expired pin(s).`,
      ephemeral: true,
    }
  })

  command("archive-expired", {
    description: "Archive expired shows and post a quiet staff summary",
  }, async (ctx) => {
    if (!canAdminOnly(ctx)) {
      return permissionDenied()
    }
    const result = await archiveExpiredShows(ctx, { logStaff: true })
    return {
      content: `Archived ${result.archived} expired show(s) and unpinned ${result.unpinned}.`,
      ephemeral: true,
    }
  })
})
