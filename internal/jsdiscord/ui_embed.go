package jsdiscord

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

// EmbedBuilder holds the state for a Discord embed being constructed via ui.embed().
type EmbedBuilder struct {
	title       string
	description string
	color       int
	fields      []discordgo.MessageEmbedField
	footerText  string
	authorName  string
	timestamp   bool
}

var embedAvailable = []string{"title", "description", "color", "field", "fields", "footer", "author", "timestamp", "build"}

func newEmbedBuilder(vm *goja.Runtime, title string) goja.Value {
	b := &EmbedBuilder{title: title}

	target := vm.NewObject()
	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			switch property {
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
						panic(vm.NewTypeError("ui.embed: maximum 25 fields exceeded"))
					}
					b.fields = append(b.fields, discordgo.MessageEmbedField{Name: name, Value: value, Inline: inline})
					return receiver
				})
			case "fields":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					items := argArrayMaps(vm, call.Argument(0))
					for _, item := range items {
						if len(b.fields) >= 25 {
							panic(vm.NewTypeError("ui.embed: maximum 25 fields exceeded"))
						}
						b.fields = append(b.fields, discordgo.MessageEmbedField{
							Name:   fmtStr(item["name"]),
							Value:  fmtStr(item["value"]),
							Inline: fmtBool(item["inline"]),
						})
					}
					return receiver
				})
			case "footer":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.footerText = argString(call, 0)
					return receiver
				})
			case "author":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.authorName = argString(call, 0)
					return receiver
				})
			case "timestamp":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					b.timestamp = true
					return receiver
				})
			case "build":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(b.build())
				})
			default:
				checkMethod(vm, "ui.embed", property, embedAvailable)
				return goja.Undefined() // unreachable
			}
		},
	})

	return vm.ToValue(proxy)
}

func (b *EmbedBuilder) build() *discordgo.MessageEmbed {
	e := &discordgo.MessageEmbed{}
	if b.title != "" {
		e.Title = b.title
	}
	if b.description != "" {
		e.Description = b.description
	}
	if b.color != 0 {
		e.Color = b.color
	}
	if len(b.fields) > 0 {
		e.Fields = make([]*discordgo.MessageEmbedField, len(b.fields))
		for i := range b.fields {
			f := b.fields[i]
			e.Fields[i] = &f
		}
	}
	if b.footerText != "" {
		e.Footer = &discordgo.MessageEmbedFooter{Text: b.footerText}
	}
	if b.authorName != "" {
		e.Author = &discordgo.MessageEmbedAuthor{Name: b.authorName}
	}
	if b.timestamp {
		e.Timestamp = time.Now().Format(time.RFC3339)
	}
	return e
}
