package jsdiscord

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/runtimebridge"
	"github.com/go-go-golems/go-go-goja/pkg/runtimeowner"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type commandDraft struct {
	name        string
	commandType string
	spec        map[string]any
	handler     goja.Callable
}

type eventDraft struct {
	name    string
	handler goja.Callable
}

type componentDraft struct {
	customID string
	handler  goja.Callable
}

type modalDraft struct {
	customID string
	handler  goja.Callable
}

type autocompleteDraft struct {
	commandName string
	optionName  string
	handler     goja.Callable
}

type subcommandDraft struct {
	rootName string
	name     string
	spec     map[string]any
	handler  goja.Callable
}

type botDraft struct {
	moduleName      string
	store           *MemoryStore
	metadata        map[string]any
	commands        []*commandDraft
	userCommands    []*commandDraft
	messageCommands []*commandDraft
	subcommands     []*subcommandDraft
	events          []*eventDraft
	components      []*componentDraft
	modals          []*modalDraft
	autocompletes   []*autocompleteDraft
}

type DiscordOps struct {
	GuildFetch         func(context.Context, string) (map[string]any, error)
	RoleList           func(context.Context, string) ([]map[string]any, error)
	RoleFetch          func(context.Context, string, string) (map[string]any, error)
	ThreadFetch        func(context.Context, string) (map[string]any, error)
	ThreadJoin         func(context.Context, string) error
	ThreadLeave        func(context.Context, string) error
	ThreadStart        func(context.Context, string, any) (map[string]any, error)
	ChannelSend        func(context.Context, string, any) error
	ChannelFetch       func(context.Context, string) (map[string]any, error)
	ChannelSetTopic    func(context.Context, string, string) error
	ChannelSetSlowmode func(context.Context, string, int) error
	MessageFetch       func(context.Context, string, string) (map[string]any, error)
	MessageList        func(context.Context, string, any) ([]map[string]any, error)
	MessageEdit        func(context.Context, string, string, any) error
	MessageDelete      func(context.Context, string, string) error
	MessageReact       func(context.Context, string, string, string) error
	MessagePin         func(context.Context, string, string) error
	MessageUnpin       func(context.Context, string, string) error
	MessageListPinned  func(context.Context, string) ([]map[string]any, error)
	MessageBulkDelete  func(context.Context, string, any) error
	MemberFetch        func(context.Context, string, string) (map[string]any, error)
	MemberList         func(context.Context, string, any) ([]map[string]any, error)
	MemberAddRole      func(context.Context, string, string, string) error
	MemberRemoveRole   func(context.Context, string, string, string) error
	MemberSetTimeout   func(context.Context, string, string, any) error
	MemberKick         func(context.Context, string, string, any) error
	MemberBan          func(context.Context, string, string, any) error
	MemberUnban        func(context.Context, string, string) error
}

type DispatchRequest struct {
	Name        string
	Args        map[string]any
	Values      any
	Command     map[string]any
	Interaction map[string]any
	Message     map[string]any
	Before      map[string]any
	User        map[string]any
	Guild       map[string]any
	Channel     map[string]any
	Member      map[string]any
	Reaction    map[string]any
	Me          map[string]any
	Metadata    map[string]any
	Config      map[string]any
	Component   map[string]any
	Modal       map[string]any
	Focused     map[string]any
	Discord     *DiscordOps
	Reply       func(context.Context, any) error
	FollowUp    func(context.Context, any) error
	Edit        func(context.Context, any) error
	Defer       func(context.Context, any) error
	ShowModal   func(context.Context, any) error
}

type BotHandle struct {
	vm                   *goja.Runtime
	dispatchCommand      goja.Callable
	dispatchSubcommand   goja.Callable
	dispatchEvent        goja.Callable
	dispatchComponent    goja.Callable
	dispatchModal        goja.Callable
	dispatchAutocomplete goja.Callable
	describe             goja.Callable
}

func CompileBot(vm *goja.Runtime, value goja.Value) (*BotHandle, error) {
	if vm == nil {
		return nil, fmt.Errorf("discord bot compile: vm is nil")
	}
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, fmt.Errorf("discord bot compile: value is nil")
	}
	obj := value.ToObject(vm)
	if obj == nil {
		return nil, fmt.Errorf("discord bot compile: value is not an object")
	}
	dispatchCommand, ok := goja.AssertFunction(obj.Get("dispatchCommand"))
	if !ok {
		return nil, fmt.Errorf("discord bot compile: missing dispatchCommand method")
	}
	dispatchSubcommand, ok := goja.AssertFunction(obj.Get("dispatchSubcommand"))
	if !ok {
		return nil, fmt.Errorf("discord bot compile: missing dispatchSubcommand method")
	}
	dispatchEvent, ok := goja.AssertFunction(obj.Get("dispatchEvent"))
	if !ok {
		return nil, fmt.Errorf("discord bot compile: missing dispatchEvent method")
	}
	dispatchComponent, ok := goja.AssertFunction(obj.Get("dispatchComponent"))
	if !ok {
		return nil, fmt.Errorf("discord bot compile: missing dispatchComponent method")
	}
	dispatchModal, ok := goja.AssertFunction(obj.Get("dispatchModal"))
	if !ok {
		return nil, fmt.Errorf("discord bot compile: missing dispatchModal method")
	}
	dispatchAutocomplete, ok := goja.AssertFunction(obj.Get("dispatchAutocomplete"))
	if !ok {
		return nil, fmt.Errorf("discord bot compile: missing dispatchAutocomplete method")
	}
	describe, ok := goja.AssertFunction(obj.Get("describe"))
	if !ok {
		return nil, fmt.Errorf("discord bot compile: missing describe method")
	}
	return &BotHandle{
		vm:                   vm,
		dispatchCommand:      dispatchCommand,
		dispatchSubcommand:   dispatchSubcommand,
		dispatchEvent:        dispatchEvent,
		dispatchComponent:    dispatchComponent,
		dispatchModal:        dispatchModal,
		dispatchAutocomplete: dispatchAutocomplete,
		describe:             describe,
	}, nil
}

