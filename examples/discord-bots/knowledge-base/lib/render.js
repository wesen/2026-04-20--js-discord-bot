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

function formatTags(tags) {
  const list = Array.isArray(tags) ? tags.map((tag) => String(tag || "").trim()).filter(Boolean) : []
  return list.length > 0 ? list.map((tag) => `#${tag}`).join(" ") : "(none)"
}

function formatSource(entry) {
  if (!entry || !entry.source) {
    return "(unknown)"
  }
  const source = entry.source
  const parts = []
  if (source.kind) parts.push(source.kind)
  if (source.guildId) parts.push(`guild ${source.guildId}`)
  if (source.channelId) parts.push(`channel ${source.channelId}`)
  if (source.messageId) parts.push(`message ${source.messageId}`)
  if (source.jumpUrl) parts.push(source.jumpUrl)
  if (source.note) parts.push(source.note)
  return parts.join(" • ") || "(unknown)"
}

function knowledgeEmbed(entry) {
  return {
    title: entry.title || "Untitled knowledge",
    description: entry.summary || entry.body || "",
    color: statusColor(entry.status),
    fields: [
      { name: "Status", value: String(entry.status || "draft"), inline: true },
      { name: "Confidence", value: confidenceLabel(entry.confidence), inline: true },
      { name: "Tags", value: formatTags(entry.tags), inline: false },
      { name: "Aliases", value: formatAliases(entry.aliases), inline: false },
      { name: "Source", value: formatSource(entry), inline: false },
    ],
  }
}

function knowledgeAnnouncement(entry, verb) {
  const action = verb || "Saved"
  return {
    content: `${action} knowledge entry **${entry.title}** (${entry.status}).`,
    embeds: [knowledgeEmbed(entry)],
  }
}

function searchResults(query, entries) {
  const results = Array.isArray(entries) ? entries : []
  if (results.length === 0) {
    return {
      content: `No knowledge entries matched ${query}.`,
      ephemeral: true,
    }
  }
  return {
    content: `Found ${results.length} knowledge entr${results.length === 1 ? "y" : "ies"} for ${query}.`,
    embeds: [
      {
        title: `Search results for ${query}`,
        description: results.map((entry) => renderResultLine(entry)).join("\n"),
        color: 0x5865F2,
      },
    ],
    ephemeral: true,
  }
}

function recentResults(title, entries) {
  const results = Array.isArray(entries) ? entries : []
  return {
    content: title,
    embeds: [
      {
        title,
        description: results.length > 0 ? results.map((entry) => renderResultLine(entry)).join("\n") : "No entries yet.",
        color: 0x5865F2,
      },
    ],
    ephemeral: true,
  }
}

function renderResultLine(entry) {
  const status = String(entry.status || "draft")
  const confidence = confidenceLabel(entry.confidence)
  const tagText = Array.isArray(entry.tags) && entry.tags.length > 0 ? ` ${formatTags(entry.tags)}` : ""
  return `• **${entry.title || entry.slug || "Untitled"}** — ${status} · ${confidence}${tagText}`
}

function reviewQueue(entries, status) {
  const items = Array.isArray(entries) ? entries : []
  const label = status || "draft"
  return {
    content: `Review queue for ${label}`,
    embeds: [
      {
        title: `Review queue: ${label}`,
        description: items.length > 0 ? items.map((entry) => renderReviewLine(entry)).join("\n") : `No ${label} entries found.`,
        color: 0x5865F2,
      },
    ],
    ephemeral: true,
  }
}

function renderReviewLine(entry) {
  const parts = [
    `• **${entry.title || entry.slug || "Untitled"}**`,
    `id: ${entry.id}`,
    `status: ${entry.status}`,
  ]
  if (entry.reviewNote) {
    parts.push(`note: ${entry.reviewNote}`)
  }
  return parts.join(" · ")
}

function confidenceLabel(confidence) {
  const value = Number(confidence || 0)
  return `${Math.round(Math.max(0, Math.min(1, value)) * 100)}%`
}

function formatAliases(aliases) {
  const list = Array.isArray(aliases) ? aliases.map((alias) => String(alias || "").trim()).filter(Boolean) : []
  return list.length > 0 ? list.join(", ") : "(none)"
}

module.exports = {
  statusColor,
  formatTags,
  formatSource,
  knowledgeEmbed,
  knowledgeAnnouncement,
  searchResults,
  recentResults,
  reviewQueue,
  confidenceLabel,
  formatAliases,
}
