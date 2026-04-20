package jsdiscord

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectScriptReturnsBotDescriptor(t *testing.T) {
	scriptPath := writeMultiHostBot(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command, event, configure }) => {
			configure({ name: "knowledge-base", description: "Search internal docs" })
			command("kb-search", { description: "Search docs" }, async () => ({ content: "ok" }))
			event("ready", async () => null)
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
}

func TestNewMultiHostRejectsDuplicateCommandNames(t *testing.T) {
	scriptA := writeNamedBot(t, "knowledge-base", "shared-cmd")
	scriptB := writeNamedBot(t, "support", "shared-cmd")
	_, err := NewMultiHost(context.Background(), []string{scriptA, scriptB})
	if err == nil {
		t.Fatalf("expected duplicate command error")
	}
}

func TestNewMultiHostRejectsDuplicateBotNames(t *testing.T) {
	scriptA := writeNamedBot(t, "knowledge-base", "kb-search")
	scriptB := writeNamedBot(t, "knowledge-base", "kb-article")
	_, err := NewMultiHost(context.Background(), []string{scriptA, scriptB})
	if err == nil {
		t.Fatalf("expected duplicate bot name error")
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

func writeMultiHostBot(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "bot.js")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}

func writeNamedBot(t *testing.T, name, command string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "index.js")
	source := `const { defineBot } = require("discord"); module.exports = defineBot(({ command, configure }) => { configure({ name: "` + name + `" }); command("` + command + `", async () => ({ content: "ok" })) })`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}
