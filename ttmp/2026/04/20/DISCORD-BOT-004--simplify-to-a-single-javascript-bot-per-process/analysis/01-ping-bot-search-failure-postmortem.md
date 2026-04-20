---
Title: Ping Bot Search Failure Postmortem
Ticket: DISCORD-BOT-004
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/modules/timer/timer.go
      Note: Provided timer module that the example bot should use instead of assuming browser globals
    - Path: examples/discord-bots/ping/index.js
      Note: Example bot code that originally assumed setTimeout and now uses the timer module
    - Path: internal/jsdiscord/bot.go
      Note: Promise settlement and rejected-promise formatting path that originally collapsed JS Error objects to map[]
    - Path: internal/jsdiscord/host.go
      Note: Interaction dispatch path and logging around command failures
    - Path: internal/jsdiscord/runtime_test.go
      Note: Regression coverage for timer-backed async commands and clearer promise rejection details
ExternalSources: []
Summary: 'Explain why the example ping bot’s `/search` command failed with `promise rejected: map[]`, what the real root cause was, and what changed to make the failure both fixed and easier to debug.'
LastUpdated: 2026-04-20T18:20:00-04:00
WhatFor: Give maintainers a clear postmortem for the ping bot search failure, including the operator-visible symptoms, the runtime mechanics behind the bug, and the follow-up fixes.
WhenToUse: Use when reviewing the ping bot regression, improving runtime diagnostics, or explaining why example bots must use provided runtime modules instead of assuming browser globals.
---


# Ping Bot Search Failure Postmortem

## Goal

Explain what was happening when the example `ping` bot failed on `/search architecture`, why the error message was misleading, and what changed in both the example bot and the runtime to fix the problem and improve future debugging.

## Executive Summary

The bug was **not** in Discord command routing, slash-command registration, or promise settlement itself.

The real problem was much simpler:

- the example `ping` bot implemented a delay helper with `setTimeout(...)`
- our JavaScript runtime is **not a browser runtime** and does **not** provide a global `setTimeout`
- the `/search` command deferred successfully, then hit `await sleep(2000)`
- that caused the async JavaScript handler to reject
- the Go host correctly noticed the rejected promise, but formatted the rejection so poorly that the operator only saw:

```text
promise rejected: map[]
```

So there were really **two issues at once**:

1. an **example bot bug**: assuming browser globals exist
2. a **runtime diagnostics bug**: losing the useful JavaScript error details when a promise rejected

## Operator-visible symptom

The operator reported a run that looked healthy up to the command dispatch point:

```text
bot=ping commands=["announce","echo","feedback","ping","search"] events=["guildCreate","messageCreate","ready"] script=/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js
2026-04-20T17:41:56.775154119-04:00 INF internal/bot/bot.go:119 > synced discord application commands commands=["ping","echo","feedback","search","announce"] scope=guild:586274407350272042
synced 5 commands
2026-04-20T17:41:57.342009044-04:00 INF internal/bot/bot.go:166 > discord bot connected bot_script= user=llm-bot user_id=1324847363872784414
2026-04-20T17:41:57.342725093-04:00 INF internal/jsdiscord/bot.go:863 > js discord bot connected jsKind=event jsName=ready meta.category=examples meta.description="Discord JS API showcase with buttons, modals, and autocomplete" meta.name=ping script=/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js user=llm-bot
2026-04-20T17:41:57.343566843-04:00 INF internal/jsdiscord/bot.go:863 > joined guild guild=slono guildId=586274407350272042 jsKind=event jsName=guildCreate meta.category=examples meta.description="Discord JS API showcase with buttons, modals, and autocomplete" meta.name=ping
2026-04-20T17:42:06.659254367-04:00 ERR internal/bot/bot.go:197 > failed to dispatch interaction to javascript bot error="promise rejected: map[]"
```

That created the impression that something deep in the host/runtime had broken. In reality, the core host path was mostly working. The failure was in the example bot’s own async logic.

## User impact

### What the operator experienced

- the bot started successfully
- commands synced successfully
- `/search` was available in Discord
- invoking `/search architecture` failed instead of returning results
- the log message was too vague to explain why

### Why this was especially confusing

This example bot is supposed to act as executable documentation for the JavaScript Discord API. When the showcase bot fails on a normal command path, it undermines confidence in:

- the example repository
- the JS runtime integration
- the async dispatch path
- the promise settlement behavior

## What actually happened

The relevant code in the example bot looked like this before the fix:

```js
const { defineBot } = require("discord")

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))
```

And later:

```js
await ctx.defer({ ephemeral: true })
await ctx.edit({
  content: `Searching for ${ctx.args.query}...`,
  ephemeral: true,
})
await sleep(2000)
```

That logic assumes the runtime has a global `setTimeout`, like a browser or Node-style environment.

But this project’s JavaScript environment is a **Goja-based embedded runtime** with an explicit module surface. It exposes capabilities through provided modules. It does **not** automatically expose browser globals.

This project already had a valid timing primitive available through the timer module:

```js
const { sleep } = require("timer")
```

So the actual failing chain was:

1. the operator ran `/search architecture`
2. the handler entered successfully
3. `ctx.defer(...)` succeeded
4. the first `ctx.edit(...)` succeeded
5. the command reached `await sleep(2000)`
6. the local `sleep(...)` implementation attempted `setTimeout(...)`
7. `setTimeout` was not defined in this runtime
8. the async handler rejected its promise
9. the Go host waited for the promise and saw it reject
10. the rejection formatter exported the JS error object too bluntly, producing `map[]`
11. the operator saw `promise rejected: map[]` instead of the real JavaScript error