func (h *BotHandle) Describe(ctx context.Context) (map[string]any, error) {
	if h == nil {
		return nil, fmt.Errorf("discord bot handle is nil")
	}
	bindings, ok := runtimebridge.Lookup(h.vm)
	if !ok || bindings.Owner == nil {
		return nil, fmt.Errorf("discord bot requires runtime owner bindings")
	}
	ret, err := bindings.Owner.Call(ctx, "discord.bot.describe", func(context.Context, *goja.Runtime) (any, error) {
		value, err := h.describe(goja.Undefined())
		if err != nil {
			return nil, err
		}
		if goja.IsUndefined(value) || goja.IsNull(value) {
			return map[string]any{}, nil
		}
		if exported, ok := value.Export().(map[string]any); ok {
			return exported, nil
		}
		return map[string]any{"value": value.Export()}, nil
	})
	if err != nil {
		return nil, err
	}
	result, _ := ret.(map[string]any)
	return result, nil
}

func (h *BotHandle) DispatchCommand(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchCommand, request)
}

func (h *BotHandle) DispatchSubcommand(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchSubcommand, request)
}

func (h *BotHandle) DispatchEvent(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchEvent, request)
}

func (h *BotHandle) DispatchComponent(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchComponent, request)
}

func (h *BotHandle) DispatchModal(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchModal, request)
}

func (h *BotHandle) DispatchAutocomplete(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchAutocomplete, request)
}

func (h *BotHandle) dispatch(ctx context.Context, fn goja.Callable, request DispatchRequest) (any, error) {
	if h == nil {
		return nil, fmt.Errorf("discord bot handle is nil")
	}
	bindings, ok := runtimebridge.Lookup(h.vm)
	if !ok || bindings.Owner == nil {
		return nil, fmt.Errorf("discord bot requires runtime owner bindings")
	}
	ret, err := bindings.Owner.Call(ctx, "discord.bot.dispatch", func(callCtx context.Context, vm *goja.Runtime) (any, error) {
		input := buildDispatchInput(vm, callCtx, request)
		result, err := fn(goja.Undefined(), input)
		if err != nil {
			return nil, err
		}
		if goja.IsUndefined(result) || goja.IsNull(result) {
			return nil, nil
		}
		return result.Export(), nil
	})
	if err != nil {
		return nil, err
	}
	return settleValue(ctx, bindings.Owner, ret)
}

func settleValue(ctx context.Context, owner runtimeowner.Runner, value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case *goja.Promise:
		return waitForPromise(ctx, owner, v)
	case goja.Value:
		return settleValue(ctx, owner, v.Export())
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			settled, err := settleValue(ctx, owner, item)
			if err != nil {
				return nil, err
			}
			out[i] = settled
		}
		return out, nil
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			settled, err := settleValue(ctx, owner, item)
			if err != nil {
				return nil, err
			}
			out[key] = settled
		}
		return out, nil
	default:
		return value, nil
	}
}

type promiseSnapshot struct {
	State  goja.PromiseState
	Result any
	Text   string
}

func waitForPromise(ctx context.Context, owner runtimeowner.Runner, promise *goja.Promise) (any, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		ret, err := owner.Call(ctx, "discord.bot.promise-state", func(_ context.Context, vm *goja.Runtime) (any, error) {
			result := promise.Result()
			return promiseSnapshot{
				State:  promise.State(),
				Result: exportSettledValue(result),
				Text:   describeSettledValue(vm, result),
			}, nil
		})
		if err != nil {
			return nil, err
		}
		snapshot, ok := ret.(promiseSnapshot)
		if !ok {
			return nil, fmt.Errorf("discord bot promise snapshot has unexpected type %T", ret)
		}
		switch snapshot.State {
		case goja.PromiseStatePending:
			time.Sleep(5 * time.Millisecond)
		case goja.PromiseStateRejected:
			message := strings.TrimSpace(snapshot.Text)
			if message == "" {
				message = fmt.Sprint(snapshot.Result)
			}
			return nil, fmt.Errorf("promise rejected: %s", message)
		case goja.PromiseStateFulfilled:
			return snapshot.Result, nil
		}
	}
}

func exportSettledValue(value goja.Value) any {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	return value.Export()
}

func describeSettledValue(vm *goja.Runtime, value goja.Value) string {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return ""
	}
	if obj := value.ToObject(vm); obj != nil {
		if stack := strings.TrimSpace(safeValueString(vm, obj.Get("stack"))); stack != "" {
			return stack
		}
	}
	if text := strings.TrimSpace(safeValueString(vm, value)); text != "" && text != "[object Object]" {
		return text
	}
	return strings.TrimSpace(fmt.Sprint(value.Export()))
}

