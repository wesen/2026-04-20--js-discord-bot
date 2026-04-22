package jsdiscord

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestUIModuleLoadsAndExports(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)

	expected := []string{
		"message", "embed", "card",
		"button", "select", "userSelect", "roleSelect", "channelSelect", "mentionableSelect",
		"form", "flow",
		"row", "pager", "actions", "confirm",
		"ok", "error", "emptyResults",
	}
	for _, name := range expected {
		val := exports.Get(name)
		if val == nil || goja.IsUndefined(val) {
			t.Errorf("ui module missing export: %s", name)
		}
	}
}

func TestUIModuleOkHelper(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	okFn := mustAssertFunc(t, exports.Get("ok"))

	result, err := okFn(goja.Undefined(), vm.ToValue("Done!"))
	require.NoError(t, err)

	nr, ok := result.Export().(*normalizedResponse)
	require.True(t, ok)
	require.Equal(t, "Done!", nr.Content)
	require.True(t, nr.Ephemeral)
}

func TestUIModuleErrorHelper(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	errFn := mustAssertFunc(t, exports.Get("error"))

	result, err := errFn(goja.Undefined(), vm.ToValue("Something broke"))
	require.NoError(t, err)

	nr, ok := result.Export().(*normalizedResponse)
	require.True(t, ok)
	require.Equal(t, "⚠️ Something broke", nr.Content)
	require.True(t, nr.Ephemeral)
}

func TestUIModuleEmptyResultsHelper(t *testing.T) {
	vm := goja.New()
	exports := loadUIForTest(t, vm)
	emptyFn := mustAssertFunc(t, exports.Get("emptyResults"))

	// With query
	result, err := emptyFn(goja.Undefined(), vm.ToValue("test query"))
	require.NoError(t, err)
	nr := result.Export().(*normalizedResponse)
	require.Contains(t, nr.Content, "test query")
	require.True(t, nr.Ephemeral)

	// Without query
	result2, err := emptyFn(goja.Undefined(), vm.ToValue(""))
	require.NoError(t, err)
	nr2 := result2.Export().(*normalizedResponse)
	require.Equal(t, "No results found.", nr2.Content)
}
