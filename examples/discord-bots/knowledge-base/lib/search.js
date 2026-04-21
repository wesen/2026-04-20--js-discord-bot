const render = require("./render")

const DEFAULT_SEARCH_LIMIT = 5
const SEARCH_COMPONENTS = {
  select: "knowledge:search:select",
  open: "knowledge:search:open",
  source: "knowledge:search:source",
  export: "knowledge:search:export",
}

function searchStateKey(ctx) {
  const guildId = String((ctx.guild && ctx.guild.id) || "dm").trim() || "dm"
  const channelId = String((ctx.channel && ctx.channel.id) || "unknown-channel").trim() || "unknown-channel"
  const userId = String((ctx.user && ctx.user.id) || (ctx.me && ctx.me.id) || "unknown-user").trim() || "unknown-user"
  return `knowledge.search.state.${guildId}.${channelId}.${userId}`
}

function loadSearchState(ctx) {
  const state = ctx.store.get(searchStateKey(ctx), null)
  if (!state || typeof state !== "object") {
    return { query: "", limit: DEFAULT_SEARCH_LIMIT, selectedId: "" }
  }
  return normalizeSearchState(state)
}

function saveSearchState(ctx, state) {
  const normalized = normalizeSearchState(state)
  ctx.store.set(searchStateKey(ctx), normalized)
  return normalized
}

function normalizeSearchState(state) {
  const source = state || {}
  return {
    query: String(source.query || "").trim(),
    limit: clampLimit(source.limit, DEFAULT_SEARCH_LIMIT),
    selectedId: String(source.selectedId || "").trim(),
  }
}

function clampLimit(limit, fallback) {
  const n = Number(limit)
  if (!Number.isFinite(n) || n <= 0) {
    return fallback
  }
  return Math.min(Math.floor(n), 25)
}

function rememberSearchView(ctx, query, limit, selectedId) {
  return saveSearchState(ctx, {
    query,
    limit,
    selectedId,
  })
}

function stateFromSearchCommand(ctx, query, limit, entries) {
  const selectedId = entries && entries[0] && entries[0].id ? entries[0].id : ""
  return rememberSearchView(ctx, query, limit, selectedId)
}

function searchEntries(ctx, store) {
  const state = loadSearchState(ctx)
  if (!store || !state.query) {
    return []
  }
  return store.search(ctx.config, state.query, state.limit)
}

function currentSearchEntry(ctx, store) {
  const state = loadSearchState(ctx)
  const entries = searchEntries(ctx, store)
  if (!entries || entries.length === 0) {
    return null
  }
  if (state.selectedId) {
    const selected = entries.find((entry) => entry && entry.id === state.selectedId)
    if (selected) {
      return selected
    }
  }
  const first = entries[0]
  if (first && first.id) {
    saveSearchState(ctx, { query: state.query, limit: state.limit, selectedId: first.id })
  }
  return first || null
}

function selectedSearchEntryId(ctx, store) {
  const state = loadSearchState(ctx)
  if (state.selectedId) {
    return state.selectedId
  }
  const entries = searchEntries(ctx, store)
  if (!entries || entries.length === 0) {
    return ""
  }
  const first = entries[0]
  if (first && first.id) {
    saveSearchState(ctx, { query: state.query, limit: state.limit, selectedId: first.id })
    return first.id
  }
  return ""
}

function setSearchSelection(ctx, entryId) {
  const state = loadSearchState(ctx)
  return saveSearchState(ctx, { query: state.query, limit: state.limit, selectedId: entryId })
}

function buildSearchMessage(query, entries, state) {
  const items = Array.isArray(entries) ? entries : []
  const queryText = String(query || "").trim()
  const selectedId = String(state && state.selectedId || "").trim()
  const selectedEntry = items.find((entry) => entry && entry.id === selectedId) || items[0] || null
  const total = items.length
  const activeState = {
    query: queryText,
    limit: clampLimit(state && state.limit, DEFAULT_SEARCH_LIMIT),
    selectedId: selectedEntry && selectedEntry.id ? selectedEntry.id : selectedId,
  }
  const components = buildSearchComponents(items, activeState)
  if (!selectedEntry) {
    return {
      content: `No knowledge entries matched ${queryText}.`,
      ephemeral: true,
    }
  }
  const selectedIndex = items.findIndex((entry) => entry && entry.id === selectedEntry.id)
  return {
    content: `Found ${total} knowledge entr${total === 1 ? "y" : "ies"} for ${queryText}.`,
    embeds: [renderSearchResultCard(selectedEntry, {
      query: queryText,
      total,
      position: selectedIndex >= 0 ? selectedIndex + 1 : 1,
    })],
    components,
    ephemeral: true,
  }
}