func safeValueString(vm *goja.Runtime, value goja.Value) string {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return ""
	}
	if ex, ok := value.Export().(error); ok {
		return ex.Error()
	}
	if obj := value.ToObject(vm); obj != nil {
		if fn, ok := goja.AssertFunction(obj.Get("toString")); ok {
			if ret, err := fn(value); err == nil && !goja.IsUndefined(ret) && !goja.IsNull(ret) {
				return ret.String()
			}
		}
	}
	return value.String()
}

func newBotDraft(state *RuntimeState) *botDraft {
	return &botDraft{
		moduleName:      state.ModuleName(),
		store:           state.Store(),
		metadata:        map[string]any{},
		commands:        []*commandDraft{},
		userCommands:    []*commandDraft{},
		messageCommands: []*commandDraft{},
		subcommands:     []*subcommandDraft{},
		events:          []*eventDraft{},
		components:      []*componentDraft{},
		modals:          []*modalDraft{},
		autocompletes:   []*autocompleteDraft{},
	}
}

func (d *botDraft) command(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
		panic(vm.NewGoError(fmt.Errorf("discord.command expects command(name, [spec], handler)")))
	}
	name := strings.TrimSpace(call.Arguments[0].String())
	if name == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.command name is empty")))
	}
	var spec map[string]any
	var handler goja.Callable
	var ok bool
	if len(call.Arguments) == 2 {
		handler, ok = goja.AssertFunction(call.Arguments[1])
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("discord.command %q handler is not a function", name)))
		}
	} else {
		spec = exportMap(call.Arguments[1])
		handler, ok = goja.AssertFunction(call.Arguments[2])
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("discord.command %q handler is not a function", name)))
		}
	}
	cmdType := ""
	if spec != nil {
		cmdType = strings.TrimSpace(fmt.Sprint(spec["type"]))
	}
	switch strings.ToLower(cmdType) {
	case "user":
		d.userCommands = append(d.userCommands, &commandDraft{name: name, commandType: "user", spec: spec, handler: handler})
	case "message":
		d.messageCommands = append(d.messageCommands, &commandDraft{name: name, commandType: "message", spec: spec, handler: handler})
	default:
		d.commands = append(d.commands, &commandDraft{name: name, commandType: "chat_input", spec: spec, handler: handler})
	}
	return goja.Undefined()
}

func (d *botDraft) userCommand(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 2 {
		panic(vm.NewGoError(fmt.Errorf("discord.userCommand expects userCommand(name, handler)")))
	}
	name := strings.TrimSpace(call.Arguments[0].String())
	if name == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.userCommand name is empty")))
	}
	handler, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("discord.userCommand %q handler is not a function", name)))
	}
	d.userCommands = append(d.userCommands, &commandDraft{name: name, commandType: "user", handler: handler})
	return goja.Undefined()
}

func (d *botDraft) messageCommand(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 2 {
		panic(vm.NewGoError(fmt.Errorf("discord.messageCommand expects messageCommand(name, handler)")))
	}
	name := strings.TrimSpace(call.Arguments[0].String())
	if name == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.messageCommand name is empty")))
	}
	handler, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("discord.messageCommand %q handler is not a function", name)))
	}
	d.messageCommands = append(d.messageCommands, &commandDraft{name: name, commandType: "message", handler: handler})
	return goja.Undefined()
}

func (d *botDraft) subcommand(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 3 || len(call.Arguments) > 4 {
		panic(vm.NewGoError(fmt.Errorf("discord.subcommand expects subcommand(rootName, name, [spec], handler)")))
	}
	rootName := strings.TrimSpace(call.Arguments[0].String())
	if rootName == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.subcommand rootName is empty")))
	}
	name := strings.TrimSpace(call.Arguments[1].String())
	if name == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.subcommand name is empty")))
	}
	var spec map[string]any
	var handler goja.Callable
	var ok bool
	if len(call.Arguments) == 3 {
		handler, ok = goja.AssertFunction(call.Arguments[2])
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("discord.subcommand %q/%q handler is not a function", rootName, name)))
		}
	} else {
		spec = exportMap(call.Arguments[2])
		handler, ok = goja.AssertFunction(call.Arguments[3])
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("discord.subcommand %q/%q handler is not a function", rootName, name)))
		}
	}
	d.subcommands = append(d.subcommands, &subcommandDraft{rootName: rootName, name: name, spec: spec, handler: handler})
	return goja.Undefined()
}

func (d *botDraft) event(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 2 {
		panic(vm.NewGoError(fmt.Errorf("discord.event expects event(name, handler)")))
	}
	name := strings.TrimSpace(call.Arguments[0].String())
	if name == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.event name is empty")))
	}
	handler, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("discord.event %q handler is not a function", name)))
	}
	d.events = append(d.events, &eventDraft{name: name, handler: handler})
	return goja.Undefined()
}

func (d *botDraft) component(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 2 {
		panic(vm.NewGoError(fmt.Errorf("discord.component expects component(customId, handler)")))
	}
	customID := strings.TrimSpace(call.Arguments[0].String())
	if customID == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.component customId is empty")))
	}
	handler, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("discord.component %q handler is not a function", customID)))
	}
	d.components = append(d.components, &componentDraft{customID: customID, handler: handler})
	return goja.Undefined()
}

func (d *botDraft) modal(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 2 {
		panic(vm.NewGoError(fmt.Errorf("discord.modal expects modal(customId, handler)")))
	}
	customID := strings.TrimSpace(call.Arguments[0].String())
	if customID == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.modal customId is empty")))
	}
	handler, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("discord.modal %q handler is not a function", customID)))
	}
	d.modals = append(d.modals, &modalDraft{customID: customID, handler: handler})
	return goja.Undefined()
}

