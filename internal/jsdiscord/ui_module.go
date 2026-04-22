package jsdiscord

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
)

// UIRegistrar registers the "ui" native Goja module that provides type-safe
// Discord UI builder primitives using ES6 Proxy traps.
type UIRegistrar struct{}

func (r *UIRegistrar) ID() string { return "discord-ui-registrar" }

func (r *UIRegistrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
	if reg == nil {
		return fmt.Errorf("require registry is nil")
	}
	reg.RegisterNativeModule("ui", uiLoader)
	return nil
}

// uiLoader is the Goja native module loader for require("ui").
// It exports all builder constructors and helper functions.
func uiLoader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	// Builders
	_ = exports.Set("message", func(call goja.FunctionCall) goja.Value {
		return newMessageBuilder(vm)
	})
	_ = exports.Set("embed", func(call goja.FunctionCall) goja.Value {
		title := argString(call, 0)
		return newEmbedBuilder(vm, title)
	})
	_ = exports.Set("button", func(call goja.FunctionCall) goja.Value {
		customId := argString(call, 0)
		label := argString(call, 1)
		style := argString(call, 2)
		return newButtonBuilder(vm, customId, label, style)
	})
	_ = exports.Set("select", func(call goja.FunctionCall) goja.Value {
		customId := argString(call, 0)
		return newSelectBuilder(vm, customId)
	})
	_ = exports.Set("userSelect", func(call goja.FunctionCall) goja.Value {
		customId := argString(call, 0)
		return newTypedSelectBuilder(vm, customId, "userSelect")
	})
	_ = exports.Set("roleSelect", func(call goja.FunctionCall) goja.Value {
		customId := argString(call, 0)
		return newTypedSelectBuilder(vm, customId, "roleSelect")
	})
	_ = exports.Set("channelSelect", func(call goja.FunctionCall) goja.Value {
		customId := argString(call, 0)
		return newTypedSelectBuilder(vm, customId, "channelSelect")
	})
	_ = exports.Set("mentionableSelect", func(call goja.FunctionCall) goja.Value {
		customId := argString(call, 0)
		return newTypedSelectBuilder(vm, customId, "mentionableSelect")
	})
	_ = exports.Set("form", func(call goja.FunctionCall) goja.Value {
		customId := argString(call, 0)
		title := argString(call, 1)
		return newFormBuilder(vm, customId, title)
	})
	_ = exports.Set("flow", func(call goja.FunctionCall) goja.Value {
		return newFlowHelper(vm, call)
	})

	// Helpers
	_ = exports.Set("row", func(call goja.FunctionCall) goja.Value {
		return uiRow(vm, call)
	})
	_ = exports.Set("pager", func(call goja.FunctionCall) goja.Value {
		return uiPager(vm, call)
	})
	_ = exports.Set("actions", func(call goja.FunctionCall) goja.Value {
		return uiActions(vm, call)
	})
	_ = exports.Set("confirm", func(call goja.FunctionCall) goja.Value {
		return uiConfirm(vm, call)
	})
	_ = exports.Set("card", func(call goja.FunctionCall) goja.Value {
		title := argString(call, 0)
		return newCardBuilder(vm, title)
	})
	_ = exports.Set("ok", func(call goja.FunctionCall) goja.Value {
		content := argString(call, 0)
		return vm.ToValue(&normalizedResponse{Content: content, Ephemeral: true})
	})
	_ = exports.Set("error", func(call goja.FunctionCall) goja.Value {
		content := argString(call, 0)
		return vm.ToValue(&normalizedResponse{Content: "⚠️ " + content, Ephemeral: true})
	})
	_ = exports.Set("emptyResults", func(call goja.FunctionCall) goja.Value {
		query := argString(call, 0)
		msg := "No results found."
		if query != "" {
			msg = "No results found for **" + query + "**."
		}
		return vm.ToValue(&normalizedResponse{Content: msg, Ephemeral: true})
	})
}

// argString extracts argument i as a string, defaulting to "".
func argString(call goja.FunctionCall, i int) string {
	if i >= len(call.Arguments) || call.Arguments[i] == nil || goja.IsUndefined(call.Arguments[i]) {
		return ""
	}
	return call.Arguments[i].String()
}

// argInt extracts argument i as an int, defaulting to 0.
func argInt(call goja.FunctionCall, i int) int {
	if i >= len(call.Arguments) || call.Arguments[i] == nil || goja.IsUndefined(call.Arguments[i]) {
		return 0
	}
	v := call.Arguments[i].Export()
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// argBool extracts argument i as a bool, defaulting to false.
func argBool(call goja.FunctionCall, i int) bool {
	if i >= len(call.Arguments) || call.Arguments[i] == nil || goja.IsUndefined(call.Arguments[i]) {
		return false
	}
	b, ok := call.Arguments[i].Export().(bool)
	return ok && b
}

// argArrayMaps extracts a JS array of objects into []map[string]any.
func argArrayMaps(vm *goja.Runtime, arg goja.Value) []map[string]any {
	if arg == nil || goja.IsUndefined(arg) {
		return nil
	}
	obj, ok := arg.(*goja.Object)
	if !ok {
		return nil
	}
	export := obj.Export()
	arr, ok := export.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}
