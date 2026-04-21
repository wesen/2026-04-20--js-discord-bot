const database = require("database")

const DEFAULT_DB_PATH = "./examples/discord-bots/knowledge-base/data/knowledge.sqlite"
const DEFAULT_REVIEW_LIMIT = 5
const DEFAULT_SEARCH_LIMIT = 8
const SEED_ENTRIES = [
  {
    title: "How to add knowledge",
    summary: "Use /teach or /remember to save a useful explanation as a draft entry.",
    body: "Open the teach modal, add a title, summary, body, and tags, then submit. The bot will save a draft entry that the channel can review later.",
    tags: ["onboarding", "knowledge", "capture"],
    aliases: ["teach the bot", "remember this"],
    source: {
      kind: "seed",
      note: "Onboarding entry created by the bot",
    },
  },
  {
    title: "How to search knowledge",
    summary: "Use /ask or /kb-search to look up canonical answers and draft notes.",
    body: "The bot searches verified entries first, then review and draft entries. Use /kb-article to open one entry by id or slug.",
    tags: ["onboarding", "search", "retrieval"],
    aliases: ["kb search", "article lookup"],
    source: {
      kind: "seed",
      note: "Onboarding entry created by the bot",
    },
  },
  {
    title: "How to review knowledge",
    summary: "Use /kb-review, /kb-verify, /kb-stale, and /kb-reject to curate the queue.",
    body: "Drafts stay visible until a reviewer promotes them to verified, marks them stale, or rejects them with a reason.",
    tags: ["onboarding", "review", "curation"],
    aliases: ["moderate knowledge", "queue review"],
    source: {
      kind: "seed",
      note: "Onboarding entry created by the bot",
    },
  },
]