func (d *botDraft) autocomplete(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 3 {
		panic(vm.NewGoError(fmt.Errorf("discord.autocomplete expects autocomplete(commandName, optionName, handler)")))
	}
	commandName := strings.TrimSpace(call.Arguments[0].String())
	if commandName == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.autocomplete command name is empty")))
	}
	optionName := strings.TrimSpace(call.Arguments[1].String())
	if optionName == "" {
		panic(vm.NewGoError(fmt.Errorf("discord.autocomplete option name is empty")))
	}
	handler, ok := goja.AssertFunction(call.Arguments[2])
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("discord.autocomplete %q/%q handler is not a function", commandName, optionName)))
	}
	d.autocompletes = append(d.autocompletes, &autocompleteDraft{commandName: commandName, optionName: optionName, handler: handler})
	return goja.Undefined()
}

func (d *botDraft) configure(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 1 {
		panic(vm.NewGoError(fmt.Errorf("discord.configure expects configure(options)")))
	}
	options := exportMap(call.Arguments[0])
	for key, value := range options {
		d.metadata[key] = value
	}
	return goja.Undefined()
}

func (d *botDraft) finalize(vm *goja.Runtime) goja.Value {
	bot := vm.NewObject()
	_ = bot.Set("kind", "discord.bot")
	_ = bot.Set("metadata", cloneMap(d.metadata))
	_ = bot.Set("commands", commandSnapshotsFromDrafts(append(append(append([]*commandDraft(nil), d.commands...), d.userCommands...), d.messageCommands...)))
	_ = bot.Set("events", eventSnapshotsFromDrafts(d.events))
	_ = bot.Set("components", componentSnapshotsFromDrafts(d.components))
	_ = bot.Set("modals", modalSnapshotsFromDrafts(d.modals))
	_ = bot.Set("autocompletes", autocompleteSnapshotsFromDrafts(d.autocompletes))

	commands := append([]*commandDraft(nil), d.commands...)
	userCommands := append([]*commandDraft(nil), d.userCommands...)
	messageCommands := append([]*commandDraft(nil), d.messageCommands...)
	subcommands := append([]*subcommandDraft(nil), d.subcommands...)
	events := append([]*eventDraft(nil), d.events...)
	components := append([]*componentDraft(nil), d.components...)
	modals := append([]*modalDraft(nil), d.modals...)
	autocompletes := append([]*autocompleteDraft(nil), d.autocompletes...)
	store := d.store
	metadata := cloneMap(d.metadata)
	moduleName := d.moduleName

	_ = bot.Set("describe", func(goja.FunctionCall) goja.Value {
		return vm.ToValue(map[string]any{
			"kind":          "discord.bot",
			"metadata":      cloneMap(metadata),
			"commands":      commandSnapshotsFromDrafts(append(append(append([]*commandDraft(nil), commands...), userCommands...), messageCommands...)),
			"events":        eventSnapshotsFromDrafts(events),
			"components":    componentSnapshotsFromDrafts(components),
			"modals":        modalSnapshotsFromDrafts(modals),
			"autocompletes": autocompleteSnapshotsFromDrafts(autocompletes),
		})
	})
	_ = bot.Set("dispatchCommand", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchCommand expects one input object")))
		}
		input := objectFromValue(vm, call.Arguments[0])
		name := strings.TrimSpace(input.Get("name").String())
		if name == "" {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchCommand input name is empty")))
		}
		command := findCommand(commands, name)
		if command == nil {
			command = findCommand(userCommands, name)
		}
		if command == nil {
			command = findCommand(messageCommands, name)
		}
		if command == nil {
			panic(vm.NewGoError(fmt.Errorf("discord bot %q has no command named %q", moduleName, name)))
		}
		ctx := buildContext(vm, store, input, "command", name, metadata)
		result, err := command.handler(goja.Undefined(), ctx)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return result
	})
	_ = bot.Set("dispatchSubcommand", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchSubcommand expects one input object")))
		}
		input := objectFromValue(vm, call.Arguments[0])
		rootName := strings.TrimSpace(input.Get("rootName").String())
		subName := strings.TrimSpace(input.Get("subName").String())
		if rootName == "" || subName == "" {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchSubcommand input rootName and subName are required")))
		}
		sub := findSubcommand(subcommands, rootName, subName)
		if sub == nil {
			panic(vm.NewGoError(fmt.Errorf("discord bot %q has no subcommand %q/%q", moduleName, rootName, subName)))
		}
		ctx := buildContext(vm, store, input, "subcommand", rootName+"/"+subName, metadata)
		result, err := sub.handler(goja.Undefined(), ctx)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return result
	})
	_ = bot.Set("dispatchEvent", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchEvent expects one input object")))
		}
		input := objectFromValue(vm, call.Arguments[0])
		name := strings.TrimSpace(input.Get("name").String())
		if name == "" {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchEvent input name is empty")))
		}
		matches := findEvents(events, name)
		if len(matches) == 0 {
			panic(vm.NewGoError(fmt.Errorf("discord bot %q has no event named %q", moduleName, name)))
		}
		ctx := buildContext(vm, store, input, "event", name, metadata)
		results := make([]any, 0, len(matches))
		for _, ev := range matches {
			result, err := ev.handler(goja.Undefined(), ctx)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			if !goja.IsUndefined(result) && !goja.IsNull(result) {
				results = append(results, result.Export())
			}
		}
		return vm.ToValue(results)
	})
	_ = bot.Set("dispatchComponent", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchComponent expects one input object")))
		}
		input := objectFromValue(vm, call.Arguments[0])
		name := strings.TrimSpace(input.Get("name").String())
		if name == "" {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchComponent input name is empty")))
		}
		component := findComponent(components, name)
		if component == nil {
			panic(vm.NewGoError(fmt.Errorf("discord bot %q has no component handler for %q", moduleName, name)))
		}
		ctx := buildContext(vm, store, input, "component", name, metadata)
		result, err := component.handler(goja.Undefined(), ctx)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return result
	})
	_ = bot.Set("dispatchModal", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchModal expects one input object")))
		}
		input := objectFromValue(vm, call.Arguments[0])
		name := strings.TrimSpace(input.Get("name").String())
		if name == "" {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchModal input name is empty")))
		}
		modal := findModal(modals, name)
		if modal == nil {
			panic(vm.NewGoError(fmt.Errorf("discord bot %q has no modal handler for %q", moduleName, name)))
		}
		ctx := buildContext(vm, store, input, "modal", name, metadata)
		result, err := modal.handler(goja.Undefined(), ctx)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return result
	})
	_ = bot.Set("dispatchAutocomplete", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchAutocomplete expects one input object")))
		}
		input := objectFromValue(vm, call.Arguments[0])
		commandName := strings.TrimSpace(input.Get("name").String())
		focused := exportMap(input.Get("focused"))
		optionName := strings.TrimSpace(fmt.Sprint(focused["name"]))
		if commandName == "" || optionName == "" {
			panic(vm.NewGoError(fmt.Errorf("discord.bot.dispatchAutocomplete requires command name and focused option name")))
		}
		autocomplete := findAutocomplete(autocompletes, commandName, optionName)
		if autocomplete == nil {
			panic(vm.NewGoError(fmt.Errorf("discord bot %q has no autocomplete handler for %q/%q", moduleName, commandName, optionName)))
		}
		ctx := buildContext(vm, store, input, "autocomplete", commandName+":"+optionName, metadata)
		result, err := autocomplete.handler(goja.Undefined(), ctx)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return result
	})
	return bot
}

