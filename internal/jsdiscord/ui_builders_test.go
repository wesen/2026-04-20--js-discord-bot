package jsdiscord

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

// ── Shared test helpers ─────────────────────────────────────────────────────

func loadUIForTest(t *testing.T, vm *goja.Runtime) *goja.Object {
	t.Helper()
	moduleObj := vm.NewObject()
	_ = moduleObj.Set("exports", vm.NewObject())
	uiLoader(vm, moduleObj)
	return moduleObj.Get("exports").(*goja.Object)
}

func mustAssertFunc(t *testing.T, val goja.Value) func(goja.Value, ...goja.Value) (goja.Value, error) {
	t.Helper()
	fn, ok := goja.AssertFunction(val)
	require.True(t, ok, "expected a function, got %v", val)
	return fn
}

func mustCall(t *testing.T, obj *goja.Object, method string, args ...goja.Value) goja.Value {
	t.Helper()
	fn := obj.Get(method)
	require.NotNil(t, fn, "method %s not found", method)
	callable, ok := goja.AssertFunction(fn)
	require.True(t, ok, "method %s is not callable", method)
	result, err := callable(obj, args...)
	require.NoError(t, err, "calling method %s failed", method)
	return result
}

// tryCall attempts to call a method and returns (result, error).
// It does not fail the test on error — the caller decides.
func tryCall(t *testing.T, obj *goja.Object, method string, args ...goja.Value) (result goja.Value, err error) { //nolint:nonamedreturns
	t.Helper()
	// Proxy Get trap panics propagate — catch them
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	fn := obj.Get(method)
	if fn == nil {
		return nil, fmt.Errorf("method %s not found", method)
	}
	callable, ok := goja.AssertFunction(fn)
	if !ok {
		return nil, fmt.Errorf("method %s is not callable", method)
	}
	result, err = callable(obj, args...)
	return
}

func panicMessage(v any) string {
	s := fmt.Sprint(v)
	s = strings.TrimPrefix(s, "Error: ")
	s = strings.TrimPrefix(s, "RuntimeError: ")
	return strings.TrimSpace(s)
}

// ── Embed builder tests ─────────────────────────────────────────────────────

func TestEmbedBuilderBasicChain(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	builder, _ := embedFn(goja.Undefined(), vm.ToValue("Test Embed"))
	obj := builder.(*goja.Object)

	builder = mustCall(t, obj, "description", vm.ToValue("A description"))
	builder = mustCall(t, builder.(*goja.Object), "color", vm.ToValue(0xFF0000))
	result := mustCall(t, builder.(*goja.Object), "build")

	emb, ok := result.Export().(*discordgo.MessageEmbed)
	require.True(t, ok)
	require.Equal(t, "Test Embed", emb.Title)
	require.Equal(t, "A description", emb.Description)
	require.Equal(t, 0xFF0000, emb.Color)
}

func TestEmbedBuilderFields(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	builder, _ := embedFn(goja.Undefined(), vm.ToValue("Card"))
	obj := builder.(*goja.Object)

	builder = mustCall(t, obj, "field", vm.ToValue("Name"), vm.ToValue("Value"), vm.ToValue(true))
	builder = mustCall(t, builder.(*goja.Object), "field", vm.ToValue("Status"), vm.ToValue("Active"), vm.ToValue(false))
	result := mustCall(t, builder.(*goja.Object), "build")

	emb := result.Export().(*discordgo.MessageEmbed)
	require.Len(t, emb.Fields, 2)
	require.Equal(t, "Name", emb.Fields[0].Name)
	require.True(t, emb.Fields[0].Inline)
	require.Equal(t, "Status", emb.Fields[1].Name)
	require.False(t, emb.Fields[1].Inline)
}

func TestEmbedBuilderFooterAndAuthor(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	builder, _ := embedFn(goja.Undefined(), vm.ToValue("Title"))
	obj := builder.(*goja.Object)

	builder = mustCall(t, obj, "footer", vm.ToValue("Page 1/3"))
	builder = mustCall(t, builder.(*goja.Object), "author", vm.ToValue("Alice"))
	result := mustCall(t, builder.(*goja.Object), "build")

	emb := result.Export().(*discordgo.MessageEmbed)
	require.Equal(t, "Page 1/3", emb.Footer.Text)
	require.Equal(t, "Alice", emb.Author.Name)
}

func TestEmbedBuilderTimestamp(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	builder, _ := embedFn(goja.Undefined(), vm.ToValue("Log"))
	builder = mustCall(t, builder.(*goja.Object), "timestamp")
	result := mustCall(t, builder.(*goja.Object), "build")

	emb := result.Export().(*discordgo.MessageEmbed)
	require.NotEmpty(t, emb.Timestamp)
}

func TestEmbedBuilderMax25Fields(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	builder, _ := embedFn(goja.Undefined(), vm.ToValue("Big"))
	obj := builder.(*goja.Object)
	for i := 0; i < 25; i++ {
		builder = mustCall(t, obj, "field", vm.ToValue(fmt.Sprintf("F%d", i)), vm.ToValue("V"))
		obj = builder.(*goja.Object)
	}

	// 26th field should return an error
	_, err := tryCall(t, obj, "field", vm.ToValue("Extra"), vm.ToValue("V"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "maximum 25 fields")
}

func TestEmbedBuilderWrongParentError(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	builder, _ := embedFn(goja.Undefined(), vm.ToValue("Test"))
	obj := builder.(*goja.Object)

	// .ephemeral() belongs to ui.message, not ui.embed
	require.Panics(t, func() {
		obj.Get("ephemeral")
	})
}

func TestEmbedBuilderUnknownMethod(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	builder, _ := embedFn(goja.Undefined(), vm.ToValue("Test"))
	obj := builder.(*goja.Object)

	require.Panics(t, func() {
		obj.Get("totallyFakeMethod")
	})
}

// ── Message builder tests ───────────────────────────────────────────────────

func TestMessageBuilderBasicChain(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))

	builder, _ := msgFn(goja.Undefined())
	obj := builder.(*goja.Object)

	builder = mustCall(t, obj, "content", vm.ToValue("Hello world"))
	builder = mustCall(t, builder.(*goja.Object), "ephemeral")
	result := mustCall(t, builder.(*goja.Object), "build")

	nr := result.Export().(*normalizedResponse)
	require.Equal(t, "Hello world", nr.Content)
	require.True(t, nr.Ephemeral)
}

