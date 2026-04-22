package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
)

type Host struct {
	scriptPath    string
	runtime       *engine.Runtime
	handle        *BotHandle
	runtimeConfig map[string]any
}

func NewHost(ctx context.Context, scriptPath string) (*Host, error) {
	if strings.TrimSpace(scriptPath) == "" {
		return nil, fmt.Errorf("discord bot script path is empty")
	}
	absScript, err := filepath.Abs(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("resolve script path: %w", err)
	}
	factory, err := engine.NewBuilder(
		engine.WithModuleRootsFromScript(absScript, engine.DefaultModuleRootsOptions()),
	).WithModules(engine.DefaultRegistryModules()).
		WithRuntimeModuleRegistrars(NewRegistrar(Config{}), &UIRegistrar{}).
		WithRequireOptions(require.WithGlobalFolders(filepath.Dir(absScript), filepath.Join(filepath.Dir(absScript), "node_modules"))).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build js runtime: %w", err)
	}
	rt, err := factory.NewRuntime(ctx)
	if err != nil {
		return nil, fmt.Errorf("create js runtime: %w", err)
	}
	value, err := rt.Require.Require(absScript)
	if err != nil {
		_ = rt.Close(context.Background())
		return nil, fmt.Errorf("load js bot script: %w", err)
	}
	handle, err := CompileBot(rt.VM, value)
	if err != nil {
		_ = rt.Close(context.Background())
		return nil, fmt.Errorf("compile js bot: %w", err)
	}
	return &Host{scriptPath: absScript, runtime: rt, handle: handle, runtimeConfig: map[string]any{}}, nil
}

func (h *Host) SetRuntimeConfig(config map[string]any) {
	if h == nil {
		return
	}
	h.runtimeConfig = cloneMap(config)
}

func (h *Host) Close(ctx context.Context) error {
	if h == nil || h.runtime == nil {
		return nil
	}
	return h.runtime.Close(ctx)
}

func (h *Host) Describe(ctx context.Context) (map[string]any, error) {
	if h == nil || h.handle == nil {
		return nil, fmt.Errorf("discord js host is nil")
	}
	return h.handle.Describe(ctx)
}

func (h *Host) ApplicationCommands(ctx context.Context) ([]*discordgo.ApplicationCommand, error) {
	desc, err := h.Describe(ctx)
	if err != nil {
		return nil, err
	}
	rawCommands := commandSnapshots(desc["commands"])
	commands := make([]*discordgo.ApplicationCommand, 0, len(rawCommands))
	for _, raw := range rawCommands {
		snapshot, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		command, err := applicationCommandFromSnapshot(snapshot)
		if err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	return commands, nil
}
