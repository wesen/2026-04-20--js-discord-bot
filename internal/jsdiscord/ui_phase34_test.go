package jsdiscord

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

// ── Button builder tests ────────────────────────────────────────────────────

func TestButtonBuilderBasicChain(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	buttonFn := mustAssertFunc(t, exports.Get("button"))

	builder, _ := buttonFn(goja.Undefined(), vm.ToValue("btn-1"), vm.ToValue("Click Me"), vm.ToValue("primary"))
	obj := builder.(*goja.Object)

	builder = mustCall(t, obj, "disabled")
	result := mustCall(t, builder.(*goja.Object), "build")

	btn, ok := result.Export().(discordgo.Button)
	require.True(t, ok)
	require.Equal(t, "btn-1", btn.CustomID)
	require.Equal(t, "Click Me", btn.Label)
	require.Equal(t, discordgo.PrimaryButton, btn.Style)
	require.True(t, btn.Disabled)
}

func TestButtonBuilderStyles(t *testing.T) {
	tests := []struct {
		input    string
		expected discordgo.ButtonStyle
	}{
		{"", discordgo.PrimaryButton},
		{"primary", discordgo.PrimaryButton},
		{"secondary", discordgo.SecondaryButton},
		{"success", discordgo.SuccessButton},
		{"green", discordgo.SuccessButton},
		{"danger", discordgo.DangerButton},
		{"red", discordgo.DangerButton},
		{"link", discordgo.LinkButton},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			vm := goja.New()
			exports := loadUIForTest(t, vm)
			buttonFn := mustAssertFunc(t, exports.Get("button"))

			builder, _ := buttonFn(goja.Undefined(), vm.ToValue("id"), vm.ToValue("Label"), vm.ToValue(tc.input))
			result := mustCall(t, builder.(*goja.Object), "build")

			btn := result.Export().(discordgo.Button)
			require.Equal(t, tc.expected, btn.Style)
		})
	}
}

func TestButtonBuilderLinkWithUrl(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	buttonFn := mustAssertFunc(t, exports.Get("button"))

	builder, _ := buttonFn(goja.Undefined(), vm.ToValue("btn-link"), vm.ToValue("Visit"), vm.ToValue("link"))
	builder = mustCall(t, builder.(*goja.Object), "url", vm.ToValue("https://example.com"))
	result := mustCall(t, builder.(*goja.Object), "build")

	btn := result.Export().(discordgo.Button)
	require.Equal(t, discordgo.LinkButton, btn.Style)
	require.Equal(t, "https://example.com", btn.URL)
	require.Empty(t, btn.CustomID) // Link buttons don't have custom IDs
}

func TestButtonBuilderEmoji(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	buttonFn := mustAssertFunc(t, exports.Get("button"))

	builder, _ := buttonFn(goja.Undefined(), vm.ToValue("e1"), vm.ToValue("React"), vm.ToValue(""))
	builder = mustCall(t, builder.(*goja.Object), "emoji", vm.ToValue("👍"))
	result := mustCall(t, builder.(*goja.Object), "build")

	btn := result.Export().(discordgo.Button)
	require.NotNil(t, btn.Emoji)
	require.Equal(t, "👍", btn.Emoji.Name)
}

func TestButtonBuilderWrongParent(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	buttonFn := mustAssertFunc(t, exports.Get("button"))

	builder, _ := buttonFn(goja.Undefined(), vm.ToValue("b1"), vm.ToValue("X"), vm.ToValue(""))
	obj := builder.(*goja.Object)

	_, err := tryCall(t, obj, "ephemeral")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ui.button")
}

func TestButtonBuilderUnknownMethod(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	buttonFn := mustAssertFunc(t, exports.Get("button"))

	builder, _ := buttonFn(goja.Undefined(), vm.ToValue("b1"), vm.ToValue("X"), vm.ToValue(""))
	obj := builder.(*goja.Object)

	_, err := tryCall(t, obj, "bogus")
	require.Error(t, err)
}

