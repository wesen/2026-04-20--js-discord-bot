const { defineBot } = require("discord")
const poker = require("./lib/poker")

function commandError(message) {
  return {
    content: message,
    ephemeral: true,
  }
}

function helpEmbed() {
  return {
    title: "Poker bot help",
    description: "This bot supports two play styles:\n1. Video poker hand management\n2. Hold'em hand ranking and action advice",
    color: 0x5865F2,
    fields: [
      {
        name: "Commands",
        value: [
          "`/poker-deal` — deal a fresh five-card hand",
          "`/poker-draw keep: 1,3,5` — keep cards and redraw once",
          "`/poker-score` — show your current hand",
          "`/poker-reset` — clear the current hand",
          "`/poker-rank cards: As Kd Qh Jc Tc` — evaluate a 5–7 card hand",
          "`/poker-action hole: As Kd board: Qh Jc Tc 2d 3s` — get action advice",
        ].join("\n"),
        inline: false,
      },
      {
        name: "Examples",
        value: [
          "`/poker-deal`",
          "`/poker-draw keep: 1,5`",
          "`/poker-rank cards: As Kd Qh Jc Tc`",
          "`/poker-action hole: As Kd board: Qh Jc Tc 2d 3s pot: 120 to_call: 30`",
        ].join("\n"),
        inline: false,
      },
      {
        name: "Buttons",
        value: "Use the quick-action buttons below to deal, score, rank, or open the action-advice modal.",
        inline: false,
      },
    ],
  }
}

function helpComponents() {
  return [
    {
      type: "actionRow",
      components: [
        {
          type: "button",
          style: "primary",
          label: "Deal",
          customId: "poker:help:deal",
        },
        {
          type: "button",
          style: "secondary",
          label: "Score",
          customId: "poker:help:score",
        },
        {
          type: "button",
          style: "secondary",
          label: "Rank",
          customId: "poker:help:rank",
        },
        {
          type: "button",
          style: "secondary",
          label: "Action",
          customId: "poker:help:action",
        },
        {
          type: "button",
          style: "danger",
          label: "Reset",
          customId: "poker:help:reset",
        },
      ],
    },
  ]
}

function helpResponse() {
  return {
    content: "Use the buttons for quick actions, or run the slash commands directly.",
    embeds: [helpEmbed()],
    components: helpComponents(),
  }
}

function dealResponse(ctx) {
  const state = poker.pokerState(ctx)
  const hand = poker.startRound(state)
  const best = poker.chooseBestHand(hand)
  return {
    content: poker.renderHandSummary(hand, best) + "\nUse /poker-draw to keep cards and draw replacements.",
    embeds: [{
      title: "Poker hand dealt",
      description: `**Hand**: ${poker.formatHand(hand)}\n**Best hand**: ${best.name}\n**State**: round ${state.get("round", 0)}`,
      color: 0x5865F2,
    }],
  }
}

function scoreResponse(ctx) {
  const summary = poker.stateSummary(poker.pokerState(ctx))
  if (!summary) {
    return commandError("No active hand. Use /poker-deal first.")
  }
  return {
    content: poker.renderHandSummary(summary.hand, summary.best),
    embeds: [{
      title: "Poker hand status",
      description: `**Hand**: ${poker.formatHand(summary.hand)}\n**Best hand**: ${summary.best.name}\n**Draw used**: ${summary.drawn ? "yes" : "no"}\n**Round**: ${summary.round}`,
      color: 0xFEE75C,
    }],
  }
}

function resetResponse(ctx) {
  poker.resetRound(poker.pokerState(ctx))
  return {
    content: "Poker state reset.",
    ephemeral: true,
  }
}

function rankResponse(cardsText) {
  const cards = poker.parseCardList(cardsText)
  if (cards.length < 5) {
    throw new Error("poker-rank expects at least five cards")
  }
  poker.validateNoDuplicateCards(cards)
  const best = poker.chooseBestHand(cards)
  return {
    content: poker.renderHandSummary(best.cards, best),
    embeds: [{
      title: "Poker hand rank",
      description: `**Input**: ${poker.formatHand(cards)}\n**Best hand**: ${best.name}\n**Made hand**: ${poker.formatHand(best.cards)}`,
      color: 0x3498DB,
    }],
  }
}

function actionResponse(input) {
  const result = poker.suggestAction({
    hole: input.hole,
    board: input.board,
    toCall: input.toCall,
    pot: input.pot,
  })
  return {
    content: poker.renderAdvice(result.hole, result.board, result),
    embeds: [{
      title: "Poker action advice",
      description: `**Action**: ${result.action.toUpperCase()}\n**Reason**: ${result.reason}\n**Hole**: ${poker.formatHand(result.hole)}${result.board.length ? `\n**Board**: ${poker.formatHand(result.board)}` : ""}${result.bestHand ? `\n**Best hand**: ${result.bestHand.name}\n**Made hand**: ${poker.formatHand(result.bestHand.cards)}` : ""}`,
      color: 0x9B59B6,
    }],
  }
}

