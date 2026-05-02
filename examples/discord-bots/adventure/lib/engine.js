const prompts = require("./prompts")
const llm = require("./llm")
const { validateScenePatch, validateInterpretedAction } = require("./schema")

function publicSession(session) {
  return {
    id: session.id,
    adventure_id: session.adventureId,
    mode: session.mode,
    turn: session.turn,
    current_scene_id: session.currentSceneId,
    stats: session.stats,
    inventory: session.inventory,
    flags: session.flags,
  }
}

function publicScene(scene) {
  if (!scene) return null
  return {
    id: scene.id,
    title: scene.title,
    narration: scene.narration,
    choices: (scene.choices || []).map((choice) => ({ id: choice.id, label: choice.label, next_hint: choice.nextHint })),
  }
}

function hpKey(stats) {
  return Object.keys(stats || {}).find((key) => key.toLowerCase() === "hp") || "hp"
}

function secondaryStatKey(stats) {
  return Object.keys(stats || {}).find((key) => key.toLowerCase() !== "hp") || "spirit"
}

function hasMassiveHarm(choice) {
  const stats = choice && choice.proposedEffects && choice.proposedEffects.stats ? choice.proposedEffects.stats : {}
  return Object.keys(stats).some((key) => Number(stats[key] || 0) <= -4)
}

function ensureDangerousChoice(scene, session) {
  if (!scene || scene.ending && scene.ending.isFinal) return scene
  const choices = scene.choices || []
  if (choices.some(hasMassiveHarm)) return scene
  const stats = session && session.stats ? session.stats : {}
  const secondary = secondaryStatKey(stats)
  const target = choices[choices.length - 1]
  if (!target) return scene
  target.label = target.label || "Risk everything"
  target.proposedEffects = Object.assign({}, target.proposedEffects || {}, {
    stats: Object.assign({}, (target.proposedEffects && target.proposedEffects.stats) || {}, { [secondary]: -5 }),
  })
  target.nextHint = target.nextHint || `A disastrous gamble that could shatter your ${secondary}.`
  return scene
}

function secondaryCollapseNarration(stat, actorAction) {
  const key = String(stat || "spirit").toLowerCase()
  const upper = key.toUpperCase()
  const action = actorAction || "your last choice"
  if (key === "groove") return `The cost of ${action} hits the beat like a cracked mirror. Your GROOVE falls to zero. The bassline stumbles, the lights lose their color, and the dance floor forgets how to move. With no rhythm left to carry you, the party-fantasy collapses into silence. Your adventure ends here.`
  if (key === "mana") return `The cost of ${action} burns through the last blue spark inside you. Your MANA falls to zero. The spellwork unthreads, every charm goes cold, and the enchanted world closes its doors. Your adventure ends here.`
  if (key === "oxygen") return `The cost of ${action} empties the last breath from your lungs. Your OXYGEN falls to zero. Stars smear into a dark tide, the alarms fade, and the void takes the final word. Your adventure ends here.`
  if (key === "comfort") return `The cost of ${action} drains the last warmth from the world. Your COMFORT falls to zero. The tea goes bitter, the hearthlight gutters, and every safe place becomes strange. Your adventure ends here.`
  if (key === "focus") return `The cost of ${action} shatters your final thread of attention. Your FOCUS falls to zero. Clues blur, motives invert, and the case becomes a maze with no center. Your adventure ends here.`
  if (key === "swagger") return `The cost of ${action} strips away your final ounce of bravado. Your SWAGGER falls to zero. The crew sees the tremor in your grin, the sea turns its back, and the legend ends mid-boast. Your adventure ends here.`
  if (key === "signal") return `The cost of ${action} floods every channel with static. Your SIGNAL falls to zero. Neon glyphs fracture, systems desync, and your presence drops from the network forever. Your adventure ends here.`
  return `The cost of ${action} reaches deeper than flesh. Your ${upper} falls to zero, and the thread that held this journey together snaps in a way only this world could understand. Your adventure ends here.`
}

