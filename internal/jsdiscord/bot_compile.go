package jsdiscord

import (
	"context"
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/runtimebridge"
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
	RootName    string
	SubName     string
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
			// Bot didn't register this event; silently skip (not an error)
			return vm.ToValue([]any{})
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
		if command.commandType != "" {
			snapshot["type"] = command.commandType
		}
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
