package jsdiscord

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
)

// RuntimeStateContextKey is the engine context key for the discord runtime state.
// It is set during module registration so that runtime inspectors or future
// extensions can retrieve the RuntimeState from the module context.
const RuntimeStateContextKey = "discord.runtime"

type Config struct {
	ModuleName string
}

type Registrar struct {
	config Config
}

func NewRegistrar(config Config) *Registrar {
	return &Registrar{config: config}
}

func (r *Registrar) ID() string {
	return "discord-js-registrar"
}

func (r *Registrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
	if reg == nil {
		return fmt.Errorf("require registry is nil")
	}
	moduleName := strings.TrimSpace(r.config.ModuleName)
	if moduleName == "" {
		moduleName = "discord"
	}
	state := NewRuntimeState(moduleName)
	if ctx != nil {
		ctx.SetValue(RuntimeStateContextKey, state)
	}
	reg.RegisterNativeModule(state.ModuleName(), state.Loader)
	return nil
}

type RuntimeState struct {
	moduleName string
	store      *MemoryStore
}

func NewRuntimeState(moduleName string) *RuntimeState {
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		moduleName = "discord"
	}
	return &RuntimeState{moduleName: moduleName, store: NewMemoryStore()}
}

func (s *RuntimeState) ModuleName() string {
	if s == nil || strings.TrimSpace(s.moduleName) == "" {
		return "discord"
	}
	return s.moduleName
}

func (s *RuntimeState) Store() *MemoryStore {
	if s == nil {
		return NewMemoryStore()
	}
	if s.store == nil {
		s.store = NewMemoryStore()
	}
	return s.store
}

func (s *RuntimeState) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)
	_ = exports.Set("defineBot", func(call goja.FunctionCall) goja.Value {
		return s.defineBot(vm, call)
	})

	// Polyfill jsverbs metadata functions so bot scripts can coexist
	// with __verb__ / __section__ / __package__ declarations.
	for _, name := range []string{"__package__", "__section__", "__verb__", "doc"} {
		if v := vm.Get(name); v == nil || goja.IsUndefined(v) {
			_ = vm.Set(name, func(goja.FunctionCall) goja.Value { return goja.Undefined() })
		}
	}
}

func (s *RuntimeState) defineBot(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 1 {
		panic(vm.NewGoError(fmt.Errorf("discord.defineBot expects defineBot(builderFn)")))
	}
	builder, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("discord.defineBot builder is not a function")))
	}
	draft := newBotDraft(s)
	api := vm.NewObject()
	_ = api.Set("command", func(call goja.FunctionCall) goja.Value { return draft.command(vm, call) })
	_ = api.Set("userCommand", func(call goja.FunctionCall) goja.Value { return draft.userCommand(vm, call) })
	_ = api.Set("messageCommand", func(call goja.FunctionCall) goja.Value { return draft.messageCommand(vm, call) })
	_ = api.Set("subcommand", func(call goja.FunctionCall) goja.Value { return draft.subcommand(vm, call) })
	_ = api.Set("event", func(call goja.FunctionCall) goja.Value { return draft.event(vm, call) })
	_ = api.Set("component", func(call goja.FunctionCall) goja.Value { return draft.component(vm, call) })
	_ = api.Set("modal", func(call goja.FunctionCall) goja.Value { return draft.modal(vm, call) })
	_ = api.Set("autocomplete", func(call goja.FunctionCall) goja.Value { return draft.autocomplete(vm, call) })
	_ = api.Set("configure", func(call goja.FunctionCall) goja.Value { return draft.configure(vm, call) })
	if _, err := builder(goja.Undefined(), api); err != nil {
		panic(vm.NewGoError(err))
	}
	return draft.finalize(vm)
}
