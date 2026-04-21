const { defineBot } = require("discord")
const { fetchAllMessages } = require("./lib/fetcher")
const { renderArchive, sanitize } = require("./lib/renderer")

module.exports = defineBot(({ command, messageCommand, event, configure }) => {
  configure({
    name: "archive-helper",
    description: "Download channels and threads as Markdown archives",
    category: "utilities",
    run: {
      fields: {
        default_limit: {
          type: "integer",
          help: "Default maximum messages to archive per request",
          default: 500,
        },
      },
    },
  })

  event("ready", async (ctx) => {
    ctx.log.info("archive-helper bot ready", {
      user: ctx.me && ctx.me.username,
    })
  })

  command("archive-channel", {
    description: "Archive messages from the current channel as Markdown",
    options: {
      limit: {
        type: "integer",
        description: "Maximum messages to archive (default: 500)",
        required: false,
      },
      before_message_id: {
        type: "string",
        description: "Archive messages before this message ID (optional time anchor)",
        required: false,
      },
    },
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id
    if (!channelId) {
      return { content: "This command must be run in a channel.", ephemeral: true }
    }

    await ctx.defer({ ephemeral: true })

    const maxMessages = ctx.args.limit || ctx.config.default_limit || 500
    const beforeId = ctx.args.before_message_id || null

    // Fetch channel info for metadata
    let channel
    try {
      channel = await ctx.discord.channels.fetch(channelId)
    } catch (err) {
      await ctx.edit({ content: `Could not fetch channel info: ${err.message || err}`, ephemeral: true })
      return
    }

    // Fetch messages with pagination
    let messages
    try {
      messages = await fetchAllMessages(ctx, channelId, maxMessages, beforeId)
    } catch (err) {
      await ctx.edit({ content: `Failed to fetch messages: ${err.message || err}`, ephemeral: true })
      return
    }

    if (messages.length === 0) {
      await ctx.edit({ content: "No messages found in this channel.", ephemeral: true })
      return
    }

    // Render to Markdown
    const markdown = renderArchive(channel, null, messages)

    // Send as file attachment
    try {
      await ctx.discord.channels.send(channelId, {
        content: `📄 Archive of #${channel.name}: ${messages.length} messages`,
        files: [
          {
            name: `${sanitize(channel.name)}--archive.md`,
            content: markdown,
          },
        ],
      })
    } catch (err) {
      await ctx.edit({ content: `Failed to send archive file: ${err.message || err}`, ephemeral: true })
      return
    }

    // Update the deferred reply
    await ctx.edit({
      content: `Archived ${messages.length} messages from #${channel.name}.`,
      ephemeral: true,
    })
  })

  messageCommand("Archive Thread", async (ctx) => {
    const targetMessage = ctx.args.target
    if (!targetMessage || !targetMessage.id) {
      return { content: "Could not resolve the target message.", ephemeral: true }
    }

    // The message's channelID IS the thread ID when invoked inside a thread
    const threadId = targetMessage.channelID

    await ctx.defer({ ephemeral: true })

    // Fetch thread info
    let thread
    try {
      thread = await ctx.discord.threads.fetch(threadId)
    } catch (err) {
      await ctx.edit({ content: "This command only works inside threads. Right-click a message in a thread and choose Apps → Archive Thread.", ephemeral: true })
      return
    }

    // Fetch parent channel for metadata
    const channel = await ctx.discord.channels.fetch(thread.parentID)

    // Fetch all messages from the thread
    const maxMessages = ctx.config.default_limit || 500
    const messages = await fetchAllMessages(ctx, threadId, maxMessages)

    if (messages.length === 0) {
      await ctx.edit({ content: "No messages found in this thread.", ephemeral: true })
      return
    }

    // Render to Markdown
    const markdown = renderArchive(channel, thread, messages)

    // Send as file attachment to the current channel (the thread itself)
    await ctx.discord.channels.send(threadId, {
      content: `📄 Archive of thread "${thread.name}": ${messages.length} messages`,
      files: [
        {
          name: `${sanitize(channel.name)}--${sanitize(thread.name)}--archive.md`,
          content: markdown,
        },
      ],
    })

    await ctx.edit({
      content: `Archived ${messages.length} messages from thread "${thread.name}".`,
      ephemeral: true,
    })
  })
})
