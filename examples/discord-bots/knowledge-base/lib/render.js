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
  return formatSourceCitation(entry)
}

function formatSourceCitation(entry) {
  if (!entry || !entry.source) {
    return "(unknown)"
  }
  const source = entry.source
  const parts = []
  if (source.kind) parts.push(source.kind)
  if (source.jumpUrl) parts.push(`[jump](${source.jumpUrl})`)
  if (source.guildId) parts.push(`guild ${source.guildId}`)
  if (source.channelId) parts.push(`channel ${source.channelId}`)
  if (source.messageId) parts.push(`message ${source.messageId}`)
  return parts.join(" • ") || "(unknown)"
}

function formatSourceDetails(entry) {
  if (!entry || !entry.source) {
    return "(unknown)"
  }
  const source = entry.source
  const lines = []
  if (source.kind) lines.push(`kind: ${source.kind}`)
  if (source.guildId) lines.push(`guild: ${source.guildId}`)
  if (source.channelId) lines.push(`channel: ${source.channelId}`)
  if (source.messageId) lines.push(`message: ${source.messageId}`)
  if (source.authorId) lines.push(`author: ${source.authorId}`)
  if (source.jumpUrl) lines.push(`jump: ${source.jumpUrl}`)
  if (source.note) lines.push(`note: ${source.note}`)
  return lines.join("\n") || "(unknown)"
}

function knowledgeEmbed(entry) {
  return {
    title: entry.title || "Untitled knowledge",
    description: entry.summary || entry.body || "",
    color: statusColor(entry.status),
    fields: [
      { name: "Status", value: String(entry.status || "draft"), inline: true },
      { name: "Confidence", value: confidenceLabel(entry.confidence), inline: true },
      { name: "Canonical source", value: canonicalSourceLabel(entry), inline: false },
      { name: "Tags", value: formatTags(entry.tags), inline: false },
      { name: "Aliases", value: formatAliases(entry.aliases), inline: false },
      { name: "Source citation", value: formatSourceCitation(entry), inline: false },
      { name: "Source details", value: formatSourceDetails(entry), inline: false },
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
    embeds: results.map((entry, index) => searchResultCard(entry, {
      query,
      total: results.length,
      position: index + 1,
    })),
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

function searchResultCard(entry, meta) {
  const embed = knowledgeEmbed(entry)
  const fields = Array.isArray(embed.fields) ? embed.fields.slice() : []
  fields.unshift({ name: "Entry ID", value: String(entry && entry.id || "(unknown)"), inline: false })
  if (meta && Array.isArray(meta.relatedEntries) && meta.relatedEntries.length > 0) {
    fields.push({ name: "Related entries", value: meta.relatedEntries.map(formatRelatedEntry).join("\n"), inline: false })
  }
  const next = { ...embed, fields }
  if (meta) {
    const footerParts = []
    if (meta.query) footerParts.push(`Search: ${meta.query}`)
    if (meta.position && meta.total) footerParts.push(`Result ${meta.position}/${meta.total}`)
    if (meta.position && !meta.total) footerParts.push(`Result ${meta.position}`)
    if (meta.total && !meta.position) footerParts.push(`Total ${meta.total}`)
    if (meta.page && meta.pageCount) footerParts.push(`Page ${meta.page}/${meta.pageCount}`)
    else if (meta.page) footerParts.push(`Page ${meta.page}`)
    if (footerParts.length > 0) {
      next.footer = { text: footerParts.join(" · ") }
    }
  }
  return next
}

function sourceDetailsEmbed(entry, title) {
  if (!entry) {
    return {
      title: title || "Source details",
      description: "(unknown)",
      color: 0x95A5A6,
    }
  }
  return {
    title: title || `${entry.title || "Untitled knowledge"} source`,
    description: formatSourceDetails(entry),
    color: statusColor(entry.status),
    fields: [
      { name: "Entry ID", value: String(entry.id || "(unknown)"), inline: false },
      { name: "Status", value: String(entry.status || "draft"), inline: true },
      { name: "Citation", value: formatSourceCitation(entry), inline: false },
    ],
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

function canonicalSourceLabel(entry) {
  if (!entry) {
    return "(unknown)"
  }
  if (String(entry.status || "").toLowerCase() === "verified") {
    return "Verified canonical entry"
  }
  if (entry.source && String(entry.source.kind || "").toLowerCase() === "seed") {
    return "Seeded canonical entry"
  }
  return "Candidate entry"
}

function formatRelatedEntry(entry) {
  if (!entry) {
    return "(unknown)"
  }
  const title = String(entry.title || entry.slug || entry.id || "Untitled").trim()
  const status = String(entry.status || "draft").trim()
  return `• **${title}** — ${status}`
}

function formatAliases(aliases) {
  const list = Array.isArray(aliases) ? aliases.map((alias) => String(alias || "").trim()).filter(Boolean) : []
  return list.length > 0 ? list.join(", ") : "(none)"
}

module.exports = {
  statusColor,
  formatTags,
  formatSource,
  formatSourceCitation,
  formatSourceDetails,
  knowledgeEmbed,
  knowledgeAnnouncement,
  searchResults,
  recentResults,
  searchResultCard,
  sourceDetailsEmbed,
  reviewQueue,
  confidenceLabel,
  formatAliases,
}
