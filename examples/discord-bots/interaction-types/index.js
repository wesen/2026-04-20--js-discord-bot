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

  // Root command for /fun — defines the Discord command structure with subcommands
  command("fun", {
    description: "Fun and games",
    options: {
      roll: {
        type: "sub_command",
        description: "Roll a die",
        options: {
          sides: { type: "integer", description: "Number of sides (default 6)", default: 6 }
        }
      },
      coin: {
        type: "sub_command",
        description: "Flip a coin"
      }
    }
  }, async (ctx) => {
    return { content: "Please use `/fun roll` or `/fun coin`" }
  })

  // Subcommand handlers for /fun roll and /fun coin
  subcommand("fun", "roll", {
    description: "Roll a die"
  }, async (ctx) => {
    const sides = ctx.args.sides || 6
    const result = Math.floor(Math.random() * sides) + 1
    return {
      content: `🎲 You rolled a **${result}** (1-${sides})`,
      ephemeral: true
    }
  })

  subcommand("fun", "coin", {
    description: "Flip a coin"
  }, async (ctx) => {
    const result = Math.random() < 0.5 ? "Heads" : "Tails"
    return {
      content: `🪙 **${result}**!`,
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
