const { defineBot } = require("discord")
const ui = require("ui")
const { createStore } = require("./lib/store")
const engine = require("./lib/engine")
const render = require("./lib/render")
const llm = require("./lib/llm")

const store = createStore()

function userId(ctx) {
  return String((ctx.user && ctx.user.id) || "")
}

function channelId(ctx) {
  return String((ctx.channel && ctx.channel.id) || (ctx.interaction && ctx.interaction.channelID) || "")
}

function guildId(ctx) {
  return String((ctx.guild && ctx.guild.id) || (ctx.interaction && ctx.interaction.guildID) || "")
}

function ensureStore(ctx) {
  console.log("[adventure] ensureStore", JSON.stringify({ config: ctx.config || {} }))
  store.ensure(ctx.config || {})
}

function messageTurn(ctx) {
  const content = String((ctx.message && ctx.message.content) || "")
  const match = content.match(/Turn\s+(\d+)/i)
  return match ? Number(match[1]) : null
}

function makeProgressEditor(ctx, session, title, details) {
  let lastLength = 0
  let edits = 0
  return (event) => {
    const text = String((event && event.text) || "")
    if (!text || event.done) return
    if (edits >= 8 || text.length - lastLength < 160) return
    lastLength = text.length
    edits += 1
    const nextDetails = Object.assign({}, details || {}, { streamText: text })
    const message = nextDetails.scene
      ? render.pendingActionMessage(session, nextDetails.scene, nextDetails)
      : render.loadingMessage(session, title, nextDetails)
    ctx.edit(message)
  }
}

function requireOwnedSession(ctx, options) {
  ensureStore(ctx)
  const session = options && options.allowCompleted ? store.findLatestSessionInChannel(channelId(ctx)) : store.findActiveSessionInChannel(channelId(ctx))
  if (!session) return { ok: false, error: "No adventure session in this channel. Use /adventure-start first." }
  // Adventure sessions are channel-scoped collaborative games. The starter is
  // recorded as owner for reset/debug provenance, but other channel members may
  // choose actions and submit free-form moves.
  if (options && options.rejectStaleMessage) {
    const turn = messageTurn(ctx)
    if (turn !== null && turn !== Number(session.turn || 0)) {
      return { ok: false, error: "That scene is stale. Use /adventure-resume to see the latest scene." }
    }
  }
  const seed = store.getSeed(session.adventureId)
  const scene = store.getCurrentScene(session)
  return { ok: true, session, seed, scene }
}

async function showHistory(ctx, direction) {
  console.log("[adventure] history", JSON.stringify({ userId: userId(ctx), channelId: channelId(ctx), direction }))
  const loaded = requireOwnedSession(ctx, { allowCompleted: true })
  if (!loaded.ok) return render.errorMessage(loaded.error)
  const currentTurn = messageTurn(ctx)
  let targetTurn = loaded.session.turn
  if (currentTurn !== null) {
    if (direction === "prev") targetTurn = Math.max(0, currentTurn - 1)
    if (direction === "next") targetTurn = Math.min(loaded.session.turn, currentTurn + 1)
    if (direction === "current") targetTurn = loaded.session.turn
  }
  const scene = store.getSceneByTurn(loaded.session, targetTurn) || loaded.scene
  return render.sceneMessage(loaded.session, scene, { history: targetTurn !== loaded.session.turn })
}

function knownSecondaryStat(prompt) {
  const text = String(prompt || "").toLowerCase()
  if (/disco|party|dance|club|funk|groove|music|dj/.test(text)) return "groove"
  if (/space|star|cosmic|alien|ship|planet|underwater|undersea|ocean|sea|deep|submarine|diving|dive|reef|aquatic/.test(text)) return "oxygen"
  if (/cozy|cat|bakery|tea|garden|wholesome/.test(text)) return "comfort"
  if (/detective|mystery|noir|case|clue/.test(text)) return "focus"
  if (/pirate|ship|island|treasure/.test(text)) return "swagger"
  if (/wizard|magic|spell|arcane|fantasy/.test(text)) return "mana"
  if (/robot|cyber|neon|hacker|computer/.test(text)) return "signal"
  return ""
}