// ── Select builder tests ────────────────────────────────────────────────────

func TestSelectBuilderBasicChain(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	selectFn := mustAssertFunc(t, exports.Get("select"))

	builder, _ := selectFn(goja.Undefined(), vm.ToValue("menu-1"))
	obj := builder.(*goja.Object)

	builder = mustCall(t, obj, "placeholder", vm.ToValue("Choose one..."))
	builder = mustCall(t, builder.(*goja.Object), "option", vm.ToValue("Red"), vm.ToValue("red"), vm.ToValue("The color red"))
	builder = mustCall(t, builder.(*goja.Object), "option", vm.ToValue("Blue"), vm.ToValue("blue"))
	builder = mustCall(t, builder.(*goja.Object), "minValues", vm.ToValue(1))
	builder = mustCall(t, builder.(*goja.Object), "maxValues", vm.ToValue(3))
	result := mustCall(t, builder.(*goja.Object), "build")

	menu := result.Export().(discordgo.SelectMenu)
	require.Equal(t, "menu-1", menu.CustomID)
	require.Equal(t, "Choose one...", menu.Placeholder)
	require.Equal(t, discordgo.StringSelectMenu, menu.MenuType)
	require.Len(t, menu.Options, 2)
	require.Equal(t, "Red", menu.Options[0].Label)
	require.Equal(t, "red", menu.Options[0].Value)
	require.Equal(t, "The color red", menu.Options[0].Description)
	require.Equal(t, "Blue", menu.Options[1].Label)
	require.Equal(t, 1, *menu.MinValues)
	require.Equal(t, 3, menu.MaxValues)
}

func TestSelectBuilderOptionsFromArray(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	selectFn := mustAssertFunc(t, exports.Get("select"))

	builder, _ := selectFn(goja.Undefined(), vm.ToValue("pick"))
	obj := builder.(*goja.Object)

	arr := vm.NewArray()
	opts := []map[string]any{
		{"label": "A", "value": "a"},
		{"label": "B", "value": "b", "description": "Option B"},
	}
	for i, o := range opts {
		obj2 := vm.NewObject()
		for k, v := range o {
			_ = obj2.Set(k, v)
		}
		_ = arr.Set(fmt.Sprintf("%d", i), obj2)
	}

	builder = mustCall(t, obj, "options", arr)
	result := mustCall(t, builder.(*goja.Object), "build")

	menu := result.Export().(discordgo.SelectMenu)
	require.Len(t, menu.Options, 2)
	require.Equal(t, "Option B", menu.Options[1].Description)
}

