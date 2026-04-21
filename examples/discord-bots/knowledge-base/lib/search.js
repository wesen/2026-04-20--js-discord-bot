const render = require("./render")

const DEFAULT_SEARCH_LIMIT = 5
const MAX_SEARCH_RESULTS = 50
const SEARCH_COMPONENTS = {
  select: "knowledge:search:select",
  previous: "knowledge:search:previous",
  next: "knowledge:search:next",
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
    return { query: "", limit: DEFAULT_SEARCH_LIMIT, page: 1, selectedId: "" }
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
    page: clampPage(source.page, 1),
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

function clampPage(page, fallback) {
  const n = Number(page)
  if (!Number.isFinite(n) || n <= 0) {
    return fallback || 1
  }
  return Math.min(Math.floor(n), 10)
}

function rememberSearchView(ctx, query, limit, selectedId, page) {
  return saveSearchState(ctx, {
    query,
    limit,
    selectedId,
    page,
  })
}

function stateFromSearchCommand(ctx, query, limit, entries) {
  const selectedId = entries && entries[0] && entries[0].id ? entries[0].id : ""
  return rememberSearchView(ctx, query, limit, selectedId, 1)
}

function setSearchSelection(ctx, entryId) {
  const state = loadSearchState(ctx)
  return saveSearchState(ctx, {
    query: state.query,
    limit: state.limit,
    page: state.page,
    selectedId: entryId,
  })
}

function setSearchPage(ctx, page) {
  const state = loadSearchState(ctx)
  return saveSearchState(ctx, {
    query: state.query,
    limit: state.limit,
    page,
    selectedId: "",
  })
}

function shiftSearchPage(ctx, delta) {
  const state = loadSearchState(ctx)
  return setSearchPage(ctx, clampPage(Number(state.page || 1) + Number(delta || 0), 1))
}

function searchEntriesForState(ctx, store, state) {
  const query = String(state && state.query || "").trim()
  if (!store || !query) {
    return []
  }
  const limit = clampLimit(state && state.limit, DEFAULT_SEARCH_LIMIT)
  const page = clampPage(state && state.page, 1)
  const fetchLimit = Math.min(Math.max(limit * (page + 1), limit), MAX_SEARCH_RESULTS)
  return store.search(ctx.config, query, fetchLimit)
}

function pageWindow(entries, limit, page) {
  const items = Array.isArray(entries) ? entries : []
  const safeLimit = clampLimit(limit, DEFAULT_SEARCH_LIMIT)
  const safePage = clampPage(page, 1)
  const maxPage = Math.max(1, Math.ceil(items.length / safeLimit))
  const currentPage = Math.min(safePage, maxPage)
  const start = (currentPage - 1) * safeLimit
  const currentEntries = items.slice(start, start + safeLimit)
  const hasNext = items.length > currentPage * safeLimit
  const hasPrevious = currentPage > 1
  return {
    entries: currentEntries,
    page: currentPage,
    pageCount: Math.max(maxPage, currentPage),
    hasNext,
    hasPrevious,
  }
}

function searchView(ctx, store) {
  const state = loadSearchState(ctx)
  const allResults = searchEntriesForState(ctx, store, state)
  const window = pageWindow(allResults, state.limit, state.page)
  const selectedId = String(state.selectedId || "").trim()
  const selectedEntry = window.entries.find((entry) => entry && entry.id === selectedId) || window.entries[0] || null
  const selectedIndex = selectedEntry ? window.entries.findIndex((entry) => entry && entry.id === selectedEntry.id) : -1
  const relatedEntries = selectedEntry ? relatedEntryHints(selectedEntry, window.entries) : []
  const nextState = saveSearchState(ctx, {
    query: state.query,
    limit: state.limit,
    page: window.page,
    selectedId: selectedEntry && selectedEntry.id ? selectedEntry.id : "",
  })
  return {
    state: nextState,
    query: state.query,
    allResults,
    pageEntries: window.entries,
    selectedEntry,
    selectedIndex: selectedIndex >= 0 ? selectedIndex + 1 : 1,
    relatedEntries,
    page: window.page,
    pageCount: window.pageCount,
    hasNext: window.hasNext,
    hasPrevious: window.hasPrevious,
  }
}

function buildSearchMessage(view) {
  const state = normalizeSearchState(view && view.state)
  const entries = Array.isArray(view && view.pageEntries) ? view.pageEntries : []
  const selectedEntry = view && view.selectedEntry ? view.selectedEntry : null
  const query = String(view && view.query || state.query || "").trim()
  if (!selectedEntry) {
    return {
      content: `No knowledge entries matched ${query}.`,
      ephemeral: true,
    }
  }
  return {
    content: `Found ${view.allResults ? view.allResults.length : entries.length} knowledge entr${(view.allResults ? view.allResults.length : entries.length) === 1 ? "y" : "ies"} for ${query}.`,
    embeds: [renderSearchResultCard(selectedEntry, {
      query,
      total: view.allResults ? view.allResults.length : entries.length,
      position: view.selectedIndex,
      page: view.page,
      pageCount: view.pageCount,
      relatedEntries: view.relatedEntries,
    })],
    components: buildSearchComponents(entries, state, {
      hasPrevious: Boolean(view.hasPrevious),
      hasNext: Boolean(view.hasNext),
    }),
    ephemeral: true,
  }
}

function buildSearchComponents(entries, state, controls) {
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
      { type: "button", style: "secondary", label: "Previous", customId: SEARCH_COMPONENTS.previous },
      { type: "button", style: "secondary", label: "Next", customId: SEARCH_COMPONENTS.next },
    ],
  })

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

