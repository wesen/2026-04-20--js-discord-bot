package jsdiscord

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

// methodOwner maps every known DSL method to the builder it belongs to.
// Used by each Proxy Get trap to produce "wrong parent" error messages.
var methodOwner = map[string]string{
	// MessageBuilder
	"content":  "ui.message()",
	"ephemeral": "ui.message()",
	"tts":      "ui.message()",
	"embed":    "ui.message()",
	"row":      "ui.message()",
	"file":     "ui.message()",

	// EmbedBuilder / CardBuilder
	"description": "ui.embed()",
	"color":      "ui.embed()",
	"field":      "ui.embed()",
	"fields":     "ui.embed()",
	"footer":     "ui.embed()",
	"author":     "ui.embed()",
	"timestamp":  "ui.embed()",
	"meta":       "ui.card()",

	// ButtonBuilder
	"disabled": "ui.button()",
	"emoji":    "ui.button()",

	// SelectBuilder
	"option":      "ui.select()",
	"options":     "ui.select()",
	"minValues":   "ui.select()",
	"maxValues":   "ui.select()",

	// FormBuilder
	"text":       "ui.form()",
	"textarea":   "ui.form()",
	"required":   "ui.form()",
	"value":      "ui.form()",
	"placeholder": "ui.form()",
	"min":        "ui.form()",
	"max":        "ui.form()",
}

// wrongParentError produces a Goja TypeError explaining that `method` belongs
// to a different builder.
func wrongParentError(vm *goja.Runtime, builderName, method string) goja.Value {
	owner, ok := methodOwner[method]
	if !ok {
		owner = "(unknown builder)"
	}
	panic(vm.NewTypeError(
		fmt.Sprintf("%s: .%s() is not available here. You probably meant to call this on %s.",
			builderName, method, owner)))
}

// unknownMethodError produces a Goja TypeError for a method that doesn't
// exist on any builder.
func unknownMethodError(vm *goja.Runtime, builderName, method string, available []string) goja.Value {
	panic(vm.NewTypeError(
		fmt.Sprintf("%s: unknown method .%s(). Available: %s.",
			builderName, method, strings.Join(available, ", "))))
}

// typeMismatchError produces a Goja TypeError explaining that a raw JS object
// was passed where a builder was expected.
func typeMismatchError(vm *goja.Runtime, builderName, method, expected string, got goja.Value) goja.Value {
	gotType := "unknown"
	if goja.IsUndefined(got) || goja.IsNull(got) {
		gotType = "undefined"
	} else if _, ok := got.Export().(map[string]any); ok {
		gotType = "object"
	} else {
		gotType = got.String()
	}
	panic(vm.NewTypeError(
		fmt.Sprintf("%s.%s: expected %s, got %s.",
			builderName, method, expected, gotType)))
}

// checkMethod is called by every Proxy Get trap. It handles the common
// wrong-parent and unknown-method branches after the builder's own methods
// have been checked via switch.
//
// Returns false if the method is unrecognized (caller should panic).
func checkMethod(vm *goja.Runtime, builderName, property string, available []string) {
	if _, ok := methodOwner[property]; ok {
		wrongParentError(vm, builderName, property)
	}
	unknownMethodError(vm, builderName, property, available)
}