func TestSelectBuilderMax25Options(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	selectFn := mustAssertFunc(t, exports.Get("select"))

	builder, _ := selectFn(goja.Undefined(), vm.ToValue("big"))
	obj := builder.(*goja.Object)

	for i := 0; i < 25; i++ {
		builder = mustCall(t, obj, "option", vm.ToValue(fmt.Sprintf("Opt%d", i)), vm.ToValue(fmt.Sprintf("v%d", i)))
		obj = builder.(*goja.Object)
	}

	_, err := tryCall(t, obj, "option", vm.ToValue("Extra"), vm.ToValue("x"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "maximum 25")
}

func TestSelectBuilderOptionRequiresLabelAndValue(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	selectFn := mustAssertFunc(t, exports.Get("select"))

	builder, _ := selectFn(goja.Undefined(), vm.ToValue("s1"))
	obj := builder.(*goja.Object)

	_, err := tryCall(t, obj, "option", vm.ToValue(""), vm.ToValue("val"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "label and value")
}

// ── Typed select tests ──────────────────────────────────────────────────────

func TestUserSelectBuilder(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	fn := mustAssertFunc(t, exports.Get("userSelect"))

	builder, _ := fn(goja.Undefined(), vm.ToValue("pick-user"))
	builder = mustCall(t, builder.(*goja.Object), "placeholder", vm.ToValue("Select a user"))
	result := mustCall(t, builder.(*goja.Object), "build")

	menu := result.Export().(discordgo.SelectMenu)
	require.Equal(t, discordgo.UserSelectMenu, menu.MenuType)
	require.Equal(t, "pick-user", menu.CustomID)
}

func TestRoleSelectBuilder(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	fn := mustAssertFunc(t, exports.Get("roleSelect"))

	builder, _ := fn(goja.Undefined(), vm.ToValue("pick-role"))
	result := mustCall(t, builder.(*goja.Object), "build")

	menu := result.Export().(discordgo.SelectMenu)
	require.Equal(t, discordgo.RoleSelectMenu, menu.MenuType)
}

func TestChannelSelectBuilder(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	fn := mustAssertFunc(t, exports.Get("channelSelect"))

	builder, _ := fn(goja.Undefined(), vm.ToValue("pick-channel"))
	result := mustCall(t, builder.(*goja.Object), "build")

	menu := result.Export().(discordgo.SelectMenu)
	require.Equal(t, discordgo.ChannelSelectMenu, menu.MenuType)
}

func TestMentionableSelectBuilder(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	fn := mustAssertFunc(t, exports.Get("mentionableSelect"))

	builder, _ := fn(goja.Undefined(), vm.ToValue("pick-mention"))
	result := mustCall(t, builder.(*goja.Object), "build")

	menu := result.Export().(discordgo.SelectMenu)
	require.Equal(t, discordgo.MentionableSelectMenu, menu.MenuType)
}

func TestTypedSelectRejectsOption(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	fn := mustAssertFunc(t, exports.Get("userSelect"))

	builder, _ := fn(goja.Undefined(), vm.ToValue("pick-user"))
	obj := builder.(*goja.Object)

	// .option() should not exist on typed selects
	_, err := tryCall(t, obj, "option", vm.ToValue("A"), vm.ToValue("a"))
	require.Error(t, err)
}

// ── Message + components integration ────────────────────────────────────────

func TestMessageWithButtonsInRow(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))
	buttonFn := mustAssertFunc(t, exports.Get("button"))

	b1, _ := buttonFn(goja.Undefined(), vm.ToValue("approve"), vm.ToValue("✅ Approve"), vm.ToValue("success"))

	b2, _ := buttonFn(goja.Undefined(), vm.ToValue("reject"), vm.ToValue("❌ Reject"), vm.ToValue("danger"))

	msg, _ := msgFn(goja.Undefined())
	msg = mustCall(t, msg.(*goja.Object), "content", vm.ToValue("Review this:"))
	msg = mustCall(t, msg.(*goja.Object), "row", b1, b2)
	msg = mustCall(t, msg.(*goja.Object), "ephemeral")
	result := mustCall(t, msg.(*goja.Object), "build")

	nr := result.Export().(*normalizedResponse)
	require.Len(t, nr.Components, 1)
	row := nr.Components[0].(discordgo.ActionsRow)
	require.Len(t, row.Components, 2)
	btn1 := row.Components[0].(discordgo.Button)
	require.Equal(t, "approve", btn1.CustomID)
	require.Equal(t, discordgo.SuccessButton, btn1.Style)
}

func TestMessageWithSelectInRow(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))
	selectFn := mustAssertFunc(t, exports.Get("select"))

	sel, _ := selectFn(goja.Undefined(), vm.ToValue("color-pick"))
	sel = mustCall(t, sel.(*goja.Object), "option", vm.ToValue("Red"), vm.ToValue("red"))

	msg, _ := msgFn(goja.Undefined())
	msg = mustCall(t, msg.(*goja.Object), "row", sel)
	result := mustCall(t, msg.(*goja.Object), "build")

	nr := result.Export().(*normalizedResponse)
	require.Len(t, nr.Components, 1)
	row := nr.Components[0].(discordgo.ActionsRow)
	require.Len(t, row.Components, 1)
	menu := row.Components[0].(discordgo.SelectMenu)
	require.Equal(t, discordgo.StringSelectMenu, menu.MenuType)
	require.Len(t, menu.Options, 1)
}

