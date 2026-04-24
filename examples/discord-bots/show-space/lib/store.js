const database = require("database")
const { loadSeedShows, normalizeShow } = require("./shows")
const { formatDisplayDate, localDateISO, todayISO, compareDateISO } = require("./dates")

function trimText(value) {
  return String(value || "").trim()
}

function boolValue(value, fallback) {
  if (typeof value === "boolean") {
    return value
  }
  if (value === undefined || value === null || value === "") {
    return Boolean(fallback)
  }
  const text = String(value).trim().toLowerCase()
  if (["1", "true", "yes", "y", "on"].includes(text)) {
    return true
  }
  if (["0", "false", "no", "n", "off"].includes(text)) {
    return false
  }
  return Boolean(fallback)
}

function createShowStore() {
  let configuredPath = ""
  let initialized = false

  function ensure(config) {
    const dbPath = trimText(config && (config.dbPath || config.db_path))
    if (!dbPath) {
      return false
    }
    if (initialized && configuredPath === dbPath) {
      return true
    }

    try {
      database.close()
    } catch (err) {
      // ignored; the module may not have been initialized yet
    }

    database.configure("sqlite3", dbPath)
    configuredPath = dbPath
    initialized = true
    ensureSchema()
    seedIfNeeded(config)
    return true
  }

  function ensureSchema() {
    database.exec(`
      CREATE TABLE IF NOT EXISTS shows (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        artist TEXT NOT NULL,
        date TEXT NOT NULL,
        doors_time TEXT NOT NULL DEFAULT '',
        age TEXT NOT NULL DEFAULT '',
        price TEXT NOT NULL DEFAULT '',
        notes TEXT NOT NULL DEFAULT '',
        status TEXT NOT NULL DEFAULT 'confirmed',
        discord_message_id TEXT NOT NULL DEFAULT '',
        discord_channel_id TEXT NOT NULL DEFAULT '',
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL
      )
    `)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_shows_date ON shows(date ASC, id ASC)`)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_shows_status ON shows(status, date ASC)`)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_shows_message ON shows(discord_channel_id, discord_message_id)`)
  }

  function seedIfNeeded(config) {
    const seedFlag = config && (config.seedFromJson !== undefined ? config.seedFromJson : config.seed_from_json)
    if (!boolValue(seedFlag, true)) {
      return
    }
    const count = querySingle(`SELECT COUNT(1) AS count FROM shows`, [])
    if (Number(count.count || 0) > 0) {
      return
    }
    const seeds = loadSeedShows()
    for (const seed of seeds) {
      const normalized = normalizeShow(seed, { timeZone: config && config.timeZone, referenceDate: new Date() })
      if (!normalized.ok) {
        continue
      }
      insertShowRecord(normalized.show, { status: normalized.show.status || "confirmed" })
    }
  }

  function querySingle(sql, args) {
    const rows = database.query(sql, ...(args || []))
    if (!Array.isArray(rows) || rows.length === 0) {
      return {}
    }
    return rows[0] || {}
  }

  function showFromRow(row, config) {
    if (!row) {
      return null
    }
    const dateISO = trimText(row.date)
    const date = dateISO ? new Date(`${dateISO}T00:00:00`) : new Date()
    return {
      id: row.id,
      artist: trimText(row.artist),
      dateISO,
      displayDate: formatDisplayDate(date, config && config.timeZone),
      doorsTime: trimText(row.doors_time),
      ageRestriction: trimText(row.age),
      price: trimText(row.price),
      notes: trimText(row.notes),
      status: trimText(row.status) || "confirmed",
      discordChannelId: trimText(row.discord_channel_id),
      discordMessageId: trimText(row.discord_message_id),
      source: "database",
      createdAt: trimText(row.created_at),
      updatedAt: trimText(row.updated_at),
    }
  }

  function insertShowRecord(show, patch) {
    const now = new Date().toISOString()
    const current = { ...show, ...(patch || {}), createdAt: trimText(show.createdAt) || now, updatedAt: now }
    database.exec(
      `
        INSERT INTO shows (
          artist, date, doors_time, age, price, notes, status, discord_message_id, discord_channel_id, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `,
      trimText(current.artist),
      trimText(current.dateISO),
      trimText(current.doorsTime),
      trimText(current.ageRestriction),
      trimText(current.price),
      trimText(current.notes),
      trimText(current.status) || "confirmed",
      trimText(current.discordMessageId),
      trimText(current.discordChannelId),
      current.createdAt,
      current.updatedAt,
    )
    const inserted = querySingle(`SELECT last_insert_rowid() AS id`, [])
    return getShowByRowID(inserted.id)
  }

  function getShowByRowID(id) {
    const rows = database.query(`SELECT * FROM shows WHERE id = ? LIMIT 1`, id)
    if (!Array.isArray(rows) || rows.length === 0) {
      return null
    }
    return rows[0] ? showFromRow(rows[0], {}) : null
  }

  function updateShowRecord(config, id, patch) {
    ensure(config)
    const current = getShow(config, id)
    if (!current) {
      return { ok: false, error: `No show found for ID ${JSON.stringify(String(id))}.` }
    }
    const next = { ...current, ...(patch || {}), id: current.id }
    const normalized = normalizeShow({
      id: current.id,
      artist: next.artist,
      dateISO: next.dateISO,
      doors_time: next.doorsTime,
      age: next.ageRestriction,
      price: next.price,
      notes: next.notes,
      status: next.status,
      discord_channel_id: next.discordChannelId,
      discord_message_id: next.discordMessageId,
    }, { timeZone: config && config.timeZone, referenceDate: new Date() })
    if (!normalized.ok) {
      return normalized
    }
    database.exec(
      `
        UPDATE shows
        SET artist = ?, date = ?, doors_time = ?, age = ?, price = ?, notes = ?, status = ?, discord_message_id = ?, discord_channel_id = ?, updated_at = ?
        WHERE id = ?
      `,
      trimText(normalized.show.artist),
      trimText(normalized.show.dateISO),
      trimText(normalized.show.doorsTime),
      trimText(normalized.show.ageRestriction),
      trimText(normalized.show.price),
      trimText(normalized.show.notes),
      trimText(normalized.show.status) || "confirmed",
      trimText(normalized.show.discordMessageId),
      trimText(normalized.show.discordChannelId),
      new Date().toISOString(),
      current.id,
    )
    return { ok: true, show: getShow(config, current.id) }
  }

  function listUpcoming(config, limit) {
    ensure(config)
    const today = todayISO(new Date())
    const max = Number.isFinite(Number(limit)) && Number(limit) > 0 ? Number(limit) : 25
    const rows = database.query(
      `SELECT * FROM shows WHERE status = 'confirmed' AND date >= ? ORDER BY date ASC, id ASC LIMIT ?`,
      today,
      max,
    )
    return (Array.isArray(rows) ? rows : []).map((row) => showFromRow(row, config))
  }

  function listPast(config, limit) {
    ensure(config)
    const today = todayISO(new Date())
    const max = Number.isFinite(Number(limit)) && Number(limit) > 0 ? Number(limit) : 5
    const rows = database.query(
      `
        SELECT * FROM shows
        WHERE status = 'archived' OR date < ?
        ORDER BY date DESC, id DESC
        LIMIT ?
      `,
      today,
      max,
    )
    return (Array.isArray(rows) ? rows : []).map((row) => showFromRow(row, config))
  }

  function getShow(config, id) {
    ensure(config)
    const numericID = Number(id)
    if (!Number.isFinite(numericID)) {
      return null
    }
    const row = querySingle(`SELECT * FROM shows WHERE id = ? LIMIT 1`, [numericID])
    return row && row.id ? showFromRow(row, config) : null
  }

  function createShow(config, raw) {
    ensure(config)
    const normalized = normalizeShow(raw, { timeZone: config && config.timeZone, referenceDate: new Date() })
    if (!normalized.ok) {
      return normalized
    }
    const inserted = insertShowRecord(normalized.show, { status: normalized.show.status || "confirmed" })
    if (!inserted) {
      return { ok: false, error: "Failed to insert show." }
    }
    return { ok: true, show: inserted }
  }

  function attachDiscordMessage(config, id, channelId, messageId) {
    return updateShowRecord(config, id, {
      discordChannelId: trimText(channelId),
      discordMessageId: trimText(messageId),
    })
  }

  function cancelShow(config, id) {
    return updateShowRecord(config, id, { status: "cancelled" })
  }

  function archiveShow(config, id) {
    return updateShowRecord(config, id, { status: "archived" })
  }

  function archiveByDiscordMessage(config, channelId, messageId) {
    ensure(config)
    const row = querySingle(
      `SELECT * FROM shows WHERE discord_channel_id = ? AND discord_message_id = ? LIMIT 1`,
      [trimText(channelId), trimText(messageId)],
    )
    if (!row || !row.id) {
      return { ok: false, error: "No show found for the pinned Discord message." }
    }
    return archiveShow(config, row.id)
  }

  function findExpiredShows(config, referenceDate) {
    ensure(config)
    const today = localDateISO(referenceDate instanceof Date ? referenceDate : new Date())
    const rows = database.query(
      `SELECT * FROM shows WHERE status = 'confirmed' AND date < ? ORDER BY date ASC, id ASC`,
      today,
    )
    return (Array.isArray(rows) ? rows : []).map((row) => showFromRow(row, config))
  }

  function archiveExpiredShows(config, referenceDate) {
    const expired = findExpiredShows(config, referenceDate)
    const archived = []
    for (const show of expired) {
      const result = archiveShow(config, show.id)
      if (result && result.ok) {
        archived.push(result.show)
      }
    }
    return archived
  }

  function countShows(config) {
    ensure(config)
    const row = querySingle(`SELECT COUNT(1) AS count FROM shows`, [])
    return Number(row.count || 0)
  }

  return {
    ensure,
    isEnabled() {
      return initialized
    },
    countShows,
    listUpcoming,
    listPast,
    getShow,
    createShow,
    attachDiscordMessage,
    cancelShow,
    archiveShow,
    archiveByDiscordMessage,
    findExpiredShows,
    archiveExpiredShows,
  }
}

module.exports = {
  createShowStore,
}
