package jsdiscord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

// MessageBuilder holds the state for a Discord message payload being
// constructed via ui.message().
type MessageBuilder struct {
	content    string
	ephemeral  bool
	tts        bool
	embeds     []*discordgo.MessageEmbed
	components []discordgo.MessageComponent
	files      []*discordgo.File
	followUp   bool
}

var messageAvailable = []string{"content", "ephemeral", "tts", "embed", "row", "file", "followUp", "build"}

func newMessageBuilder(vm *goja.Runtime) goja.Value {
	b := &MessageBuilder{}

	target := vm.NewObject()
	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			switch property {
			case "content":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.content = argString(call, 0)
					return receiver
				})
			case "ephemeral":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.ephemeral = true
					return receiver
				})
			case "tts":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.tts = true
					return receiver
				})
			case "embed":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					arg := call.Argument(0)
					emb := extractEmbed(vm, arg)
					if len(b.embeds) >= 10 {
						panic(vm.NewTypeError("ui.message: maximum 10 embeds exceeded"))
					}
					b.embeds = append(b.embeds, emb)
					return receiver
				})
			case "row":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					if len(b.components) >= 5 {
						panic(vm.NewTypeError("ui.message: maximum 5 component rows exceeded"))
					}
					row := buildRowFromArgs(vm, call.Arguments)
					b.components = append(b.components, row)
					return receiver
				})
			case "file":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					name := argString(call, 0)
					content := argString(call, 1)
					if name == "" {
						panic(vm.NewTypeError("ui.message.file: name is required"))
					}
					b.files = append(b.files, &discordgo.File{Name: name, Reader: nil})
					// Store content as a string; the pipeline will handle it.
					// For now, files with Reader=nil are placeholders.
					_ = content
					return receiver
				})
			case "followUp":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.followUp = true
					return receiver
				})
			case "build":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(b.build())
				})
			default:
				checkMethod(vm, "ui.message", property, messageAvailable)
				return goja.Undefined() // unreachable
			}
		},
	})

	return vm.ToValue(proxy)
}

func (b *MessageBuilder) build() *normalizedResponse {
	return &normalizedResponse{
		Content:    b.content,
		Ephemeral:  b.ephemeral,
		TTS:        b.tts,
		Embeds:     b.embeds,
		Components: b.components,
		Files:      b.files,
		FollowUp:   b.followUp,
	}
}

// extractEmbed extracts a *discordgo.MessageEmbed from a builder proxy or panics
// with a type error if the argument is not a ui.embed() builder.
func extractEmbed(vm *goja.Runtime, arg goja.Value) *discordgo.MessageEmbed {
	// First: try already-built *discordgo.MessageEmbed
	if !goja.IsUndefined(arg) && arg != nil {
		if emb, ok := arg.Export().(*discordgo.MessageEmbed); ok {
			return emb
		}
	}

	obj, ok := arg.(*goja.Object)
	if !ok {
		typeMismatchError(vm, "ui.message", "embed", "a ui.embed() builder", arg)
		return nil // unreachable
	}
	buildVal := obj.Get("build")
	if buildVal == nil || goja.IsUndefined(buildVal) {
		typeMismatchError(vm, "ui.message", "embed", "a ui.embed() builder", arg)
		return nil // unreachable
	}
	buildFn, ok := goja.AssertFunction(buildVal)
	if !ok {
		typeMismatchError(vm, "ui.message", "embed", "a ui.embed() builder", arg)
		return nil // unreachable
	}
	result, err := buildFn(goja.Undefined())
	if err != nil {
		panic(err)
	}
	emb, ok := result.Export().(*discordgo.MessageEmbed)
	if !ok {
		typeMismatchError(vm, "ui.message", "embed", "a ui.embed() builder", arg)
		return nil // unreachable
	}
	return emb
}

// buildRowFromArgs takes variadic arguments and builds a discordgo.ActionsRow.
// It accepts builder proxies, already-built components, and pre-built action rows.
// If an argument is an ActionsRow (e.g. ui.pager()), its child components are flattened
// into the returned row so callers can write `.row(ui.pager(...))` without nesting rows.
func buildRowFromArgs(vm *goja.Runtime, args []goja.Value) discordgo.ActionsRow {
	var components []discordgo.MessageComponent
	for _, arg := range args {
		components = append(components, extractRowComponents(vm, arg)...)
	}
	if len(components) > 5 {
		panic(vm.NewTypeError("ui.message.row: maximum 5 components per row"))
	}
	return discordgo.ActionsRow{Components: components}
}

// extractRowComponents extracts one or more discordgo.MessageComponent values from
// a builder proxy, an already-built Go component, or an already-built row.
func extractRowComponents(vm *goja.Runtime, arg goja.Value) []discordgo.MessageComponent {
	// First: try already-built Go types
	if !goja.IsUndefined(arg) && arg != nil {
		exported := arg.Export()
		switch v := exported.(type) {
		case discordgo.ActionsRow:
			return v.Components
		case *discordgo.ActionsRow:
			if v != nil {
				return v.Components
			}
		case discordgo.MessageComponent:
			return []discordgo.MessageComponent{v}
		case discordgo.Button:
			return []discordgo.MessageComponent{v}
		case *discordgo.Button:
			if v != nil {
				return []discordgo.MessageComponent{*v}
			}
		case discordgo.SelectMenu:
			return []discordgo.MessageComponent{v}
		case *discordgo.SelectMenu:
			if v != nil {
				return []discordgo.MessageComponent{*v}
			}
		}
	}

	obj, ok := arg.(*goja.Object)
	if !ok {
		typeMismatchError(vm, "ui.message.row", "component", "a ui.button() or ui.select() builder", arg)
		return nil // unreachable
	}
	buildVal := obj.Get("build")
	if buildVal == nil || goja.IsUndefined(buildVal) {
		typeMismatchError(vm, "ui.message.row", "component", "a ui.button() or ui.select() builder", arg)
		return nil // unreachable
	}
	buildFn, ok := goja.AssertFunction(buildVal)
	if !ok {
		typeMismatchError(vm, "ui.message.row", "component", "a ui.button() or ui.select() builder", arg)
		return nil // unreachable
	}
	result, err := buildFn(goja.Undefined())
	if err != nil {
		panic(err)
	}
	exported := result.Export()
	switch v := exported.(type) {
	case discordgo.ActionsRow:
		return v.Components
	case *discordgo.ActionsRow:
		if v != nil {
			return v.Components
		}
	case discordgo.MessageComponent:
		return []discordgo.MessageComponent{v}
	case discordgo.Button:
		return []discordgo.MessageComponent{v}
	case discordgo.SelectMenu:
		return []discordgo.MessageComponent{v}
	}
	typeMismatchError(vm, "ui.message.row", "component", "a ui.button() or ui.select() builder", arg)
	return nil // unreachable
}

// extractEmbedBuilder attempts to get an EmbedBuilder from a proxy argument.
// Used by ui.card() which inherits embed methods.
func extractEmbedBuilder(vm *goja.Runtime, arg goja.Value) *EmbedBuilder {
	// The argument is a Proxy — we can't directly access the Go struct.
	// Instead we use .build() and trust the result.
	_ = vm
	return nil // not used directly
}

// fmtStr is a helper for map[string]any values.
func fmtStr(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

// fmtBool is a helper for map[string]any values.
func fmtBool(v any) bool {
	b, _ := v.(bool)
	return b
}