function cleanStatName(value) {
  const cleaned = String(value || "").toLowerCase().replace(/[^a-z0-9_ -]/g, "").trim().replace(/[ -]+/g, "_").slice(0, 18)
  if (!cleaned || cleaned === "hp" || cleaned === "health") return "resolve"
  return cleaned
}

function themedSecondaryStat(prompt) {
  const known = knownSecondaryStat(prompt)
  if (known) return known
  const completed = llm.completeJson({
    purpose: "secondary_stat_name",
    system: "Return only valid JSON. Choose one evocative non-HP resource/stat name for a lightweight adventure game.",
    user: JSON.stringify({
      task: "Name the secondary stat for this adventure premise. It should be one short word, coherent with the premise, and not HP/health.",
      premise: String(prompt || "").slice(0, 1000),
      schema: { stat: "one lowercase word, e.g. oxygen, mana, focus, groove" },
    }),
    metadata: { promptKind: "adventure-start" },
  })
  if (completed.ok && completed.value && completed.value.stat) return cleanStatName(completed.value.stat)
  console.log("[adventure] secondary stat LLM fallback failed", JSON.stringify({ error: completed.error || "unknown" }))
  return "resolve"
}

async function regenerateStoryboard(ctx) {
  console.log("[adventure] regenerateStoryboard", JSON.stringify({ userId: userId(ctx), channelId: channelId(ctx) }))
  const loaded = requireOwnedSession(ctx, { allowCompleted: true })
  if (!loaded.ok) return render.errorMessage(loaded.error)
  if (loaded.session.status !== "completed") return render.errorMessage("The latest adventure in this channel has not reached a coda yet.")
  await ctx.defer({ ephemeral: false })
  await ctx.edit({ content: "Generating a storyboard from the completed adventure..." })
  const result = engine.regenerateStoryboard(store, loaded.session)
  if (!result.ok) {
    await ctx.edit(render.errorMessage(`Could not generate storyboard: ${result.error}`))
    return
  }
  await ctx.edit(render.storyboardMessage(loaded.session, result.storyboard))
}

async function startAdventure(ctx) {
  console.log("[adventure] startAdventure", JSON.stringify({ userId: userId(ctx), channelId: channelId(ctx), args: ctx.args || {} }))
  ensureStore(ctx)
  await ctx.defer({ ephemeral: false })
  const seedId = String((ctx.args && ctx.args.seed) || "haunted-gate").trim() || "haunted-gate"
  const mode = String((ctx.args && ctx.args.mode) || "party").trim() || "party"
  const userSeed = String((ctx.args && (ctx.args.prompt || ctx.args.info || ctx.args.seed_info)) || "").trim().slice(0, 1000)
  const seed = store.getSeed(seedId)
  if (!seed) {
    await ctx.edit(render.errorMessage(`Unknown adventure seed: ${seedId}`))
    return
  }
  if (userSeed) {
    seed.initialStats = Object.assign({}, seed.initialStats || {}, { [themedSecondaryStat(userSeed)]: Number((seed.initialStats && seed.initialStats.sanity) || 6) })
    delete seed.initialStats.sanity
  }
  store.resetActive(userId(ctx), channelId(ctx))
  const session = store.createSession({ seed, ownerUserId: userId(ctx), guildId: guildId(ctx), channelId: channelId(ctx), mode, userTheme: userSeed })
  console.log("[adventure] session created", JSON.stringify({ sessionId: session.id, seedId: seed.id, turn: session.turn }))
  await ctx.edit(render.loadingMessage(session, "Opening the gate..."))
  const generated = engine.generateScene({
    store,
    seed,
    session,
    currentScene: null,
    input: { kind: "start", opening_prompt: userSeed || seed.openingPrompt, user_seed: userSeed, override_seed_tone: Boolean(userSeed), actor: (ctx.user && (ctx.user.username || ctx.user.id)) || userId(ctx) },
    onChunk: makeProgressEditor(ctx, session, "Opening the gate...", { actor: (ctx.user && (ctx.user.username || ctx.user.id)) || userId(ctx) }),
  })
  if (!generated.ok) {
    console.log("[adventure] opening scene generation failed", JSON.stringify({ sessionId: session.id, error: generated.error }))
    await ctx.edit(render.errorMessage(`Could not generate opening scene: ${generated.error}`))
    return
  }
  console.log("[adventure] opening scene generated", JSON.stringify({ sessionId: session.id, sceneId: generated.scene && generated.scene.id }))
  await ctx.edit(render.sceneMessage(generated.session || session, generated.scene, { exported: generated.exported, storyboard: generated.storyboard }))
}