## Why the command got as far as it did

This is an important detail.

The command did **not** fail immediately on entry, because the bug was in the delay helper, not in command registration or dispatch. That means earlier async work could still succeed.

So the flow was:

- slash command dispatch: working
- context creation: working
- deferral: working
- initial response edit: working
- custom delay implementation: broken

That partial success made the failure feel more mysterious than it really was.

## Root cause

### Primary root cause

The example `ping` bot assumed a browser-style global API:

- `setTimeout(...)`

But the runtime contract for this project is **module-based**, not browser-global-based.

The example should have used the provided timer module instead of assuming global timers existed.

### Secondary root cause

The runtime’s promise rejection reporting did not preserve enough JavaScript error information.

Rejected `Error` objects were being reduced through Go-side export formatting, which often collapsed useful exception details into:

```text
map[]
```

So a clear JavaScript runtime failure became an opaque host-side message.

## Contributing factors

### 1. Example code looked normal at a glance

This line is very common in browser and Node-adjacent JavaScript:

```js
new Promise((resolve) => setTimeout(resolve, ms))
```

That made it easy to miss the runtime assumption during review.

### 2. The runtime already had a correct timer abstraction

Because the correct solution already existed, the bug was not a missing platform feature. It was a mismatch between:

- what the example assumed
- what the runtime contract actually guarantees

### 3. The error message hid the important clue

If the operator had seen something like:

```text
ReferenceError: setTimeout is not defined
```

then the root cause would have been obvious almost immediately.

Instead, the log showed:

```text
promise rejected: map[]
```

which pointed attention toward promise machinery instead of the actual JS exception.

## Why the host/runtime itself was mostly fine

It is worth separating the parts that were working from the parts that were not.

### Working correctly

- bot discovery
- slash-command description and sync
- command routing into JS
- async handler invocation
- deferred response support
- response edit support
- detection that the top-level promise rejected

### Not good enough

- diagnostics for rejected JS promises
- example bot correctness with respect to runtime assumptions

This distinction matters because the postmortem is not “the async runtime was broken.” The async runtime did its job well enough to surface the rejection. It just surfaced it poorly.

## Fix that was applied

### Example bot fix

The example bot was changed from:

```js
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))
```

to:

```js
const { sleep } = require("timer")
```

This aligns the example with the actual runtime contract.

### Runtime diagnostics fix

Promise rejection handling in the Go host was improved so that rejection reporting now snapshots:

- the exported settled value
- a VM-side string/stack rendering of the rejected value

That means future JavaScript errors should produce a real message whenever possible, instead of collapsing to `map[]`.

### Logging improvement

Interaction dispatch logging was also improved so it is easier to tell:

- which script handled the interaction
- which command or custom ID was involved
- which interaction type failed

## Why `require("timer")` is the right abstraction

The timer module is the correct solution because it expresses the runtime contract explicitly.

Instead of assuming a global environment, the bot now declares its dependency on a provided capability:

```js
const { sleep } = require("timer")
```

That is better because it is:

- explicit
- portable within this runtime
- aligned with how other host capabilities are exposed
- easier to test and reason about than implicit globals

## What made the bug-report symptom misleading

The misleading part was the final host error string, not the actual runtime behavior.

The error suggested:

- maybe a malformed Go map
- maybe a bad payload normalization path
- maybe a strange promise export bug

But the real issue was just a missing JavaScript global.

This is a good example of why diagnostics quality matters: bad errors change where maintainers look first.

## Preventive lessons

### 1. Example bots are part of the API surface

The example repository is not “just demo code.” It teaches authors what patterns are safe. If an example assumes the wrong runtime model, users will copy the wrong pattern.

### 2. Embedded JS runtimes must document environment assumptions clearly

Bot authors should know which of these are true:

- provided modules
- available globals
- browser-only APIs
- Node-only APIs
- host-specific helper surfaces

### 3. Promise rejection paths must preserve JS-native information

If a JS exception reaches the Go host, the host should try hard to keep:

- error message
- stack string, if available
- command/context metadata

Otherwise debugging becomes much slower than necessary.

## Suggested long-term follow-ups

### Documentation follow-up

Add a short “runtime environment assumptions” section to the example-bot docs that explains:

- use `require("timer")` for delays
- do not assume browser globals like `setTimeout`
- rely on provided host modules and documented context helpers

### Testing follow-up

Keep explicit tests for:

- timer-backed async command behavior
- promise rejection messages that include useful JS text

### Logging follow-up

If needed, add even richer structured logs around command lifecycle events:

- command dispatch started
- command deferred
- command completed
- command rejected

That may or may not be worth doing at info level, but the debugging path is now much clearer.

## Short version for future maintainers

If you only remember one thing from this postmortem, it should be this:

> The `/search` failure was caused by an example bot using `setTimeout` in a runtime that does not provide browser globals. The confusing `promise rejected: map[]` message was a separate diagnostics issue that hid the real JavaScript exception.

## Related

- `reference/02-diary.md`
- `reference/01-single-bot-runner-reference-and-migration-notes.md`
- `design-doc/01-single-javascript-bot-per-process-architecture-and-implementation-guide.md`
