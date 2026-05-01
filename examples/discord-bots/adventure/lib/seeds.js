const HAUNTED_GATE = {
  id: "haunted-gate",
  title: "The Haunted Gate",
  genre: "gothic fantasy",
  tone: "eerie, concise, mysterious, lightly dangerous",
  initialStats: { hp: 8, sanity: 6 },
  inventoryVocab: ["iron_key", "lantern", "bone_charm", "silver_thread"],
  flagVocab: [
    "opened_gate",
    "spirit_curious",
    "spirit_respected",
    "attempted_key",
    "heard_knocking",
    "entered_courtyard",
  ],
  constraints: {
    maxChoices: 4,
    minChoices: 2,
    maxAsciiLines: 12,
    maxNarrationChars: 900,
    statDeltaMin: -3,
    statDeltaMax: 3,
  },
  openingPrompt: "The player stands before an old iron gate in cold rain. Something behind the gate knocks three times, then whispers as if it already knows the player's name.",
}

function allSeeds() {
  return [HAUNTED_GATE]
}

function findSeed(id) {
  const needle = String(id || "haunted-gate").trim() || "haunted-gate"
  return allSeeds().find((seed) => seed.id === needle) || HAUNTED_GATE
}

module.exports = { HAUNTED_GATE, allSeeds, findSeed }
