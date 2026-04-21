package jsdiscord

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
)

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
	setObjectField(vm, input, "rootName", request.RootName)
	setObjectField(vm, input, "subName", request.SubName)
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
