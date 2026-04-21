const DEFAULT_REVIEW_LIMIT = 5
const REVIEW_COMPONENTS = {
  select: "knowledge:review:select",
  verify: "knowledge:review:verify",
  edit: "knowledge:review:edit",
  stale: "knowledge:review:stale",
  reject: "knowledge:review:reject",
  source: "knowledge:review:source",
}

function reviewStateKey(ctx) {
  const guildId = String((ctx.guild && ctx.guild.id) || "dm").trim() || "dm"
  const channelId = String((ctx.channel && ctx.channel.id) || "unknown-channel").trim() || "unknown-channel"
  const userId = String((ctx.user && ctx.user.id) || (ctx.me && ctx.me.id) || "unknown-user").trim() || "unknown-user"
  return `knowledge.review.state.${guildId}.${channelId}.${userId}`
}

function loadReviewState(ctx) {
  const state = ctx.store.get(reviewStateKey(ctx), null)
  if (!state || typeof state !== "object") {
    return { status: "draft", limit: DEFAULT_REVIEW_LIMIT, selectedId: "" }
  }
  return normalizeReviewState(state)
}

function saveReviewState(ctx, state) {
  const normalized = normalizeReviewState(state)
  ctx.store.set(reviewStateKey(ctx), normalized)
  return normalized
}

function normalizeReviewState(state) {
  const source = state || {}
  const status = normalizeStatus(source.status)
  const limit = clampLimit(source.limit, DEFAULT_REVIEW_LIMIT)
  const selectedId = String(source.selectedId || "").trim()
  return { status, limit, selectedId }
}

function normalizeStatus(value) {
  const status = String(value || "draft").trim().toLowerCase()
  switch (status) {
    case "draft":
    case "review":
    case "verified":
    case "stale":
    case "rejected":
      return status
    default:
      return "draft"
  }
}

function clampLimit(limit, fallback) {
  const n = Number(limit)
  if (!Number.isFinite(n) || n <= 0) {
    return fallback
  }
  return Math.min(Math.floor(n), 25)
}

function currentReviewEntry(ctx, store) {
  const state = loadReviewState(ctx)
  if (!store) {
    return null
  }
  if (state.selectedId) {
    const selected = store.getEntry(state.selectedId)
    if (selected) {
      return selected
    }
  }
  const entries = store.listByStatus(ctx.config, state.status, state.limit)
  if (!entries || entries.length === 0) {
    return null
  }
  const next = entries[0]
  if (next && next.id) {
    saveReviewState(ctx, { status: state.status, limit: state.limit, selectedId: next.id })
  }
  return next || null
}

function selectedReviewEntryId(ctx, store) {
  const state = loadReviewState(ctx)
  if (state.selectedId) {
    return state.selectedId
  }
  const entries = store.listByStatus(ctx.config, state.status, state.limit)
  if (!entries || entries.length === 0) {
    return ""
  }
  const first = entries[0]
  if (first && first.id) {
    saveReviewState(ctx, { status: state.status, limit: state.limit, selectedId: first.id })
    return first.id
  }
  return ""
}

function setReviewSelection(ctx, entryId) {
  const state = loadReviewState(ctx)
  const next = saveReviewState(ctx, { status: state.status, limit: state.limit, selectedId: entryId })
  return next
}

function rememberReviewView(ctx, status, limit, selectedId) {
  return saveReviewState(ctx, { status, limit, selectedId })
}

function queueResponse(entries, state, render) {
  return render.reviewQueue(entries, state)
}

function renderCurrentReviewQueue(ctx, store, render) {
  const state = loadReviewState(ctx)
  const entries = store.listByStatus(ctx.config, state.status, state.limit)
  const selectedId = state.selectedId || (entries[0] && entries[0].id) || ""
  saveReviewState(ctx, { status: state.status, limit: state.limit, selectedId })
  return queueResponse(entries, { status: state.status, limit: state.limit, selectedId }, render)
}

function renderReviewQueueForStatus(ctx, store, render, status, limit) {
  const entries = store.listByStatus(ctx.config, status, limit)
  const selectedId = (entries[0] && entries[0].id) || ""
  saveReviewState(ctx, { status, limit, selectedId })
  return queueResponse(entries, { status, limit, selectedId }, render)
}

