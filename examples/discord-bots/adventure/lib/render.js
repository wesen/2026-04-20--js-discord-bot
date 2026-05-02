const ui = require("ui")

function statLine(session) {
  const stats = session && session.stats ? session.stats : {}
  const parts = Object.keys(stats).sort().map((key) => `${key.toUpperCase()}: ${stats[key]}`)
  const inventory = Array.isArray(session && session.inventory) && session.inventory.length > 0 ? session.inventory.join(", ") : "empty"
  return `${parts.join("  ")}\nInventory: ${inventory}`
}

function actionLine(scene) {
  const action = scene && scene.action ? scene.action : null
  if (!action || !action.label) return ""
  const actor = action.actor ? `\nBy: ${action.actor}` : ""
  return `Action: ${action.label}${actor}`
}

function sceneContent(session, scene) {
  const displayState = scene && scene.snapshot ? Object.assign({}, session || {}, scene.snapshot) : session
  const art = scene.asciiArt ? `${scene.asciiArt}\n\n` : ""
  const body = `${art}${scene.narration || ""}`.trim()
  const action = actionLine(scene)
  return [
    "```",
    `╔═ ${scene.title || "Adventure"}`,
    `Turn ${scene && scene.turn !== undefined ? scene.turn : session.turn}`,
    "",
    action,
    action ? "" : null,
    body.slice(0, action ? 1450 : 1600),
    "",
    statLine(displayState),
    "```",
  ].filter((line) => line !== null).join("\n").slice(0, 1900)
}

function codaContent(session, scene, exported) {
  const scenes = exported && Array.isArray(exported.scenes) ? exported.scenes : []
  const summary = scene && scene.ending && scene.ending.summary ? scene.ending.summary : "The adventure has ended."
  const lookback = scenes.length > 0
    ? scenes.map((item) => {
        const action = item.action && item.action.label ? ` — ${item.action.label}${item.action.actor ? ` (${item.action.actor})` : ""}` : ""
        return `Turn ${item.turn}: ${item.title || "Untitled"}${action}`
      }).join("\n")
    : "Use ← Previous to look back through the adventure."
  return [
    sceneContent(session, scene),
    "",
    "**Coda**",
    summary.slice(0, 500),
    "",
    "**Look back**",
    lookback.slice(0, 700),
    "",
    "Use the navigation buttons to scroll through previous scenes.",
  ].join("\n").slice(0, 2000)
}

function sceneMessage(session, scene, options) {
  const builder = ui.message().content(sceneContent(session, scene))
  const viewingHistory = Boolean(options && options.history)
  if (viewingHistory) {
    const rows = []
    if (scene.turn > 0) rows.push(ui.button("adv:history:prev", "← Previous", "secondary"))
    if (scene.turn < session.turn) rows.push(ui.button("adv:history:next", "Next →", "secondary"))
    if (scene.turn !== session.turn) rows.push(ui.button("adv:history:current", session.status === "completed" ? "Coda" : "Current", "primary"))
    if (rows.length > 0) builder.row(...rows)
    return builder.build()
  }
  if (scene && scene.ending && scene.ending.isFinal) {
    const storyboard = options && options.storyboard
    const content = codaContent(session, scene, options && options.exported) + (storyboard && storyboard.imageUrl ? `\n\nStoryboard: ${storyboard.imageUrl}` : storyboard && storyboard.file ? "\n\nStoryboard image generated." : "")
    const ret = { content }
    if (storyboard && storyboard.file) ret.files = [storyboard.file]
    if (scene.turn > 0) ret.components = [{ type: "row", components: [{ type: "button", customId: "adv:history:prev", label: "← Previous", style: "secondary" }] }]
    return ret
  }
  const nav = []
  if (scene && scene.turn > 0) nav.push(ui.button("adv:history:prev", "← Previous", "secondary"))
  if (nav.length > 0) builder.row(...nav)
  const choices = (scene.choices || []).slice(0, 4)
  if (choices.length > 0) {
    const buttons = choices.map((choice, index) => ui.button(`adv:choice:${index}`, choice.label, index === 0 ? "primary" : "secondary"))
    builder.row(...buttons)
  }
  builder.row(ui.button("adv:freeform", "Try something else…", "secondary"))
  if (options && options.ephemeral) builder.ephemeral()
  return builder.build()
}

function partialJsonString(jsonText, key) {
  const text = String(jsonText || "")
  const marker = `"${key}"`
  const start = text.indexOf(marker)
  if (start < 0) return ""
  const colon = text.indexOf(":", start + marker.length)
  if (colon < 0) return ""
  const quote = text.indexOf('"', colon + 1)
  if (quote < 0) return ""
  let out = ""
  let escaped = false
  for (let i = quote + 1; i < text.length; i++) {
    const ch = text[i]
    if (escaped) {
      if (ch === "n") out += "\n"
      else if (ch === "t") out += "\t"
      else out += ch
      escaped = false
      continue
    }
    if (ch === "\\") {
      escaped = true
      continue
    }
    if (ch === '"') break
    out += ch
  }
  return out.trim()
}

function streamingScenePreview(jsonText) {
  const title = partialJsonString(jsonText, "title")
  const ascii = partialJsonString(jsonText, "ascii_art")
  const narration = partialJsonString(jsonText, "narration")
  const lines = []
  if (title) lines.push(title)
  if (ascii) lines.push(ascii.split(/\r?\n/).slice(0, 8).join("\n"))
  if (narration) lines.push(narration.slice(0, 700))
  return lines.join("\n\n")
}

function loadingMessage(session, text, details) {
  const action = details && details.action ? `Action: ${details.action}` : ""
  const actor = details && details.actor ? `By: ${details.actor}` : ""
  const streamText = details && details.streamText ? streamingScenePreview(details.streamText) : ""
  return ui.message()
    .content([
      "```",
      `╔═ ${text || "The story shifts..."}`,
      `Turn ${session ? session.turn : "?"}`,
      "",
      action,
      actor,
      "",
      "The mist curls while the next scene is written...",
      streamText,
      "```",
    ].filter((line) => line !== "").join("\n").slice(0, 1900))
    .build()
}

function pendingActionMessage(session, scene, details) {
  const action = details && details.action ? String(details.action) : "Something else"
  const actor = details && details.actor ? String(details.actor) : "Someone"
  const streamText = details && details.streamText ? streamingScenePreview(details.streamText) : ""
  const content = [
    sceneContent(session, scene),
    "",
    `**Action chosen:** ${action}`,
    `**By:** ${actor}`,
    "",
    "_Writing the next scene..._",
    streamText ? "```\n" + streamText.trim() + "\n```" : "",
  ].filter(Boolean).join("\n").slice(0, 2000)
  return ui.message().content(content).build()
}

function codaMessage(session, scene, exported) {
  return sceneMessage(session, scene, { exported })
}

function storyboardMessage(session, storyboard) {
  const content = storyboard && storyboard.imageUrl
    ? `Storyboard for adventure ${session.id}: ${storyboard.imageUrl}`
    : `Storyboard for adventure ${session.id} generated.`
  const ret = { content }
  if (storyboard && storyboard.file) ret.files = [storyboard.file]
  return ret
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

module.exports = { sceneMessage, codaMessage, loadingMessage, pendingActionMessage, storyboardMessage, errorMessage, stateMessage }
