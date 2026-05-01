package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dop251/goja"
	noderequire "github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/stretchr/testify/require"
)

type adventureLLMMockRegistrar struct{}

func (adventureLLMMockRegistrar) ID() string { return "adventure-llm-mock" }

func (adventureLLMMockRegistrar) RegisterRuntimeModules(_ *engine.RuntimeModuleContext, reg *noderequire.Registry) error {
	reg.RegisterNativeModule("adventure_llm", func(vm *goja.Runtime, moduleObj *goja.Object) {
		exports := moduleObj.Get("exports").(*goja.Object)
		_ = exports.Set("completeJson", func(call goja.FunctionCall) goja.Value {
			input, _ := call.Argument(0).Export().(map[string]any)
			purpose := fmt.Sprint(input["purpose"])
			metadata, _ := input["metadata"].(map[string]any)
			turn := int64(0)
			if raw, ok := metadata["turn"]; ok {
				switch v := raw.(type) {
				case int64:
					turn = v
				case int:
					turn = int64(v)
				case float64:
					turn = int64(v)
				}
			}
			if purpose == "interpret_action" {
				return vm.ToValue(map[string]any{
					"ok":       true,
					"provider": "mock",
					"text":     `{"interpreted_action":{"summary":"The player improvises.","kind":"other","target":"gate","risk":"low","proposed_effects":{"flags":{"improvised":true}},"response_hint":"Acknowledge the improvisation."}}`,
				})
			}
			text := fmt.Sprintf(`{"scene_patch":{"scene":{"id":"mock-scene-%d","title":"Mock Scene %d","ascii_art":"/\\","narration":"Scene for turn %d.","choices":[{"id":"continue","label":"Continue","proposed_effects":{"flags":{"continued":true}},"next_hint":"forward"},{"id":"wait","label":"Wait","proposed_effects":{"stats":{"sanity":1}},"next_hint":"pause"}]},"engine_notes":{"mock":true}}}`, turn, turn, turn)
			return vm.ToValue(map[string]any{"ok": true, "provider": "mock", "text": text})
		})
	})
	return nil
}

