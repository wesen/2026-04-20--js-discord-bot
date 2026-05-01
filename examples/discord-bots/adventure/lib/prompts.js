function sceneSystemPrompt() {
  return [
    "You generate content for a Discord ASCII choose-your-own-adventure game.",
    "Return only valid JSON. Do not wrap it in Markdown unless unavoidable.",
    "The engine owns canonical state. You may propose effects, but do not claim they are applied.",
    "Keep scenes concise and Discord-friendly.",
    "Return 2 to 4 concrete choices.",
    "Use small ASCII art, at most 12 lines, at most 80 columns.",
  ].join("\n")
}

function sceneUserPrompt({ seed, session, currentScene, input, recentHistory }) {
  return JSON.stringify({
    task: "Generate the next scene_patch JSON object.",
    schema: {
      scene_patch: {
        scene: {
          id: "short-stable-scene-id",
          title: "Scene title",
          ascii_art: "small ASCII art",
          narration: "short second-person scene narration",
          choices: [
            {
              id: "choice_id",
              label: "Button label",
              requires: {},
              proposed_effects: { stats: {}, flags: {}, add_inventory: [] },
              next_hint: "what this points toward",
            },
          ],
        },
        engine_notes: { mood: "", continuity: "" },
      },
    },
    seed,
    session,
    current_scene: currentScene || null,
    player_input: input,
    recent_history: recentHistory || [],
  }, null, 2)
}

function actionSystemPrompt() {
  return [
    "You interpret free-form player input for a Discord choose-your-own-adventure game.",
    "Return only valid JSON. Do not apply state changes; propose effects only.",
  ].join("\n")
}

function actionUserPrompt({ seed, session, currentScene, text }) {
  return JSON.stringify({
    task: "Interpret the player's free-form action as interpreted_action JSON.",
    schema: {
      interpreted_action: {
        summary: "short summary",
        kind: "dialogue|movement|inspection|item_use|other",
        target: "target if any",
        risk: "low|medium|high",
        matched_choice_id: "optional existing choice id",
        proposed_effects: { stats: {}, flags: {}, add_inventory: [] },
        response_hint: "what the next scene should acknowledge",
      },
    },
    seed,
    session,
    current_scene: currentScene || null,
    player_text: text,
  }, null, 2)
}

module.exports = { sceneSystemPrompt, sceneUserPrompt, actionSystemPrompt, actionUserPrompt }
