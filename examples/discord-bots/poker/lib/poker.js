const SUIT_SYMBOLS = {
  c: "♣",
  d: "♦",
  h: "♥",
  s: "♠",
}

const VALUE_TO_RANK = {
  2: "2",
  3: "3",
  4: "4",
  5: "5",
  6: "6",
  7: "7",
  8: "8",
  9: "9",
  10: "T",
  11: "J",
  12: "Q",
  13: "K",
  14: "A",
}

const RANK_TO_VALUE = {
  "2": 2,
  "3": 3,
  "4": 4,
  "5": 5,
  "6": 6,
  "7": 7,
  "8": 8,
  "9": 9,
  "T": 10,
  "J": 11,
  "Q": 12,
  "K": 13,
  "A": 14,
}

const HAND_NAMES = [
  "High card",
  "One pair",
  "Two pair",
  "Three of a kind",
  "Straight",
  "Flush",
  "Full house",
  "Four of a kind",
  "Straight flush",
]

function makeDeck() {
  const suits = ["c", "d", "h", "s"]
  const ranks = ["2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"]
  const deck = []
  for (let i = 0; i < suits.length; i += 1) {
    for (let j = 0; j < ranks.length; j += 1) {
      deck.push(makeCard(ranks[j], suits[i]))
    }
  }
  return deck
}

function shuffle(deck) {
  const copy = deck.slice()
  for (let i = copy.length - 1; i > 0; i -= 1) {
    const j = Math.floor(Math.random() * (i + 1))
    const tmp = copy[i]
    copy[i] = copy[j]
    copy[j] = tmp
  }
  return copy
}

function makeCard(rank, suit) {
  const normalizedRank = String(rank || "").toUpperCase()
  const normalizedSuit = String(suit || "").toLowerCase()
  if (!RANK_TO_VALUE[normalizedRank]) {
    throw new Error(`Unsupported rank ${rank}`)
  }
  if (!SUIT_SYMBOLS[normalizedSuit]) {
    throw new Error(`Unsupported suit ${suit}`)
  }
  return {
    rank: normalizedRank,
    suit: normalizedSuit,
    value: RANK_TO_VALUE[normalizedRank],
    text: `${normalizedRank}${SUIT_SYMBOLS[normalizedSuit]}`,
    code: `${normalizedRank}${normalizedSuit}`,
  }
}

function normalizeUnicodeSuits(text) {
  return String(text || "")
    .replace(/♠/g, "S")
    .replace(/♣/g, "C")
    .replace(/♥/g, "H")
    .replace(/♦/g, "D")
}

function parseCardToken(token) {
  let text = normalizeUnicodeSuits(token).trim()
  if (!text) {
    throw new Error("Empty card token")
  }
  text = text.replace(/\s+/g, "")
  text = text.toUpperCase()
  if (text.indexOf("10") === 0) {
    text = "T" + text.slice(2)
  }
  if (text.length !== 2) {
    throw new Error(`Invalid card token ${token}`)
  }
  const rank = text.charAt(0)
  const suit = text.charAt(1).toLowerCase()
  return makeCard(rank, suit)
}

function parseCardList(text) {
  if (Array.isArray(text)) {
    const cards = []
    for (let i = 0; i < text.length; i += 1) {
      cards.push(parseCardValue(text[i]))
    }
    return cards
  }
  const raw = String(text || "").trim()
  if (!raw) {
    return []
  }
  const tokens = raw.split(/[\s,]+/)
  const cards = []
  for (let i = 0; i < tokens.length; i += 1) {
    if (!tokens[i]) {
      continue
    }
    cards.push(parseCardToken(tokens[i]))
  }
  return cards
}

function parseCardValue(value) {
  if (!value) {
    throw new Error("Empty card value")
  }
  if (typeof value === "string") {
    return parseCardToken(value)
  }
  if (typeof value === "object" && value.rank && value.suit) {
    return makeCard(value.rank, value.suit)
  }
  if (typeof value === "object" && value.code) {
    return parseCardToken(value.code)
  }
  throw new Error(`Unsupported stored card value ${String(value)}`)
}

function parseStoredCards(value) {
  if (!value) {
    return []
  }
  if (Array.isArray(value)) {
    const cards = []
    for (let i = 0; i < value.length; i += 1) {
      cards.push(parseCardValue(value[i]))
    }
    return cards
  }
  return parseCardList(value)
}

function cardKey(card) {
  return `${card.rank}${card.suit}`
}

function cardText(card) {
  return card.text || `${card.rank}${SUIT_SYMBOLS[card.suit] || card.suit}`
}

function formatHand(cards) {
  const out = []
  for (let i = 0; i < cards.length; i += 1) {
    out.push(cardText(cards[i]))
  }
  return out.join(" ")
}

