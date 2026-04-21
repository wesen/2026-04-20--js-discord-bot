package jsdiscord

import "testing"

func TestApplicationCommandFromSnapshotSupportsAutocompleteAndConstraints(t *testing.T) {
	cmd, err := applicationCommandFromSnapshot(map[string]any{
		"name": "echo",
		"spec": map[string]any{
			"description": "Echo text",
			"options": map[string]any{
				"text": map[string]any{
					"type": "string", "description": "Text", "required": true,
					"autocomplete": true, "minLength": 2, "maxLength": 100,
				},
				"count": map[string]any{"type": "integer", "description": "Count", "minValue": 1, "maxValue": 10},
			},
		},
	})
	if err != nil {
		t.Fatalf("applicationCommandFromSnapshot: %v", err)
	}
	if cmd.Name != "echo" || cmd.Description != "Echo text" {
		t.Fatalf("unexpected command: %#v", cmd)
	}
	if len(cmd.Options) != 2 {
		t.Fatalf("options = %d", len(cmd.Options))
	}
	if !cmd.Options[0].Autocomplete {
		t.Fatalf("expected autocomplete option: %#v", cmd.Options[0])
	}
	if cmd.Options[0].MinLength == nil || *cmd.Options[0].MinLength != 2 || cmd.Options[0].MaxLength != 100 {
		t.Fatalf("unexpected string constraints: %#v", cmd.Options[0])
	}
	if cmd.Options[1].MinValue == nil || *cmd.Options[1].MinValue != 1 || cmd.Options[1].MaxValue != 10 {
		t.Fatalf("unexpected number constraints: %#v", cmd.Options[1])
	}
}