// ── Form builder tests ──────────────────────────────────────────────────────

func TestFormBuilderBasicChain(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	formFn := mustAssertFunc(t, exports.Get("form"))

	builder, _ := formFn(goja.Undefined(), vm.ToValue("submit-form"), vm.ToValue("My Form"))
	obj := builder.(*goja.Object)

	builder = mustCall(t, obj, "text", vm.ToValue("name"), vm.ToValue("Name"))
	builder = mustCall(t, builder.(*goja.Object), "required")
	builder = mustCall(t, builder.(*goja.Object), "placeholder", vm.ToValue("Enter your name"))
	builder = mustCall(t, builder.(*goja.Object), "textarea", vm.ToValue("desc"), vm.ToValue("Description"))
	builder = mustCall(t, builder.(*goja.Object), "value", vm.ToValue("Default text"))
	builder = mustCall(t, builder.(*goja.Object), "min", vm.ToValue(10))
	builder = mustCall(t, builder.(*goja.Object), "max", vm.ToValue(500))
	result := mustCall(t, builder.(*goja.Object), "build")

	payload := result.Export().(map[string]any)
	require.Equal(t, "submit-form", payload["customId"])
	require.Equal(t, "My Form", payload["title"])

	rows, ok := payload["components"].([]discordgo.MessageComponent)
	require.True(t, ok)
	require.Len(t, rows, 2)

	// First row: text input
	row1 := rows[0].(discordgo.ActionsRow)
	field1 := row1.Components[0].(discordgo.TextInput)
	require.Equal(t, "Name", field1.Label)
	require.Equal(t, "name", field1.CustomID)
	require.Equal(t, discordgo.TextInputShort, field1.Style)
	require.True(t, field1.Required)
	require.Equal(t, "Enter your name", field1.Placeholder)

	// Second row: textarea
	row2 := rows[1].(discordgo.ActionsRow)
	field2 := row2.Components[0].(discordgo.TextInput)
	require.Equal(t, "Description", field2.Label)
	require.Equal(t, "desc", field2.CustomID)
	require.Equal(t, discordgo.TextInputParagraph, field2.Style)
	require.Equal(t, "Default text", field2.Value)
	require.Equal(t, 10, field2.MinLength)
	require.Equal(t, 500, field2.MaxLength)
}