func loadAdventureTestBot(t *testing.T) *BotHandle {
	t.Helper()
	repoRoot := repoRootJSDiscord(t)
	scriptPath := filepath.Join(repoRoot, "examples", "discord-bots", "adventure", "index.js")
	factory, err := engine.NewBuilder(
		engine.WithModuleRootsFromScript(scriptPath, engine.DefaultModuleRootsOptions()),
	).
		WithModules(engine.DefaultRegistryModulesNamed("database")).
		WithRuntimeModuleRegistrars(NewRegistrar(Config{}), &UIRegistrar{}, adventureLLMMockRegistrar{}).
		Build()
	require.NoError(t, err)
	rt, err := factory.NewRuntime(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	value, err := rt.Require.Require(scriptPath)
	require.NoError(t, err)
	handle, err := CompileBot(rt.VM, value)
	require.NoError(t, err)
	return handle
}

func TestAdventureRejectsOtherPlayerWithoutBreakingOwnerSession(t *testing.T) {
	handle := loadAdventureTestBot(t)
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "adventure.sqlite")
	base := DispatchRequest{
		Config:  map[string]any{"sessionDbPath": dbPath},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	}

	var startEdits []any
	_, err := handle.DispatchCommand(ctx, mergeDispatch(base, DispatchRequest{
		Name:  "adventure-start",
		User:  UserSnapshot{ID: "player-1", Username: "Player One"},
		Args:  map[string]any{"seed": "haunted-gate"},
		Defer: func(context.Context, any) error { return nil },
		Edit: func(_ context.Context, value any) error {
			startEdits = append(startEdits, value)
			return nil
		},
	}))
	require.NoError(t, err)
	require.Len(t, startEdits, 1)
	require.Contains(t, fmt.Sprint(startEdits[0]), "Mock Scene 0")

	otherResult, err := handle.DispatchComponent(ctx, mergeDispatch(base, DispatchRequest{
		Name:    "adv:choice:0",
		User:    UserSnapshot{ID: "player-2", Username: "Player Two"},
		Message: &MessageSnapshot{Content: "Turn 0", ChannelID: "channel-1", GuildID: "guild-1"},
	}))
	require.NoError(t, err)
	otherResponse, ok := otherResult.(*normalizedResponse)
	require.True(t, ok, "other result type = %T", otherResult)
	require.Contains(t, otherResponse.Content, "belongs to another player")
	require.True(t, otherResponse.Ephemeral)

	var ownerEdits []any
	_, err = handle.DispatchComponent(ctx, mergeDispatch(base, DispatchRequest{
		Name:    "adv:choice:0",
		User:    UserSnapshot{ID: "player-1", Username: "Player One"},
		Message: &MessageSnapshot{Content: "Turn 0", ChannelID: "channel-1", GuildID: "guild-1"},
		Defer:   func(context.Context, any) error { return nil },
		Edit: func(_ context.Context, value any) error {
			ownerEdits = append(ownerEdits, value)
			return nil
		},
	}))
	require.NoError(t, err)
	require.Len(t, ownerEdits, 1)
	ownerEdit := fmt.Sprint(ownerEdits[0])
	require.Contains(t, ownerEdit, "Mock Scene 1")
	require.Contains(t, ownerEdit, "Turn 1")

	stateResult, err := handle.DispatchCommand(ctx, mergeDispatch(base, DispatchRequest{
		Name: "adventure-state",
		User: UserSnapshot{ID: "player-1", Username: "Player One"},
	}))
	require.NoError(t, err)
	stateText := fmt.Sprint(stateResult)
	require.Contains(t, stateText, "player-1")
	require.Contains(t, stateText, "currentSceneId")
	require.Contains(t, stateText, "mock-scene-1")
	require.NotContains(t, stateText, "player-2")
}

func TestAdventureRejectsStaleOwnerButton(t *testing.T) {
	handle := loadAdventureTestBot(t)
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "adventure.sqlite")
	base := DispatchRequest{Config: map[string]any{"sessionDbPath": dbPath}, Guild: map[string]any{"id": "guild-1"}, Channel: map[string]any{"id": "channel-1"}}

	_, err := handle.DispatchCommand(ctx, mergeDispatch(base, DispatchRequest{
		Name:  "adventure-start",
		User:  UserSnapshot{ID: "player-1"},
		Defer: func(context.Context, any) error { return nil },
		Edit:  func(context.Context, any) error { return nil },
	}))
	require.NoError(t, err)

	_, err = handle.DispatchComponent(ctx, mergeDispatch(base, DispatchRequest{
		Name: "adv:choice:0", User: UserSnapshot{ID: "player-1"}, Message: &MessageSnapshot{Content: "Turn 0"},
		Defer: func(context.Context, any) error { return nil }, Edit: func(context.Context, any) error { return nil },
	}))
	require.NoError(t, err)

	staleResult, err := handle.DispatchComponent(ctx, mergeDispatch(base, DispatchRequest{
		Name: "adv:choice:1", User: UserSnapshot{ID: "player-1"}, Message: &MessageSnapshot{Content: "Turn 0"},
	}))
	require.NoError(t, err)
	require.True(t, strings.Contains(fmt.Sprint(staleResult), "stale"), "result = %v", staleResult)
}

func mergeDispatch(base, override DispatchRequest) DispatchRequest {
	out := base
	if override.Name != "" {
		out.Name = override.Name
	}
	if override.Args != nil {
		out.Args = override.Args
	}
	if override.User.ID != "" {
		out.User = override.User
	}
	if override.Message != nil {
		out.Message = override.Message
	}
	if override.Defer != nil {
		out.Defer = override.Defer
	}
	if override.Edit != nil {
		out.Edit = override.Edit
	}
	return out
}
