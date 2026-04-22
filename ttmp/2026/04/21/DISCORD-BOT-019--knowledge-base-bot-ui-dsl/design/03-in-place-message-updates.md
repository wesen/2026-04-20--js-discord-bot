---
Title: In-Place Message Updates for Component Interactions
Ticket: DISCORD-BOT-019
Status: active
Topics: ui-dsl, interactions, update-message
---

# In-Place Message Updates for Component Interactions

## Problem

When a user clicks a button or selects an option in a Discord message (e.g., `/demo-cards` product selector, `/demo-search` pagination), the bot **creates a new message** each time instead of **updating the existing message in-place**. This produces a flood of messages that makes interactive UIs unusable.

Example: clicking "Next ▶" in the search pager creates a brand-new message with page 2, instead of editing the page 1 message to show page 2.

## Root Cause

In `host_responses.go`, `interactionResponder.Reply()` always responds with type 4 (`InteractionResponseChannelMessageWithSource`):

```go
// Line 98-99 — ALWAYS creates a new message
err = r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,  // type 4
    Data: data,
})
```

But Discord provides **type 7 (`InteractionResponseUpdateMessage`)** specifically for component interactions — it edits the message the component was attached to, in-place.

## Discord's Interaction Response Types

From [Discord docs](https://docs.discord.com/developers/interactions/receiving-and-responding#interaction-response-object-interaction-callback-type):

| Type | Value | When to use |
|------|-------|-------------|
| `CHANNEL_MESSAGE_WITH_SOURCE` | 4 | Respond to a slash command with a new message |
| `DEFERRED_CHANNEL_MESSAGE_WITH_SOURCE` | 5 | ACK a slash command, edit later (shows "thinking...") |
| `UPDATE_MESSAGE` | 7 | **Component interaction: edit the parent message in-place** |
| `DEFERRED_UPDATE_MESSAGE` | 6 | **Component interaction: ACK, edit later (no loading state)** |

Types 6 and 7 only work for `MESSAGE_COMPONENT` interactions (buttons, selects). Types 4 and 5 work for slash commands.

## Current Architecture

```
Host.dispatchMessageComponentInteraction()
  → creates interactionResponder (no awareness of interaction type)
  → handle.DispatchComponent(ctx, req)
    → JS handler returns a payload
  → emitEventResult(ctx, responder.Reply, result)
    → responder.Reply(ctx, payload)
      → InteractionRespond(type=4) ← ALWAYS creates new message
```

The `interactionResponder` doesn't know what kind of interaction it's handling. It always uses type 4.

## Proposed Solution

### Option A: Auto-detect in `interactionResponder.Reply()` (minimal change)

The `interactionResponder` already has access to `r.interaction`. Check `r.interaction.Type`:

```go
func (r *interactionResponder) Reply(ctx context.Context, payload any) error {
    // ... existing ack check ...
    
    responseType := discordgo.InteractionResponseChannelMessageWithSource // type 4
    if r.interaction.Type == discordgo.InteractionMessageComponent {
        responseType = discordgo.InteractionResponseUpdateMessage // type 7
    }
    
    err = r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
        Type: responseType,
        Data: data,
    })
}
```

**Pros:** Minimal code change. All component interactions automatically update in-place.
**Cons:** No way for the bot to opt-out (sometimes you want a new message from a component click).

### Option B: Add `.update()` to the message builder DSL

Add a flag to the DSL that the dispatch pipeline reads:

```js
// Default: component clicks update in-place
ui.message()
  .content("Page 2")
  .embed(...)
  .row(ui.pager(...))
  .build()   // → *normalizedResponse with UpdateMessage=true (default for components)

// Opt-out: force a new message
ui.message()
  .content("Done!")
  .followUp()   // explicit: creates a new message
  .build()
```

**Pros:** Explicit control. JS bot authors can choose.
**Cons:** More complex. Need to thread a flag through `normalizedResponse` → dispatch pipeline.

### Option C: Hybrid — auto-detect + DSL override

- Component interactions default to type 7 (update in-place)
- `ui.message().followUp()` forces type 4 (new followup message)
- Slash commands always use type 4

This is the best of both worlds. The common case (clicking buttons in a pager/card gallery) "just works" without any JS changes, and bots that need new messages can opt in.

## How Component Updates Work in Discord

When a user clicks a button on message M:

1. Discord sends `InteractionCreate` with `type: MESSAGE_COMPONENT` (type 3)
2. Bot responds within 3 seconds with `type: 7 (UPDATE_MESSAGE)` + new message data
3. Discord replaces message M's content/embeds/components with the new data
4. The message stays in the same position — no new message created

If the bot needs more than 3 seconds:

1. Respond with `type: 6 (DEFERRED_UPDATE_MESSAGE)` — user sees nothing (no loading state)
2. Later, use `Edit` (PATCH webhook `@original`) to update the message

### Important: ephemeral messages

Ephemeral messages (ephemeral: true) are per-user. Updating them works the same way — the update only affects the user who clicked.

### Important: the interaction token is valid for 15 minutes

Followup messages and edits are possible for 15 minutes after the initial interaction.

## Implementation Plan

### Step 1: Auto-detect in `interactionResponder` (quick fix)

1. Check `r.interaction.Type` in `Reply()` and `Defer()`
2. Use type 7 for `InteractionMessageComponent`, type 4 otherwise
3. All existing bots benefit immediately — no JS changes needed

### Step 2: Add `.followUp()` to message builder (DSL enhancement)

1. Add `followUp bool` to `normalizedResponse`
2. Add `.followUp()` method to the message builder Proxy
3. In `Reply()`, check `normalizedResponse.FollowUp` to override auto-detection
4. JS bots can explicitly create new messages from component handlers

### Step 3: Add `ui.defer()` helper for long operations

1. Returns a `*normalizedResponse` with just `Deferred: true`
2. Pipeline maps this to type 6 (deferred update) for components
3. Bot calls `ctx.edit()` later with the actual content

## Impact on Existing Bots

- **Knowledge-base bot:** No changes needed. Component handlers (search select, pager buttons, action buttons) will automatically update in-place.
- **Ping bot:** No changes (doesn't use components).
- **UI-showcase bot:** No changes needed. All interactive screens (search, review, cards, pager) will automatically update in-place.

## Sources

- `sources/discord-receiving-and-responding.md` — Discord docs on interaction response types, followup messages, editing
- `sources/discord-interactions-overview.md` — Discord docs on interaction lifecycle, webhook mechanics
