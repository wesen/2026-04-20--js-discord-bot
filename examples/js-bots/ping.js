const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "ping-bot", runtime: "js" })

  command("ping", {
    description: "Reply with pong from the JavaScript Discord bot API"
  }, async () => {
    return {
      content: "pong",
      embeds: [
        {
          title: "Pong",
          description: "JavaScript handled this slash command.",
          color: 0x5865F2,
        }
      ],
      components: [
        {
          type: "actionRow",
          components: [
            {
              type: "button",
              style: "link",
              label: "Project repo",
              url: "https://github.com/manuel/wesen"
            }
          ]
        }
      ]
    }
  })

  command("echo", {
    description: "Echo text back from JavaScript",
    options: {
      text: {
        type: "string",
        description: "Text to echo back",
        required: true,
      }
    }
  }, async (ctx) => {
    await ctx.defer({ ephemeral: true })
    await ctx.edit({
      content: ctx.args.text,
      embeds: [
        {
          title: "Echo",
          description: "Edited deferred response from JavaScript.",
          color: 0x57F287,
        }
      ]
    })
    await ctx.followUp({
      content: "Follow-up from JavaScript",
      ephemeral: true,
    })
  })

  event("ready", async (ctx) => {
    ctx.log.info("js discord bot connected", {
      user: ctx.me && ctx.me.username,
      script: ctx.metadata && ctx.metadata.scriptPath,
    })
  })

  event("guildCreate", async (ctx) => {
    ctx.log.info("joined guild", {
      guild: ctx.guild && ctx.guild.name,
      guildId: ctx.guild && ctx.guild.id,
    })
  })

  event("messageCreate", async (ctx) => {
    const content = (ctx.message && ctx.message.content || "").trim()
    if (content === "!pingjs") {
      await ctx.reply({
        content: "pong from messageCreate",
        embeds: [
          {
            title: "Message event",
            description: "This came from a normal Discord message.",
            color: 0xFEE75C,
          }
        ]
      })
    }
  })
})
