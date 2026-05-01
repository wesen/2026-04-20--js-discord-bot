const database = require("database")

const DEFAULT_DB_PATH = "./examples/discord-bots/custom-kb/data/custom-kb.sqlite"

function createLinkStore() {
  let configuredPath = ""
  let initialized = false

  function ensure(config) {
    const dbPath = configValue(config, ["dbPath", "db_path"], DEFAULT_DB_PATH)
    if (initialized && configuredPath === dbPath) return
    try { database.close() } catch (err) {}
    database.configure("sqlite3", dbPath)
    configuredPath = dbPath
    initialized = true
    ensureSchema()
  }

  function ensureSchema() {
    database.exec(`
      CREATE TABLE IF NOT EXISTS kb_links (
        id TEXT PRIMARY KEY,
        url TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        summary TEXT NOT NULL DEFAULT '',
        tags_json TEXT NOT NULL DEFAULT '[]',
        added_by TEXT NOT NULL DEFAULT '',
        guild_id TEXT NOT NULL DEFAULT '',
        channel_id TEXT NOT NULL DEFAULT '',
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL
      )
    `)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_kb_links_updated_at ON kb_links(updated_at DESC)`)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_kb_links_title ON kb_links(title)`)
  }

  function addLink(config, payload) {
    ensure(config)
    const normalized = normalizeLink(payload)
    const existing = findByUrl(config, normalized.url)
    if (existing) {
      database.exec(
        `UPDATE kb_links SET title = ?, summary = ?, tags_json = ?, added_by = ?, guild_id = ?, channel_id = ?, updated_at = ? WHERE url = ?`,
        normalized.title, normalized.summary, json(normalized.tags), normalized.addedBy, normalized.guildId, normalized.channelId, now(), normalized.url
      )
      return getLink(config, existing.id)
    }
    database.exec(
      `INSERT INTO kb_links (id, url, title, summary, tags_json, added_by, guild_id, channel_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      normalized.id, normalized.url, normalized.title, normalized.summary, json(normalized.tags), normalized.addedBy, normalized.guildId, normalized.channelId, normalized.createdAt, normalized.updatedAt
    )
    return getLink(config, normalized.id)
  }

  function search(config, query, limit) {
    ensure(config)
    const n = clampLimit(limit, 10)
    const q = trimText(query)
    if (!q) return listRecent(config, n)
    const like = `%${q.toLowerCase()}%`
    return mapRows(database.query(
      `SELECT * FROM kb_links
       WHERE lower(title) LIKE ? OR lower(summary) LIKE ? OR lower(url) LIKE ? OR lower(tags_json) LIKE ?
       ORDER BY updated_at DESC LIMIT ?`,
      like, like, like, like, n
    ))
  }

  function listRecent(config, limit) {
    ensure(config)
    return mapRows(database.query(`SELECT * FROM kb_links ORDER BY updated_at DESC LIMIT ?`, clampLimit(limit, 10)))
  }

  function getLink(config, id) {
    ensure(config)
    const rows = database.query(`SELECT * FROM kb_links WHERE id = ? LIMIT 1`, trimText(id))
    return rows.length ? mapRow(rows[0]) : null
  }

  function findByUrl(config, url) {
    ensure(config)
    const rows = database.query(`SELECT * FROM kb_links WHERE url = ? LIMIT 1`, trimText(url))
    return rows.length ? mapRow(rows[0]) : null
  }

  function count(config) {
    ensure(config)
    const rows = database.query(`SELECT COUNT(1) AS count FROM kb_links`)
    return Number(rows[0] && rows[0].count || 0)
  }

  return { ensure: ensure, addLink: addLink, search: search, listRecent: listRecent, getLink: getLink, findByUrl: findByUrl, count: count, configPath: function() { return configuredPath } }
}

function normalizeLink(payload) {
  const createdAt = now()
  const url = normalizeUrl(payload.url)
  if (!url) throw new Error("URL is required")
  return {
    id: payload.id || makeID("link"),
    url,
    title: trimText(payload.title) || url,
    summary: trimText(payload.summary),
    tags: normalizeTags(payload.tags),
    addedBy: trimText(payload.addedBy),
    guildId: trimText(payload.guildId),
    channelId: trimText(payload.channelId),
    createdAt,
    updatedAt: createdAt,
  }
}

function mapRows(rows) { return (rows || []).map(mapRow) }
function mapRow(row) {
  return {
    id: String(row.id || ""),
    url: String(row.url || ""),
    title: String(row.title || ""),
    summary: String(row.summary || ""),
    tags: parseJSON(row.tags_json, []),
    addedBy: String(row.added_by || ""),
    guildId: String(row.guild_id || ""),
    channelId: String(row.channel_id || ""),
    createdAt: String(row.created_at || ""),
    updatedAt: String(row.updated_at || ""),
  }
}
function normalizeUrl(value) {
  const url = trimText(value)
  if (!url) return ""
  if (/^https?:\/\//i.test(url)) return url
  return `https://${url}`
}
function normalizeTags(value) {
  if (Array.isArray(value)) return value.map(trimText).filter(Boolean)
  return trimText(value).split(/[,#]/).map(trimText).filter(Boolean)
}
function parseJSON(value, fallback) { try { return JSON.parse(value || "") } catch (err) { return fallback } }
function json(value) { return JSON.stringify(value || []) }
function trimText(value) { return String(value == null ? "" : value).trim() }
function now() { return new Date().toISOString() }
function makeID(prefix) { return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}` }
function clampLimit(value, fallback) { const n = Number(value || fallback); return Math.max(1, Math.min(25, Number.isFinite(n) ? n : fallback)) }
function configValue(config, keys, fallback) { for (const k of keys) if (config && config[k] != null && String(config[k]).trim()) return String(config[k]).trim(); return fallback }

module.exports = { createLinkStore: createLinkStore }
