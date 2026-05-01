function sceneSystemPrompt() {
  return [
    "You generate content for a Discord ASCII choose-your-own-adventure game.",
    "Return only valid JSON. Do not wrap it in Markdown unless unavoidable.",
    "The engine owns canonical state. You may propose effects, but do not claim they are applied.",
    "Keep scenes concise and Discord-friendly.",
    "Return 2 to 4 concrete choices unless the story has reached a logical ending.",
    "When the story should end, set scene_patch.ending.is_final=true and include a concise ending summary.",
    "Use small ASCII art, at most 12 lines, at most 80 columns.",
  ].join("\n")
}

function sceneUserPrompt({ seed, session, currentScene, input, recentHistory }) {
  const userTheme = (input && input.user_seed) || (session && session.flags && session.flags.user_theme) || ""
  const seedForPrompt = userTheme
    ? Object.assign({}, seed || {}, {
        genre: userTheme,
        tone: userTheme,
        openingPrompt: userTheme,
        original_seed_id: seed && seed.id,
      })
    : seed
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
        ending: { is_final: false, summary: "Only set when the adventure reaches a logical ending." },
        engine_notes: { mood: "", continuity: "" },
      },
    },
    seed: seedForPrompt,
    user_starting_context_policy: "If user_starting_context or session.flags.user_theme is present, treat it as the primary premise, genre, vocabulary, and tone for EVERY turn. Re-theme the adventure around it even if it conflicts with the default seed genre/tone/opening prompt. Do not drift back to the default Haunted Gate/gothic horror framing unless the user theme asks for that. Keep only the engine constraints and safety boundaries from the seed.",
    session,
    current_scene: currentScene || null,
    player_input: input,
    user_starting_context: userTheme,
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