function createKnowledgeStore() {
  let configuredPath = ""
  let initialized = false

  function ensure(config) {
    const dbPath = configValue(config, ["dbPath", "db_path"], DEFAULT_DB_PATH)
    if (initialized && configuredPath === dbPath) {
      return
    }

    try {
      database.close()
    } catch (err) {
      // Ignored on purpose: the module may not have been configured yet.
    }

    database.configure("sqlite3", dbPath)
    configuredPath = dbPath
    initialized = true
    ensureSchema()
    ensureSeedData(config)
  }

  function ensureSchema() {
    database.exec(`
      CREATE TABLE IF NOT EXISTS knowledge_entries (
        id TEXT PRIMARY KEY,
        slug TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        summary TEXT NOT NULL,
        body TEXT NOT NULL,
        status TEXT NOT NULL,
        confidence REAL NOT NULL DEFAULT 0,
        tags_json TEXT NOT NULL DEFAULT '[]',
        aliases_json TEXT NOT NULL DEFAULT '[]',
        source_kind TEXT NOT NULL DEFAULT 'capture',
        source_guild_id TEXT NOT NULL DEFAULT '',
        source_channel_id TEXT NOT NULL DEFAULT '',
        source_message_id TEXT NOT NULL DEFAULT '',
        source_author_id TEXT NOT NULL DEFAULT '',
        source_jump_url TEXT NOT NULL DEFAULT '',
        source_note TEXT NOT NULL DEFAULT '',
        reviewed_by TEXT NOT NULL DEFAULT '',
        review_note TEXT NOT NULL DEFAULT '',
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        verified_at TEXT,
        stale_at TEXT,
        rejected_at TEXT
      )
    `)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_knowledge_entries_status ON knowledge_entries(status)`)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_knowledge_entries_updated_at ON knowledge_entries(updated_at DESC)`)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_knowledge_entries_source_message ON knowledge_entries(source_channel_id, source_message_id)`)
    database.exec(`
      CREATE TABLE IF NOT EXISTS knowledge_entry_versions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        entry_id TEXT NOT NULL,
        version INTEGER NOT NULL,
        title TEXT NOT NULL,
        summary TEXT NOT NULL,
        body TEXT NOT NULL,
        status TEXT NOT NULL,
        edited_by TEXT NOT NULL DEFAULT '',
        note TEXT NOT NULL DEFAULT '',
        created_at TEXT NOT NULL,
        FOREIGN KEY(entry_id) REFERENCES knowledge_entries(id) ON DELETE CASCADE
      )
    `)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_knowledge_entry_versions_entry_id ON knowledge_entry_versions(entry_id, version DESC)`)
  }

  function ensureSeedData(config) {
    if (!configValue(config, ["seedEntries", "seed_entries"], true)) {
      return
    }
    const count = querySingle(`SELECT COUNT(1) AS count FROM knowledge_entries`, [])
    if (Number(count.count || 0) > 0) {
      return
    }
    for (const seed of SEED_ENTRIES) {
      insertEntry({
        title: seed.title,
        summary: seed.summary,
        body: seed.body,
        tags: seed.tags,
        aliases: seed.aliases,
        status: "verified",
        confidence: 1,
        source: seed.source,
        reviewedBy: "system",
        reviewNote: "Seed entry",
      })
    }
  }

  function saveCandidate(config, candidate) {
    ensure(config)
    if (!candidate || !trimText(candidate.title) || !trimText(candidate.body)) {
      return null
    }
    const source = candidate.source || {}
    const existing = findBySource(source.channelId || source.sourceChannelId, source.messageId || source.sourceMessageId)
    if (existing) {
      return existing
    }
    return insertEntry({
      title: candidate.title,
      summary: candidate.summary || candidate.body,
      body: candidate.body,
      tags: candidate.tags || [],
      aliases: candidate.aliases || [],
      confidence: typeof candidate.confidence === "number" ? candidate.confidence : 0,
      status: candidate.status || "draft",
      source,
      reviewedBy: candidate.reviewedBy || "",
      reviewNote: candidate.reviewNote || "",
    })
  }

  function saveManualEntry(config, payload) {
    ensure(config)
    const entry = insertEntry({
      title: payload.title,
      summary: payload.summary,
      body: payload.body,
      tags: payload.tags || [],
      aliases: payload.aliases || [],
      confidence: typeof payload.confidence === "number" ? payload.confidence : 0.8,
      status: payload.status || "draft",
      source: payload.source || {},
      reviewedBy: payload.reviewedBy || "",
      reviewNote: payload.reviewNote || "",
    })
    return entry
  }

  function insertEntry(entry) {
    const normalized = normalizeEntry(entry)
    database.exec(
      `
        INSERT INTO knowledge_entries (
          id, slug, title, summary, body, status, confidence,
          tags_json, aliases_json,
          source_kind, source_guild_id, source_channel_id, source_message_id, source_author_id, source_jump_url, source_note,
          reviewed_by, review_note,
          created_at, updated_at, verified_at, stale_at, rejected_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `,
      normalized.id,
      normalized.slug,
      normalized.title,
      normalized.summary,
      normalized.body,
      normalized.status,
      normalized.confidence,
      json(normalized.tags),
      json(normalized.aliases),
      normalized.source.kind,
      normalized.source.guildId,
      normalized.source.channelId,
      normalized.source.messageId,
      normalized.source.authorId,
      normalized.source.jumpUrl,
      normalized.source.note,
      normalized.reviewedBy,
      normalized.reviewNote,
      normalized.createdAt,
      normalized.updatedAt,
      normalized.verifiedAt,
      normalized.staleAt,
      normalized.rejectedAt,
    )
    appendVersion(normalized, 1, normalized.reviewedBy, normalized.reviewNote)
    return getEntry(normalized.id)
  }

  function updateEntry(config, identifier, patch) {
    ensure(config)
    const current = getEntry(identifier)
    if (!current) {
      return null
    }
    const next = normalizeEntry({
      ...current,
      ...patch,
      id: current.id,
      slug: patch && patch.slug ? patch.slug : current.slug,
      source: patch && patch.source ? patch.source : current.source,
      createdAt: current.createdAt,
      updatedAt: nowISO(),
      version: current.version,
      reviewedBy: patch && patch.reviewedBy ? patch.reviewedBy : current.reviewedBy,
      reviewNote: patch && patch.reviewNote ? patch.reviewNote : current.reviewNote,
    })
    database.exec(
      `
        UPDATE knowledge_entries
        SET slug = ?, title = ?, summary = ?, body = ?, status = ?, confidence = ?, tags_json = ?, aliases_json = ?,
            source_kind = ?, source_guild_id = ?, source_channel_id = ?, source_message_id = ?, source_author_id = ?,
            source_jump_url = ?, source_note = ?, reviewed_by = ?, review_note = ?, updated_at = ?, verified_at = ?, stale_at = ?, rejected_at = ?
        WHERE id = ?
      `,
      next.slug,
      next.title,
      next.summary,
      next.body,
      next.status,
      next.confidence,
      json(next.tags),
      json(next.aliases),
      next.source.kind,
      next.source.guildId,
      next.source.channelId,
      next.source.messageId,
      next.source.authorId,
      next.source.jumpUrl,
      next.source.note,
      next.reviewedBy,
      next.reviewNote,
      next.updatedAt,
      next.verifiedAt,
      next.staleAt,
      next.rejectedAt,
      current.id,
    )
    const version = Number(current.version || 0) + 1
    appendVersion(next, version, next.reviewedBy, next.reviewNote)
    return getEntry(current.id)
  }

  function setStatus(config, identifier, status, reviewedBy, reviewNote) {
    const patch = { status, reviewedBy: reviewedBy || "", reviewNote: reviewNote || "" }
    if (status === "verified") {
      patch.verifiedAt = nowISO()
      patch.staleAt = null
      patch.rejectedAt = null
    } else if (status === "stale") {
      patch.staleAt = nowISO()
      patch.verifiedAt = null
      patch.rejectedAt = null
    } else if (status === "rejected") {
      patch.rejectedAt = nowISO()
      patch.verifiedAt = null
      patch.staleAt = null
    }
    return updateEntry(config, identifier, patch)
  }

  function getEntry(identifier) {
    const text = trimText(identifier)
    if (!text) {
      return null
    }
    const row = querySingle(
      `SELECT * FROM knowledge_entries WHERE id = ? OR slug = ? LIMIT 1`,
      [text, text],
    )
    return row ? decodeEntry(row) : null
  }

  function findBySource(channelId, messageId) {
    channelId = trimText(channelId)
    messageId = trimText(messageId)
    if (!channelId || !messageId) {
      return null
    }
    const row = querySingle(
      `SELECT * FROM knowledge_entries WHERE source_channel_id = ? AND source_message_id = ? LIMIT 1`,
      [channelId, messageId],
    )
    return row ? decodeEntry(row) : null
  }

  function listByStatus(config, status, limit) {
    ensure(config)
    const rows = database.query(
      `SELECT * FROM knowledge_entries WHERE status = ? ORDER BY updated_at DESC LIMIT ?`,
      trimText(status) || "draft",
      clampLimit(limit, DEFAULT_REVIEW_LIMIT),
    )
    return rowsToEntries(rows)
  }

  function recent(config, limit) {
    ensure(config)
    const rows = database.query(
      `SELECT * FROM knowledge_entries ORDER BY updated_at DESC LIMIT ?`,
      clampLimit(limit, DEFAULT_SEARCH_LIMIT),
    )
    return rowsToEntries(rows)
  }

  function search(config, query, limit) {
    ensure(config)
    const needle = trimText(query).toLowerCase()
    if (!needle) {
      return []
    }
    const rows = database.query(`SELECT * FROM knowledge_entries WHERE status != 'rejected' ORDER BY updated_at DESC LIMIT 250`)
    const entries = rowsToEntries(rows)
      .map((entry) => ({ entry, score: scoreEntry(entry, needle) }))
      .filter((item) => item.score > 0)
      .sort((a, b) => {
        if (b.score !== a.score) return b.score - a.score
        if (statusRank(a.entry.status) !== statusRank(b.entry.status)) return statusRank(a.entry.status) - statusRank(b.entry.status)
        return compareText(b.entry.updatedAt, a.entry.updatedAt)
      })
      .slice(0, clampLimit(limit, DEFAULT_SEARCH_LIMIT))
      .map((item) => item.entry)
    return entries
  }

  function listCounts(config) {
    ensure(config)
    const rows = database.query(`SELECT status, COUNT(1) AS count FROM knowledge_entries GROUP BY status`)
    const counts = { draft: 0, review: 0, verified: 0, stale: 0, rejected: 0 }
    for (const row of rows || []) {
      const status = trimText(row.status)
      if (status) {
        counts[status] = Number(row.count || 0)
      }
    }
    return counts
  }

  function appendVersion(entry, version, editedBy, note) {
    database.exec(
      `
        INSERT INTO knowledge_entry_versions (
          entry_id, version, title, summary, body, status, edited_by, note, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
      `,
      entry.id,
      version,
      entry.title,
      entry.summary,
      entry.body,
      entry.status,
      trimText(editedBy) || "",
      trimText(note) || "",
      nowISO(),
    )
  }

  function normalizeEntry(entry) {
    const createdAt = trimText(entry && entry.createdAt) || nowISO()
    const updatedAt = trimText(entry && entry.updatedAt) || createdAt
    const title = trimText(entry && entry.title) || "Untitled knowledge"
    const summary = trimText(entry && entry.summary) || trimText(entry && entry.body) || title
    const body = trimText(entry && entry.body) || summary
    const slug = uniqueSlug(slugify(entry && entry.slug ? entry.slug : title), entry && entry.id)
    const source = normalizeSource(entry && entry.source)
    return {
      id: trimText(entry && entry.id) || makeId(),
      slug,
      title,
      summary,
      body,
      status: normalizeStatus(entry && entry.status),
      confidence: normalizeConfidence(entry && entry.confidence),
      tags: normalizeTextList(entry && entry.tags),
      aliases: normalizeTextList(entry && entry.aliases),
      source,
      reviewedBy: trimText(entry && entry.reviewedBy) || "",
      reviewNote: trimText(entry && entry.reviewNote) || "",
      createdAt,
      updatedAt,
      verifiedAt: normalizeTimestamp(entry && entry.verifiedAt),
      staleAt: normalizeTimestamp(entry && entry.staleAt),
      rejectedAt: normalizeTimestamp(entry && entry.rejectedAt),
      version: Number(entry && entry.version || 0),
    }
  }

  function normalizeSource(source) {
    const raw = source || {}
    const guildId = trimText(raw.guildId || raw.guildID || raw.guild_id)
    const channelId = trimText(raw.channelId || raw.channelID || raw.channel_id)
    const messageId = trimText(raw.messageId || raw.messageID || raw.message_id)
    const authorId = trimText(raw.authorId || raw.authorID || raw.author_id)
    return {
      kind: trimText(raw.kind) || "capture",
      guildId,
      channelId,
      messageId,
      authorId,
      jumpUrl: trimText(raw.jumpUrl || raw.jumpURL || raw.jump_url),
      note: trimText(raw.note),
    }
  }

  function normalizeStatus(status) {
    const normalized = trimText(status).toLowerCase()
    switch (normalized) {
      case "draft":
      case "review":
      case "verified":
      case "stale":
      case "rejected":
        return normalized
      default:
        return "draft"
    }
  }

  function normalizeConfidence(value) {
    const n = Number(value)
    if (Number.isNaN(n)) {
      return 0
    }
    if (n < 0) return 0
    if (n > 1) return 1
    return n
  }

  function normalizeTextList(value) {
    if (Array.isArray(value)) {
      return value.map(trimText).filter(Boolean)
    }
    const text = trimText(value)
    if (!text) {
      return []
    }
    return text
      .split(/[;,\n]/g)
      .map((part) => trimText(part))
      .filter(Boolean)
  }

  function normalizeTimestamp(value) {
    const text = trimText(value)
    return text || null
  }

  function decodeEntry(row) {
    return {
      id: trimText(row.id),
      slug: trimText(row.slug),
      title: trimText(row.title),
      summary: trimText(row.summary),
      body: trimText(row.body),
      status: trimText(row.status),
      confidence: Number(row.confidence || 0),
      tags: parseJSONList(row.tags_json),
      aliases: parseJSONList(row.aliases_json),
      source: {
        kind: trimText(row.source_kind) || "capture",
        guildId: trimText(row.source_guild_id),
        channelId: trimText(row.source_channel_id),
        messageId: trimText(row.source_message_id),
        authorId: trimText(row.source_author_id),
        jumpUrl: trimText(row.source_jump_url),
        note: trimText(row.source_note),
      },
      reviewedBy: trimText(row.reviewed_by),
      reviewNote: trimText(row.review_note),
      createdAt: trimText(row.created_at),
      updatedAt: trimText(row.updated_at),
      verifiedAt: trimText(row.verified_at),
      staleAt: trimText(row.stale_at),
      rejectedAt: trimText(row.rejected_at),
      version: versionFor(row.id),
    }
  }

  function versionFor(entryId) {
    const row = querySingle(
      `SELECT version FROM knowledge_entry_versions WHERE entry_id = ? ORDER BY version DESC LIMIT 1`,
      [entryId],
    )
    return row ? Number(row.version || 0) : 0
  }

  function rowsToEntries(rows) {
    return (rows || []).map((row) => decodeEntry(row))
  }

  function querySingle(sql, args) {
    const rows = database.query(sql, ...(args || []))
    if (!rows || rows.length === 0) {
      return null
    }
    return rows[0]
  }

  function clampLimit(limit, fallback) {
    const n = Number(limit)
    if (!Number.isFinite(n) || n <= 0) {
      return fallback
    }
    return Math.min(Math.floor(n), 50)
  }

  function scoreEntry(entry, needle) {
    const searchable = [
      entry.title,
      entry.summary,
      entry.body,
      entry.slug,
      entry.tags.join(" "),
      entry.aliases.join(" "),
    ].join(" ").toLowerCase()
    if (!searchable.includes(needle)) {
      return 0
    }
    let score = 0
    if (entry.title.toLowerCase().includes(needle)) score += 5
    if (entry.summary.toLowerCase().includes(needle)) score += 4
    if (entry.body.toLowerCase().includes(needle)) score += 2
    if (entry.slug.toLowerCase().includes(needle)) score += 3
    if (entry.tags.some((tag) => tag.toLowerCase().includes(needle))) score += 2
    if (entry.aliases.some((alias) => alias.toLowerCase().includes(needle))) score += 2
    if (entry.status === "verified") score += 1.5
    if (entry.status === "review") score += 0.5
    score += Math.min(entry.confidence || 0, 1)
    return score
  }

  function statusRank(status) {
    switch (status) {
      case "verified": return 0
      case "review": return 1
      case "draft": return 2
      case "stale": return 3
      default: return 4
    }
  }

  function compareText(a, b) {
    return String(a || "").localeCompare(String(b || ""))
  }

  function uniqueSlug(baseSlug, entryId) {
    const normalized = slugify(baseSlug || "knowledge")
    const currentId = trimText(entryId)
    let slug = normalized || "knowledge"
    let suffix = 2
    while (true) {
      const row = querySingle(`SELECT id FROM knowledge_entries WHERE slug = ? LIMIT 1`, [slug])
      if (!row || trimText(row.id) === currentId) {
        return slug
      }
      slug = `${normalized}-${suffix}`
      suffix += 1
    }
  }

  function json(value) {
    return JSON.stringify(value || [])
  }

  return {
    ensure,
    saveCandidate,
    saveManualEntry,
    updateEntry,
    setStatus,
    getEntry,
    findBySource,
    listByStatus,
    recent,
    search,
    listCounts,
    clampLimit,
    configPath: () => configuredPath,
  }
}

function configValue(config, names, fallback) {
  for (const name of names || []) {
    if (!name) continue
    if (config && Object.prototype.hasOwnProperty.call(config, name) && config[name] !== undefined && config[name] !== null && String(config[name]).trim() !== "") {
      return config[name]
    }
  }
  return fallback
}

function trimText(value) {
  if (value === undefined || value === null) {
    return ""
  }
  return String(value).trim()
}

function nowISO() {
  return new Date().toISOString()
}

function makeId() {
  return `kb_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`
}

function slugify(value) {
  const text = trimText(value).toLowerCase()
  const slug = text
    .replace(/['"`]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
  return slug || "knowledge"
}

function parseJSONList(raw) {
  if (Array.isArray(raw)) {
    return raw.map(trimText).filter(Boolean)
  }
  const text = trimText(raw)
  if (!text) {
    return []
  }
  try {
    const parsed = JSON.parse(text)
    if (Array.isArray(parsed)) {
      return parsed.map(trimText).filter(Boolean)
    }
  } catch (err) {
    // fall through
  }
  return []
}

module.exports = { createKnowledgeStore, DEFAULT_DB_PATH }
