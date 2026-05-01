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
  const scene = store.saveScene(session, validation.scene)
  const finalSession = scene.ending && scene.ending.isFinal ? store.finishSession(session) : session
  const exported = scene.ending && scene.ending.isFinal ? store.exportSession(finalSession) : null
  return { ok: true, scene, session: finalSession, exported }
}

function applyChoice(store, session, scene, choiceIndex) {
  const choice = scene && scene.choices ? scene.choices[choiceIndex] : null
  if (!choice) return { ok: false, error: "That choice is no longer available." }
  const nextSession = store.advanceSession(session, choice.proposedEffects || {})
  return { ok: true, session: nextSession, input: { kind: "choice", choice_id: choice.id, label: choice.label, effects: choice.proposedEffects || {}, next_hint: choice.nextHint || "" } }
}

function interpretFreeform({ store, seed, session, currentScene, text, onChunk }) {
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
  return { ok: true, session: nextSession, action: validation.action, input: { kind: "freeform", text, interpreted_action: validation.action, effects: validation.action.proposedEffects || {} } }
}

module.exports = { generateScene, applyChoice, interpretFreeform }
