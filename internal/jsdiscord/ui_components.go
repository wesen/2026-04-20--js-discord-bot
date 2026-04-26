package jsdiscord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

// ── Button builder ──────────────────────────────────────────────────────────

// ButtonBuilder holds the state for a Discord button component.
type ButtonBuilder struct {
	customID string
	label    string
	style    discordgo.ButtonStyle
	disabled bool
	emoji    *discordgo.ComponentEmoji
	url      string
}

var buttonAvailable = []string{"disabled", "emoji", "url", "build"}

// resolveButtonStyle converts a style string or number to a ButtonStyle.
func resolveButtonStyle(raw string) discordgo.ButtonStyle {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "primary":
		return discordgo.PrimaryButton
	case "secondary":
		return discordgo.SecondaryButton
	case "success", "green":
		return discordgo.SuccessButton
	case "danger", "red":
		return discordgo.DangerButton
	case "link":
		return discordgo.LinkButton
	default:
		return discordgo.PrimaryButton
	}
}

func newButtonBuilder(vm *goja.Runtime, customID, label, style string) goja.Value {
	b := &ButtonBuilder{
		customID: customID,
		label:    label,
		style:    resolveButtonStyle(style),
	}

	target := vm.NewObject()
	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			switch property {
			case "disabled":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.disabled = true
					return receiver
				})
			case "emoji":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					name := argString(call, 0)
					if name == "" {
						panic(vm.NewTypeError("ui.button.emoji: name is required"))
					}
					b.emoji = &discordgo.ComponentEmoji{Name: name}
					return receiver
				})
			case "url":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.url = argString(call, 0)
					b.style = discordgo.LinkButton
					return receiver
				})
			case "build":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(b.build())
				})
			default:
				checkMethod(vm, "ui.button", property, buttonAvailable)
				return goja.Undefined()
			}
		},
	})

	return vm.ToValue(proxy)
}

func (b *ButtonBuilder) build() discordgo.Button {
	btn := discordgo.Button{
		Style:    b.style,
		Label:    b.label,
		Disabled: b.disabled,
	}
	if b.style == discordgo.LinkButton {
		btn.URL = b.url
	} else {
		btn.CustomID = b.customID
	}
	if b.emoji != nil {
		btn.Emoji = b.emoji
	}
	return btn
}

// ── Select builder ──────────────────────────────────────────────────────────

// SelectBuilder holds the state for a Discord select menu component.
type SelectBuilder struct {
	customID    string
	placeholder string
	options     []discordgo.SelectMenuOption
	minValues   int
	maxValues   int
	disabled    bool
	menuType    discordgo.SelectMenuType
}

var selectAvailable = []string{"placeholder", "option", "options", "optionEntries", "minValues", "maxValues", "disabled", "build"}

// Typed selects (user, role, channel, mentionable) have a reduced method set.
var typedSelectAvailable = []string{"placeholder", "minValues", "maxValues", "disabled", "build"}

func newSelectBuilder(vm *goja.Runtime, customID string) goja.Value {
	return newSelectBuilderWithType(vm, customID, discordgo.StringSelectMenu)
}

func newTypedSelectBuilder(vm *goja.Runtime, customID, selectType string) goja.Value {
	var menuType discordgo.SelectMenuType
	switch selectType {
	case "userSelect":
		menuType = discordgo.UserSelectMenu
	case "roleSelect":
		menuType = discordgo.RoleSelectMenu
	case "channelSelect":
		menuType = discordgo.ChannelSelectMenu
	case "mentionableSelect":
		menuType = discordgo.MentionableSelectMenu
	default:
		menuType = discordgo.StringSelectMenu
	}
	return newSelectBuilderWithType(vm, customID, menuType)
}

