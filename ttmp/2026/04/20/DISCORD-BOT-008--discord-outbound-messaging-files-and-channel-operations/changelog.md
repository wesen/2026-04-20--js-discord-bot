# Changelog

## 2026-04-20

Created ticket `DISCORD-BOT-008` for richer outbound Discord operations from JavaScript, including files, arbitrary sends, and message/channel helpers.

Implemented the first outbound-operations slice: added request-scoped `ctx.discord.channels.send(...)` and `ctx.discord.messages.edit/delete/react(...)`, expanded payload normalization with `files` and `replyTo`, added runtime tests for outbound operations, and updated the example bot to send a report message through the host layer.
