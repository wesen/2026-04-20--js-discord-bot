const { defineBot } = require("discord")
const ui = require("ui")
const { createShowCatalog, loadSeedShows, normalizeShow } = require("./lib/shows")
const { createShowStore } = require("./lib/store")
const { localDateISO } = require("./lib/dates")
const { cleanRoleList, canAdminOnly, canManageShows, permissionDenied } = require("./lib/permissions")
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

function yesNo(value) {
  return value ? "yes" : "no"
}

function guildId(ctx) {
  return trimText((ctx && ctx.guild && ctx.guild.id) || "")
}

function guildName(ctx) {
  return trimText((ctx && ctx.guild && ctx.guild.name) || guildId(ctx) || "(no guild)")
}

function userId(ctx) {
  return trimText((ctx && ctx.user && ctx.user.id) || (ctx && ctx.member && ctx.member.id) || "")
}

function userName(ctx) {
  return trimText((ctx && ctx.user && ctx.user.username) || (ctx && ctx.user && ctx.user.globalName) || (ctx && ctx.member && ctx.member.nick) || userId(ctx) || "(unknown user)")
}

function memberId(ctx) {
  return trimText((ctx && ctx.member && ctx.member.id) || "")
}

function configuredRoleIds(ctx) {
  const config = ctx && ctx.config ? ctx.config : {}
  return [config.adminRoleId, config.bookerRoleId]
    .map((roleId) => String(roleId || "").trim())
    .filter(Boolean)
}

function intersectRoleIds(left, right) {
  const leftList = Array.isArray(left) ? left : []
  const rightList = Array.isArray(right) ? right : []
  return leftList.filter((value) => rightList.includes(value))
}

function subtractRoleIds(left, right) {
  const leftList = Array.isArray(left) ? left : []
  const rightList = Array.isArray(right) ? right : []
  return leftList.filter((value) => !rightList.includes(value))
}

function bulletList(items, emptyText) {
  const list = Array.isArray(items) ? items : []
  return list.length > 0 ? list.join("\n") : emptyText
}

function truncatedBulletList(items, emptyText, limit) {
  const list = Array.isArray(items) ? items : []
  const visible = list.slice(0, limit)
  const extra = list.length - visible.length
  const base = bulletList(visible, emptyText)
  return extra > 0 ? `${base}\n…and ${extra} more` : base
}

function formatRoleLine(role) {
  if (!role) {
    return "• (unknown role) — (no id)"
  }
  const markers = []
  if (role.member) {
    markers.push("member")
  }
  if (role.configured) {
    markers.push("configured")
  }
  const markerText = markers.length > 0 ? ` [${markers.join(", ")}]` : ""
  return `• ${role.name || "(unnamed)"} — ${role.id || "(no id)"}${markerText}`
}

async function collectDebugSnapshot(ctx) {
  const config = ctx && ctx.config ? ctx.config : {}
  const guildID = guildId(ctx)
  const roleIDs = cleanRoleList(ctx)
  const allowedRoleIds = configuredRoleIds(ctx)
  let guildRoles = []
  let guildRoleError = ""

  if (guildID && ctx && ctx.discord && ctx.discord.roles && typeof ctx.discord.roles.list === "function") {
    try {
      const listedRoles = await ctx.discord.roles.list(guildID)
      guildRoles = Array.isArray(listedRoles) ? listedRoles : []
    } catch (err) {
      guildRoleError = String(err || "Unable to load guild roles.")
    }
  }

  const roleByID = new Map(
    Array.isArray(guildRoles)
      ? guildRoles
        .map((role) => [trimText(role && role.id), role])
        .filter(([roleID]) => Boolean(roleID))
      : []
  )

  const memberRoles = roleIDs.map((roleID) => {
    const role = roleByID.get(roleID)
    return {
      id: roleID,
      name: role ? (trimText(role.name) || "(unnamed)") : "(unknown role)",
      member: true,
      configured: allowedRoleIds.includes(roleID),
    }
  })

  const guildRoleDetails = Array.isArray(guildRoles)
    ? guildRoles
      .map((role) => {
        const roleID = trimText(role && role.id)
        if (!roleID) {
          return null
        }
        return {
          id: roleID,
          name: trimText(role && role.name) || "(unnamed)",
          member: roleIDs.includes(roleID),
          configured: allowedRoleIds.includes(roleID),
        }
      })
      .filter(Boolean)
    : []

  return {
    debugEnabled: debugEnabled(ctx),
    guildID,
    guildName: guildName(ctx),
    userID: userId(ctx),
    userName: userName(ctx),
    memberID: memberId(ctx),
    memberRoles,
    memberRoleIDs: roleIDs,
    allowedRoleIds,
    guildRoles: guildRoleDetails,
    guildRoleError,
    canManageShows: canManageShows(ctx),
    canAdminOnly: canAdminOnly(ctx),
    config,
  }
}

