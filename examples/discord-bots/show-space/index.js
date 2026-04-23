const { defineBot } = require("discord")
const { createShowCatalog, loadSeedShows, normalizeShow } = require("./lib/shows")
const { localDateISO } = require("./lib/dates")
const { canAdminOnly, canManageShows, permissionDenied } = require("./lib/permissions")
const { showAnnouncementPayload, upcomingShowsText } = require("./lib/render")

const catalog = createShowCatalog(loadSeedShows(), {
  timeZone: "America/New_York",
})

function configChannelId(ctx, key) {
  return String((ctx && ctx.config && ctx.config[key]) || "").trim()
}

function safeMessageList(messageList) {
  return Array.isArray(messageList) ? messageList : []
}

function commandError(error) {
  return { content: `❌ ${String(error || "Something went wrong.")}`, ephemeral: true }
}

function resolveAnnouncementMatch(messages, show) {
  const wantedTitle = `🎵 ${show.artist} — ${show.displayDate}`
  return safeMessageList(messages).find((message) => {
    if (!message) {
      return false
    }
    const embeds = Array.isArray(message.embeds) ? message.embeds : []
    return embeds.some((embed) => String((embed && embed.title) || "").trim() === wantedTitle)
  }) || null
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
      },
    },
  })

  event("ready", async (ctx) => {
    ctx.log.info("show-space bot ready", {
      user: ctx.me && ctx.me.username,
      script: ctx.metadata && ctx.metadata.scriptPath,
      shows: catalog.listAll().length,
    })
  })

  command("upcoming", {
    description: "Show upcoming shows",
  }, async () => {
    const shows = catalog.listUpcoming(25)
    return {
      content: upcomingShowsText(shows),
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
    const normalized = normalizeShow({
      artist: ctx.args.artist,
      date: ctx.args.date,
      doors_time: ctx.args.doors_time,
      age_restriction: ctx.args.age_restriction,
      price: ctx.args.price,
      notes: ctx.args.notes,
      source: "announcement",
    }, { timeZone: String((ctx.config && ctx.config.timeZone) || "") })
    if (!normalized.ok) {
      return commandError(normalized.error)
    }

    const channelId = configChannelId(ctx, "upcomingShowsChannelId")
    if (!channelId) {
      return commandError("upcomingShowsChannelId is not configured.")
    }

    const saved = catalog.addShow(normalized.show)
    if (!saved.ok) {
      return commandError(saved.error)
    }

    await ctx.discord.channels.send(channelId, showAnnouncementPayload(saved.show))
    const recent = await ctx.discord.messages.list(channelId, { limit: 5 })
    const matched = resolveAnnouncementMatch(recent, saved.show)
    if (!matched) {
      return commandError("Posted the announcement, but could not find the message to pin.")
    }
    await ctx.discord.messages.pin(channelId, matched.id)
    const attached = catalog.attachDiscordMessage(saved.show.id, channelId, matched.id)
    if (!attached.ok) {
      return commandError(attached.error)
    }
    return {
      content: "✅ Posted and pinned in #upcoming-shows",
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
    const now = new Date()
    const todayISO = localDateISO(now)
    let removed = 0
    for (const message of safeMessageList(pinned)) {
      const embeds = Array.isArray(message.embeds) ? message.embeds : []
      const embed = embeds[0] || {}
      const title = String((embed && embed.title) || "")
      const dateMatch = title.match(/[—-]\s*(.+)$/)
      if (!dateMatch) {
        continue
      }
      const parsedDate = new Date(String(dateMatch[1] || "").trim())
      if (Number.isNaN(parsedDate.getTime())) {
        continue
      }
      const dateISO = localDateISO(parsedDate)
      if (dateISO && todayISO && dateISO < todayISO) {
        await ctx.discord.messages.unpin(channelId, message.id)
        removed += 1
      }
    }
    return {
      content: `Removed ${removed} expired pin(s).`,
      ephemeral: true,
    }
  })
})
