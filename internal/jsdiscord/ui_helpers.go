package jsdiscord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

// uiRow creates an ActionsRow from builder proxy arguments.
// Each argument must be a button/select builder with a .build() method.
func uiRow(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	row := buildRowFromArgs(vm, call.Arguments)
	return vm.ToValue(row)
}

// uiPager creates an ActionsRow with Previous/Next navigation buttons.
// Supports optional controls object: { hasPrevious: bool, hasNext: bool }
func uiPager(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	prevID := argString(call, 0)
	nextID := argString(call, 1)

	if prevID == "" {
		prevID = "prev"
	}
	if nextID == "" {
		nextID = "next"
	}

	// Check for optional controls object (3rd arg)
	hasPrevious := true
	hasNext := true
	controls := argMap(vm, call.Argument(2))
	if controls != nil {
		if v, ok := controls["hasPrevious"]; ok {
			hasPrevious = fmtBool(v)
		}
		if v, ok := controls["hasNext"]; ok {
			hasNext = fmtBool(v)
		}
	}

	prevBtn := discordgo.Button{
		Style:    discordgo.SecondaryButton,
		Label:    "◀ Previous",
		CustomID: prevID,
		Disabled: !hasPrevious,
	}
	nextBtn := discordgo.Button{
		Style:    discordgo.SecondaryButton,
		Label:    "Next ▶",
		CustomID: nextID,
		Disabled: !hasNext,
	}

	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{prevBtn, nextBtn},
	}
	return vm.ToValue(row)
}

// uiActions creates an ActionsRow from an array of action definitions.
// Each definition is {id, label, style?}.
func uiActions(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	items := argArrayMaps(vm, call.Argument(0))
	if len(items) == 0 {
		panic(vm.NewTypeError("ui.actions: at least one action is required"))
	}
	if len(items) > 5 {
		panic(vm.NewTypeError("ui.actions: maximum 5 actions per row"))
	}

	components := make([]discordgo.MessageComponent, 0, len(items))
	for _, item := range items {
		id := fmtStr(item["id"])
		label := fmtStr(item["label"])
		if id == "" || label == "" {
			panic(vm.NewTypeError("ui.actions: each action requires id and label"))
		}
		style := resolveButtonStyle(fmtStr(item["style"]))
		if style == discordgo.LinkButton {
			// Actions use customIds, not URLs
			style = discordgo.PrimaryButton
		}
		components = append(components, discordgo.Button{
			Style:    style,
			Label:    label,
			CustomID: id,
		})
	}

	row := discordgo.ActionsRow{Components: components}
	return vm.ToValue(row)
}

// uiConfirm returns a *normalizedResponse with an embed and confirm/cancel buttons.
// Supports two calling conventions:
//   ui.confirm(message, confirmId, cancelId)            — Go-style (positional)
//   ui.confirm(confirmId, cancelId, {body, title, ...}) — JS-style (options object)
func uiConfirm(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	arg0 := argString(call, 0)
	arg1 := argString(call, 1)

	// Try to detect JS-style: ui.confirm(confirmId, cancelId, {body, title, ...})
	opts := argMap(vm, call.Argument(2))
	if opts != nil {
		// JS-style: (confirmId, cancelId, {body, title, confirmLabel, cancelLabel, confirmStyle})
		confirmID := arg0
		cancelID := arg1
		body := fmtStr(opts["body"])
		title := fmtStr(opts["title"])
		if title == "" {
			title = "Confirm"
		}
		if body == "" {
			body = "Are you sure?"
		}
		confirmLabel := fmtStr(opts["confirmLabel"])
		if confirmLabel == "" {
			confirmLabel = "Confirm"
		}
		cancelLabel := fmtStr(opts["cancelLabel"])
		if cancelLabel == "" {
			cancelLabel = "Cancel"
		}
		confirmStyle := discordgo.DangerButton
		if styleStr := fmtStr(opts["confirmStyle"]); styleStr != "" {
			confirmStyle = resolveButtonStyle(styleStr)
		}

		embed := &discordgo.MessageEmbed{
			Title:       title,
			Description: body,
			Color:       0xFEE75C,
		}
		row := discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Style: confirmStyle, Label: confirmLabel, CustomID: confirmID},
				discordgo.Button{Style: discordgo.SecondaryButton, Label: cancelLabel, CustomID: cancelID},
			},
		}
		return vm.ToValue(&normalizedResponse{
			Content:    body,
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{row},
			Ephemeral:  true,
		})
	}

	// Go-style: (message, confirmId, cancelId)
	cancelID := argString(call, 2)
	message := arg0
	confirmID := arg1

	if message == "" {
		message = "Are you sure?"
	}
	if confirmID == "" {
		confirmID = "confirm"
	}
	if cancelID == "" {
		cancelID = "cancel"
	}

	embed := &discordgo.MessageEmbed{
		Description: message,
		Color:       0xFEE75C,
	}

	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{Style: discordgo.DangerButton, Label: "Confirm", CustomID: confirmID},
			discordgo.Button{Style: discordgo.SecondaryButton, Label: "Cancel", CustomID: cancelID},
		},
	}

	return vm.ToValue(&normalizedResponse{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{row},
		Ephemeral:  true,
	})
}