function buildEntryModal(entry) {
  const safe = entry || {}
  return {
    customId: "knowledge:review:edit",
    title: `Edit ${safe.title || "knowledge entry"}`,
    components: [
      textInputRow("title", "Title", "short", safe.title || "", true, 3, 100),
      textInputRow("summary", "Summary", "paragraph", safe.summary || "", true, 10, 300),
      textInputRow("body", "Body", "paragraph", safe.body || "", true, 20, 4000),
      textInputRow("tags", "Tags (comma-separated)", "short", Array.isArray(safe.tags) ? safe.tags.join(", ") : "", false, 0, 200),
      textInputRow("aliases", "Aliases (comma-separated)", "short", Array.isArray(safe.aliases) ? safe.aliases.join(", ") : "", false, 0, 200),
      textInputRow("source", "Source URL or note", "short", sourceInputValue(safe), false, 0, 300),
    ],
  }
}

function textInputRow(customId, label, style, value, required, minLength, maxLength) {
  const input = {
    type: "textInput",
    customId,
    label,
    style,
    required: Boolean(required),
    value: String(value || ""),
  }
  if (minLength) {
    input.minLength = minLength
  }
  if (maxLength) {
    input.maxLength = maxLength
  }
  return {
    type: "actionRow",
    components: [input],
  }
}

function sourceInputValue(entry) {
  const source = entry && entry.source ? entry.source : {}
  const parts = []
  if (source.jumpUrl) {
    parts.push(source.jumpUrl)
  }
  if (source.note) {
    parts.push(source.note)
  }
  return parts.join("\n")
}

function parseReviewSourceInput(value) {
  const text = String(value || "").trim()
  if (!text) {
    return { jumpUrl: "", note: "" }
  }
  const lines = text.split(/\r?\n/).map((part) => part.trim()).filter(Boolean)
  const url = lines.find((line) => /^https?:\/\//i.test(line)) || ""
  const note = lines.filter((line) => line !== url).join(" ")
  return { jumpUrl: url, note }
}

function reviewSelectionLabel(entry) {
  if (!entry) {
    return "No entry selected"
  }
  return `${entry.title} · ${entry.status} · ${confidenceLabel(entry.confidence)}`
}

function confidenceLabel(confidence) {
  const value = Number(confidence || 0)
  return `${Math.round(Math.max(0, Math.min(1, value)) * 100)}%`
}

function normalizeTags(raw) {
  if (Array.isArray(raw)) {
    return raw.map((item) => String(item || "").trim()).filter(Boolean)
  }
  return String(raw || "")
    .split(/[;,\n]/g)
    .map((part) => String(part || "").trim())
    .filter(Boolean)
}

function renderSelectedEntryCard(entry, queueMeta) {
  if (!entry) {
    return {
      title: "No review entry selected",
      description: "Select a knowledge candidate from the dropdown to review it.",
      color: 0x95A5A6,
    }
  }
  return {
    title: entry.title || "Untitled knowledge",
    description: entry.summary || entry.body || "",
    color: statusColor(entry.status),
    fields: [
      { name: "Entry ID", value: String(entry.id || "(unknown)"), inline: false },
      { name: "Status", value: String(entry.status || "draft"), inline: true },
      { name: "Confidence", value: confidenceLabel(entry.confidence), inline: true },
      { name: "Tags", value: formatList(entry.tags), inline: false },
      { name: "Aliases", value: formatList(entry.aliases), inline: false },
      { name: "Source", value: formatSource(entry), inline: false },
      { name: "Version", value: String(entry.version || 0), inline: true },
    ],
    footer: {
      text: queueMeta ? `Queue: ${queueMeta.status} · ${queueMeta.limit} entries` : "Knowledge steward review",
    },
  }
}

function statusColor(status) {
  switch (String(status || "draft").toLowerCase()) {
    case "verified":
      return 0x57F287
    case "review":
      return 0x5865F2
    case "stale":
      return 0x95A5A6
    case "rejected":
      return 0xED4245
    default:
      return 0xFEE75C
  }
}

function formatList(items) {
  const list = Array.isArray(items) ? items.map((item) => String(item || "").trim()).filter(Boolean) : []
  return list.length > 0 ? list.join(", ") : "(none)"
}

function formatSource(entry) {
  if (!entry || !entry.source) {
    return "(unknown)"
  }
  const source = entry.source
  const parts = []
  if (source.guildId) parts.push(`guild ${source.guildId}`)
  if (source.channelId) parts.push(`channel ${source.channelId}`)
  if (source.messageId) parts.push(`message ${source.messageId}`)
  if (source.jumpUrl) parts.push(source.jumpUrl)
  if (source.note) parts.push(source.note)
  return parts.join(" • ") || "(unknown)"
}

function buildReviewComponents(entries, state) {
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
          customId: REVIEW_COMPONENTS.select,
          placeholder: "Choose a knowledge entry to review",
          options: selectOptions,
        },
      ],
    })
  }

  components.push({
    type: "actionRow",
    components: [
      { type: "button", style: "success", label: "Verify", customId: REVIEW_COMPONENTS.verify },
      { type: "button", style: "secondary", label: "Edit", customId: REVIEW_COMPONENTS.edit },
      { type: "button", style: "secondary", label: "Source", customId: REVIEW_COMPONENTS.source },
      { type: "button", style: "primary", label: "Stale", customId: REVIEW_COMPONENTS.stale },
      { type: "button", style: "danger", label: "Reject", customId: REVIEW_COMPONENTS.reject },
    ],
  })

  return components
}

