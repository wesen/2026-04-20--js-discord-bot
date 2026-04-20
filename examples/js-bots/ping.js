const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "ping-bot", runtime: "js" })

  command("ping", {
    description: "Reply with pong from the JavaScript Discord bot API"
  }, async () => {
    return { content: "pong" }
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
    return { content: ctx.args.text }
  })

  event("ready", async (ctx) => {
    ctx.log.info("js discord bot connected", {
      user: ctx.me && ctx.me.username,
      script: ctx.metadata && ctx.metadata.scriptPath,
    })
  })
})