function rankModal() {
  return {
    customId: "poker:rank:submit",
    title: "Rank a poker hand",
    components: [
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "cards",
            label: "Cards",
            style: "paragraph",
            required: true,
            placeholder: "As Kd Qh Jc Tc",
          },
        ],
      },
    ],
  }
}

function actionModal() {
  return {
    customId: "poker:action:submit",
    title: "Hold'em action advice",
    components: [
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "hole",
            label: "Hole cards",
            style: "short",
            required: true,
            placeholder: "As Kd",
          },
        ],
      },
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "board",
            label: "Board cards",
            style: "paragraph",
            required: false,
            placeholder: "Qh Jc Tc 2d 3s",
          },
        ],
      },
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "to_call",
            label: "Amount to call",
            style: "short",
            required: false,
            placeholder: "30",
          },
        ],
      },
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "pot",
            label: "Pot size",
            style: "short",
            required: false,
            placeholder: "120",
          },
        ],
      },
    ],
  }
}

function modalActionResponse(values) {
  return actionResponse({
    hole: values.hole,
    board: values.board,
    toCall: values.to_call,
    pot: values.pot,
  })
}

module.exports = defineBot(({ command, component, event, modal, configure }) => {
  configure({
    name: "poker",
    description: "Video poker hands and hold'em action advice",
    category: "games",
  })

  command("poker-help", {
    description: "Show poker bot commands and examples",
  }, async () => helpResponse())

  command("poker-deal", {
    description: "Deal a fresh five-card poker hand",
  }, async (ctx) => dealResponse(ctx))

  command("poker-draw", {
    description: "Keep selected cards and redraw the rest once",
    options: {
      keep: {
        type: "string",
        description: "Card positions to keep, e.g. 1,3,5",
      },
    },
  }, async (ctx) => {
    try {
      const state = poker.pokerState(ctx)
      const keep = poker.parsePositions(ctx.args.keep)
      const hand = poker.redrawRound(state, keep)
      const best = poker.chooseBestHand(hand)
      return {
        content: poker.renderHandSummary(hand, best) + `\nKept positions: ${poker.prettyPositions(keep)}`,
        embeds: [{
          title: "Poker draw complete",
          description: `**Kept**: ${poker.prettyPositions(keep)}\n**Hand**: ${poker.formatHand(hand)}\n**Best hand**: ${best.name}`,
          color: 0x57F287,
        }],
      }
    } catch (err) {
      return commandError(err.message)
    }
  })

  command("poker-score", {
    description: "Show the current poker hand stored for this channel and user",
  }, async (ctx) => scoreResponse(ctx))

  command("poker-reset", {
    description: "Forget the current poker hand for this channel and user",
  }, async (ctx) => resetResponse(ctx))

  command("poker-rank", {
    description: "Evaluate a five- to seven-card poker hand",
    options: {
      cards: {
        type: "string",
        description: "Cards like As Kd Qh Jc Tc",
        required: true,
      },
    },
  }, async (ctx) => {
    try {
      return rankResponse(ctx.args.cards)
    } catch (err) {
      return commandError(err.message)
    }
  })

  command("poker-action", {
    description: "Recommend a Texas Hold'em action from hole cards and board cards",
    options: {
      hole: {
        type: "string",
        description: "Two hole cards like As Kd",
        required: true,
      },
      board: {
        type: "string",
        description: "Optional board cards like Qh Jc Tc 2d 3s",
      },
      to_call: {
        type: "number",
        description: "Optional amount needed to call",
      },
      pot: {
        type: "number",
        description: "Optional pot size",
      },
    },
  }, async (ctx) => {
    try {
      return actionResponse({
        hole: ctx.args.hole,
        board: ctx.args.board,
        toCall: ctx.args.to_call,
        pot: ctx.args.pot,
      })
    } catch (err) {
      return commandError(err.message)
    }
  })

  component("poker:help:deal", async (ctx) => dealResponse(ctx))

  component("poker:help:score", async (ctx) => scoreResponse(ctx))

  component("poker:help:rank", async (ctx) => {
    await ctx.showModal(rankModal())
  })

  component("poker:help:action", async (ctx) => {
    await ctx.showModal(actionModal())
  })

  component("poker:help:reset", async (ctx) => resetResponse(ctx))

  modal("poker:rank:submit", async (ctx) => {
    try {
      return rankResponse(ctx.values.cards)
    } catch (err) {
      return commandError(err.message)
    }
  })

  modal("poker:action:submit", async (ctx) => {
    try {
      return modalActionResponse(ctx.values)
    } catch (err) {
      return commandError(err.message)
    }
  })

  event("ready", async (ctx) => {
    ctx.log.info("poker bot ready", {
      user: ctx.me && ctx.me.username,
      bot: ctx.metadata && ctx.metadata.name,
    })
  })

  event("messageCreate", async (ctx) => {
    const content = String((ctx.message && ctx.message.content) || "").trim()
    if (content === "!poker") {
      await ctx.reply({
        content: "Poker bot is online. Try /poker-help, /poker-deal, /poker-score, /poker-rank, or /poker-action.",
      })
    }
  })
})