func exportMap(value goja.Value) map[string]any {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return map[string]any{}
	}
	if exported, ok := value.Export().(map[string]any); ok {
		return exported
	}
	return map[string]any{"value": value.Export()}
}

func cloneMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func commandSnapshotsFromDrafts(commands []*commandDraft) []map[string]any {
	out := make([]map[string]any, 0, len(commands))
	for _, command := range commands {
		snapshot := map[string]any{"name": command.name}
		if len(command.spec) > 0 {
			snapshot["spec"] = cloneMap(command.spec)
		}
		out = append(out, snapshot)
	}
	return out
}

func eventSnapshotsFromDrafts(events []*eventDraft) []map[string]any {
	out := make([]map[string]any, 0, len(events))
	for _, event := range events {
		out = append(out, map[string]any{"name": event.name})
	}
	return out
}

func componentSnapshotsFromDrafts(components []*componentDraft) []map[string]any {
	out := make([]map[string]any, 0, len(components))
	for _, component := range components {
		out = append(out, map[string]any{"customId": component.customID})
	}
	return out
}

func modalSnapshotsFromDrafts(modals []*modalDraft) []map[string]any {
	out := make([]map[string]any, 0, len(modals))
	for _, modal := range modals {
		out = append(out, map[string]any{"customId": modal.customID})
	}
	return out
}

func autocompleteSnapshotsFromDrafts(autocompletes []*autocompleteDraft) []map[string]any {
	out := make([]map[string]any, 0, len(autocompletes))
	for _, autocomplete := range autocompletes {
		out = append(out, map[string]any{"commandName": autocomplete.commandName, "optionName": autocomplete.optionName})
	}
	return out
}

func findCommand(commands []*commandDraft, name string) *commandDraft {
	for _, command := range commands {
		if command != nil && command.name == name {
			return command
		}
	}
	return nil
}

func findSubcommand(subcommands []*subcommandDraft, rootName, name string) *subcommandDraft {
	for _, sub := range subcommands {
		if sub != nil && sub.rootName == rootName && sub.name == name {
			return sub
		}
	}
	return nil
}

func findEvents(events []*eventDraft, name string) []*eventDraft {
	matches := make([]*eventDraft, 0, 1)
	for _, event := range events {
		if event != nil && event.name == name {
			matches = append(matches, event)
		}
	}
	return matches
}

func findComponent(components []*componentDraft, customID string) *componentDraft {
	for _, component := range components {
		if component != nil && component.customID == customID {
			return component
		}
	}
	return nil
}

func findModal(modals []*modalDraft, customID string) *modalDraft {
	for _, modal := range modals {
		if modal != nil && modal.customID == customID {
			return modal
		}
	}
	return nil
}

func findAutocomplete(autocompletes []*autocompleteDraft, commandName, optionName string) *autocompleteDraft {
	for _, autocomplete := range autocompletes {
		if autocomplete != nil && autocomplete.commandName == commandName && autocomplete.optionName == optionName {
			return autocomplete
		}
	}
	return nil
}

func objectFromValue(vm *goja.Runtime, value goja.Value) *goja.Object {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return vm.NewObject()
	}
	obj := value.ToObject(vm)
	if obj == nil {
		return vm.NewObject()
	}
	return obj
}