function formatTagsShort(tags) {
  const list = Array.isArray(tags) ? tags.map((tag) => String(tag || "").trim()).filter(Boolean) : []
  return list.slice(0, 3).join(" ")
}

function truncateLabel(value) {
  const text = String(value || "").trim()
  if (text.length <= 95) {
    return text
  }
  return `${text.slice(0, 92).trim()}...`
}

function reviewReply(entry, verb, queueMeta) {
  if (!entry) {
    return {
      content: "No review entry is currently selected.",
      ephemeral: true,
    }
  }
  return {
    content: `${verb} **${entry.title || "knowledge entry"}** (${entry.status}).`,
    embeds: [renderSelectedEntryCard(entry, queueMeta)],
    ephemeral: true,
  }
}

function reviewSourceReply(entry) {
  if (!entry) {
    return {
      content: "No review entry is currently selected.",
      ephemeral: true,
    }
  }
  return {
    content: `Source for **${entry.title || "knowledge entry"}**`,
    embeds: [
      {
        title: entry.title || "Untitled knowledge",
        description: formatSource(entry),
        color: statusColor(entry.status),
      },
    ],
    ephemeral: true,
  }
}

function buildQueueMessage(entries, state) {
  const items = Array.isArray(entries) ? entries : []
  const selectedId = String(state && state.selectedId || "").trim()
  const selectedEntry = items.find((entry) => entry && entry.id === selectedId) || items[0] || null
  const queueMeta = {
    status: state && state.status ? state.status : "draft",
    limit: state && state.limit ? state.limit : DEFAULT_REVIEW_LIMIT,
  }
  const components = buildReviewComponents(items, { ...queueMeta, selectedId: selectedEntry && selectedEntry.id })
  return {
    content: items.length > 0
      ? `Review queue for ${queueMeta.status} (${items.length} entr${items.length === 1 ? "y" : "ies"})`
      : `No ${queueMeta.status} entries found.`,
    embeds: [renderSelectedEntryCard(selectedEntry, queueMeta)],
    components,
    ephemeral: true,
  }
}

function stateFromQueueCommand(ctx, status, limit, entries) {
  const state = {
    status: normalizeStatus(status),
    limit: clampLimit(limit, DEFAULT_REVIEW_LIMIT),
    selectedId: entries && entries[0] && entries[0].id ? entries[0].id : "",
  }
  saveReviewState(ctx, state)
  return state
}

function entryToEditPatch(values) {
  const raw = values || {}
  const source = parseReviewSourceInput(raw.source)
  return {
    title: String(raw.title || "").trim(),
    summary: String(raw.summary || "").trim(),
    body: String(raw.body || "").trim(),
    tags: normalizeTags(raw.tags),
    aliases: normalizeTags(raw.aliases),
    source: {
      jumpUrl: source.jumpUrl,
      note: source.note,
    },
  }
}

module.exports = {
  REVIEW_COMPONENTS,
  reviewStateKey,
  loadReviewState,
  saveReviewState,
  currentReviewEntry,
  selectedReviewEntryId,
  setReviewSelection,
  rememberReviewView,
  renderCurrentReviewQueue,
  renderReviewQueueForStatus,
  buildEntryModal,
  reviewSelectionLabel,
  reviewReply,
  reviewSourceReply,
  buildQueueMessage,
  stateFromQueueCommand,
  entryToEditPatch,
  normalizeStatus,
  clampLimit,
  parseReviewSourceInput,
}
