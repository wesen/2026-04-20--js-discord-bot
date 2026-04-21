function shouldCaptureMessage(ctx, config) {
  const message = (ctx && ctx.message) || {}
  const author = message.author || {}
  if (!message.content || author.bot) {
    return false
  }
  if (!configValue(config, ["captureEnabled", "capture_enabled"], true)) {
    return false
  }
  const allowedChannels = parseIdList(configValue(config, ["captureChannels", "capture_channels"], ""))
  if (allowedChannels.length > 0) {
    const channelId = String(message.channelID || message.channelId || (ctx.channel && ctx.channel.id) || "").trim()
    if (!channelId || !allowedChannels.includes(channelId)) {
      return false
    }
  }
  return true
}

function captureFromMessage(ctx, config) {
  if (!shouldCaptureMessage(ctx, config)) {
    return null
  }
  const message = ctx.message || {}
  const raw = normalizeWhitespace(String(message.content || ""))
  const score = scoreMessage(raw)
  const threshold = Number(configValue(config, ["captureThreshold", "capture_threshold"], 0.65))
  if (score < threshold) {
    return null
  }

  const guildId = String((ctx.guild && ctx.guild.id) || message.guildID || message.guildId || "").trim()
  const channelId = String((ctx.channel && ctx.channel.id) || message.channelID || message.channelId || "").trim()
  const messageId = String(message.id || message.messageID || message.messageId || "").trim()
  const author = message.author || ctx.user || {}
  const title = deriveTitle(raw)
  const summary = deriveSummary(raw)
  const tags = deriveTags(raw)
  const aliases = deriveAliases(raw, title)

  return {
    title,
    summary,
    body: raw,
    tags,
    aliases,
    confidence: score,
    status: "draft",
    source: {
      kind: "capture",
      guildId,
      channelId,
      messageId,
      authorId: String(author.id || "").trim(),
      jumpUrl: buildJumpUrl(guildId, channelId, messageId),
      note: "Captured from a Discord message",
    },
  }
}

function captureFromModal(values, ctx) {
  const rawValues = values || {}
  const title = normalizeWhitespace(rawValues.title || rawValues.name || "")
  const summary = normalizeWhitespace(rawValues.summary || rawValues.excerpt || "")
  const body = normalizeWhitespace(rawValues.body || rawValues.details || rawValues.content || summary)
  const tags = parseIdList(rawValues.tags || rawValues.tag || "")
  const aliases = parseIdList(rawValues.aliases || rawValues.alias || "")
  const sourceUrl = normalizeWhitespace(rawValues.source || rawValues.sourceUrl || rawValues.url || "")
  const sourceNote = normalizeWhitespace(rawValues.sourceNote || rawValues.note || "")
  if (!title || !body) {
    return null
  }
  const guildId = String((ctx.guild && ctx.guild.id) || "").trim()
  const channelId = String((ctx.channel && ctx.channel.id) || "").trim()
  const userId = String((ctx.user && ctx.user.id) || "").trim()
  return {
    title,
    summary: summary || body,
    body,
    tags,
    aliases,
    confidence: 0.8,
    status: "draft",
    source: {
      kind: "manual",
      guildId,
      channelId,
      messageId: "",
      authorId: userId,
      jumpUrl: sourceUrl,
      note: sourceNote || "Submitted through /teach",
    },
  }
}

function buildJumpUrl(guildId, channelId, messageId) {
  if (!guildId || !channelId || !messageId) {
    return ""
  }
  return `https://discord.com/channels/${guildId}/${channelId}/${messageId}`
}

function scoreMessage(text) {
  const content = normalizeWhitespace(text)
  if (!content) {
    return 0
  }
  let score = 0.2
  if (/```/.test(content)) score += 0.35
  if (/^\s*(use|run|set|install|configure|fix|remember|note|save|export|deploy)\b/im.test(content)) score += 0.25
  if (/\b(the fix is|the answer is|we should|you can|for example|try this)\b/i.test(content)) score += 0.15
  if (/https?:\/\//i.test(content)) score += 0.1
  if (/\b(sqlite|database|bot|discord|javascript|goja|command|workflow|review)\b/i.test(content)) score += 0.05
  if (content.length > 160) score += 0.05
  return Math.min(score, 1)
}

function deriveTitle(text) {
  const normalized = normalizeWhitespace(text)
  if (!normalized) {
    return "Untitled knowledge"
  }
  const firstLine = normalized.split(/[\n\.\?!]/)[0].trim()
  const title = firstLine.length > 72 ? `${firstLine.slice(0, 69).trim()}...` : firstLine
  return title || "Untitled knowledge"
}

function deriveSummary(text) {
  const normalized = normalizeWhitespace(text)
  if (!normalized) {
    return ""
  }
  return normalized.length > 260 ? `${normalized.slice(0, 257).trim()}...` : normalized
}

function deriveTags(text) {
  const content = normalizeWhitespace(text).toLowerCase()
  const tags = new Set()
  if (/discord|slash command|guild|channel/.test(content)) tags.add("discord")
  if (/sqlite|database|sql/.test(content)) tags.add("database")
  if (/javascript|js|goja/.test(content)) tags.add("javascript")
  if (/command|modal|button|component|autocomplete/.test(content)) tags.add("interaction")
  if (/review|verify|stale|draft|queue/.test(content)) tags.add("curation")
  if (/runbook|workflow|how to|remember|teach/.test(content)) tags.add("knowledge")
  if (/bot|automation/.test(content)) tags.add("bot")
  return Array.from(tags)
}

function deriveAliases(text, title) {
  const aliases = new Set()
  if (title) {
    aliases.add(title.toLowerCase())
  }
  const content = normalizeWhitespace(text).toLowerCase()
  const match = content.match(/(?:remember|teach|search|review|capture|record)\s+([a-z0-9][a-z0-9\s-]{2,48})/i)
  if (match && match[1]) {
    aliases.add(normalizeWhitespace(match[1]).toLowerCase())
  }
  return Array.from(aliases)
}

function parseIdList(raw) {
  if (Array.isArray(raw)) {
    return raw.map(normalizeWhitespace).filter(Boolean)
  }
  const text = normalizeWhitespace(raw)
  if (!text) {
    return []
  }
  return text
    .split(/[\s,;]+/g)
    .map((part) => normalizeWhitespace(part))
    .filter(Boolean)
}

function normalizeWhitespace(value) {
  if (value === undefined || value === null) {
    return ""
  }
  return String(value).replace(/\s+/g, " ").trim()
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

module.exports = {
  shouldCaptureMessage,
  captureFromMessage,
  captureFromModal,
  deriveTitle,
  deriveSummary,
  deriveTags,
  deriveAliases,
  scoreMessage,
}