func newSelectBuilderWithType(vm *goja.Runtime, customID string, menuType discordgo.SelectMenuType) goja.Value {
	isTypedSelect := menuType != discordgo.StringSelectMenu

	b := &SelectBuilder{
		customID: customID,
		menuType: menuType,
	}

	available := selectAvailable
	if isTypedSelect {
		available = typedSelectAvailable
	}

	target := vm.NewObject()
	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			switch property {
			case "placeholder":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.placeholder = argString(call, 0)
					return receiver
				})
			case "option":
				if isTypedSelect {
					// Typed selects don't have options — wrong-parent or unknown
					checkMethod(vm, "ui."+selectName(b.menuType), property, available)
					return goja.Undefined()
				}
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					label := argString(call, 0)
					value := argString(call, 1)
					description := argString(call, 2)
					if label == "" || value == "" {
						panic(vm.NewTypeError("ui.select.option: label and value are required"))
					}
					if len(b.options) >= 25 {
						panic(vm.NewTypeError("ui.select: maximum 25 options exceeded"))
					}
					opt := discordgo.SelectMenuOption{Label: label, Value: value}
					if description != "" {
						opt.Description = description
					}
					b.options = append(b.options, opt)
					return receiver
				})
			case "options":
				if isTypedSelect {
					checkMethod(vm, "ui."+selectName(b.menuType), property, available)
					return goja.Undefined()
				}
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					items := argArrayMaps(vm, call.Argument(0))
					for _, item := range items {
						if len(b.options) >= 25 {
							panic(vm.NewTypeError("ui.select: maximum 25 options exceeded"))
						}
						label := fmtStr(item["label"])
						value := fmtStr(item["value"])
						if label == "" || value == "" {
							panic(vm.NewTypeError("ui.select.options: each option requires label and value"))
						}
						opt := discordgo.SelectMenuOption{Label: label, Value: value}
						if desc, ok := item["description"]; ok {
							opt.Description = fmtStr(desc)
						}
						if def, ok := item["default"]; ok {
							opt.Default = fmtBool(def)
						}
						b.options = append(b.options, opt)
					}
					return receiver
				})
			case "optionEntries":
				if isTypedSelect {
					checkMethod(vm, "ui."+selectName(b.menuType), property, available)
					return goja.Undefined()
				}
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					items := argArrayMaps(vm, call.Argument(0))
					selectedID := argString(call, 1)
					for _, item := range items {
						if len(b.options) >= 25 {
							panic(vm.NewTypeError("ui.select: maximum 25 options exceeded"))
						}
						label := fmtStr(item["label"])
						if label == "" {
							label = fmtStr(item["title"])
						}
						if label == "" {
							label = fmtStr(item["name"])
						}
						value := fmtStr(item["value"])
						if value == "" {
							value = fmtStr(item["id"])
						}
						if label == "" || value == "" {
							panic(vm.NewTypeError("ui.select.optionEntries: each entry requires label or value"))
						}
						description := fmtStr(item["description"])
						opt := discordgo.SelectMenuOption{Label: truncateString(label, 100), Value: value}
						if description != "" {
							opt.Description = truncateString(description, 100)
						}
						if value == selectedID {
							opt.Default = true
						}
						b.options = append(b.options, opt)
					}
					return receiver
				})
			case "minValues":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.minValues = argInt(call, 0)
					return receiver
				})
			case "maxValues":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.maxValues = argInt(call, 0)
					return receiver
				})
			case "disabled":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.disabled = true
					return receiver
				})
			case "build":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(b.build())
				})
			default:
				builderName := "ui.select"
				if isTypedSelect {
					builderName = "ui." + selectName(b.menuType)
				}
				checkMethod(vm, builderName, property, available)
				return goja.Undefined()
			}
		},
	})

	return vm.ToValue(proxy)
}

func (b *SelectBuilder) build() discordgo.SelectMenu {
	menu := discordgo.SelectMenu{
		MenuType:    b.menuType,
		CustomID:    b.customID,
		Placeholder: b.placeholder,
		MinValues:   &b.minValues,
		MaxValues:   b.maxValues,
		Disabled:    b.disabled,
	}
	if b.options != nil {
		menu.Options = b.options
	}
	return menu
}

func selectName(menuType discordgo.SelectMenuType) string {
	switch menuType { //nolint:exhaustive
	case discordgo.UserSelectMenu:
		return "userSelect"
	case discordgo.RoleSelectMenu:
		return "roleSelect"
	case discordgo.ChannelSelectMenu:
		return "channelSelect"
	case discordgo.MentionableSelectMenu:
		return "mentionableSelect"
	default:
		return "select"
	}
}

// truncateString truncates text to maxLen with ellipsis.
func truncateString(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return strings.TrimSpace(text[:maxLen-3]) + "..."
}