function terminalForDepletedStats(store, session, input) {
  const stats = session && session.stats ? session.stats : {}
  const hp = hpKey(stats)
  const secondary = secondaryStatKey(stats)
  const hpValue = Number(stats[hp] || 0)
  const secondaryValue = Number(stats[secondary] || 0)
  if (hpValue > 0 && secondaryValue > 0) return null

  const actorAction = input && (input.label || input.text || input.summary) ? String(input.label || input.text || input.summary) : "your last choice"
  const secondaryUpper = secondary.toUpperCase()
  const hpDeath = hpValue <= 0
  const title = hpDeath ? "The Last Breath" : `${secondaryUpper} Breaks`
  const narration = hpDeath
    ? `The cost of ${actorAction} lands with final force. Your HP falls to zero. The world narrows to a thin bright line, then goes quiet. Your adventure ends here.`
    : secondaryCollapseNarration(secondary, actorAction)
  store.addAudit({ sessionId: session.id, turn: session.turn, kind: "scene_patch", input, llmRequest: {}, llmResponseText: "", parsed: {}, validation: { ok: true, terminal: true }, appliedEffects: input && input.effects ? input.effects : {} })
  const scene = store.saveScene(session, {
    id: `${session.id}_turn_${session.turn}_ending`,
    title,
    asciiArt: hpDeath ? "  x_x\n /| |\\\n  / \\" : "  . . .\n .  *  .\n  '---'",
    narration,
    choices: [],
    engineNotes: { terminal: hpDeath ? "hp_depleted" : `${secondary}_depleted` },
    ending: { isFinal: true, summary: narration },
    rawPatch: { scene_patch: { ending: { is_final: true, summary: narration } } },
  })
  const finalSession = store.finishSession(session)
  return { ok: true, scene, session: finalSession, exported: store.exportSession(finalSession) }
}

function generateStoryboard(store, exported) {
  if (!exported || !Array.isArray(exported.scenes) || exported.scenes.length === 0) return null
  const scenes = exported.scenes.map((scene) => ({ turn: scene.turn, title: scene.title, narration: String(scene.narration || "").slice(0, 220) }))
  const prompt = [
    "Create one cohesive illustrated storyboard image for this completed choose-your-own-adventure.",
    "Show 4-6 panels in a single wide image, with clear visual progression from beginning to ending.",
    "No readable text, no captions, no UI, no speech bubbles. Evocative fantasy/adventure illustration style.",
    JSON.stringify({ scenes }, null, 2),
  ].join("\n")
  const request = {
    purpose: "adventure_storyboard",
    system: "You generate a single image storyboard from a completed adventure story.",
    user: prompt,
    metadata: { sessionId: exported.session && exported.session.id, turn: exported.session && exported.session.turn },
  }
  console.log("[adventure] storyboard prompt", JSON.stringify({ sessionId: exported.session && exported.session.id, sceneCount: scenes.length, prompt: prompt.slice(0, 1200) }))
  const result = llm.generateImage(request)
  if (store && exported.session && exported.session.id) {
    store.addAudit({
      sessionId: exported.session.id,
      turn: exported.session.turn,
      kind: result.ok ? "storyboard_image" : "storyboard_image_error",
      input: { scenes, prompt },
      llmRequest: request,
      llmResponseText: result.imageUrl || result.text || result.error || "",
      parsed: result.raw || result,
      validation: { ok: Boolean(result.ok), errors: result.ok ? [] : [result.error || "image generation failed"] },
      appliedEffects: {},
    })
  }
  if (!result.ok) return { ok: false, error: result.error }
  return imageAttachmentFromURL(result.imageUrl)
}

function imageAttachmentFromURL(imageUrl) {
  const text = String(imageUrl || "")
  const match = text.match(/^data:(image\/[a-zA-Z0-9.+-]+);base64,(.+)$/)
  if (!match) return { ok: true, imageUrl: text }
  const ext = match[1].includes("jpeg") || match[1].includes("jpg") ? "jpg" : match[1].includes("webp") ? "webp" : "png"
  return { ok: true, file: { name: `adventure-storyboard.${ext}`, content: match[2], contentType: match[1], encoding: "base64" } }
}

function finishResultWithStoryboard(result, store) {
  if (!result || !result.exported) return result
  const storyboard = generateStoryboard(store, result.exported)
  if (storyboard && storyboard.ok) result.storyboard = storyboard
  else if (storyboard && storyboard.error) result.storyboardError = storyboard.error
  return result
}

