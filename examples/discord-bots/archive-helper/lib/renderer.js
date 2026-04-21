// Markdown rendering for Discord messages

function renderArchive(channel, thread, messages) {
  const lines = []

  // Frontmatter
  lines.push("---")
  lines.push(`source: "discord"`)
  lines.push(`server: "${escapeYaml(channel.guildID ? "Guild " + channel.guildID : "unknown")}"`)
  lines.push(`server_id: "${channel.guildID || "unknown"}"`)
  lines.push(`channel: "${escapeYaml(channel.name || "unknown")}"`)
  lines.push(`channel_id: "${channel.id}"`)
  if (thread) {
    lines.push(`thread: "${escapeYaml(thread.name || "unknown")}"`)
    lines.push(`thread_id: "${thread.id}"`)
  }
  lines.push(`archived_at: "${new Date().toISOString()}"`)
  lines.push(`message_count: ${messages.length}`)
  lines.push("---")
  lines.push("")

  // Messages
  for (const message of messages) {
    lines.push(renderMessage(message))
    lines.push("")
  }

  return lines.join("\n")
}

function renderMessage(message) {
  const author = message.author && message.author.username || "unknown"
  const timestamp = message.timestamp || new Date().toISOString()
  const formattedTime = String(timestamp).replace("T", " ").slice(0, 19) + " UTC"
  const edited = message.editedTimestamp ? " *(edited)*" : ""

  const header = `**${author}** *(${formattedTime})*:${edited}`
  let body = discordToMarkdown(String(message.content || ""))

  // If content is empty and this is a thread starter placeholder, show a note
  if (!body.trim() && message.type === 21) {
    body = "*(thread starter — original message may be in parent channel)*"
  }

  // Render attachments
  const attachmentLines = renderAttachments(message.attachments)

  // Render embeds
  const embedLines = renderEmbeds(message.embeds)

  const parts = [header]
  if (body.trim()) parts.push(body)
  if (attachmentLines) parts.push(attachmentLines)
  if (embedLines) parts.push(embedLines)

  return parts.join("\n")
}

function renderAttachments(attachments) {
  if (!attachments || attachments.length === 0) return null
  const lines = attachments.map(att => {
    if (att.contentType && att.contentType.startsWith("image/")) {
      return `![${att.filename}](${att.url})`
    }
    return `[📎 ${att.filename}](${att.url})`
  })
  return lines.join("\n")
}

function renderEmbeds(embeds) {
  if (!embeds || embeds.length === 0) return null
  const lines = []
  for (const embed of embeds) {
    lines.push("> ---")
    if (embed.title) {
      if (embed.url) {
        lines.push(`> **[${embed.title}](${embed.url})**`)
      } else {
        lines.push(`> **${embed.title}**`)
      }
    }
    if (embed.description) {
      lines.push("> " + embed.description.replace(/\n/g, "\n> "))
    }
    lines.push("> ---")
  }
  return lines.join("\n")
}

function discordToMarkdown(content) {
  if (!content) return ""

  // Mentions
  content = content.replace(/<@(\d+)>/g, "@user-$1")
  content = content.replace(/<#(\d+)>/g, "#channel-$1")
  content = content.replace(/<@&(\d+)>/g, "@role-$1")
  content = content.replace(/<:(\w+):(\d+)>/g, ":$1:")
  content = content.replace(/<a:(\w+):(\d+)>/g, ":$1:")

  // Multi-line quotes: >>> \n... -> > line1\n> line2
  content = content.replace(/^>>>(\s*)\n?((?:.|\n)*)/gm, (match, space, quote) => {
    return quote.split("\n").map(line => "> " + line).join("\n")
  })

  return content
}

function sanitize(name) {
  return String(name || "unknown")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 60)
}

function escapeYaml(value) {
  return String(value || "").replace(/"/g, '\\"')
}

module.exports = { renderArchive, renderMessage, discordToMarkdown, sanitize, escapeYaml }
