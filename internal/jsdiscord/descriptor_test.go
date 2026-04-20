package jsdiscord

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectScriptReturnsBotDescriptor(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command, event, component, modal, autocomplete, configure }) => {
			configure({
				name: "knowledge-base",
				description: "Search internal docs",
				run: { fields: { indexPath: { type: "string", help: "Docs index" } } }
			})
			command("kb-search", { description: "Search docs", options: { query: { type: "string", autocomplete: true } } }, async () => ({ content: "ok" }))
			event("ready", async () => null)
			component("kb:open", async () => ({ content: "opened" }))
			modal("kb:create", async () => ({ content: "created" }))
			autocomplete("kb-search", "query", async () => [])
		})
	`)
	info, err := InspectScript(context.Background(), scriptPath)
	if err != nil {
		t.Fatalf("InspectScript: %v", err)
	}
	if info.Name != "knowledge-base" {
		t.Fatalf("name = %q", info.Name)
	}
	if len(info.Commands) != 1 || info.Commands[0].Name != "kb-search" {
		t.Fatalf("commands = %#v", info.Commands)
	}
	if len(info.Events) != 1 || info.Events[0].Name != "ready" {
		t.Fatalf("events = %#v", info.Events)
	}
	if len(info.Components) != 1 || info.Components[0].CustomID != "kb:open" {
		t.Fatalf("components = %#v", info.Components)
	}
	if len(info.Modals) != 1 || info.Modals[0].CustomID != "kb:create" {
		t.Fatalf("modals = %#v", info.Modals)
	}
	if len(info.Autocompletes) != 1 || info.Autocompletes[0].CommandName != "kb-search" || info.Autocompletes[0].OptionName != "query" {
		t.Fatalf("autocompletes = %#v", info.Autocompletes)
	}
	if info.RunSchema == nil || len(info.RunSchema.Sections) != 1 || len(info.RunSchema.Sections[0].Fields) != 1 || info.RunSchema.Sections[0].Fields[0].Name != "indexPath" {
		t.Fatalf("run schema = %#v", info.RunSchema)
	}
}

func TestFallbackBotNameUsesDirectoryForIndex(t *testing.T) {
	dir := t.TempDir()
	botDir := filepath.Join(dir, "support")
	if err := os.MkdirAll(botDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	scriptPath := filepath.Join(botDir, "index.js")
	if err := os.WriteFile(scriptPath, []byte(`const { defineBot } = require("discord"); module.exports = defineBot(({ command }) => { command("support-ticket", async () => ({ content: "ok" })) })`), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	info, err := InspectScript(context.Background(), scriptPath)
	if err != nil {
		t.Fatalf("InspectScript: %v", err)
	}
	if info.Name != "support" {
		t.Fatalf("name = %q", info.Name)
	}
}