function generateScene({ store, seed, session, currentScene, input, onChunk }) {
  console.log("[adventure] generateScene", JSON.stringify({ sessionId: session.id, turn: session.turn, inputKind: input && input.kind }))
  const request = {
    purpose: "scene_patch",
    system: prompts.sceneSystemPrompt(),
    user: prompts.sceneUserPrompt({
      seed,
      session: publicSession(session),
      currentScene: publicScene(currentScene),
      input,
      recentHistory: [],
    }),
    metadata: { sessionId: session.id, turn: session.turn, adventureId: session.adventureId },
  }
  const completed = llm.completeJson(request, onChunk)
  console.log("[adventure] generateScene llm result", JSON.stringify({ sessionId: session.id, ok: completed.ok, error: completed.error || "" }))
  if (!completed.ok) {
    store.addAudit({ sessionId: session.id, turn: session.turn, kind: "scene_patch_error", input, llmRequest: request, llmResponseText: completed.rawText || "", parsed: completed.parsed || {}, validation: { ok: false, errors: [completed.error] } })
    return { ok: false, error: completed.error }
  }
  const validation = validateScenePatch(completed.value, seed)
  console.log("[adventure] generateScene validation", JSON.stringify({ sessionId: session.id, ok: validation.ok, errors: validation.errors || [] }))
  store.addAudit({ sessionId: session.id, turn: session.turn, kind: "scene_patch", input, llmRequest: request, llmResponseText: completed.rawText, parsed: completed.value, validation, appliedEffects: input && input.effects ? input.effects : {} })
  if (!validation.ok) {
    return { ok: false, error: validation.errors.join("; ") }
  }
  if (!validation.scene.id) {
    validation.scene.id = `${session.id}_turn_${session.turn}`
  }
  ensureDangerousChoice(validation.scene, session)
  const scene = store.saveScene(session, validation.scene)
  const finalSession = scene.ending && scene.ending.isFinal ? store.finishSession(session) : session
  const exported = scene.ending && scene.ending.isFinal ? store.exportSession(finalSession) : null
  return finishResultWithStoryboard({ ok: true, scene, session: finalSession, exported }, store)
}

function applyChoice(store, session, scene, choiceIndex, actor) {
  const choice = scene && scene.choices ? scene.choices[choiceIndex] : null
  if (!choice) return { ok: false, error: "That choice is no longer available." }
  const nextSession = store.advanceSession(session, choice.proposedEffects || {})
  const input = { kind: "choice", choice_id: choice.id, label: choice.label, actor: actor || "", effects: choice.proposedEffects || {}, next_hint: choice.nextHint || "" }
  const terminal = terminalForDepletedStats(store, nextSession, input)
  if (terminal) return finishResultWithStoryboard(terminal, store)
  return { ok: true, session: nextSession, input }
}

function interpretFreeform({ store, seed, session, currentScene, text, actor, onChunk }) {
  console.log("[adventure] interpretFreeform", JSON.stringify({ sessionId: session.id, turn: session.turn, textLength: String(text || "").length }))
  const request = {
    purpose: "interpret_action",
    system: prompts.actionSystemPrompt(),
    user: prompts.actionUserPrompt({ seed, session: publicSession(session), currentScene: publicScene(currentScene), text }),
    metadata: { sessionId: session.id, turn: session.turn, adventureId: session.adventureId },
  }
  const completed = llm.completeJson(request, onChunk)
  console.log("[adventure] interpretFreeform llm result", JSON.stringify({ sessionId: session.id, ok: completed.ok, error: completed.error || "" }))
  if (!completed.ok) {
    store.addAudit({ sessionId: session.id, turn: session.turn, kind: "interpret_action_error", input: { text }, llmRequest: request, llmResponseText: completed.rawText || "", parsed: completed.parsed || {}, validation: { ok: false, errors: [completed.error] } })
    return { ok: false, error: completed.error }
  }
  const validation = validateInterpretedAction(completed.value)
  store.addAudit({ sessionId: session.id, turn: session.turn, kind: "interpret_action", input: { text }, llmRequest: request, llmResponseText: completed.rawText, parsed: completed.value, validation, appliedEffects: validation.action ? validation.action.proposedEffects : {} })
  if (!validation.ok) return { ok: false, error: validation.errors.join("; ") }
  const nextSession = store.advanceSession(session, validation.action.proposedEffects || {})
  const input = { kind: "freeform", text, actor: actor || "", interpreted_action: validation.action, effects: validation.action.proposedEffects || {} }
  const terminal = terminalForDepletedStats(store, nextSession, input)
  if (terminal) return Object.assign({ action: validation.action }, finishResultWithStoryboard(terminal, store))
  return { ok: true, session: nextSession, action: validation.action, input }
}

module.exports = { generateScene, applyChoice, interpretFreeform }
