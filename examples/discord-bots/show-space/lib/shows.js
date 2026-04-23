const { compareDateISO, isPastDate, parseShowDate, todayISO } = require("./dates")

function trimText(value) {
  return String(value || "").trim()
}

function clone(show) {
  return JSON.parse(JSON.stringify(show))
}

function normalizeShow(raw, options) {
  const value = raw || {}
  const displayDateSource = trimText(value.dateDisplay || value.displayDate || value.dateISO)
  const dateInput = trimText(value.date || value.dateISO || displayDateSource)
  const parsed = parseShowDate(dateInput, options)
  if (!parsed.ok) {
    return { ok: false, error: parsed.error }
  }
  const artist = trimText(value.artist)
  if (!artist) {
    return { ok: false, error: "Artist is required." }
  }

  return {
    ok: true,
    show: {
      id: trimText(value.id),
      artist,
      dateISO: parsed.dateISO,
      displayDate: parsed.displayDate,
      doorsTime: trimText(value.doors_time || value.doorsTime),
      ageRestriction: trimText(value.age_restriction || value.age || value.ageRestriction),
      price: trimText(value.price),
      notes: trimText(value.notes),
      status: trimText(value.status) || "confirmed",
      discordChannelId: trimText(value.discord_channel_id || value.discordChannelId),
      discordMessageId: trimText(value.discord_message_id || value.discordMessageId),
      source: trimText(value.source) || "seed",
      createdAt: trimText(value.createdAt) || new Date().toISOString(),
      updatedAt: trimText(value.updatedAt) || new Date().toISOString(),
    },
  }
}

function createShowCatalog(seedShows, options) {
  const config = options || {}
  const state = []
  let sequence = 0

  function generateID(prefix) {
    sequence += 1
    return `${prefix}-${sequence}`
  }

  function addNormalizedShow(raw, source) {
    const normalized = normalizeShow({ ...raw, source: source || raw && raw.source }, config)
    if (!normalized.ok) {
      return normalized
    }
    const show = normalized.show
    if (!show.id) {
      show.id = generateID(show.source === "announcement" ? "announce" : "show")
    }
    show.createdAt = trimText(show.createdAt) || new Date().toISOString()
    show.updatedAt = trimText(show.updatedAt) || show.createdAt
    state.push(show)
    return { ok: true, show: clone(show) }
  }

  const seeds = Array.isArray(seedShows) ? seedShows : []
  seeds.forEach((seed) => {
    const result = addNormalizedShow(seed, "seed")
    if (!result.ok) {
      throw new Error(result.error)
    }
  })

  function listAll() {
    return state.slice().sort((a, b) => compareDateISO(a.dateISO, b.dateISO)).map(clone)
  }

  function listUpcoming(limit) {
    const max = Number.isFinite(Number(limit)) && Number(limit) > 0 ? Number(limit) : state.length
    return state
      .filter((show) => show.status === "confirmed" && !isPastDate(show.dateISO))
      .sort((a, b) => compareDateISO(a.dateISO, b.dateISO) || trimText(a.artist).localeCompare(trimText(b.artist)))
      .slice(0, max)
      .map(clone)
  }

  function getShow(id) {
    const needle = trimText(id)
    if (!needle) {
      return null
    }
    const found = state.find((show) => show.id === needle)
    return found ? clone(found) : null
  }

  function addShow(raw) {
    const result = addNormalizedShow(raw, "announcement")
    if (!result.ok) {
      return result
    }
    return { ok: true, show: result.show }
  }

  function updateShow(id, patch) {
    const needle = trimText(id)
    if (!needle) {
      return { ok: false, error: "Show ID is required." }
    }
    const index = state.findIndex((show) => show.id === needle)
    if (index === -1) {
      return { ok: false, error: `No show found for ID ${JSON.stringify(needle)}.` }
    }
    const next = { ...state[index], ...patch, updatedAt: new Date().toISOString() }
    const normalized = normalizeShow(next, config)
    if (!normalized.ok) {
      return normalized
    }
    normalized.show.id = state[index].id
    state[index] = normalized.show
    return { ok: true, show: clone(normalized.show) }
  }

  function findByDiscordMessage(channelId, messageId) {
    const channelNeedle = trimText(channelId)
    const messageNeedle = trimText(messageId)
    if (!messageNeedle) {
      return null
    }
    const found = state.find((show) => trimText(show.discordMessageId) === messageNeedle && (!channelNeedle || trimText(show.discordChannelId) === channelNeedle))
    return found ? clone(found) : null
  }

  function archiveByDiscordMessage(channelId, messageId) {
    const found = findByDiscordMessage(channelId, messageId)
    if (!found) {
      return { ok: false, error: "No show found for the pinned Discord message." }
    }
    return updateShow(found.id, { status: "archived" })
  }

  function attachDiscordMessage(id, channelId, messageId) {
    return updateShow(id, {
      discordChannelId: trimText(channelId),
      discordMessageId: trimText(messageId),
    })
  }

  function cancelShow(id) {
    return updateShow(id, { status: "cancelled" })
  }

  function archiveShow(id) {
    return updateShow(id, { status: "archived" })
  }

  function listPast(limit) {
    const max = Number.isFinite(Number(limit)) && Number(limit) > 0 ? Number(limit) : 5
    return state
      .filter((show) => show.status === "archived" || isPastDate(show.dateISO))
      .sort((a, b) => compareDateISO(b.dateISO, a.dateISO) || trimText(a.artist).localeCompare(trimText(b.artist)))
      .slice(0, max)
      .map(clone)
  }

  function findExpiredShows(referenceDate) {
    const today = referenceDate instanceof Date ? todayISO(referenceDate) : todayISO()
    return state.filter((show) => show.status === "confirmed" && String(show.dateISO || "") < today).map(clone)
  }

  function markExpiredArchived(referenceDate) {
    const expired = findExpiredShows(referenceDate)
    expired.forEach((show) => {
      const index = state.findIndex((entry) => entry.id === show.id)
      if (index !== -1) {
        state[index] = { ...state[index], status: "archived", updatedAt: new Date().toISOString() }
      }
    })
    return expired
  }

  return {
    listAll,
    listUpcoming,
    listPast,
    getShow,
    addShow,
    updateShow,
    attachDiscordMessage,
    cancelShow,
    archiveShow,
    findByDiscordMessage,
    archiveByDiscordMessage,
    findExpiredShows,
    markExpiredArchived,
    _state: state,
  }
}

function loadSeedShows() {
  const seedShows = require("../shows.json")
  if (Array.isArray(seedShows)) {
    return seedShows
  }
  if (seedShows && Array.isArray(seedShows.default)) {
    return seedShows.default
  }
  return []
}

module.exports = {
  createShowCatalog,
  loadSeedShows,
  normalizeShow,
}
