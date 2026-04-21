function registerReactionPromotions({ event }, store, render) {
  event("reactionAdd", async (ctx) => {
    const result = promoteCandidateFromReaction(ctx, store, render)
    if (!result) {
      return null
    }
    return result
  })
}

function promoteCandidateFromReaction(ctx, store, render) {
  const reaction = ctx.reaction || {}
  const emoji = reaction.emoji || {}
  const emojiLabel = emojiLabelForReaction(emoji)
  if (!emojiLabel || !isPromoteEmoji(ctx.config, emojiLabel)) {
    return null
  }
  if (!isTrustedReviewer(ctx, ctx.config)) {
    return null
  }

  const message = ctx.message || {}
  const channelId = String(message.channelID || message.channelId || (ctx.channel && ctx.channel.id) || "").trim()
  const messageId = String(message.id || message.messageID || message.messageId || "").trim()
  if (!channelId || !messageId) {
    return null
  }

  const entry = store.findBySource(channelId, messageId)
  if (!entry) {
    return null
  }

  const nextStatus = nextPromotionStatus(entry.status)
  if (!nextStatus) {
    return null
  }

  const actor = String((ctx.user && ctx.user.id) || "").trim() || "trusted-reactor"
  const updated = store.setStatus(ctx.config, entry.id, nextStatus, actor, `Promoted via ${emojiLabel} reaction`)
  if (!updated) {
    return null
  }

  return {
    content: `Promoted **${updated.title}** from ${entry.status} to ${updated.status} via ${emojiLabel}.`,
    embeds: [render.knowledgeEmbed(updated)],
  }
}

function isPromoteEmoji(config, emojiLabel) {
  const configured = parseList(configValue(config, ["reactionPromoteEmojis", "reaction_promote_emojis"], "🧠,📌"))
  if (configured.length === 0) {
    return false
  }
  const normalized = String(emojiLabel || "").trim()
  return configured.includes(normalized)
}

function isTrustedReviewer(ctx, config) {
  const trustedUsers = parseList(configValue(config, ["trustedReviewerIds", "trusted_reviewer_ids"], ""))
  const trustedRoles = parseList(configValue(config, ["trustedReviewerRoleIds", "trusted_reviewer_role_ids"], ""))
  if (trustedUsers.length === 0 && trustedRoles.length === 0) {
    return true
  }

  const userId = String((ctx.user && ctx.user.id) || "").trim()
  if (userId && trustedUsers.includes(userId)) {
    return true
  }

  const roles = Array.isArray(ctx.member && ctx.member.roles) ? ctx.member.roles.map((role) => String(role || "").trim()).filter(Boolean) : []
  return roles.some((roleId) => trustedRoles.includes(roleId))
}

function nextPromotionStatus(status) {
  switch (String(status || "draft").toLowerCase()) {
    case "draft":
      return "review"
    case "review":
      return "verified"
    case "stale":
    case "rejected":
      return "review"
    default:
      return ""
  }
}

function emojiLabelForReaction(emoji) {
  if (!emoji) {
    return ""
  }
  if (String(emoji.name || "").trim()) {
    return String(emoji.name).trim()
  }
  if (String(emoji.id || "").trim()) {
    return String(emoji.id).trim()
  }
  return ""
}

function parseList(raw) {
  if (Array.isArray(raw)) {
    return raw.map((item) => String(item || "").trim()).filter(Boolean)
  }
  return String(raw || "")
    .split(/[;,\n]/g)
    .map((item) => String(item || "").trim())
    .filter(Boolean)
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
  registerReactionPromotions,
  promoteCandidateFromReaction,
  isPromoteEmoji,
  isTrustedReviewer,
  nextPromotionStatus,
  emojiLabelForReaction,
}
