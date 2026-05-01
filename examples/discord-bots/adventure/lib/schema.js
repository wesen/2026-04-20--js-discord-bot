function trimText(value) {
  return String(value || "").trim()
}

function safeJsonParse(text, fallback) {
  if (text === undefined || text === null || text === "") return fallback
  try {
    return JSON.parse(String(text))
  } catch (_) {
    return fallback
  }
}

function jsonString(value) {
  return JSON.stringify(value === undefined ? null : value)
}

function extractJsonText(text) {
  const raw = String(text || "").trim()
  if (!raw) return ""
  const fenced = raw.match(/```(?:json)?\s*([\s\S]*?)```/i)
  if (fenced) return fenced[1].trim()
  const firstBrace = raw.indexOf("{")
  const lastBrace = raw.lastIndexOf("}")
  if (firstBrace >= 0 && lastBrace > firstBrace) {
    return raw.slice(firstBrace, lastBrace + 1)
  }
  return raw
}

function parseLLMJson(text) {
  const jsonText = extractJsonText(text)
  if (!jsonText) {
    return { ok: false, error: "LLM response was empty", value: null, jsonText: "" }
  }
  try {
    return { ok: true, value: JSON.parse(jsonText), jsonText }
  } catch (err) {
    return { ok: false, error: `Could not parse LLM JSON: ${err.message || err}`, value: null, jsonText }
  }
}

function normalizeChoice(choice, index) {
  const id = trimText(choice && choice.id) || `choice_${index + 1}`
  const label = trimText(choice && choice.label)
  return {
    id: id.replace(/[^a-zA-Z0-9_-]/g, "_").slice(0, 48) || `choice_${index + 1}`,
    label: label.slice(0, 80),
    requires: choice && typeof choice.requires === "object" && choice.requires ? choice.requires : {},
    proposedEffects: choice && typeof choice.proposed_effects === "object" && choice.proposed_effects ? choice.proposed_effects : {},
    nextHint: trimText(choice && choice.next_hint).slice(0, 120),
  }
}

function validateScenePatch(value, seed) {
  const errors = []
  const patch = value && value.scene_patch
  const scene = patch && patch.scene
  if (!scene || typeof scene !== "object") {
    errors.push("Missing scene_patch.scene")
  }
  const constraints = (seed && seed.constraints) || {}
  const minChoices = Number(constraints.minChoices || 2)
  const maxChoices = Number(constraints.maxChoices || 4)
  const title = trimText(scene && scene.title).slice(0, 120)
  const narration = trimText(scene && scene.narration).slice(0, Number(constraints.maxNarrationChars || 900))
  const asciiArt = trimAscii(trimText(scene && scene.ascii_art), Number(constraints.maxAsciiLines || 12))
  const rawChoices = Array.isArray(scene && scene.choices) ? scene.choices : []
  const choices = rawChoices.map(normalizeChoice).filter((choice) => choice.label)
  if (!title) errors.push("Scene title is required")
  if (!narration) errors.push("Scene narration is required")
  if (choices.length < minChoices) errors.push(`At least ${minChoices} choices are required`)
  if (choices.length > maxChoices) choices.length = maxChoices
  return {
    ok: errors.length === 0,
    errors,
    scene: {
      id: trimText(scene && scene.id).slice(0, 80),
      title,
      asciiArt,
      narration,
      choices,
      engineNotes: patch && typeof patch.engine_notes === "object" && patch.engine_notes ? patch.engine_notes : {},
      rawPatch: value || {},
    },
  }
}

function validateInterpretedAction(value) {
  const action = value && value.interpreted_action
  if (!action || typeof action !== "object") {
    return { ok: false, errors: ["Missing interpreted_action"], action: null }
  }
  return {
    ok: true,
    errors: [],
    action: {
      summary: trimText(action.summary).slice(0, 300),
      kind: trimText(action.kind).slice(0, 60),
      target: trimText(action.target).slice(0, 80),
      risk: trimText(action.risk).slice(0, 40),
      matchedChoiceId: trimText(action.matched_choice_id).slice(0, 80),
      proposedEffects: action.proposed_effects && typeof action.proposed_effects === "object" ? action.proposed_effects : {},
      responseHint: trimText(action.response_hint).slice(0, 300),
    },
  }
}

function trimAscii(text, maxLines) {
  const lines = String(text || "").split(/\r?\n/).slice(0, maxLines || 12)
  return lines.map((line) => line.slice(0, 80)).join("\n")
}

module.exports = {
  trimText,
  safeJsonParse,
  jsonString,
  parseLLMJson,
  validateScenePatch,
  validateInterpretedAction,
}
