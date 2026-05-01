package jsdiscord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/stretchr/testify/require"
)

func TestOpenRouterModuleCompleteJson(t *testing.T) {
	var captured struct {
		Authorization string
		Title         string
		Model         string
		MaxTokens     int
		Temperature   float64
		Messages      []openRouterChatMessage
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		captured.Authorization = r.Header.Get("Authorization")
		captured.Title = r.Header.Get("X-Title")
		var requestBody openRouterChatRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))
		captured.Model = requestBody.Model
		captured.MaxTokens = requestBody.MaxTokens
		captured.Temperature = requestBody.Temperature
		captured.Messages = requestBody.Messages
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"test",
			"model":"test-model",
			"choices":[{"message":{"role":"assistant","content":"{\"scene_patch\":{}}"}}],
			"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}
		}`))
	}))
	defer server.Close()

	t.Setenv("OPENROUTER_API_KEY", "test-key")
	t.Setenv("OPENROUTER_BASE_URL", server.URL)
	t.Setenv("OPENROUTER_MODEL", "go-owned-model")
	t.Setenv("OPENROUTER_MAX_TOKENS", "42")
	t.Setenv("OPENROUTER_TEMPERATURE", "0.25")
	t.Setenv("OPENROUTER_APP_TITLE", "unit-test-title")

	factory, err := engine.NewBuilder().WithRuntimeModuleRegistrars(&OpenRouterRegistrar{}).Build()
	require.NoError(t, err)
	rt, err := factory.NewRuntime(t.Context())
	require.NoError(t, err)
	defer func() { _ = rt.Close(t.Context()) }()

	value, err := rt.Require.Require("adventure_llm")
	require.NoError(t, err)
	module := value.ToObject(rt.VM)
	complete, ok := goja.AssertFunction(module.Get("completeJson"))
	require.True(t, ok)
	resultValue, err := complete(goja.Undefined(), rt.VM.ToValue(map[string]any{
		"purpose": "scene_patch",
		"system":  "Return JSON only.",
		"user":    "Create a room.",
		"model":   "js-must-not-control-this",
	}))
	require.NoError(t, err)
	result := resultValue.Export().(map[string]any)
	require.Equal(t, true, result["ok"])
	require.Equal(t, "openrouter", result["provider"])
	require.Equal(t, "{\"scene_patch\":{}}", result["text"])
	require.Equal(t, "Bearer test-key", captured.Authorization)
	require.Equal(t, "unit-test-title", captured.Title)
	require.Equal(t, "go-owned-model", captured.Model)
	require.Equal(t, 42, captured.MaxTokens)
	require.Equal(t, 0.25, captured.Temperature)
	require.Len(t, captured.Messages, 2)
	require.Equal(t, "system", captured.Messages[0].Role)
	require.Equal(t, "user", captured.Messages[1].Role)
}

func TestOpenRouterModuleRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	factory, err := engine.NewBuilder().WithRuntimeModuleRegistrars(&OpenRouterRegistrar{}).Build()
	require.NoError(t, err)
	rt, err := factory.NewRuntime(t.Context())
	require.NoError(t, err)
	defer func() { _ = rt.Close(t.Context()) }()

	value, err := rt.Require.Require("adventure_llm")
	require.NoError(t, err)
	module := value.ToObject(rt.VM)
	complete, ok := goja.AssertFunction(module.Get("completeJson"))
	require.True(t, ok)
	resultValue, err := complete(goja.Undefined(), rt.VM.ToValue(map[string]any{"user": "hello"}))
	require.NoError(t, err)
	result := resultValue.Export().(map[string]any)
	require.Equal(t, false, result["ok"])
	require.Contains(t, result["error"], "OPENROUTER_API_KEY")
}
