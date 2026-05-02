---
Title: Initial Brainstorm
Ticket: adventure
Status: active
Topics:
    - discord
    - game
    - adventure
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-01T13:09:39.036478-07:00
WhatFor: ""
WhenToUse: ""
---

# Initial Brainstorm

Goal: create a lightweight choose-your-own-adventure game inside the Discord bot, using ASCII/monospace art for mood, scenes, maps, items, and outcomes.

## Core Product Ideas

- **Thread-based adventure sessions**: `/adventure start` creates a private or public thread where the story unfolds without cluttering the main channel.
- **Button-driven choices**: each scene posts ASCII art plus 2–4 Discord buttons; users choose paths without typing exact commands.
- **Party voting mode**: for group play, choices are voted on for 30–90 seconds; the winning option advances the story.
- **Solo mode**: one user owns the run; buttons are only accepted from that user.
- **Persistent save slots**: store scene ID, inventory, flags, HP/sanity/score, and history so players can resume later.
- **ASCII scene cards**: every scene has a small terminal-style panel: title, art, stats, inventory hints, and choices.
- **Inventory and flags**: choices can grant items/flags that unlock future branches.
- **Procedural “storylets”**: mix authored key scenes with reusable encounters selected by tags and state.

## Technical Directions To Explore

- Define adventures in YAML files so content can be written, inspected, validated, replayed, and versioned without changing Go code.
- Support live LLM-generated scenes that emit/extend the same YAML schema at runtime, rather than bypassing the game engine.
- Allow controlled free-form player input for useful context, puzzle solving, roleplay flavor, and intent capture.
- Create a small state machine: scene nodes, transitions, conditions, effects, and terminal endings.
- Use Discord interactions/buttons for choice handling; fall back to slash commands like `/adventure choose A` if needed.
- Decide whether sessions are channel-scoped, thread-scoped, or user-scoped.
- Add guardrails for expired buttons, duplicate clicks, and concurrent updates.

## High-Level Design Goal

Build the adventure system around a strict, inspectable YAML game-state and scene schema while using an LLM as a live content generator. The LLM should never be the source of truth. It proposes YAML scene patches, action interpretations, narration, and ASCII art; the Go/Discord bot engine validates those proposals, applies only allowed effects to canonical state, persists the resulting scene/state, and renders the outcome through Discord interactions.

This gives us the best of both approaches:

- **Data-driven reliability**: scenes, choices, effects, inventory, flags, and state transitions are structured and replayable.
- **LLM flexibility**: free-form player actions, dialogue, inspections, and emergent scene details can be interpreted and expanded live.
- **Operational safety**: schema validation, clamps, state ownership, and audit logs keep hallucinations from corrupting the game.
- **Authoring ergonomics**: an authored seed adventure can set genre/tone/mechanics, while the LLM fills in moment-to-moment variation.

## LLM + YAML Generation Direction

The core contract should stay data-driven: the engine consumes a structured YAML scene/adventure schema. An LLM can generate or patch that data live, but the bot should validate the YAML before applying it.

Possible model:

- Start with a seed adventure definition: genre, tone, safety limits, allowed mechanics, initial state, and schema version.
- At each turn, the engine sends the current state plus recent history to the LLM.
- The LLM returns a structured YAML scene patch: art, narration, choices, conditions, and effects.
- The bot validates the patch, clamps unsafe or invalid effects, stores it, renders it, and waits for interaction.
- Player input can be either choice-button based or free-form text, but the LLM must map free-form input back into validated game effects.

Free-form input ideas:

- `/adventure say <text>` for dialogue with NPCs.
- `/adventure do <text>` for custom actions.
- `/adventure inspect <thing>` for targeted descriptions.
- Modal prompts when a scene asks the player to type a password, name, spell, plan, etc.
- Hybrid choices: buttons for common paths plus “Try something else…” for free-form action.

Important guardrails:

- Keep canonical state outside the LLM.
- Treat LLM output as a proposal, not truth.
- Validate YAML schema and reject/repair bad output.
- Enforce max stat changes, allowed inventory mutations, known flags, and ending rules.
- Keep an audit trail of generated scene YAML for debugging/replay.

## Open Questions

- Should the first adventure be mostly authored, procedural, or hybrid?
- Should group play be collaborative voting, chaotic first-click-wins, or turn-based?
- How big should ASCII art be to look good on desktop and mobile Discord?
- Should art be embedded in scene definitions or generated/loaded from separate files?
- Do we want AI-generated content later, or keep this deterministic/authored first?