function countByValue(cards) {
  const counts = {}
  for (let i = 0; i < cards.length; i += 1) {
    const value = cards[i].value
    counts[value] = (counts[value] || 0) + 1
  }
  return counts
}

function uniqueSortedValues(cards) {
  const values = []
  for (let i = 0; i < cards.length; i += 1) {
    values.push(cards[i].value)
  }
  values.sort(function (a, b) { return a - b })
  const unique = []
  for (let i = 0; i < values.length; i += 1) {
    if (i === 0 || values[i] !== values[i - 1]) {
      unique.push(values[i])
    }
  }
  return unique
}

function isFlush(cards) {
  if (!cards.length) {
    return false
  }
  const suit = cards[0].suit
  for (let i = 1; i < cards.length; i += 1) {
    if (cards[i].suit !== suit) {
      return false
    }
  }
  return true
}

function straightHigh(cards) {
  const unique = uniqueSortedValues(cards)
  if (unique.length !== 5) {
    return 0
  }
  if (unique[4] - unique[0] === 4) {
    return unique[4]
  }
  if (unique[0] === 2 && unique[1] === 3 && unique[2] === 4 && unique[3] === 5 && unique[4] === 14) {
    return 5
  }
  return 0
}

function sortGroups(counts) {
  const values = []
  for (const key in counts) {
    if (Object.prototype.hasOwnProperty.call(counts, key)) {
      values.push(Number(key))
    }
  }
  values.sort(function (a, b) {
    const countDiff = counts[b] - counts[a]
    if (countDiff !== 0) {
      return countDiff
    }
    return b - a
  })
  return values.map(function (value) {
    return { value: value, count: counts[value] }
  })
}

function evaluateFive(cards) {
  if (!cards || cards.length !== 5) {
    throw new Error("evaluateFive expects exactly five cards")
  }
  const sorted = cards.slice().sort(function (a, b) { return b.value - a.value })
  const counts = countByValue(sorted)
  const groups = sortGroups(counts)
  const flush = isFlush(sorted)
  const straight = straightHigh(sorted)
  let category = 0
  let tiebreakers = []

  if (straight && flush) {
    category = 8
    tiebreakers = [straight]
  } else if (groups[0].count === 4) {
    category = 7
    tiebreakers = [groups[0].value, groups[1].value]
  } else if (groups[0].count === 3 && groups[1].count === 2) {
    category = 6
    tiebreakers = [groups[0].value, groups[1].value]
  } else if (flush) {
    category = 5
    tiebreakers = sorted.map(function (card) { return card.value })
  } else if (straight) {
    category = 4
    tiebreakers = [straight]
  } else if (groups[0].count === 3) {
    category = 3
    tiebreakers = [groups[0].value]
    for (let i = 0; i < sorted.length; i += 1) {
      if (sorted[i].value !== groups[0].value) {
        tiebreakers.push(sorted[i].value)
      }
    }
  } else if (groups[0].count === 2 && groups[1].count === 2) {
    category = 2
    const highPair = Math.max(groups[0].value, groups[1].value)
    const lowPair = Math.min(groups[0].value, groups[1].value)
    let kicker = 0
    for (let i = 0; i < sorted.length; i += 1) {
      const value = sorted[i].value
      if (value !== highPair && value !== lowPair) {
        kicker = value
        break
      }
    }
    tiebreakers = [highPair, lowPair, kicker]
  } else if (groups[0].count === 2) {
    category = 1
    tiebreakers = [groups[0].value]
    for (let i = 0; i < sorted.length; i += 1) {
      if (sorted[i].value !== groups[0].value) {
        tiebreakers.push(sorted[i].value)
      }
    }
  } else {
    category = 0
    tiebreakers = sorted.map(function (card) { return card.value })
  }

  return {
    category: category,
    name: HAND_NAMES[category],
    score: [category].concat(tiebreakers),
    cards: sorted,
  }
}

function compareScores(left, right) {
  const maxLen = Math.max(left.length, right.length)
  for (let i = 0; i < maxLen; i += 1) {
    const a = left[i] || 0
    const b = right[i] || 0
    if (a > b) {
      return 1
    }
    if (a < b) {
      return -1
    }
  }
  return 0
}

function chooseBestHand(cards) {
  if (!cards || cards.length < 5) {
    throw new Error("chooseBestHand expects at least five cards")
  }
  if (cards.length === 5) {
    return evaluateFive(cards)
  }
  let best = null
  const n = cards.length
  for (let a = 0; a < n - 4; a += 1) {
    for (let b = a + 1; b < n - 3; b += 1) {
      for (let c = b + 1; c < n - 2; c += 1) {
        for (let d = c + 1; d < n - 1; d += 1) {
          for (let e = d + 1; e < n; e += 1) {
            const candidate = evaluateFive([cards[a], cards[b], cards[c], cards[d], cards[e]])
            if (!best || compareScores(candidate.score, best.score) > 0) {
              best = candidate
            }
          }
        }
      }
    }
  }
  return best
}