func TestFormBuilderMax5Fields(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	formFn := mustAssertFunc(t, exports.Get("form"))

	builder, _ := formFn(goja.Undefined(), vm.ToValue("big-form"), vm.ToValue("Big"))
	obj := builder.(*goja.Object)

	for i := 0; i < 5; i++ {
		builder = mustCall(t, obj, "text", vm.ToValue(fmt.Sprintf("f%d", i)), vm.ToValue(fmt.Sprintf("F%d", i)))
		obj = builder.(*goja.Object)
	}

	_, err := tryCall(t, obj, "text", vm.ToValue("extra"), vm.ToValue("Extra"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "maximum 5 fields")
}

func TestFormBuilderFieldModifierWithoutField(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	formFn := mustAssertFunc(t, exports.Get("form"))

	builder, _ := formFn(goja.Undefined(), vm.ToValue("f1"), vm.ToValue("F"))
	obj := builder.(*goja.Object)

	// Calling .required() before .text() should error
	_, err := tryCall(t, obj, "required")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no active field")
}

func TestFormBuilderTextRequiresLabelAndId(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	formFn := mustAssertFunc(t, exports.Get("form"))

	builder, _ := formFn(goja.Undefined(), vm.ToValue("f1"), vm.ToValue("F"))
	obj := builder.(*goja.Object)

	_, err := tryCall(t, obj, "text", vm.ToValue("id"), vm.ToValue(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "label is required")
}

func TestFormBuilderWrongParent(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	formFn := mustAssertFunc(t, exports.Get("form"))

	builder, _ := formFn(goja.Undefined(), vm.ToValue("f1"), vm.ToValue("F"))
	obj := builder.(*goja.Object)

	_, err := tryCall(t, obj, "ephemeral")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ui.form")
}

// ── Helper function tests ───────────────────────────────────────────────────

func TestRowHelperWithButtons(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	rowFn := mustAssertFunc(t, exports.Get("row"))
	buttonFn := mustAssertFunc(t, exports.Get("button"))

	b1, _ := buttonFn(goja.Undefined(), vm.ToValue("a"), vm.ToValue("A"), vm.ToValue("primary"))

	b2, _ := buttonFn(goja.Undefined(), vm.ToValue("b"), vm.ToValue("B"), vm.ToValue("secondary"))

	result, _ := rowFn(goja.Undefined(), b1, b2)
	row := result.Export().(discordgo.ActionsRow)
	require.Len(t, row.Components, 2)
}

func TestPagerHelper(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	pagerFn := mustAssertFunc(t, exports.Get("pager"))

	result, _ := pagerFn(goja.Undefined(), vm.ToValue("my-prev"), vm.ToValue("my-next"))
	row := result.Export().(discordgo.ActionsRow)
	require.Len(t, row.Components, 2)

	btn1 := row.Components[0].(discordgo.Button)
	require.Equal(t, "my-prev", btn1.CustomID)
	require.Equal(t, discordgo.SecondaryButton, btn1.Style)

	btn2 := row.Components[1].(discordgo.Button)
	require.Equal(t, "my-next", btn2.CustomID)
}

func TestActionsHelper(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	actionsFn := mustAssertFunc(t, exports.Get("actions"))

	arr := vm.NewArray()
	for i, def := range []map[string]string{
		{"id": "edit", "label": "Edit"},
		{"id": "delete", "label": "Delete", "style": "danger"},
	} {
		obj := vm.NewObject()
		for k, v := range def {
			_ = obj.Set(k, v)
		}
		_ = arr.Set(fmt.Sprintf("%d", i), obj)
	}

	result, _ := actionsFn(goja.Undefined(), arr)
	row := result.Export().(discordgo.ActionsRow)
	require.Len(t, row.Components, 2)

	btn1 := row.Components[0].(discordgo.Button)
	require.Equal(t, "edit", btn1.CustomID)
	require.Equal(t, discordgo.PrimaryButton, btn1.Style)

	btn2 := row.Components[1].(discordgo.Button)
	require.Equal(t, "delete", btn2.CustomID)
	require.Equal(t, discordgo.DangerButton, btn2.Style)
}

func TestConfirmHelper(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	confirmFn := mustAssertFunc(t, exports.Get("confirm"))

	result, _ := confirmFn(goja.Undefined(), vm.ToValue("Delete everything?"), vm.ToValue("do-delete"), vm.ToValue("do-cancel"))
	nr := result.Export().(*normalizedResponse)

	require.True(t, nr.Ephemeral)
	require.Len(t, nr.Embeds, 1)
	require.Equal(t, "Delete everything?", nr.Embeds[0].Description)
	require.Len(t, nr.Components, 1)

	row := nr.Components[0].(discordgo.ActionsRow)
	require.Len(t, row.Components, 2)
	btn1 := row.Components[0].(discordgo.Button)
	require.Equal(t, "do-delete", btn1.CustomID)
	require.Equal(t, discordgo.DangerButton, btn1.Style)
	btn2 := row.Components[1].(discordgo.Button)
	require.Equal(t, "do-cancel", btn2.CustomID)
	require.Equal(t, discordgo.SecondaryButton, btn2.Style)
}

func TestConfirmHelperDefaults(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	confirmFn := mustAssertFunc(t, exports.Get("confirm"))

	result, _ := confirmFn(goja.Undefined())
	nr := result.Export().(*normalizedResponse)
	require.Equal(t, "Are you sure?", nr.Embeds[0].Description)

	row := nr.Components[0].(discordgo.ActionsRow)
	btn1 := row.Components[0].(discordgo.Button)
	require.Equal(t, "confirm", btn1.CustomID)
	btn2 := row.Components[1].(discordgo.Button)
	require.Equal(t, "cancel", btn2.CustomID)
}

// ── Card builder tests ──────────────────────────────────────────────────────

func TestCardBuilderBasic(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	cardFn := mustAssertFunc(t, exports.Get("card"))

	builder, _ := cardFn(goja.Undefined(), vm.ToValue("My Card"))
	obj := builder.(*goja.Object)

	builder = mustCall(t, obj, "description", vm.ToValue("Card description"))
	builder = mustCall(t, builder.(*goja.Object), "color", vm.ToValue(0x5865F2))
	builder = mustCall(t, builder.(*goja.Object), "meta", vm.ToValue("Status"), vm.ToValue("Active"))
	builder = mustCall(t, builder.(*goja.Object), "meta", vm.ToValue("Priority"), vm.ToValue("High"))
	result := mustCall(t, builder.(*goja.Object), "build")

	emb := result.Export().(*discordgo.MessageEmbed)
	require.Equal(t, "My Card", emb.Title)
	require.Equal(t, "Card description", emb.Description)
	require.Equal(t, 0x5865F2, emb.Color)
	require.Len(t, emb.Fields, 2)
	require.Equal(t, "Status", emb.Fields[0].Name)
	require.Equal(t, "Active", emb.Fields[0].Value)
	require.True(t, emb.Fields[0].Inline)
	require.Equal(t, "Priority", emb.Fields[1].Name)
	require.True(t, emb.Fields[1].Inline)
}

func TestCardBuilderMetaDefaultsToNA(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	cardFn := mustAssertFunc(t, exports.Get("card"))

	builder, _ := cardFn(goja.Undefined(), vm.ToValue("Card"))
	builder = mustCall(t, builder.(*goja.Object), "meta", vm.ToValue("Empty"))
	result := mustCall(t, builder.(*goja.Object), "build")

	emb := result.Export().(*discordgo.MessageEmbed)
	require.Equal(t, "N/A", emb.Fields[0].Value)
}

func TestCardBuilderWrongParent(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	cardFn := mustAssertFunc(t, exports.Get("card"))

	builder, _ := cardFn(goja.Undefined(), vm.ToValue("Card"))
	obj := builder.(*goja.Object)

	_, err := tryCall(t, obj, "ephemeral")
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "ui.card")
}

// ── Flow helper tests ───────────────────────────────────────────────────────

func TestFlowHelperIdGeneration(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	flowFn := mustAssertFunc(t, exports.Get("flow"))

	builder, _ := flowFn(goja.Undefined(), vm.ToValue("search"))
	obj := builder.(*goja.Object)

	result := mustCall(t, obj, "id", vm.ToValue("select"))
	require.Equal(t, "search:select", result.String())

	result = mustCall(t, obj, "id", vm.ToValue("prev"))
	require.Equal(t, "search:prev", result.String())
}

func TestFlowHelperPagerIds(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	flowFn := mustAssertFunc(t, exports.Get("flow"))

	builder, _ := flowFn(goja.Undefined(), vm.ToValue("browse"))
	obj := builder.(*goja.Object)

	result := mustCall(t, obj, "pagerIds")
	resultObj := result.(*goja.Object)
	require.Equal(t, "browse:prev", resultObj.Get("prev").String())
	require.Equal(t, "browse:next", resultObj.Get("next").String())
}

func TestFlowHelperComponentIds(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	flowFn := mustAssertFunc(t, exports.Get("flow"))

	builder, _ := flowFn(goja.Undefined(), vm.ToValue("review"))
	obj := builder.(*goja.Object)

	arr := vm.NewArray()
	o := vm.NewObject()
	_ = o.Set("name", "approve")
	arr.Set("0", o)
	o2 := vm.NewObject()
	_ = o2.Set("name", "reject")
	arr.Set("1", o2)

	result := mustCall(t, obj, "componentIds", arr)
	resultObj := result.(*goja.Object)
	require.Equal(t, "review:approve", resultObj.Get("approve").String())
	require.Equal(t, "review:reject", resultObj.Get("reject").String())
}

func TestFlowHelperRequiresNamespace(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	flowFn := mustAssertFunc(t, exports.Get("flow"))

	_, err := flowFn(goja.Undefined(), vm.ToValue(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "namespace is required")
}

// ── End-to-end integration: full message with everything ────────────────────

func TestFullMessageWithEmbedButtonsAndSelect(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)

	// Build an embed
	embedFn := mustAssertFunc(t, exports.Get("embed"))
	emb, _ := embedFn(goja.Undefined(), vm.ToValue("Results"))
	emb = mustCall(t, emb.(*goja.Object), "description", vm.ToValue("Found 5 items"))
	emb = mustCall(t, emb.(*goja.Object), "color", vm.ToValue(0x5865F2))

	// Build buttons
	buttonFn := mustAssertFunc(t, exports.Get("button"))
	b1, _ := buttonFn(goja.Undefined(), vm.ToValue("refresh"), vm.ToValue("🔄 Refresh"), vm.ToValue("secondary"))

	b2, _ := buttonFn(goja.Undefined(), vm.ToValue("cancel"), vm.ToValue("Cancel"), vm.ToValue("danger"))

	// Build a select
	selectFn := mustAssertFunc(t, exports.Get("select"))
	sel, _ := selectFn(goja.Undefined(), vm.ToValue("page-select"))
	sel = mustCall(t, sel.(*goja.Object), "placeholder", vm.ToValue("Jump to page..."))
	sel = mustCall(t, sel.(*goja.Object), "option", vm.ToValue("Page 1"), vm.ToValue("1"))
	sel = mustCall(t, sel.(*goja.Object), "option", vm.ToValue("Page 2"), vm.ToValue("2"))

	// Build the message
	msgFn := mustAssertFunc(t, exports.Get("message"))
	msg, _ := msgFn(goja.Undefined())
	msg = mustCall(t, msg.(*goja.Object), "content", vm.ToValue("Search results:"))
	msg = mustCall(t, msg.(*goja.Object), "embed", emb)
	msg = mustCall(t, msg.(*goja.Object), "row", b1, b2)
	msg = mustCall(t, msg.(*goja.Object), "row", sel)
	msg = mustCall(t, msg.(*goja.Object), "ephemeral")
	result := mustCall(t, msg.(*goja.Object), "build")

	nr := result.Export().(*normalizedResponse)
	require.Equal(t, "Search results:", nr.Content)
	require.True(t, nr.Ephemeral)
	require.Len(t, nr.Embeds, 1)
	require.Len(t, nr.Components, 2)

	// First row: buttons
	row1 := nr.Components[0].(discordgo.ActionsRow)
	require.Len(t, row1.Components, 2)

	// Second row: select
	row2 := nr.Components[1].(discordgo.ActionsRow)
	menu := row2.Components[0].(discordgo.SelectMenu)
	require.Len(t, menu.Options, 2)

	// Pass through normalizePayload
	normalized, err := normalizePayload(nr)
	require.NoError(t, err)
	require.Equal(t, nr.Content, normalized.Content)
	require.Len(t, normalized.Components, 2)
}
