const { defineBot } = require("discord")
const createKnowledgeStore = require("./lib/store").createKnowledgeStore
const capture = require("./lib/capture")
const render = require("./lib/render")
const registerKnowledgeBot = require("./lib/register-knowledge-bot")

const store = createKnowledgeStore()

module.exports = defineBot(({ command, event, modal, configure }) => {
  configure({
    name: "knowledge-base",
    description: "Listen to Discord chat, record candidate knowledge, and curate it as a shared memory",
    category: "knowledge",
    run: {
      fields: {
        dbPath: {
          type: "string",
          help: "SQLite path for the knowledge store",
          default: "./examples/discord-bots/knowledge-base/data/knowledge.sqlite",
        },
        captureEnabled: {
          type: "bool",
          help: "Enable passive capture from messageCreate events",
          default: true,
        },
        captureThreshold: {
          type: "number",
          help: "Minimum confidence required to save a passive capture",
          default: 0.65,
        },
        captureChannels: {
          type: "string",
          help: "Optional comma-separated channel IDs to allow for passive capture",
          default: "",
        },
        reviewLimit: {
          type: "integer",
          help: "Number of entries to show in review lists",
          default: 5,
        },
        seedEntries: {
          type: "bool",
          help: "Seed onboarding entries the first time the SQLite store is created",
          default: true,
        },
      },
    },
  })

  registerKnowledgeBot({ command, event, modal }, store, capture, render)
})
