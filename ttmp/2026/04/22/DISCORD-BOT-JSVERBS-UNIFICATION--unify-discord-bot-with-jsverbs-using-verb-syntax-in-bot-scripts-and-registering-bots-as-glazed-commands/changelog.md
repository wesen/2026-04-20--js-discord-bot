# Changelog

## 2026-04-22

- Initial workspace created


## 2026-04-22

Created comprehensive architecture analysis document comparing discord-bot and jsverbs. Identified four-phase unification plan: (1) glazed list command, (2) glazed run command per-bot, (3) __verb__ polyfill in Discord runtime, (4) jsverbs scan over bot repositories.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command.go — Current raw cobra implementation to be replaced


## 2026-04-22

Updated design doc with BareCommand approach for __verb__(run): fields declared in __verb__ metadata become CLI flags, parsed by Glazed, converted via runtimeFieldInternalName(), and injected into the bot as ctx.config. Replaces RunSchema + manual flag parsing entirely.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/run_static_args.go — Manual flag parsing replaced by Glazed parsing

