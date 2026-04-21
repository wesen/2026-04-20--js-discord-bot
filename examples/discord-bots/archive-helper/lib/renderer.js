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

  const header = `**${author}** *(${formattedTime})*:`
  const body = discordToMarkdown(String(message.content || ""))

  return [header, body].join("\n")
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