function currentSearchEntry(ctx, store) {
  return searchView(ctx, store).selectedEntry
}

function selectedSearchEntryId(ctx, store) {
  const entry = currentSearchEntry(ctx, store)
  return entry && entry.id ? entry.id : ""
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
      page: queryMeta && queryMeta.page,
      pageCount: queryMeta && queryMeta.pageCount,
      relatedEntries: queryMeta && queryMeta.relatedEntries,
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
  return render.searchResultCard(entry, meta)
}

function relatedEntryHints(entry, pool) {
  if (!entry) {
    return []
  }
  const items = Array.isArray(pool) ? pool : []
  const baseTags = new Set(normalizeList(entry.tags).map((tag) => tag.toLowerCase()))
  const baseAliases = new Set(normalizeList(entry.aliases).map((alias) => alias.toLowerCase()))
  const baseWords = new Set(tokenizeText([entry.title, entry.summary].join(" ")).map((word) => word.toLowerCase()))

  const scored = []
  for (const candidate of items) {
    if (!candidate || candidate.id === entry.id) {
      continue
    }
    let score = 0
    const candidateTags = normalizeList(candidate.tags).map((tag) => tag.toLowerCase())
    const candidateAliases = normalizeList(candidate.aliases).map((alias) => alias.toLowerCase())
    const candidateWords = new Set(tokenizeText([candidate.title, candidate.summary].join(" ")).map((word) => word.toLowerCase()))
    for (const tag of candidateTags) {
      if (baseTags.has(tag)) score += 3
    }
    for (const alias of candidateAliases) {
      if (baseAliases.has(alias)) score += 2
    }
    for (const word of candidateWords) {
      if (baseWords.has(word)) score += 1
    }
    if (score > 0) {
      scored.push({ candidate, score })
    }
  }

  return scored
    .sort((a, b) => b.score - a.score)
    .slice(0, 3)
    .map(({ candidate }) => ({
      title: candidate.title || candidate.slug || candidate.id,
      status: candidate.status,
      id: candidate.id,
    }))
}

function searchAutocompleteSuggestions(ctx, store) {
  const focused = String((ctx.focused && ctx.focused.value) || "").trim()
  return autocompleteSuggestions(store.recent(ctx.config, MAX_SEARCH_RESULTS), focused, { includeTags: true, includeAliases: true, includeSummaries: true })
}

function articleAutocompleteSuggestions(ctx, store) {
  const focused = String((ctx.focused && ctx.focused.value) || "").trim()
  return autocompleteSuggestions(store.recent(ctx.config, MAX_SEARCH_RESULTS), focused, { includeTags: false, includeAliases: true, includeSummaries: false })
}

function autocompleteSuggestions(entries, focused, options) {
  const query = String(focused || "").toLowerCase()
  const seen = new Set()
  const suggestions = []
  const includeTags = Boolean(options && options.includeTags)
  const includeAliases = Boolean(options && options.includeAliases)
  const includeSummaries = Boolean(options && options.includeSummaries)

  function add(name, value) {
    const label = String(name || "").trim()
    const key = String(value || label).trim()
    if (!label || !key || seen.has(key)) {
      return
    }
    if (query && !label.toLowerCase().includes(query) && !key.toLowerCase().includes(query)) {
      return
    }
    seen.add(key)
    suggestions.push({ name: truncateLabel(label), value: key })
  }

  for (const entry of entries || []) {
    if (!entry) continue
    add(entry.title || entry.slug || entry.id, entry.slug || entry.id || entry.title)
    add(`${entry.title || entry.id} — ${entry.status}`, entry.id)
    if (includeAliases) {
      for (const alias of normalizeList(entry.aliases)) {
        add(`${alias} — ${entry.title || entry.id}`, entry.id)
      }
    }
    if (includeTags) {
      for (const tag of normalizeList(entry.tags)) {
        add(`#${tag}`, tag)
      }
    }
    if (includeSummaries && entry.summary) {
      add(`${entry.title || entry.id}: ${entry.summary}`, entry.id)
    }
  }

  return suggestions.slice(0, 25)
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

function normalizeList(value) {
  if (Array.isArray(value)) {
    return value.map((item) => String(item || "").trim()).filter(Boolean)
  }
  return String(value || "")
    .split(/[;\n,]/g)
    .map((item) => String(item || "").trim())
    .filter(Boolean)
}

function tokenizeText(text) {
  return String(text || "")
    .toLowerCase()
    .split(/[^a-z0-9]+/g)
    .map((part) => part.trim())
    .filter(Boolean)
}

module.exports = {
  SEARCH_COMPONENTS,
  searchStateKey,
  loadSearchState,
  saveSearchState,
  normalizeSearchState,
  clampLimit,
  clampPage,
  rememberSearchView,
  stateFromSearchCommand,
  setSearchSelection,
  setSearchPage,
  shiftSearchPage,
  searchEntriesForState,
  pageWindow,
  searchView,
  buildSearchMessage,
  buildSearchComponents,
  searchSelectionLabel,
  currentSearchEntry,
  selectedSearchEntryId,
  searchReply,
  searchSourceReply,
  searchExportPayload,
  renderSearchResultCard,
  relatedEntryHints,
  searchAutocompleteSuggestions,
  articleAutocompleteSuggestions,
  autocompleteSuggestions,
  truncateLabel,
  formatTagsShort,
  confidenceLabel,
}