func setObjectField(vm *goja.Runtime, obj *goja.Object, name string, value any) {
	if obj == nil {
		return
	}
	if value == nil {
		_ = obj.Set(name, vm.ToValue(nil))
		return
	}
	_ = obj.Set(name, value)
}

func buildDispatchInput(vm *goja.Runtime, ctx context.Context, request DispatchRequest) *goja.Object {
	input := vm.NewObject()
	setObjectField(vm, input, "name", request.Name)
	setObjectField(vm, input, "args", request.Args)
	setObjectField(vm, input, "values", request.Values)
	setObjectField(vm, input, "command", request.Command)
	setObjectField(vm, input, "interaction", request.Interaction)
	setObjectField(vm, input, "message", request.Message)
	setObjectField(vm, input, "before", request.Before)
	setObjectField(vm, input, "user", request.User)
	setObjectField(vm, input, "guild", request.Guild)
	setObjectField(vm, input, "channel", request.Channel)
	setObjectField(vm, input, "member", request.Member)
	setObjectField(vm, input, "reaction", request.Reaction)
	setObjectField(vm, input, "me", request.Me)
	setObjectField(vm, input, "metadata", request.Metadata)
	setObjectField(vm, input, "config", request.Config)
	setObjectField(vm, input, "component", request.Component)
	setObjectField(vm, input, "modal", request.Modal)
	setObjectField(vm, input, "focused", request.Focused)
	if request.Discord != nil {
		setObjectField(vm, input, "discord", discordOpsObject(vm, ctx, request.Discord))
	}
	if request.Reply != nil {
		_ = input.Set("reply", func(message any) error { return request.Reply(ctx, message) })
	} else {
		_ = input.Set("reply", func(any) error { return nil })
	}
	if request.FollowUp != nil {
		_ = input.Set("followUp", func(message any) error { return request.FollowUp(ctx, message) })
	} else {
		_ = input.Set("followUp", func(any) error { return nil })
	}
	if request.Edit != nil {
		_ = input.Set("edit", func(message any) error { return request.Edit(ctx, message) })
	} else {
		_ = input.Set("edit", func(any) error { return nil })
	}
	if request.Defer != nil {
		_ = input.Set("defer", func(message any) error { return request.Defer(ctx, message) })
	} else {
		_ = input.Set("defer", func(any) error { return nil })
	}
	if request.ShowModal != nil {
		_ = input.Set("showModal", func(payload any) error { return request.ShowModal(ctx, payload) })
	} else {
		_ = input.Set("showModal", func(any) error { return fmt.Errorf("showModal is not available in this context") })
	}
	return input
}

func buildContext(vm *goja.Runtime, store *MemoryStore, input *goja.Object, kind, name string, metadata map[string]any) *goja.Object {
	ctx := vm.NewObject()
	setObjectField(vm, ctx, "args", input.Get("args"))
	setObjectField(vm, ctx, "options", input.Get("args"))
	setObjectField(vm, ctx, "values", input.Get("values"))
	setObjectField(vm, ctx, "command", input.Get("command"))
	setObjectField(vm, ctx, "interaction", input.Get("interaction"))
	setObjectField(vm, ctx, "message", input.Get("message"))
	setObjectField(vm, ctx, "before", input.Get("before"))
	setObjectField(vm, ctx, "user", input.Get("user"))
	setObjectField(vm, ctx, "guild", input.Get("guild"))
	setObjectField(vm, ctx, "channel", input.Get("channel"))
	setObjectField(vm, ctx, "member", input.Get("member"))
	setObjectField(vm, ctx, "reaction", input.Get("reaction"))
	setObjectField(vm, ctx, "me", input.Get("me"))
	setObjectField(vm, ctx, "metadata", input.Get("metadata"))
	setObjectField(vm, ctx, "config", input.Get("config"))
	setObjectField(vm, ctx, "component", input.Get("component"))
	setObjectField(vm, ctx, "modal", input.Get("modal"))
	setObjectField(vm, ctx, "focused", input.Get("focused"))
	setObjectField(vm, ctx, "discord", input.Get("discord"))
	_ = ctx.Set("store", storeObject(vm, store))
	_ = ctx.Set("log", loggerObject(vm, kind, name, metadata))
	if reply := input.Get("reply"); !goja.IsUndefined(reply) && !goja.IsNull(reply) {
		_ = ctx.Set("reply", reply)
	} else {
		_ = ctx.Set("reply", func(any) error { return nil })
	}
	if followUp := input.Get("followUp"); !goja.IsUndefined(followUp) && !goja.IsNull(followUp) {
		_ = ctx.Set("followUp", followUp)
	} else {
		_ = ctx.Set("followUp", func(any) error { return nil })
	}
	if edit := input.Get("edit"); !goja.IsUndefined(edit) && !goja.IsNull(edit) {
		_ = ctx.Set("edit", edit)
	} else {
		_ = ctx.Set("edit", func(any) error { return nil })
	}
	if def := input.Get("defer"); !goja.IsUndefined(def) && !goja.IsNull(def) {
		_ = ctx.Set("defer", def)
	} else {
		_ = ctx.Set("defer", func(any) error { return nil })
	}
	if showModal := input.Get("showModal"); !goja.IsUndefined(showModal) && !goja.IsNull(showModal) {
		_ = ctx.Set("showModal", showModal)
	} else {
		_ = ctx.Set("showModal", func(any) error { return fmt.Errorf("showModal is not available in this context") })
	}
	return ctx
}