async function choose(ctx, index) {
  console.log("[adventure] choose", JSON.stringify({ userId: userId(ctx), channelId: channelId(ctx), index }))
  const loaded = requireOwnedSession(ctx, { rejectStaleMessage: true })
  if (!loaded.ok) {
    console.log("[adventure] choose rejected", JSON.stringify({ index, error: loaded.error }))
    return render.errorMessage(loaded.error)
  }
  await ctx.defer({ ephemeral: false })
  const pendingChoice = loaded.scene && loaded.scene.choices ? loaded.scene.choices[index] : null
  await ctx.edit(render.pendingActionMessage(loaded.session, loaded.scene, { action: pendingChoice && pendingChoice.label, actor: (ctx.user && (ctx.user.username || ctx.user.id)) || userId(ctx) }))
  const actor = (ctx.user && (ctx.user.username || ctx.user.id)) || userId(ctx)
  const applied = engine.applyChoice(store, loaded.session, loaded.scene, index, actor)
  if (!applied.ok) {
    await ctx.edit(render.errorMessage(applied.error))
    return
  }
  if (applied.scene) {
    await ctx.edit(render.sceneMessage(applied.session, applied.scene, { exported: applied.exported, storyboard: applied.storyboard }))
    return
  }
  const generated = engine.generateScene({
    store,
    seed: loaded.seed,
    session: applied.session,
    currentScene: loaded.scene,
    input: applied.input,
    onChunk: makeProgressEditor(ctx, applied.session, "Resolving your choice...", { scene: loaded.scene, action: pendingChoice && pendingChoice.label, actor: (ctx.user && (ctx.user.username || ctx.user.id)) || userId(ctx) }),
  })
  if (!generated.ok) {
    await ctx.edit(render.errorMessage(`Could not generate next scene: ${generated.error}`))
    return
  }
  await ctx.edit(render.sceneMessage(generated.session || applied.session, generated.scene, { exported: generated.exported, storyboard: generated.storyboard }))
}

