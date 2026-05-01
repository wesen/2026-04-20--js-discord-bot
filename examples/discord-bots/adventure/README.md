# adventure

ASCII choose-your-own-adventure Discord bot.

The JavaScript bot owns Discord interaction flow, state validation, rendering, and SQLite persistence. OpenRouter access is owned by Go through `require("adventure_llm")`; JS supplies prompts only and cannot choose the model, token, base URL, or max tokens.

## Run

```bash
export DISCORD_BOT_TOKEN=...
export DISCORD_APPLICATION_ID=...
export DISCORD_GUILD_ID=...
export OPENROUTER_API_KEY=...

GOWORK=off go run ./cmd/discord-bot bots adventure run \
  --bot-repository ./examples/discord-bots \
  --sync-on-start \
  --session-db-path ./examples/discord-bots/adventure/data/adventure.sqlite
```

Go-owned OpenRouter defaults:

- model: `anthropic/claude-3.5-haiku`
- max tokens: `1200`
- temperature: `0.7`
- base URL: `https://openrouter.ai/api/v1`

Optional process-level overrides, still not visible to JS as runtime config:

- `OPENROUTER_MODEL`
- `OPENROUTER_MAX_TOKENS`
- `OPENROUTER_TEMPERATURE`
- `OPENROUTER_BASE_URL`
- `OPENROUTER_HTTP_REFERER`
- `OPENROUTER_APP_TITLE`

## Commands

- `/adventure-start` — optionally pass `prompt` with a premise, character, goal, or starting context
- `/adventure-resume`
- `/adventure-state`
- `/adventure-reset`

Buttons advance fixed choice slots. “Try something else…” opens a modal and asks the LLM to interpret the free-form action as JSON before generating the next scene.