function parsePositions(text) {
  const raw = String(text || "").trim()
  if (!raw) {
    return []
  }
  const tokens = raw.split(/[\s,]+/)
  const positions = []
  for (let i = 0; i < tokens.length; i += 1) {
    if (!tokens[i]) {
      continue
    }
    const parsed = Number(tokens[i])
    if (!isFinite(parsed) || Math.floor(parsed) !== parsed) {
      throw new Error(`Invalid hold position ${tokens[i]}`)
    }
    if (parsed < 1 || parsed > 5) {
      throw new Error(`Hold positions must be between 1 and 5, got ${parsed}`)
    }
    if (positions.indexOf(parsed) === -1) {
      positions.push(parsed)
    }
  }
  positions.sort(function (a, b) { return a - b })
  return positions
}

function prettyPositions(positions) {
  if (!positions.length) {
    return "none"
  }
  return positions.join(", ")
}

function extractCardsFromStore(raw) {
  return parseStoredCards(raw)
}

function validateNoDuplicateCards(cards) {
  const seen = {}
  for (let i = 0; i < cards.length; i += 1) {
    const key = cardKey(cards[i])
    if (seen[key]) {
      throw new Error(`Duplicate card detected: ${cardText(cards[i])}`)
    }
    seen[key] = true
  }
}

function holeStrength(hole) {
  const cards = hole.slice().sort(function (a, b) { return b.value - a.value })
  const high = cards[0]
  const low = cards[1]
  const pair = high.value === low.value
  const suited = high.suit === low.suit
  const gap = high.value - low.value

  if (pair) {
    if (high.value >= 10) {
      return { action: "raise", reason: `Premium pocket pair ${high.rank}${high.rank}` }
    }
    if (high.value >= 7) {
      return { action: "call", reason: `Middle pocket pair ${high.rank}${high.rank}` }
    }
    return { action: "call", reason: `Small pocket pair ${high.rank}${high.rank} can set-mine` }
  }

  if (high.value === 14 && low.value >= 10 && suited) {
    return { action: "raise", reason: `Ace-high suited Broadway cards` }
  }
  if (high.value >= 13 && low.value >= 11) {
    return { action: "call", reason: `Two strong Broadway cards` }
  }
  if (suited && gap <= 1 && high.value >= 10) {
    return { action: "call", reason: `Suited connectors can play well post-flop` }
  }
  if (suited && high.value >= 11) {
    return { action: "call", reason: `Suited high cards have decent playability` }
  }
  if (high.value >= 14 && low.value >= 8) {
    return { action: "call", reason: `Ace-high with a reasonable kicker` }
  }
  return { action: "fold", reason: `Unconnected offsuit hand` }
}

function boardStrength(bestHand) {
  if (!bestHand) {
    return { action: "check", reason: "No made hand yet" }
  }
  switch (bestHand.category) {
    case 8:
    case 7:
    case 6:
    case 5:
    case 4:
      return { action: "raise", reason: `Made ${bestHand.name.toLowerCase()}` }
    case 3:
      return { action: "call", reason: `Trips are strong showdown value` }
    case 2:
      return { action: "call", reason: `Two pair has solid value` }
    case 1:
      if (bestHand.score[1] >= 11) {
        return { action: "call", reason: `Top pair or better` }
      }
      return { action: "check", reason: `Weak one pair` }
    default:
      return { action: "fold", reason: `Just a high card` }
  }
}

function suggestAction(input) {
  const hole = parseCardList(input.hole)
  if (hole.length !== 2) {
    throw new Error("poker-action expects exactly two hole cards")
  }
  validateNoDuplicateCards(hole)

  const board = parseCardList(input.board)
  validateNoDuplicateCards(hole.concat(board))

  const toCall = numberValue(input.toCall, 0)
  const pot = numberValue(input.pot, 0)

  if (board.length < 3) {
    const preflop = holeStrength(hole)
    const reason = board.length === 0
      ? preflop.reason
      : `${preflop.reason} (board is still developing)`
    return {
      action: preflop.action,
      reason: reason,
      hole: hole,
      board: board,
      bestHand: null,
      toCall: toCall,
      pot: pot,
    }
  }

  const bestHand = chooseBestHand(hole.concat(board))
  const postflop = boardStrength(bestHand)
  let action = postflop.action
  let reason = postflop.reason

  if (toCall > 0 && pot > 0) {
    const pressure = toCall / Math.max(pot, 1)
    if (action === "call" && pressure > 0.35 && bestHand.category <= 2) {
      action = "fold"
      reason = "The price is too high for a marginal made hand"
    }
    if (action === "check" && pressure <= 0.2 && bestHand.category <= 1) {
      action = "call"
      reason = "Cheap call with a modest made hand"
    }
  }

  return {
    action: action,
    reason: reason,
    hole: hole,
    board: board,
    bestHand: bestHand,
    toCall: toCall,
    pot: pot,
  }
}