func storeObject(vm *goja.Runtime, store *MemoryStore) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("get", func(key string, defaultValue any) any {
		if store == nil {
			return defaultValue
		}
		return store.Get(key, defaultValue)
	})
	_ = obj.Set("set", func(key string, value any) {
		if store != nil {
			store.Set(key, value)
		}
	})
	_ = obj.Set("delete", func(key string) bool {
		if store == nil {
			return false
		}
		return store.Delete(key)
	})
	_ = obj.Set("keys", func(prefix string) []string {
		if store == nil {
			return nil
		}
		return store.Keys(prefix)
	})
	_ = obj.Set("namespace", func(parts ...string) any {
		if store == nil {
			return storeObject(vm, NewMemoryStore().Namespace(parts...))
		}
		return storeObject(vm, store.Namespace(parts...))
	})
	return obj
}

func discordOpsObject(vm *goja.Runtime, ctx context.Context, ops *DiscordOps) *goja.Object {
	root := vm.NewObject()
	guilds := vm.NewObject()
	roles := vm.NewObject()
	threads := vm.NewObject()
	channels := vm.NewObject()
	messages := vm.NewObject()
	members := vm.NewObject()
	if ops == nil {
		_ = guilds.Set("fetch", func(string) any { return map[string]any{} })
		_ = roles.Set("list", func(string) any { return []map[string]any{} })
		_ = roles.Set("fetch", func(string, string) any { return map[string]any{} })
		_ = threads.Set("fetch", func(string) any { return map[string]any{} })
		_ = threads.Set("join", func(string) error { return nil })
		_ = threads.Set("leave", func(string) error { return nil })
		_ = threads.Set("start", func(string, any) any { return map[string]any{} })
		_ = channels.Set("send", func(string, any) error { return nil })
		_ = channels.Set("fetch", func(string) any { return map[string]any{} })
		_ = channels.Set("setTopic", func(string, string) error { return nil })
		_ = channels.Set("setSlowmode", func(string, int) error { return nil })
		_ = messages.Set("fetch", func(string, string) any { return map[string]any{} })
		_ = messages.Set("list", func(string, any) any { return []map[string]any{} })
		_ = messages.Set("edit", func(string, string, any) error { return nil })
		_ = messages.Set("delete", func(string, string) error { return nil })
		_ = messages.Set("react", func(string, string, string) error { return nil })
		_ = messages.Set("pin", func(string, string) error { return nil })
		_ = messages.Set("unpin", func(string, string) error { return nil })
		_ = messages.Set("listPinned", func(string) any { return []map[string]any{} })
		_ = messages.Set("bulkDelete", func(string, any) error { return nil })
		_ = members.Set("fetch", func(string, string) any { return map[string]any{} })
		_ = members.Set("list", func(string, any) any { return []map[string]any{} })
		_ = members.Set("addRole", func(string, string, string) error { return nil })
		_ = members.Set("removeRole", func(string, string, string) error { return nil })
		_ = members.Set("timeout", func(string, string, any) error { return nil })
		_ = members.Set("kick", func(string, string, any) error { return nil })
		_ = members.Set("ban", func(string, string, any) error { return nil })
		_ = members.Set("unban", func(string, string) error { return nil })
	} else {
		_ = guilds.Set("fetch", func(guildID string) (any, error) {
			if ops.GuildFetch == nil {
				return map[string]any{}, nil
			}
			return ops.GuildFetch(ctx, guildID)
		})
		_ = roles.Set("list", func(guildID string) (any, error) {
			if ops.RoleList == nil {
				return []map[string]any{}, nil
			}
			return ops.RoleList(ctx, guildID)
		})
		_ = roles.Set("fetch", func(guildID, roleID string) (any, error) {
			if ops.RoleFetch == nil {
				return map[string]any{}, nil
			}
			return ops.RoleFetch(ctx, guildID, roleID)
		})
		_ = threads.Set("fetch", func(threadID string) (any, error) {
			if ops.ThreadFetch == nil {
				return map[string]any{}, nil
			}
			return ops.ThreadFetch(ctx, threadID)
		})
		_ = threads.Set("join", func(threadID string) error {
			if ops.ThreadJoin == nil {
				return nil
			}
			return ops.ThreadJoin(ctx, threadID)
		})
		_ = threads.Set("leave", func(threadID string) error {
			if ops.ThreadLeave == nil {
				return nil
			}
			return ops.ThreadLeave(ctx, threadID)
		})
		_ = threads.Set("start", func(channelID string, payload any) (any, error) {
			if ops.ThreadStart == nil {
				return map[string]any{}, nil
			}
			return ops.ThreadStart(ctx, channelID, payload)
		})
		_ = channels.Set("send", func(channelID string, payload any) error {
			if ops.ChannelSend == nil {
				return nil
			}
			return ops.ChannelSend(ctx, channelID, payload)
		})
		_ = channels.Set("fetch", func(channelID string) (any, error) {
			if ops.ChannelFetch == nil {
				return map[string]any{}, nil
			}
			return ops.ChannelFetch(ctx, channelID)
		})
		_ = channels.Set("setTopic", func(channelID, topic string) error {
			if ops.ChannelSetTopic == nil {
				return nil
			}
			return ops.ChannelSetTopic(ctx, channelID, topic)
		})
		_ = channels.Set("setSlowmode", func(channelID string, seconds int) error {
			if ops.ChannelSetSlowmode == nil {
				return nil
			}
			return ops.ChannelSetSlowmode(ctx, channelID, seconds)
		})
		_ = messages.Set("fetch", func(channelID, messageID string) (any, error) {
			if ops.MessageFetch == nil {
				return map[string]any{}, nil
			}
			return ops.MessageFetch(ctx, channelID, messageID)
		})
		_ = messages.Set("list", func(channelID string, payload any) (any, error) {
			if ops.MessageList == nil {
				return []map[string]any{}, nil
			}
			return ops.MessageList(ctx, channelID, payload)
		})
		_ = messages.Set("edit", func(channelID, messageID string, payload any) error {
			if ops.MessageEdit == nil {
				return nil
			}
			return ops.MessageEdit(ctx, channelID, messageID, payload)
		})
		_ = messages.Set("delete", func(channelID, messageID string) error {
			if ops.MessageDelete == nil {
				return nil
			}
			return ops.MessageDelete(ctx, channelID, messageID)
		})
		_ = messages.Set("react", func(channelID, messageID, emoji string) error {
			if ops.MessageReact == nil {
				return nil
			}
			return ops.MessageReact(ctx, channelID, messageID, emoji)
		})
		_ = messages.Set("pin", func(channelID, messageID string) error {
			if ops.MessagePin == nil {
				return nil
			}
			return ops.MessagePin(ctx, channelID, messageID)
		})
		_ = messages.Set("unpin", func(channelID, messageID string) error {
			if ops.MessageUnpin == nil {
				return nil
			}
			return ops.MessageUnpin(ctx, channelID, messageID)
		})
		_ = messages.Set("listPinned", func(channelID string) (any, error) {
			if ops.MessageListPinned == nil {
				return []map[string]any{}, nil
			}
			return ops.MessageListPinned(ctx, channelID)
		})
		_ = messages.Set("bulkDelete", func(channelID string, messageIDs any) error {
			if ops.MessageBulkDelete == nil {
				return nil
			}
			return ops.MessageBulkDelete(ctx, channelID, messageIDs)
		})
		_ = members.Set("fetch", func(guildID, userID string) (any, error) {
			if ops.MemberFetch == nil {
				return map[string]any{}, nil
			}
			return ops.MemberFetch(ctx, guildID, userID)
		})
		_ = members.Set("list", func(guildID string, payload any) (any, error) {
			if ops.MemberList == nil {
				return []map[string]any{}, nil
			}
			return ops.MemberList(ctx, guildID, payload)
		})
		_ = members.Set("addRole", func(guildID, userID, roleID string) error {
			if ops.MemberAddRole == nil {
				return nil
			}
			return ops.MemberAddRole(ctx, guildID, userID, roleID)
		})
		_ = members.Set("removeRole", func(guildID, userID, roleID string) error {
			if ops.MemberRemoveRole == nil {
				return nil
			}
			return ops.MemberRemoveRole(ctx, guildID, userID, roleID)
		})
		_ = members.Set("timeout", func(guildID, userID string, payload any) error {
			if ops.MemberSetTimeout == nil {
				return nil
			}
			return ops.MemberSetTimeout(ctx, guildID, userID, payload)
		})
		_ = members.Set("kick", func(guildID, userID string, payload any) error {
			if ops.MemberKick == nil {
				return nil
			}
			return ops.MemberKick(ctx, guildID, userID, payload)
		})
		_ = members.Set("ban", func(guildID, userID string, payload any) error {
			if ops.MemberBan == nil {
				return nil
			}
			return ops.MemberBan(ctx, guildID, userID, payload)
		})
		_ = members.Set("unban", func(guildID, userID string) error {
			if ops.MemberUnban == nil {
				return nil
			}
			return ops.MemberUnban(ctx, guildID, userID)
		})
	}
	_ = root.Set("guilds", guilds)
	_ = root.Set("roles", roles)
	_ = root.Set("threads", threads)
	_ = root.Set("channels", channels)
	_ = root.Set("messages", messages)
	_ = root.Set("members", members)
	return root
}

