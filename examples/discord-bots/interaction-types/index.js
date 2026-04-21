const { defineBot } = require("discord")

module.exports = defineBot(({ command, userCommand, messageCommand, subcommand, event, configure }) => {
  configure({
    name: "interaction-types",
    description: "Demo of all Discord application command interaction types",
    category: "examples"
  })

  // Simple slash command
  command("hello", {
    description: "Say hello"
  }, async () => {
    return { content: "Hello from the interaction types bot! 👋" }
  })

  // Slash command with options
  command("echo", {
    description: "Echo text back",
    options: {
      text: { type: "string", description: "Text to echo", required: true }
    }
  }, async (ctx) => {
    return { content: ctx.args.text }
  })

  // Root command for /admin — defines the Discord command structure
  command("admin", {
    description: "Administration commands",
    options: {
      kick: {
        type: "sub_command",
        description: "Kick a user",
        options: {
          user: { type: "user", description: "User to kick", required: true },
          reason: { type: "string", description: "Reason for kick" }
        }
      },
      ban: {
        type: "sub_command",
        description: "Ban a user",
        options: {
          user: { type: "user", description: "User to ban", required: true },
          duration: { type: "integer", description: "Ban duration in days" }
        }
      }
    }
  }, async (ctx) => {
    return { content: "Please use `/admin kick` or `/admin ban`" }
  })

  // Subcommand handlers for /admin kick and /admin ban
  subcommand("admin", "kick", {
    description: "Kick a user"
  }, async (ctx) => {
    const userId = ctx.args.user || "unknown"
    const reason = ctx.args.reason || "no reason given"
    return {
      content: `🔨 Would kick <@${userId}> for: ${reason}`,
      ephemeral: true
    }
  })

  subcommand("admin", "ban", {
    description: "Ban a user"
  }, async (ctx) => {
    const userId = ctx.args.user || "unknown"
    const duration = ctx.args.duration || 0
    return {
      content: `🔨 Would ban <@${userId}> for ${duration} day(s)`,
      ephemeral: true
    }
  })

  // User context menu command — right-click a user
  userCommand("Show Avatar", async (ctx) => {
    const target = ctx.args.target
    if (!target || !target.id) {
      return { content: "Could not resolve target user", ephemeral: true }
    }
    const avatarUrl = target.avatar
      ? `https://cdn.discordapp.com/avatars/${target.id}/${target.avatar}.png?size=512`
      : `https://cdn.discordapp.com/embed/avatars/${(parseInt(target.discriminator) || 0) % 5}.png`
    return {
      content: `**${target.username}**'s avatar:\n${avatarUrl}`,
      ephemeral: true
    }
  })

  // Message context menu command — right-click a message
  messageCommand("Quote Message", async (ctx) => {
    const target = ctx.args.target
    if (!target || !target.id) {
      return { content: "Could not resolve target message", ephemeral: true }
    }
    const author = target.author && target.author.username || "unknown"
    const content = target.content || "(empty message)"
    return {
      content: `> ${content}\n— **${author}**`,
      ephemeral: true
    }
  })

  event("ready", async (ctx) => {
    ctx.log.info("interaction-types bot ready", {
      user: ctx.me && ctx.me.username,
      script: ctx.metadata && ctx.metadata.scriptPath,
    })
  })
})
