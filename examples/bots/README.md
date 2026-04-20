# Bot CLI example repository

This directory is the local example repository for `discord-bot bots ...`.

## Quick start

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot
GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/bots
```

## Smoke commands

```bash
GOWORK=off go run ./cmd/discord-bot bots run discord greet --bot-repository ./examples/bots Manuel --excited
GOWORK=off go run ./cmd/discord-bot bots run discord banner --bot-repository ./examples/bots Manuel
GOWORK=off go run ./cmd/discord-bot bots run math multiply --bot-repository ./examples/bots 6 7
GOWORK=off go run ./cmd/discord-bot bots run nested relay --bot-repository ./examples/bots hi there
GOWORK=off go run ./cmd/discord-bot bots run issues list --bot-repository ./examples/bots acme/repo --state closed --labels bug --labels docs
GOWORK=off go run ./cmd/discord-bot bots help issues list --bot-repository ./examples/bots
```