func TestMessageBuilderWithEmbed(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	// Build an embed (don't call .build() — message.embed accepts builder proxy)
	embedBuilder, _ := embedFn(goja.Undefined(), vm.ToValue("Card"))
	embedBuilder = mustCall(t, embedBuilder.(*goja.Object), "description", vm.ToValue("A card"))
	embedBuilder = mustCall(t, embedBuilder.(*goja.Object), "color", vm.ToValue(0x57F287))

	// Build message with embed
	msgBuilder, _ := msgFn(goja.Undefined())
	msgBuilder = mustCall(t, msgBuilder.(*goja.Object), "content", vm.ToValue("Found 3 results."))
	msgBuilder = mustCall(t, msgBuilder.(*goja.Object), "embed", embedBuilder)
	msgBuilder = mustCall(t, msgBuilder.(*goja.Object), "ephemeral")
	result := mustCall(t, msgBuilder.(*goja.Object), "build")

	nr := result.Export().(*normalizedResponse)
	require.Equal(t, "Found 3 results.", nr.Content)
	require.True(t, nr.Ephemeral)
	require.Len(t, nr.Embeds, 1)
	require.Equal(t, "Card", nr.Embeds[0].Title)
	require.Equal(t, "A card", nr.Embeds[0].Description)
	require.Equal(t, 0x57F287, nr.Embeds[0].Color)
}

func TestMessageBuilderPassesThroughNormalizePayload(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))

	builder, _ := msgFn(goja.Undefined())
	builder = mustCall(t, builder.(*goja.Object), "content", vm.ToValue("test"))
	builder = mustCall(t, builder.(*goja.Object), "ephemeral")
	result := mustCall(t, builder.(*goja.Object), "build")

	nr := result.Export().(*normalizedResponse)

	// Should pass through the *normalizedResponse fast path
	normalized, err := normalizePayload(nr)
	require.NoError(t, err)
	require.Equal(t, "test", normalized.Content)
	require.True(t, normalized.Ephemeral)
}

func TestMessageBuilderRejectsRawObjectEmbed(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))

	builder, _ := msgFn(goja.Undefined())
	obj := builder.(*goja.Object)

	rawObj := vm.NewObject()
	_ = rawObj.Set("title", "Raw")

	// Raw JS object should return an error
	_, err := tryCall(t, obj, "embed", rawObj)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected")
}

func TestMessageBuilderRejectsRawObjectRow(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))

	builder, _ := msgFn(goja.Undefined())
	obj := builder.(*goja.Object)

	rawObj := vm.NewObject()
	_ = rawObj.Set("type", "button")

	// Raw JS object should return an error
	_, err := tryCall(t, obj, "row", rawObj)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected")
}

func TestMessageBuilderWrongParentError(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))

	builder, _ := msgFn(goja.Undefined())
	obj := builder.(*goja.Object)

	// .field() belongs to ui.embed
	require.Panics(t, func() {
		obj.Get("field")
	})
	// .description() belongs to ui.embed
	require.Panics(t, func() {
		obj.Get("description")
	})
}

func TestMessageBuilderUnknownMethod(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))

	builder, _ := msgFn(goja.Undefined())
	obj := builder.(*goja.Object)

	require.Panics(t, func() {
		obj.Get("nonexistent")
	})
}

// ── Error message content tests ─────────────────────────────────────────────

func TestWrongParentErrorMessageMentionsCorrectBuilder(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	builder, _ := embedFn(goja.Undefined(), vm.ToValue("X"))
	obj := builder.(*goja.Object)

	defer func() {
		r := recover()
		require.NotNil(t, r)
		msg := strings.ToLower(panicMessage(r))
		require.Contains(t, msg, "ui.embed")
		require.Contains(t, msg, "ephemeral")
		require.Contains(t, msg, "ui.message")
	}()

	obj.Get("ephemeral")
}

func TestMultipleEmbedsOnMessage(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	msgFn := mustAssertFunc(t, exports.Get("message"))
	embedFn := mustAssertFunc(t, exports.Get("embed"))

	e1, _ := embedFn(goja.Undefined(), vm.ToValue("First"))
	e1 = mustCall(t, e1.(*goja.Object), "color", vm.ToValue(0xFF0000))

	e2, _ := embedFn(goja.Undefined(), vm.ToValue("Second"))
	e2 = mustCall(t, e2.(*goja.Object), "color", vm.ToValue(0x00FF00))

	builder, _ := msgFn(goja.Undefined())
	builder = mustCall(t, builder.(*goja.Object), "embed", e1)
	builder = mustCall(t, builder.(*goja.Object), "embed", e2)
	result := mustCall(t, builder.(*goja.Object), "build")

	nr := result.Export().(*normalizedResponse)
	require.Len(t, nr.Embeds, 2)
	require.Equal(t, "First", nr.Embeds[0].Title)
	require.Equal(t, "Second", nr.Embeds[1].Title)
}
