const database = require("database")
const { allSeeds } = require("./seeds")
const { jsonString, safeJsonParse, trimText } = require("./schema")

function nowISO() {
  return new Date().toISOString()
}

function createId(prefix) {
  return `${prefix}_${Date.now().toString(36)}_${Math.floor(Math.random() * 100000).toString(36)}`
}

function createStore() {
  let configuredPath = ""
  let initialized = false

  function ensure(config) {
    const dbPath = trimText(config && (config.sessionDbPath || config.session_db_path)) || "./examples/discord-bots/adventure/data/adventure.sqlite"
    if (initialized && configuredPath === dbPath) return true
    try { database.close() } catch (_) {}
    database.configure("sqlite3", dbPath)
    configuredPath = dbPath
    initialized = true
    ensureSchema()
    seedDefaults()
    return true
  }

  function ensureSchema() {
    database.exec(`CREATE TABLE IF NOT EXISTS adventure_seeds (
      id TEXT PRIMARY KEY,
      title TEXT NOT NULL,
      genre TEXT NOT NULL DEFAULT '',
      tone TEXT NOT NULL DEFAULT '',
      initial_stats_json TEXT NOT NULL DEFAULT '{}',
      inventory_vocab_json TEXT NOT NULL DEFAULT '[]',
      flag_vocab_json TEXT NOT NULL DEFAULT '[]',
      constraints_json TEXT NOT NULL DEFAULT '{}',
      opening_prompt TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    )`)
    database.exec(`CREATE TABLE IF NOT EXISTS adventure_sessions (
      id TEXT PRIMARY KEY,
      adventure_id TEXT NOT NULL,
      owner_user_id TEXT NOT NULL,
      guild_id TEXT NOT NULL DEFAULT '',
      channel_id TEXT NOT NULL DEFAULT '',
      thread_id TEXT NOT NULL DEFAULT '',
      mode TEXT NOT NULL DEFAULT 'solo',
      turn INTEGER NOT NULL DEFAULT 0,
      current_scene_id TEXT NOT NULL DEFAULT '',
      stats_json TEXT NOT NULL DEFAULT '{}',
      inventory_json TEXT NOT NULL DEFAULT '[]',
      flags_json TEXT NOT NULL DEFAULT '{}',
      status TEXT NOT NULL DEFAULT 'active',
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    )`)
    database.exec(`CREATE TABLE IF NOT EXISTS adventure_scenes (
      id TEXT PRIMARY KEY,
      session_id TEXT NOT NULL,
      turn INTEGER NOT NULL,
      title TEXT NOT NULL,
      ascii_art TEXT NOT NULL DEFAULT '',
      narration TEXT NOT NULL DEFAULT '',
      engine_notes_json TEXT NOT NULL DEFAULT '{}',
      raw_patch_json TEXT NOT NULL DEFAULT '{}',
      stats_json TEXT NOT NULL DEFAULT '{}',
      inventory_json TEXT NOT NULL DEFAULT '[]',
      flags_json TEXT NOT NULL DEFAULT '{}',
      created_at TEXT NOT NULL
    )`)
    ensureColumn("adventure_scenes", "stats_json", "TEXT NOT NULL DEFAULT '{}'")
    ensureColumn("adventure_scenes", "inventory_json", "TEXT NOT NULL DEFAULT '[]'")
    ensureColumn("adventure_scenes", "flags_json", "TEXT NOT NULL DEFAULT '{}'")
    database.exec(`CREATE TABLE IF NOT EXISTS adventure_choices (
      id TEXT PRIMARY KEY,
      scene_id TEXT NOT NULL,
      choice_id TEXT NOT NULL,
      label TEXT NOT NULL,
      requires_json TEXT NOT NULL DEFAULT '{}',
      proposed_effects_json TEXT NOT NULL DEFAULT '{}',
      next_hint TEXT NOT NULL DEFAULT '',
      sort_order INTEGER NOT NULL DEFAULT 0,
      UNIQUE(scene_id, choice_id)
    )`)
    database.exec(`CREATE TABLE IF NOT EXISTS adventure_audit (
      id TEXT PRIMARY KEY,
      session_id TEXT NOT NULL,
      turn INTEGER NOT NULL,
      kind TEXT NOT NULL,
      input_json TEXT NOT NULL DEFAULT '{}',
      llm_request_json TEXT NOT NULL DEFAULT '{}',
      llm_response_text TEXT NOT NULL DEFAULT '',
      parsed_json TEXT NOT NULL DEFAULT '{}',
      validation_json TEXT NOT NULL DEFAULT '{}',
      applied_effects_json TEXT NOT NULL DEFAULT '{}',
      created_at TEXT NOT NULL
    )`)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_adventure_sessions_owner ON adventure_sessions(owner_user_id, status, updated_at DESC)`)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_adventure_scenes_session_turn ON adventure_scenes(session_id, turn DESC)`)
    database.exec(`CREATE INDEX IF NOT EXISTS idx_adventure_audit_session_turn ON adventure_audit(session_id, turn DESC)`)
  }

  function ensureColumn(table, column, definition) {
    const rows = database.query(`PRAGMA table_info(${table})`)
    const exists = Array.isArray(rows) && rows.some((row) => row.name === column)
    if (!exists) database.exec(`ALTER TABLE ${table} ADD COLUMN ${column} ${definition}`)
  }

  function seedDefaults() {
    for (const seed of allSeeds()) {
      const existing = one(`SELECT id FROM adventure_seeds WHERE id = ? LIMIT 1`, [seed.id])
      const ts = nowISO()
      if (existing.id) {
        database.exec(`UPDATE adventure_seeds SET title=?, genre=?, tone=?, initial_stats_json=?, inventory_vocab_json=?, flag_vocab_json=?, constraints_json=?, opening_prompt=?, updated_at=? WHERE id=?`,
          seed.title, seed.genre, seed.tone, jsonString(seed.initialStats), jsonString(seed.inventoryVocab), jsonString(seed.flagVocab), jsonString(seed.constraints), seed.openingPrompt, ts, seed.id)
      } else {
        database.exec(`INSERT INTO adventure_seeds (id,title,genre,tone,initial_stats_json,inventory_vocab_json,flag_vocab_json,constraints_json,opening_prompt,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
          seed.id, seed.title, seed.genre, seed.tone, jsonString(seed.initialStats), jsonString(seed.inventoryVocab), jsonString(seed.flagVocab), jsonString(seed.constraints), seed.openingPrompt, ts, ts)
      }
    }
  }

  function one(sql, args) {
    const rows = database.query(sql, ...(args || []))
    return Array.isArray(rows) && rows.length > 0 ? rows[0] : {}
  }

  function many(sql, args) {
    const rows = database.query(sql, ...(args || []))
    return Array.isArray(rows) ? rows : []
  }

  function getSeed(id) {
    const row = one(`SELECT * FROM adventure_seeds WHERE id = ? LIMIT 1`, [id || "haunted-gate"])
    if (!row.id) return null
    return {
      id: row.id,
      title: row.title,
      genre: row.genre,
      tone: row.tone,
      initialStats: safeJsonParse(row.initial_stats_json, {}),
      inventoryVocab: safeJsonParse(row.inventory_vocab_json, []),
      flagVocab: safeJsonParse(row.flag_vocab_json, []),
      constraints: safeJsonParse(row.constraints_json, {}),
      openingPrompt: row.opening_prompt,
    }
  }

  function createSession({ seed, ownerUserId, guildId, channelId, mode, userTheme }) {
    const id = createId("adv")
    const ts = nowISO()
    const initialFlags = userTheme ? { user_theme: String(userTheme), user_tone_override: true } : {}
    database.exec(`INSERT INTO adventure_sessions (id, adventure_id, owner_user_id, guild_id, channel_id, mode, stats_json, inventory_json, flags_json, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
      id, seed.id, ownerUserId || "", guildId || "", channelId || "", mode || "solo", jsonString(seed.initialStats || {}), "[]", jsonString(initialFlags), ts, ts)
    return getSession(id)
  }

  function getSession(id) {
    const row = one(`SELECT * FROM adventure_sessions WHERE id = ? LIMIT 1`, [id])
    return sessionFromRow(row)
  }

  function findActiveSession(ownerUserId, channelId) {
    const row = one(`SELECT * FROM adventure_sessions WHERE owner_user_id = ? AND channel_id = ? AND status = 'active' ORDER BY updated_at DESC LIMIT 1`, [ownerUserId || "", channelId || ""])
    return sessionFromRow(row)
  }

  function findActiveSessionInChannel(channelId) {
    const row = one(`SELECT * FROM adventure_sessions WHERE channel_id = ? AND status = 'active' ORDER BY updated_at DESC LIMIT 1`, [channelId || ""])
    return sessionFromRow(row)
  }

  function findLatestSessionInChannel(channelId) {
    const row = one(`SELECT * FROM adventure_sessions WHERE channel_id = ? AND status IN ('active', 'completed') ORDER BY updated_at DESC LIMIT 1`, [channelId || ""])
    return sessionFromRow(row)
  }

  function sessionFromRow(row) {
    if (!row || !row.id) return null
    return {
      id: row.id,
      adventureId: row.adventure_id,
      ownerUserId: row.owner_user_id,
      guildId: row.guild_id,
      channelId: row.channel_id,
      threadId: row.thread_id,
      mode: row.mode,
      turn: Number(row.turn || 0),
      currentSceneId: row.current_scene_id,
      stats: safeJsonParse(row.stats_json, {}),
      inventory: safeJsonParse(row.inventory_json, []),
      flags: safeJsonParse(row.flags_json, {}),
      status: row.status,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    }
  }

  function saveScene(session, scene) {
    const sceneId = scene.id || createId("scene")
    const ts = nowISO()
    database.exec(`INSERT OR REPLACE INTO adventure_scenes (id, session_id, turn, title, ascii_art, narration, engine_notes_json, raw_patch_json, stats_json, inventory_json, flags_json, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
      sceneId, session.id, session.turn, scene.title, scene.asciiArt || "", scene.narration || "", jsonString(scene.engineNotes || {}), jsonString(scene.rawPatch || {}), jsonString(session.stats || {}), jsonString(session.inventory || []), jsonString(session.flags || {}), ts)
    database.exec(`DELETE FROM adventure_choices WHERE scene_id = ?`, sceneId)
    ;(scene.choices || []).forEach((choice, index) => {
      database.exec(`INSERT INTO adventure_choices (id, scene_id, choice_id, label, requires_json, proposed_effects_json, next_hint, sort_order) VALUES (?,?,?,?,?,?,?,?)`,
        `${sceneId}:${choice.id}`, sceneId, choice.id, choice.label, jsonString(choice.requires || {}), jsonString(choice.proposedEffects || {}), choice.nextHint || "", index)
    })
    database.exec(`UPDATE adventure_sessions SET current_scene_id = ?, updated_at = ? WHERE id = ?`, sceneId, ts, session.id)
    return getScene(sceneId)
  }

  function getScene(id) {
    const row = one(`SELECT * FROM adventure_scenes WHERE id = ? LIMIT 1`, [id])
    if (!row.id) return null
    const choices = many(`SELECT * FROM adventure_choices WHERE scene_id = ? ORDER BY sort_order ASC`, [row.id]).map((choice) => ({
      id: choice.choice_id,
      label: choice.label,
      requires: safeJsonParse(choice.requires_json, {}),
      proposedEffects: safeJsonParse(choice.proposed_effects_json, {}),
      nextHint: choice.next_hint,
    }))
    const rawPatch = safeJsonParse(row.raw_patch_json, {})
    const endingRaw = rawPatch && rawPatch.scene_patch && rawPatch.scene_patch.ending ? rawPatch.scene_patch.ending : {}
    return {
      id: row.id,
      sessionId: row.session_id,
      turn: Number(row.turn || 0),
      title: row.title,
      asciiArt: row.ascii_art,
      narration: row.narration,
      engineNotes: safeJsonParse(row.engine_notes_json, {}),
      ending: { isFinal: Boolean(endingRaw.is_final || endingRaw.isFinal), summary: endingRaw.summary || "" },
      snapshot: {
        stats: safeJsonParse(row.stats_json, {}),
        inventory: safeJsonParse(row.inventory_json, []),
        flags: safeJsonParse(row.flags_json, {}),
      },
      rawPatch,
      choices,
    }
  }

  function getCurrentScene(session) {
    if (!session || !session.currentSceneId) return null
    return getScene(session.currentSceneId)
  }

  function getSceneByTurn(session, turn) {
    if (!session) return null
    const row = one(`SELECT * FROM adventure_scenes WHERE session_id = ? AND turn = ? LIMIT 1`, [session.id, Number(turn || 0)])
    return row && row.id ? getScene(row.id) : null
  }

  function advanceSession(session, effects) {
    const nextStats = Object.assign({}, session.stats || {})
    const nextFlags = Object.assign({}, session.flags || {})
    const nextInventory = Array.isArray(session.inventory) ? session.inventory.slice() : []
    const statEffects = effects && effects.stats && typeof effects.stats === "object" ? effects.stats : {}
    Object.keys(statEffects).forEach((key) => {
      const delta = Math.max(-6, Math.min(3, Number(statEffects[key] || 0)))
      nextStats[key] = Number(nextStats[key] || 0) + delta
    })
    const flagEffects = effects && effects.flags && typeof effects.flags === "object" ? effects.flags : {}
    Object.keys(flagEffects).forEach((key) => { nextFlags[key] = Boolean(flagEffects[key]) })
    const addItems = effects && Array.isArray(effects.add_inventory) ? effects.add_inventory : []
    addItems.forEach((item) => { if (!nextInventory.includes(item)) nextInventory.push(item) })
    const ts = nowISO()
    const nextTurn = Number(session.turn || 0) + 1
    database.exec(`UPDATE adventure_sessions SET turn=?, stats_json=?, inventory_json=?, flags_json=?, updated_at=? WHERE id=?`,
      nextTurn, jsonString(nextStats), jsonString(nextInventory), jsonString(nextFlags), ts, session.id)
    return getSession(session.id)
  }

  function addAudit({ sessionId, turn, kind, input, llmRequest, llmResponseText, parsed, validation, appliedEffects }) {
    database.exec(`INSERT INTO adventure_audit (id, session_id, turn, kind, input_json, llm_request_json, llm_response_text, parsed_json, validation_json, applied_effects_json, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
      createId("audit"), sessionId || "", Number(turn || 0), kind || "event", jsonString(input || {}), jsonString(llmRequest || {}), llmResponseText || "", jsonString(parsed || {}), jsonString(validation || {}), jsonString(appliedEffects || {}), nowISO())
  }

  function finishSession(session) {
    database.exec(`UPDATE adventure_sessions SET status = 'completed', updated_at = ? WHERE id = ?`, nowISO(), session.id)
    return getSession(session.id)
  }

  function exportSession(session) {
    const scenes = many(`SELECT * FROM adventure_scenes WHERE session_id = ? ORDER BY turn ASC`, [session.id]).map((scene) => ({
      id: scene.id,
      turn: Number(scene.turn || 0),
      title: scene.title,
      asciiArt: scene.ascii_art,
      narration: scene.narration,
      snapshot: {
        stats: safeJsonParse(scene.stats_json, {}),
        inventory: safeJsonParse(scene.inventory_json, []),
        flags: safeJsonParse(scene.flags_json, {}),
      },
      rawPatch: safeJsonParse(scene.raw_patch_json, {}),
      choices: many(`SELECT choice_id, label, proposed_effects_json FROM adventure_choices WHERE scene_id = ? ORDER BY sort_order ASC`, [scene.id]).map((choice) => ({
        id: choice.choice_id,
        label: choice.label,
        proposedEffects: safeJsonParse(choice.proposed_effects_json, {}),
      })),
    }))
    const audit = many(`SELECT turn, kind, input_json, applied_effects_json, created_at FROM adventure_audit WHERE session_id = ? ORDER BY created_at ASC`, [session.id]).map((row) => ({
      turn: Number(row.turn || 0),
      kind: row.kind,
      input: safeJsonParse(row.input_json, {}),
      appliedEffects: safeJsonParse(row.applied_effects_json, {}),
      createdAt: row.created_at,
    }))
    return { session, scenes, audit }
  }

  function resetActive(ownerUserId, channelId) {
    database.exec(`UPDATE adventure_sessions SET status = 'abandoned', updated_at = ? WHERE owner_user_id = ? AND channel_id = ? AND status = 'active'`, nowISO(), ownerUserId || "", channelId || "")
  }

  return { ensure, getSeed, createSession, getSession, findActiveSession, findActiveSessionInChannel, findLatestSessionInChannel, saveScene, getScene, getCurrentScene, getSceneByTurn, advanceSession, addAudit, finishSession, exportSession, resetActive }
}

module.exports = { createStore }