function debugViewTitle(view) {
  const labels = {
    summary: "Summary",
    member: "Member roles",
    guild: "Guild roles",
    config: "Config",
    checks: "Permission checks",
  }
  return labels[view] || labels.summary
}

function debugButtons(view) {
  const style = (name) => view === name ? "primary" : "secondary"
  return [
    ui.button("show-space:debug:summary", "Summary", style("summary")),
    ui.button("show-space:debug:member", "Member", style("member")),
    ui.button("show-space:debug:guild", "Guild", style("guild")),
    ui.button("show-space:debug:config", "Config", style("config")),
    ui.button("show-space:debug:checks", "Checks", style("checks")),
  ]
}

async function renderDebugMessage(ctx, view) {
  const activeView = ["summary", "member", "guild", "config", "checks"].includes(view) ? view : "summary"
  const snapshot = await collectDebugSnapshot(ctx)
  const content = `🔎 Show Space debug — ${debugViewTitle(activeView).toLowerCase()} view\nUser ID: ${snapshot.userID || "(unknown)"}\nMember ID: ${snapshot.memberID || "(none)"}\nGuild: ${snapshot.guildName} (${snapshot.guildID || "no guild"})`

  let embed = ui.embed(`Show Space Debug — ${debugViewTitle(activeView)}`)
    .color(0x5865F2)
    .description(`Use the buttons to switch between the debug views.\n\nUser: ${snapshot.userName} — ${snapshot.userID || "(unknown)"}`)

  if (activeView === "summary") {
    embed = embed
      .field("Permission checks", `canManageShows: ${yesNo(snapshot.canManageShows)}\ncanAdminOnly: ${yesNo(snapshot.canAdminOnly)}`, false)
      .field("Allowed role IDs", bulletList(snapshot.allowedRoleIds.map((roleID) => `• ${roleID}`), "(not configured)"), false)
      .field("Member role IDs", bulletList(snapshot.memberRoleIDs.map((roleID) => `• ${roleID}`), "(none)"), false)
  } else if (activeView === "member") {
    embed = embed
      .field("Member role IDs", bulletList(snapshot.memberRoleIDs.map((roleID) => `• ${roleID}`), "(none)"), false)
      .field("Resolved member roles", bulletList(snapshot.memberRoles.map(formatRoleLine), "(none)"), false)
      .field("Matched configured roles", bulletList(snapshot.memberRoles.filter((role) => role.configured).map(formatRoleLine), "(none)"), false)
  } else if (activeView === "guild") {
    embed = embed
      .field("Guild roles", truncatedBulletList(snapshot.guildRoles.map(formatRoleLine), snapshot.guildRoleError || "(no guild roles found)", 20), false)
      .field("Configured roles", bulletList(snapshot.allowedRoleIds.map((roleID) => `• ${roleID}`), "(not configured)"), false)
  } else if (activeView === "config") {
    embed = embed
      .field("Runtime config", [
        `debug: ${String(snapshot.debugEnabled)}`,
        `upcomingShowsChannelId: ${snapshot.config.upcomingShowsChannelId ? String(snapshot.config.upcomingShowsChannelId) : "(not configured)"}`,
        `announcementsChannelId: ${snapshot.config.announcementsChannelId ? String(snapshot.config.announcementsChannelId) : "(not configured)"}`,
        `staffChannelId: ${snapshot.config.staffChannelId ? String(snapshot.config.staffChannelId) : "(not configured)"}`,
        `adminRoleId: ${snapshot.config.adminRoleId ? String(snapshot.config.adminRoleId) : "(not configured)"}`,
        `bookerRoleId: ${snapshot.config.bookerRoleId ? String(snapshot.config.bookerRoleId) : "(not configured)"}`,
        `dbPath: ${snapshot.config.dbPath ? String(snapshot.config.dbPath) : "(not configured)"}`,
        `seedFromJson: ${String(snapshot.config.seedFromJson)}`,
        `timeZone: ${snapshot.config.timeZone ? String(snapshot.config.timeZone) : "(not configured)"}`,
      ].join("\n"), false)
  } else if (activeView === "checks") {
    const matchingRoleIds = intersectRoleIds(snapshot.allowedRoleIds, snapshot.memberRoleIDs)
    const missingRoleIds = subtractRoleIds(snapshot.allowedRoleIds, snapshot.memberRoleIDs)
    embed = embed
      .field("Permission result", `canManageShows: ${yesNo(snapshot.canManageShows)}\ncanAdminOnly: ${yesNo(snapshot.canAdminOnly)}`, false)
      .field("Exact matching role IDs", bulletList(matchingRoleIds.map((roleId) => `• ${roleId}`), "(none)"), false)
      .field("Configured role IDs not seen on member", bulletList(missingRoleIds.map((roleId) => `• ${roleId}`), "(none)"), false)
      .field("Intersection", bulletList(snapshot.memberRoles.filter((role) => role.configured).map(formatRoleLine), "(none)"), false)
      .field("Why the bot may deny access", matchingRoleIds.length > 0
        ? "The bot sees at least one configured role ID on your member object. If access still fails, compare the exact configured IDs, the specific command gate, and the live runtime config."
        : "The bot does not see any configured role IDs on your member object. Compare the configured admin/booker role IDs against the exact Discord role IDs in the guild.", false)
  }

  if (snapshot.guildRoleError) {
    embed = embed.field("Guild role lookup", snapshot.guildRoleError, false)
  }

  return ui.message()
    .ephemeral()
    .content(content)
    .embed(embed)
    .row(...debugButtons(activeView))
    .build()
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

module.exports = defineBot(({ command, component, event, configure }) => {
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

  command("debug", {
    description: "Open the show-space debug dashboard (requires --debug)",
  }, async (ctx) => {
    if (!debugEnabled(ctx)) {
      return {
        content: "Debug mode is disabled. Re-run the bot with --debug to use this command.",
        ephemeral: true,
      }
    }
    if (!guildId(ctx)) {
      return commandError("This command requires a guild context.")
    }
    return renderDebugMessage(ctx, "summary")
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
    if (!guildId(ctx)) {
      return commandError("This command requires a guild context.")
    }
    return renderDebugMessage(ctx, "guild")
  })

  command("debug-my-roles", {
    description: "List the roles the bot sees on your member object (requires --debug)",
  }, async (ctx) => {
    if (!debugEnabled(ctx)) {
      return {
        content: "Debug mode is disabled. Re-run the bot with --debug to use this command.",
        ephemeral: true,
      }
    }
    if (!guildId(ctx)) {
      return commandError("This command requires a guild context.")
    }
    return renderDebugMessage(ctx, "member")
  })

  component("show-space:debug:summary", async (ctx) => {
    if (!debugEnabled(ctx)) {
      return {
        content: "Debug mode is disabled. Re-run the bot with --debug to use this command.",
        ephemeral: true,
      }
    }
    if (!guildId(ctx)) {
      return commandError("This command requires a guild context.")
    }
    return renderDebugMessage(ctx, "summary")
  })

  component("show-space:debug:member", async (ctx) => {
    if (!debugEnabled(ctx)) {
      return {
        content: "Debug mode is disabled. Re-run the bot with --debug to use this command.",
        ephemeral: true,
      }
    }
    if (!guildId(ctx)) {
      return commandError("This command requires a guild context.")
    }
    return renderDebugMessage(ctx, "member")
  })

  component("show-space:debug:guild", async (ctx) => {
    if (!debugEnabled(ctx)) {
      return {
        content: "Debug mode is disabled. Re-run the bot with --debug to use this command.",
        ephemeral: true,
      }
    }
    if (!guildId(ctx)) {
      return commandError("This command requires a guild context.")
    }
    return renderDebugMessage(ctx, "guild")
  })

  component("show-space:debug:config", async (ctx) => {
    if (!debugEnabled(ctx)) {
      return {
        content: "Debug mode is disabled. Re-run the bot with --debug to use this command.",
        ephemeral: true,
      }
    }
    if (!guildId(ctx)) {
      return commandError("This command requires a guild context.")
    }
    return renderDebugMessage(ctx, "config")
  })

  component("show-space:debug:checks", async (ctx) => {
    if (!debugEnabled(ctx)) {
      return {
        content: "Debug mode is disabled. Re-run the bot with --debug to use this command.",
        ephemeral: true,
      }
    }
    if (!guildId(ctx)) {
      return commandError("This command requires a guild context.")
    }
    return renderDebugMessage(ctx, "checks")
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
      return permissionDenied(ctx, {
        requiredRoleIds: [ctx.config && ctx.config.adminRoleId, ctx.config && ctx.config.bookerRoleId],
        requirementLabel: "adminRoleId or bookerRoleId",
      })
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
      return permissionDenied(ctx, {
        requiredRoleIds: [ctx.config && ctx.config.adminRoleId, ctx.config && ctx.config.bookerRoleId],
        requirementLabel: "adminRoleId or bookerRoleId",
      })
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
      return permissionDenied(ctx, {
        requiredRoleIds: [ctx.config && ctx.config.adminRoleId, ctx.config && ctx.config.bookerRoleId],
        requirementLabel: "adminRoleId or bookerRoleId",
      })
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
      return permissionDenied(ctx, {
        requiredRoleIds: [ctx.config && ctx.config.adminRoleId],
        requirementLabel: "adminRoleId",
      })
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
      return permissionDenied(ctx, {
        requiredRoleIds: [ctx.config && ctx.config.adminRoleId],
        requirementLabel: "adminRoleId",
      })
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
      return permissionDenied(ctx, {
        requiredRoleIds: [ctx.config && ctx.config.adminRoleId],
        requirementLabel: "adminRoleId",
      })
    }
    const result = await archiveExpiredShows(ctx, { logStaff: true })
    return {
      content: `Archived ${result.archived} expired show(s) and unpinned ${result.unpinned}.`,
      ephemeral: true,
    }
  })
})
