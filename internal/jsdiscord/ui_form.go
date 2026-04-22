package jsdiscord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

// ── Form (modal) builder ────────────────────────────────────────────────────

// FormBuilder holds the state for a Discord modal form.
type FormBuilder struct {
	customID string
	title    string
	rows     []discordgo.MessageComponent
	// currentField tracks the field being configured by chain methods.
	currentField *discordgo.TextInput
}

var formAvailable = []string{"text", "textarea", "required", "value", "placeholder", "min", "max", "build"}

func newFormBuilder(vm *goja.Runtime, customID, title string) goja.Value {
	b := &FormBuilder{
		customID: customID,
		title:    title,
	}

	target := vm.NewObject()
	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			switch property {
			// text(customId, label) — adds a short text input field
			case "text":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.finishField()
					fieldID := argString(call, 0)
					label := argString(call, 1)
					if fieldID == "" {
						panic(vm.NewTypeError("ui.form.text: customId is required"))
					}
					if label == "" {
						panic(vm.NewTypeError("ui.form.text: label is required"))
					}
					if len(b.rows) >= 5 {
						panic(vm.NewTypeError("ui.form: maximum 5 fields exceeded"))
					}
					b.currentField = &discordgo.TextInput{
						CustomID: fieldID,
						Label:    label,
						Style:    discordgo.TextInputShort,
					}
					return receiver
				})
			// textarea(customId, label) — adds a paragraph text input field
			case "textarea":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.finishField()
					fieldID := argString(call, 0)
					label := argString(call, 1)
					if fieldID == "" {
						panic(vm.NewTypeError("ui.form.textarea: customId is required"))
					}
					if label == "" {
						panic(vm.NewTypeError("ui.form.textarea: label is required"))
					}
					if len(b.rows) >= 5 {
						panic(vm.NewTypeError("ui.form: maximum 5 fields exceeded"))
					}
					b.currentField = &discordgo.TextInput{
						CustomID: fieldID,
						Label:    label,
						Style:    discordgo.TextInputParagraph,
					}
					return receiver
				})
			// required() — marks the current field as required
			case "required":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if b.currentField == nil {
						panic(vm.NewTypeError("ui.form.required: no active field — call .text() or .textarea() first"))
					}
					b.currentField.Required = true
					return receiver
				})
			// value(v) — sets the default value of the current field
			case "value":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if b.currentField == nil {
						panic(vm.NewTypeError("ui.form.value: no active field — call .text() or .textarea() first"))
					}
					b.currentField.Value = argString(call, 0)
					return receiver
				})
			// placeholder(p) — sets placeholder on the current field
			case "placeholder":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if b.currentField == nil {
						panic(vm.NewTypeError("ui.form.placeholder: no active field — call .text() or .textarea() first"))
					}
					b.currentField.Placeholder = argString(call, 0)
					return receiver
				})
			// min(n) — sets minLength on the current field
			case "min":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if b.currentField == nil {
						panic(vm.NewTypeError("ui.form.min: no active field — call .text() or .textarea() first"))
					}
					b.currentField.MinLength = argInt(call, 0)
					return receiver
				})
			// max(n) — sets maxLength on the current field
			case "max":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if b.currentField == nil {
						panic(vm.NewTypeError("ui.form.max: no active field — call .text() or .textarea() first"))
					}
					b.currentField.MaxLength = argInt(call, 0)
					return receiver
				})
			case "build":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.finishField()
					return vm.ToValue(b.build())
				})
			default:
				checkMethod(vm, "ui.form", property, formAvailable)
				return goja.Undefined()
			}
		},
	})

	return vm.ToValue(proxy)
}

// finishField appends the current field (if any) to the rows.
func (b *FormBuilder) finishField() {
	if b.currentField != nil {
		field := *b.currentField
		b.rows = append(b.rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{field},
		})
		b.currentField = nil
	}
}

// build returns a map[string]any compatible with normalizeModalPayload().
func (b *FormBuilder) build() map[string]any {
	return map[string]any{
		"customId":   b.customID,
		"title":      b.title,
		"components": b.rows,
	}
}