function numberValue(value, defaultValue) {
  if (value === null || typeof value === "undefined" || value === "") {
    return defaultValue
  }
  const parsed = Number(value)
  if (!isFinite(parsed)) {
    return defaultValue
  }
  return parsed
}

function startRound(state) {
  const deck = shuffle(makeDeck())
  const hand = deck.slice(0, 5)
  const remaining = deck.slice(5)
  state.set("hand", hand.map(cardKey))
  state.set("deck", remaining.map(cardKey))
  state.set("drawn", false)
  state.set("round", numberValue(state.get("round", 0), 0) + 1)
  return hand
}

function currentHand(state) {
  return extractCardsFromStore(state.get("hand", []))
}

function currentDeck(state) {
  return extractCardsFromStore(state.get("deck", []))
}

function resetRound(state) {
  state.delete("hand")
  state.delete("deck")
  state.delete("drawn")
}

function redrawRound(state, keepPositions) {
  const hand = currentHand(state)
  const deck = currentDeck(state)
  if (hand.length !== 5) {
    throw new Error("No active poker hand. Use /poker-deal first.")
  }
  if (state.get("drawn", false)) {
    throw new Error("This round has already used its draw. Start a new hand with /poker-deal.")
  }
  if (keepPositions.length === 0) {
    keepPositions = []
  }
  const keep = {}
  for (let i = 0; i < keepPositions.length; i += 1) {
    keep[keepPositions[i]] = true
  }
  const nextHand = []
  let deckIndex = 0
  for (let position = 1; position <= 5; position += 1) {
    if (keep[position]) {
      nextHand.push(hand[position - 1])
      continue
    }
    if (deckIndex >= deck.length) {
      throw new Error("The deck does not have enough cards left to redraw")
    }
    nextHand.push(deck[deckIndex])
    deckIndex += 1
  }
  state.set("hand", nextHand.map(cardKey))
  state.set("deck", deck.slice(deckIndex).map(cardKey))
  state.set("drawn", true)
  return nextHand
}

function stateSummary(state) {
  const hand = currentHand(state)
  if (hand.length !== 5) {
    return null
  }
  return {
    hand: hand,
    best: chooseBestHand(hand),
    drawn: !!state.get("drawn", false),
    round: numberValue(state.get("round", 0), 0),
  }
}

function scopeLabel(ctx) {
  const guildId = ctx && ctx.guild && ctx.guild.id ? String(ctx.guild.id) : "dm"
  const channelId = ctx && ctx.channel && ctx.channel.id ? String(ctx.channel.id) : "channel"
  const userId = ctx && ctx.user && ctx.user.id ? String(ctx.user.id) : "user"
  return [guildId, channelId, userId].join(":")
}

function pokerState(ctx) {
  return ctx.store.namespace("poker").namespace(scopeLabel(ctx))
}

function renderHandSummary(cards, best) {
  return [
    `Hand: ${formatHand(cards)}`,
    `Best hand: ${best.name}`,
  ].join("\n")
}

function renderAdvice(hole, board, result) {
  const pieces = []
  pieces.push(`Hole: ${formatHand(hole)}`)
  if (board.length > 0) {
    pieces.push(`Board: ${formatHand(board)}`)
  }
  pieces.push(`Action: ${result.action.toUpperCase()}`)
  pieces.push(`Reason: ${result.reason}`)
  if (result.bestHand) {
    pieces.push(`Best hand: ${result.bestHand.name}`)
    pieces.push(`Made hand: ${formatHand(result.bestHand.cards)}`)
  }
  return pieces.join("\n")
}

module.exports = {
  makeDeck,
  shuffle,
  parseCardList,
  parsePositions,
  prettyPositions,
  parseStoredCards,
  validateNoDuplicateCards,
  chooseBestHand,
  evaluateFive,
  compareScores,
  startRound,
  currentHand,
  currentDeck,
  resetRound,
  redrawRound,
  stateSummary,
  pokerState,
  renderHandSummary,
  renderAdvice,
  suggestAction,
  formatHand,
  cardText,
  cardKey,
}
