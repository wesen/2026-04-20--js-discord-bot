package jsdiscord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
)

const (
	defaultOpenRouterBaseURL     = "https://openrouter.ai/api/v1"
	defaultOpenRouterModel       = "anthropic/claude-3.5-haiku"
	defaultOpenRouterMaxTokens   = 1200
	defaultOpenRouterTemperature = 0.7
)

// OpenRouterRegistrar exposes a narrow Go-owned LLM function to JavaScript bots.
// JavaScript supplies prompt content only; provider settings such as API key,
// model, token limits, base URL, and headers are owned by Go/process config.
type OpenRouterRegistrar struct{}

func (r *OpenRouterRegistrar) ID() string { return "discord-openrouter-registrar" }

func (r *OpenRouterRegistrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
	if reg == nil {
		return fmt.Errorf("require registry is nil")
	}
	reg.RegisterNativeModule("adventure_llm", openRouterLoader)
	return nil
}

func openRouterLoader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)
	client := newOpenRouterClientFromEnv()
	_ = exports.Set("completeJson", func(call goja.FunctionCall) goja.Value {
		input, err := parseOpenRouterInput(call)
		if err != nil {
			return vm.ToValue(openRouterErrorResult(err.Error(), false))
		}
		result := client.complete(context.Background(), input)
		return vm.ToValue(result)
	})
}

type openRouterClient struct {
	HTTPClient  *http.Client
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
	Referer     string
	Title       string
}

type openRouterInput struct {
	Purpose  string         `json:"purpose,omitempty"`
	System   string         `json:"system,omitempty"`
	User     string         `json:"user,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type openRouterChatRequest struct {
	Model       string                  `json:"model"`
	Messages    []openRouterChatMessage `json:"messages"`
	Temperature float64                 `json:"temperature"`
	MaxTokens   int                     `json:"max_tokens"`
}

type openRouterChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message openRouterChatMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func newOpenRouterClientFromEnv() *openRouterClient {
	return &openRouterClient{
		HTTPClient:  &http.Client{Timeout: 60 * time.Second},
		APIKey:      strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")),
		BaseURL:     envString("OPENROUTER_BASE_URL", defaultOpenRouterBaseURL),
		Model:       envString("OPENROUTER_MODEL", defaultOpenRouterModel),
		MaxTokens:   envInt("OPENROUTER_MAX_TOKENS", defaultOpenRouterMaxTokens),
		Temperature: envFloat("OPENROUTER_TEMPERATURE", defaultOpenRouterTemperature),
		Referer:     strings.TrimSpace(os.Getenv("OPENROUTER_HTTP_REFERER")),
		Title:       envString("OPENROUTER_APP_TITLE", "discord-bot-adventure"),
	}
}

func parseOpenRouterInput(call goja.FunctionCall) (openRouterInput, error) {
	if len(call.Arguments) != 1 || goja.IsUndefined(call.Arguments[0]) || goja.IsNull(call.Arguments[0]) {
		return openRouterInput{}, fmt.Errorf("adventure_llm.completeJson expects one input object")
	}
	exported := call.Arguments[0].Export()
	mapping, ok := exported.(map[string]any)
	if !ok {
		return openRouterInput{}, fmt.Errorf("adventure_llm.completeJson input must be an object")
	}
	input := openRouterInput{
		Purpose: strings.TrimSpace(fmt.Sprint(mapping["purpose"])),
		System:  strings.TrimSpace(fmt.Sprint(mapping["system"])),
		User:    strings.TrimSpace(fmt.Sprint(mapping["user"])),
	}
	if metadata, ok := mapping["metadata"].(map[string]any); ok {
		input.Metadata = metadata
	}
	if input.System == "" && input.User == "" {
		return openRouterInput{}, fmt.Errorf("adventure_llm.completeJson requires system or user prompt text")
	}
	return input, nil
}

func (c *openRouterClient) complete(ctx context.Context, input openRouterInput) map[string]any {
	if c == nil {
		return openRouterErrorResult("OpenRouter client is not configured", false)
	}
	if strings.TrimSpace(c.APIKey) == "" {
		return openRouterErrorResult("OPENROUTER_API_KEY is not configured", false)
	}
	messages := []openRouterChatMessage{}
	if input.System != "" {
		messages = append(messages, openRouterChatMessage{Role: "system", Content: input.System})
	}
	if input.User != "" {
		messages = append(messages, openRouterChatMessage{Role: "user", Content: input.User})
	}
	requestPayload := openRouterChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: c.Temperature,
		MaxTokens:   c.MaxTokens,
	}
	body, err := json.Marshal(requestPayload)
	if err != nil {
		return openRouterErrorResult("failed to encode LLM request", false)
	}
	endpoint := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return openRouterErrorResult("failed to build LLM request", false)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	if c.Referer != "" {
		req.Header.Set("HTTP-Referer", c.Referer)
	}
	if c.Title != "" {
		req.Header.Set("X-Title", c.Title)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return openRouterErrorResult("LLM request failed", true)
	}
	defer func() { _ = resp.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return openRouterErrorResult("failed to read LLM response", true)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return map[string]any{
			"ok":         false,
			"error":      "LLM request failed",
			"retryable":  resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500,
			"statusCode": resp.StatusCode,
		}
	}
	var decoded openRouterChatResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return openRouterErrorResult("failed to decode LLM response", true)
	}
	text := ""
	if len(decoded.Choices) > 0 {
		text = decoded.Choices[0].Message.Content
	}
	if strings.TrimSpace(text) == "" {
		return openRouterErrorResult("LLM response did not include message content", true)
	}
	return map[string]any{
		"ok":       true,
		"text":     text,
		"provider": "openrouter",
		"usage": map[string]any{
			"promptTokens":     decoded.Usage.PromptTokens,
			"completionTokens": decoded.Usage.CompletionTokens,
			"totalTokens":      decoded.Usage.TotalTokens,
		},
	}
}

func openRouterErrorResult(message string, retryable bool) map[string]any {
	return map[string]any{"ok": false, "error": message, "retryable": retryable}
}

func envString(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envFloat(name string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
