const ui = require("ui")

function statLine(session) {
  const stats = session && session.stats ? session.stats : {}
  const parts = Object.keys(stats).sort().map((key) => `${key.toUpperCase()}: ${stats[key]}`)
  const inventory = Array.isArray(session && session.inventory) && session.inventory.length > 0 ? session.inventory.join(", ") : "empty"
  return `${parts.join("  ")}\nInventory: ${inventory}`
}

function sceneContent(session, scene) {
  const art = scene.asciiArt ? `${scene.asciiArt}\n\n` : ""
  const body = `${art}${scene.narration || ""}`.trim()
  return [
    "```",
    `╔═ ${scene.title || "Adventure"}`,
    `Turn ${session.turn}`,
    "",
    body.slice(0, 1600),
    "",
    statLine(session),
    "```",
  ].join("\n").slice(0, 1900)
}

function sceneMessage(session, scene, options) {
  const builder = ui.message().content(sceneContent(session, scene))
  if (scene && scene.ending && scene.ending.isFinal) {
    const exported = options && options.exported ? options.exported : { session, scene }
    builder.row(ui.button("adv:ended", "Adventure complete", "secondary").disabled())
    return Object.assign(builder.build(), {
      files: [{
        name: `adventure-${session.id}.json`,
        content: JSON.stringify(exported, null, 2),
        contentType: "application/json",
      }],
    })
  }
  const choices = (scene.choices || []).slice(0, 4)
  if (choices.length > 0) {
    const buttons = choices.map((choice, index) => ui.button(`adv:choice:${index}`, choice.label, index === 0 ? "primary" : "secondary"))
    builder.row(...buttons)
  }
  builder.row(ui.button("adv:freeform", "Try something else…", "secondary"))
  if (options && options.ephemeral) builder.ephemeral()
  return builder.build()
}

function loadingMessage(session, text) {
  return ui.message()
    .content(["```", `╔═ ${text || "The story shifts..."}`, `Turn ${session ? session.turn : "?"}`, "", "The mist curls while the next scene is written...", "```"].join("\n"))
    .build()
}

function errorMessage(message) {
  return ui.message().ephemeral().content(`⚠️ ${message}`).build()
}

function stateMessage(session, scene) {
  return ui.message()
    .ephemeral()
    .content("```json\n" + JSON.stringify({ session, scene }, null, 2).slice(0, 1800) + "\n```")
    .build()
}

module.exports = { sceneMessage, loadingMessage, errorMessage, stateMessage }