func loggerObject(vm *goja.Runtime, kind, name string, metadata map[string]any) *goja.Object {
	obj := vm.NewObject()
	baseFields := map[string]any{"jsKind": kind, "jsName": name}
	for key, value := range metadata {
		baseFields["meta."+key] = value
	}
	setLogMethod := func(level string, fn func(string, map[string]any)) { _ = obj.Set(level, fn) }
	setLogMethod("info", func(msg string, fields map[string]any) {
		e := log.Info()
		applyFields(e, baseFields)
		applyFields(e, fields)
		e.Msg(msg)
	})
	setLogMethod("debug", func(msg string, fields map[string]any) {
		e := log.Debug()
		applyFields(e, baseFields)
		applyFields(e, fields)
		e.Msg(msg)
	})
	setLogMethod("warn", func(msg string, fields map[string]any) {
		e := log.Warn()
		applyFields(e, baseFields)
		applyFields(e, fields)
		e.Msg(msg)
	})
	setLogMethod("error", func(msg string, fields map[string]any) {
		e := log.Error()
		applyFields(e, baseFields)
		applyFields(e, fields)
		e.Msg(msg)
	})
	return obj
}

func applyFields(event *zerolog.Event, fields map[string]any) {
	if event == nil || len(fields) == 0 {
		return
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		event.Interface(key, fields[key])
	}
}