function buildSearchComponents(entries, state) {
  const items = Array.isArray(entries) ? entries.slice(0, 25) : []
  const selectedId = String(state && state.selectedId || "").trim()
  const selectOptions = items.map((entry) => ({
    label: truncateLabel(entry.title || entry.slug || entry.id || "Untitled"),
    value: entry.id,
    description: truncateLabel(`${entry.status} · ${confidenceLabel(entry.confidence)} · ${formatTagsShort(entry.tags)}`),
    default: entry.id === selectedId,
  }))

  const components = []
  if (selectOptions.length > 0) {
    components.push({
      type: "actionRow",
      components: [
        {
          type: "select",
          customId: SEARCH_COMPONENTS.select,
          placeholder: "Choose a knowledge entry to inspect",
          options: selectOptions,
        },
      ],
    })
  }

  components.push({
    type: "actionRow",
    components: [
      { type: "button", style: "primary", label: "Open", customId: SEARCH_COMPONENTS.open },
      { type: "button", style: "secondary", label: "Source", customId: SEARCH_COMPONENTS.source },
      { type: "button", style: "success", label: "Export", customId: SEARCH_COMPONENTS.export },
    ],
  })

  return components
}

function searchSelectionLabel(entry) {
  if (!entry) {
    return "No entry selected"
  }
  return `${entry.title || entry.slug || entry.id} · ${entry.status} · ${confidenceLabel(entry.confidence)}`
}

function searchReply(entry, queryMeta) {
  if (!entry) {
    return {
      content: "No search result is currently selected.",
      ephemeral: true,
    }
  }
  return {
    content: queryMeta && queryMeta.query ? `Selected **${entry.title || "knowledge entry"}** for ${queryMeta.query}.` : `Selected **${entry.title || "knowledge entry"}**.`,
    embeds: [renderSearchResultCard(entry, {
      query: queryMeta && queryMeta.query ? queryMeta.query : "",
      total: queryMeta && queryMeta.total,
      position: queryMeta && queryMeta.position,
    })],
    ephemeral: true,
  }
}

function searchSourceReply(entry) {
  if (!entry) {
    return {
      content: "No search result is currently selected.",
      ephemeral: true,
    }
  }
  return {
    content: `Source for **${entry.title || "knowledge entry"}**`,
    embeds: [render.sourceDetailsEmbed(entry, `${entry.title || "Untitled knowledge"} source citation`)],
    ephemeral: true,
  }
}

function searchExportPayload(entry, query) {
  if (!entry) {
    return null
  }
  const queryText = String(query || "").trim()
  return {
    content: `Shared from /ask${queryText ? ` for ${queryText}` : ""}: **${entry.title || "knowledge entry"}** (${entry.status}).`,
    embeds: [render.knowledgeEmbed(entry)],
  }
}

function renderSearchResultCard(entry, meta) {
  const card = render.searchResultCard(entry, meta)
  return card
}

function truncateLabel(value) {
  const text = String(value || "").trim()
  if (text.length <= 95) {
    return text
  }
  return `${text.slice(0, 92).trim()}...`
}

function formatTagsShort(tags) {
  const list = Array.isArray(tags) ? tags.map((tag) => String(tag || "").trim()).filter(Boolean) : []
  return list.slice(0, 3).join(" ")
}

function confidenceLabel(confidence) {
  const value = Number(confidence || 0)
  return `${Math.round(Math.max(0, Math.min(1, value)) * 100)}%`
}



module.exports = {
  SEARCH_COMPONENTS,
  searchStateKey,
  loadSearchState,
  saveSearchState,
  rememberSearchView,
  stateFromSearchCommand,
  searchEntries,
  currentSearchEntry,
  selectedSearchEntryId,
  setSearchSelection,
  buildSearchMessage,
  buildSearchComponents,
  searchSelectionLabel,
  searchReply,
  searchSourceReply,
  searchExportPayload,
  renderSearchResultCard,
  clampLimit,
}
