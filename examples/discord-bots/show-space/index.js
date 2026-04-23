const { defineBot } = require("discord")

module.exports = defineBot(({ configure, event }) => {
  configure({
    name: "show-space",
    description: "Venue operations bot for upcoming shows and pinned announcements",
    category: "venues",
  })

  event("ready", async (ctx) => {
    ctx.log.info("show-space bot ready", {
      user: ctx.me && ctx.me.username,
      script: ctx.metadata && ctx.metadata.scriptPath,
    })
  })
})
