const { defineBot } = require("discord")
const { sleep } = require("timer")

module.exports = defineBot(({ command, event, component, modal, autocomplete, configure }) => {
  configure({ name: "ping", description: "Discord JS API showcase with buttons, modals, and autocomplete", category: "examples" })

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
              style: "primary",
              label: "Open panel",
              customId: "ping:panel"
            },
            {
              type: "button",
              style: "link",
              label: "Project repo",
              url: "https://github.com/manuel/wesen"
            }
          ]
        },
        {
          type: "actionRow",
          components: [
            {
              type: "select",
              customId: "ping:topic",
              placeholder: "Choose a topic",
              options: [
                { label: "Architecture", value: "architecture" },
                { label: "Testing", value: "testing" },
                { label: "Runbooks", value: "runbooks" },
              ]
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

  command("feedback", {
    description: "Open a feedback modal"
  }, async (ctx) => {
    await ctx.showModal({
      customId: "feedback:submit",
      title: "Feedback",
      components: [
        {
          type: "actionRow",
          components: [
            {
              type: "textInput",
              customId: "summary",
              label: "Summary",
              style: "short",
              required: true,
              minLength: 5,
              maxLength: 100,
            }
          ]
        },
        {
          type: "actionRow",
          components: [
            {
              type: "textInput",
              customId: "details",
              label: "Details",
              style: "paragraph",
              maxLength: 500,
            }
          ]
        }
      ]
    })
  })

  command("search", {
    description: "Search for a topic with autocomplete",
    options: {
      query: {
        type: "string",
        description: "Topic to search for",
        required: true,
        autocomplete: true,
        minLength: 2,
        maxLength: 100,
      }
    }
  }, async (ctx) => {
    await ctx.defer({ ephemeral: true })
    await ctx.edit({
      content: `Searching for ${ctx.args.query}...`,
      ephemeral: true,
    })
    await sleep(2000)
    const query = String(ctx.args.query || "").trim().toLowerCase()
    const results = [
      { title: "Architecture", summary: "Bot wiring, handlers, and runtime layers." },
      { title: "Testing", summary: "Go tests, JS runtime coverage, and smoke checks." },
      { title: "Runbooks", summary: "How to start, sync, and operate the bot." },
    ].filter((item) => item.title.toLowerCase().includes(query) || item.summary.toLowerCase().includes(query))
    const summary = results.length > 0
      ? results.map((item) => `• **${item.title}** — ${item.summary}`).join("\n")
      : `No results found for ${ctx.args.query}.`
    await ctx.edit({
      content: `Results for ${ctx.args.query}:`,
      embeds: [
        {
          title: "Search results",
          description: summary,
          color: 0x5865F2,
        },
      ],
      ephemeral: true,
    })
  })

  command("announce", {
    description: "Send a report message through the host Discord operations layer"
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id
    if (!channelId) {
      return { content: "no channel available", ephemeral: true }
    }
    await ctx.discord.channels.send(channelId, {
      content: "report generated from JavaScript",
      files: [
        {
          name: "report.txt",
          content: "This report was created inside the JS bot runtime."
        }
      ]
    })
    return {
      content: "sent report to the current channel",
      ephemeral: true,
    }
  })

  component("ping:panel", async () => {
    return {
      content: "Panel button clicked from JavaScript",
      ephemeral: true,
    }
  })

  component("ping:topic", async (ctx) => {
    const selected = Array.isArray(ctx.values) && ctx.values.length > 0 ? ctx.values[0] : "(none)"
    return {
      content: `Selected topic: ${selected}`,
      ephemeral: true,
    }
  })

  modal("feedback:submit", async (ctx) => {
    const summary = ctx.values && ctx.values.summary || "(empty)"
    const details = ctx.values && ctx.values.details || "(none)"
    return {
      content: `Thanks for the feedback: ${summary}\nDetails: ${details}`,
      ephemeral: true,
    }
  })

  autocomplete("search", "query", async (ctx) => {
    const current = String(ctx.focused && ctx.focused.value || "")
    return [
      { name: "Architecture", value: "architecture" },
      { name: "Testing", value: "testing" },
      { name: "Runbooks", value: "runbooks" },
      { name: `Custom: ${current || "query"}`, value: current || "custom" },
    ]
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
