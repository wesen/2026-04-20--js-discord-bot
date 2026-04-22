package jsdiscord

import (
	"github.com/dop251/goja"
)

// Stub builders — will be implemented in Phase 3+.
// These are referenced by ui_module.go and must exist to compile.

func newButtonBuilder(vm *goja.Runtime, customId, label, style string) goja.Value {
	panic(vm.NewTypeError("ui.button: not yet implemented"))
}

func newSelectBuilder(vm *goja.Runtime, customId string) goja.Value {
	panic(vm.NewTypeError("ui.select: not yet implemented"))
}

func newTypedSelectBuilder(vm *goja.Runtime, customId, selectType string) goja.Value {
	panic(vm.NewTypeError("ui." + selectType + ": not yet implemented"))
}

func newFormBuilder(vm *goja.Runtime, customId, title string) goja.Value {
	panic(vm.NewTypeError("ui.form: not yet implemented"))
}

func newFlowHelper(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	panic(vm.NewTypeError("ui.flow: not yet implemented"))
}

func newCardBuilder(vm *goja.Runtime, title string) goja.Value {
	panic(vm.NewTypeError("ui.card: not yet implemented"))
}

func uiRow(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	panic(vm.NewTypeError("ui.row: not yet implemented"))
}

func uiPager(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	panic(vm.NewTypeError("ui.pager: not yet implemented"))
}

func uiActions(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	panic(vm.NewTypeError("ui.actions: not yet implemented"))
}

func uiConfirm(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	panic(vm.NewTypeError("ui.confirm: not yet implemented"))
}