module.exports = defineBot(({ command, component, modal, event, configure }) => {
  configure({
    name: "adventure",
    description: "ASCII choose-your-own-adventure bot with Go-owned OpenRouter generation",
    run: {
      storage: {
        title: "Storage",
        fields: {
          sessionDbPath: {
            type: "string",
            help: "SQLite database path for adventure sessions/scenes",
            default: "./examples/discord-bots/adventure/data/adventure.sqlite",
          },
          debugYaml: {
            type: "bool",
            help: "Legacy debug flag; JSON state is used internally",
            default: false,
          },
        },
      },
    },
  })

  event("ready", async (ctx) => {
    console.log("[adventure] ready", JSON.stringify({ metadata: ctx.metadata || {} }))
    ctx.log.info("adventure bot ready", { bot: ctx.metadata && ctx.metadata.name })
  })

  command("adventure-start", {
    description: "Start a new ASCII adventure session",
    options: {
      seed: { type: "string", description: "Adventure seed ID", required: false },
      mode: { type: "string", description: "Play mode: party", required: false },
      prompt: { type: "string", description: "Optional premise, character, goal, or starting context", required: false },
    },
  }, startAdventure)

  command("adventure-resume", {
    description: "Resume the current adventure session in this channel",
  }, async (ctx) => {
    const loaded = requireOwnedSession(ctx, { allowCompleted: true })
    if (!loaded.ok) return render.errorMessage(loaded.error)
    return render.sceneMessage(loaded.session, loaded.scene || { title: "Adventure", narration: "No scene has been generated yet.", choices: [] })
  })

  command("adventure-storyboard", {
    description: "Regenerate a storyboard image for the latest completed adventure in this channel",
  }, regenerateStoryboard)

  command("adventure-state", {
    description: "Show debug state for your active adventure session",
  }, async (ctx) => {
    const loaded = requireOwnedSession(ctx)
    if (!loaded.ok) return render.errorMessage(loaded.error)
    return render.stateMessage(loaded.session, loaded.scene)
  })

  command("adventure-reset", {
    description: "Abandon your active adventure session in this channel",
  }, async (ctx) => {
    ensureStore(ctx)
    store.resetActive(userId(ctx), channelId(ctx))
    return ui.message().ephemeral().content("Adventure session reset.").build()
  })

  component("adv:choice:0", async (ctx) => choose(ctx, 0))
  component("adv:choice:1", async (ctx) => choose(ctx, 1))
  component("adv:choice:2", async (ctx) => choose(ctx, 2))
  component("adv:choice:3", async (ctx) => choose(ctx, 3))
  component("adv:history:prev", async (ctx) => showHistory(ctx, "prev"))
  component("adv:history:next", async (ctx) => showHistory(ctx, "next"))
  component("adv:history:current", async (ctx) => showHistory(ctx, "current"))

  component("adv:freeform", async (ctx) => {
    console.log("[adventure] freeform button", JSON.stringify({ userId: userId(ctx), channelId: channelId(ctx) }))
    const loaded = requireOwnedSession(ctx, { rejectStaleMessage: true })
    if (!loaded.ok) return render.errorMessage(loaded.error)
    await ctx.showModal(
      ui.form("adv:modal:freeform", "Try something else")
        .textarea("action", "What do you try?").required().min(2).max(800)
        .build()
    )
  })

  modal("adv:modal:freeform", async (ctx) => {
    console.log("[adventure] freeform modal", JSON.stringify({ userId: userId(ctx), channelId: channelId(ctx), values: ctx.values || {} }))
    const loaded = requireOwnedSession(ctx)
    if (!loaded.ok) return render.errorMessage(loaded.error)
    await ctx.defer({ ephemeral: false })
    const text = String((ctx.values || {}).action || "").trim()
    await ctx.edit(render.pendingActionMessage(loaded.session, loaded.scene, { action: text || "free-form action", actor: (ctx.user && (ctx.user.username || ctx.user.id)) || userId(ctx) }))
    if (!text) {
      await ctx.edit(render.errorMessage("Free-form action cannot be empty."))
      return
    }
    const freeformDetails = { scene: loaded.scene, action: text || "free-form action", actor: (ctx.user && (ctx.user.username || ctx.user.id)) || userId(ctx) }
    const interpreted = engine.interpretFreeform({ store, seed: loaded.seed, session: loaded.session, currentScene: loaded.scene, text, actor: (ctx.user && (ctx.user.username || ctx.user.id)) || userId(ctx), onChunk: makeProgressEditor(ctx, loaded.session, "Interpreting your action...", freeformDetails) })
    if (!interpreted.ok) {
      await ctx.edit(render.errorMessage(`Could not interpret action: ${interpreted.error}`))
      return
    }
    if (interpreted.scene) {
      await ctx.edit(render.sceneMessage(interpreted.session, interpreted.scene, { exported: interpreted.exported, storyboard: interpreted.storyboard }))
      return
    }
    const generated = engine.generateScene({
      store,
      seed: loaded.seed,
      session: interpreted.session,
      currentScene: loaded.scene,
      input: interpreted.input,
      onChunk: makeProgressEditor(ctx, interpreted.session, "Trying something else...", freeformDetails),
    })
    if (!generated.ok) {
      await ctx.edit(render.errorMessage(`Could not generate next scene: ${generated.error}`))
      return
    }
    await ctx.edit(render.sceneMessage(generated.session || interpreted.session, generated.scene, { exported: generated.exported, storyboard: generated.storyboard }))
  })
})