// newCardBuilder returns an embed builder with an additional .meta() method
// for quickly adding inline key-value metadata fields.
func newCardBuilder(vm *goja.Runtime, title string) goja.Value {
	b := &EmbedBuilder{title: title}

	target := vm.NewObject()
	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			switch property {
			// Embed methods — delegate to the embed builder
			case "title":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.title = argString(call, 0)
					return receiver
				})
			case "description":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.description = argString(call, 0)
					return receiver
				})
			case "color":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.color = argInt(call, 0)
					return receiver
				})
			case "field":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					name := argString(call, 0)
					value := argString(call, 1)
					inline := argBool(call, 2)
					if len(b.fields) >= 25 {
						panic(vm.NewTypeError("ui.card: maximum 25 fields exceeded"))
					}
					b.fields = append(b.fields, discordgo.MessageEmbedField{Name: name, Value: value, Inline: inline})
					return receiver
				})
			case "footer":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.footerText = argString(call, 0)
					return receiver
				})
			case "timestamp":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.timestamp = true
					return receiver
				})
			// Card-specific method
			case "meta":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					name := argString(call, 0)
					value := argString(call, 1)
					if name == "" {
						panic(vm.NewTypeError("ui.card.meta: name is required"))
					}
					if len(b.fields) >= 25 {
						panic(vm.NewTypeError("ui.card: maximum 25 fields exceeded"))
					}
					displayValue := value
					if displayValue == "" {
						displayValue = "N/A"
					}
					b.fields = append(b.fields, discordgo.MessageEmbedField{
						Name:   name,
						Value:  displayValue,
						Inline: true,
					})
					return receiver
				})
			case "build":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(b.build())
				})
			default:
				// Card shares embed + meta available methods
				cardAvailable := []string{"title", "description", "color", "field", "footer", "timestamp", "meta", "build"}
				checkMethod(vm, "ui.card", property, cardAvailable)
				return goja.Undefined()
			}
		},
	})

	return vm.ToValue(proxy)
}

// newFlowHelper is a stub for Phase 5 — returns a simple namespace helper
// for generating component IDs.
func newFlowHelper(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	namespace := argString(call, 0)
	if namespace == "" {
		panic(vm.NewTypeError("ui.flow: namespace is required"))
	}

	target := vm.NewObject()
	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			switch property {
			case "id":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					suffix := argString(call, 0)
					return vm.ToValue(namespace + ":" + suffix)
				})
			case "componentIds":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					result := vm.NewObject()
					items := argArrayMaps(vm, call.Argument(0))
					for i, item := range items {
						name := fmtStr(item["name"])
						if name == "" {
							name = fmt.Sprintf("component%d", i)
						}
						_ = result.Set(name, namespace+":"+name)
					}
					return result
				})
			case "pagerIds":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					result := vm.NewObject()
					_ = result.Set("prev", namespace+":prev")
					_ = result.Set("next", namespace+":next")
					return result
				})
			default:
				checkMethod(vm, "ui.flow", property, []string{"id", "componentIds", "pagerIds"})
				return goja.Undefined()
			}
		},
	})

	return vm.ToValue(proxy)
}
