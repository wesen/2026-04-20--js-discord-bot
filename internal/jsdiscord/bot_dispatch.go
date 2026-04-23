package jsdiscord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/runtimebridge"
	"github.com/go-go-golems/go-go-goja/pkg/runtimeowner"
	"github.com/rs/zerolog/log"
)

func (h *BotHandle) DispatchCommand(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchCommand, request)
}

func (h *BotHandle) DispatchSubcommand(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchSubcommand, request)
}

func (h *BotHandle) DispatchEvent(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchEvent, request)
}

func (h *BotHandle) DispatchComponent(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchComponent, request)
}

func (h *BotHandle) DispatchModal(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchModal, request)
}

func (h *BotHandle) DispatchAutocomplete(ctx context.Context, request DispatchRequest) (any, error) {
	return h.dispatch(ctx, h.dispatchAutocomplete, request)
}

func (h *BotHandle) dispatch(ctx context.Context, fn goja.Callable, request DispatchRequest) (any, error) {
	if h == nil {
		return nil, fmt.Errorf("discord bot handle is nil")
	}
	bindings, ok := runtimebridge.Lookup(h.vm)
	if !ok || bindings.Owner == nil {
		return nil, fmt.Errorf("discord bot requires runtime owner bindings")
	}
	ret, err := bindings.Owner.Call(ctx, "discord.bot.dispatch", func(callCtx context.Context, vm *goja.Runtime) (any, error) {
		input := buildDispatchInput(vm, callCtx, request)
		result, err := fn(goja.Undefined(), input)
		if err != nil {
			return nil, err
		}
		if goja.IsUndefined(result) || goja.IsNull(result) {
			return nil, nil
		}
		return result.Export(), nil
	})
	if err != nil {
		return nil, err
	}
	return settleValue(ctx, bindings.Owner, ret)
}



// DispatchCommandAsMap dispatches a command and normalizes the result to map[string]any.
// This is used by tests that expect the old map[string]any format.
func (h *BotHandle) DispatchCommandAsMap(ctx context.Context, request DispatchRequest) (map[string]any, error) {
	result, err := h.DispatchCommand(ctx, request)
	if err != nil {
		return nil, err
	}
	return normalizeResultToMap(result)
}

// DispatchComponentAsMap dispatches a component interaction and normalizes the result.
func (h *BotHandle) DispatchComponentAsMap(ctx context.Context, request DispatchRequest) (map[string]any, error) {
	result, err := h.DispatchComponent(ctx, request)
	if err != nil {
		return nil, err
	}
	return normalizeResultToMap(result)
}

// DispatchModalAsMap dispatches a modal submission and normalizes the result.
func (h *BotHandle) DispatchModalAsMap(ctx context.Context, request DispatchRequest) (map[string]any, error) {
	result, err := h.DispatchModal(ctx, request)
	if err != nil {
		return nil, err
	}
	return normalizeResultToMap(result)
}

func normalizeResultToMap(result any) (map[string]any, error) {
	if result == nil {
		return nil, nil
	}
	switch v := result.(type) {
	case *normalizedResponse:
		return v.toMap(), nil
	case map[string]any:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected result type %T", result)
	}
}

func settleValue(ctx context.Context, owner runtimeowner.Runner, value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	log.Info().Str("component", "dispatch").Str("type", fmt.Sprintf("%T", value)).Msg("settleValue")
	switch v := value.(type) {
	case *goja.Promise:
		return waitForPromise(ctx, owner, v)
	case goja.Value:
		return settleValue(ctx, owner, v.Export())
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			settled, err := settleValue(ctx, owner, item)
			if err != nil {
				return nil, err
			}
			out[i] = settled
		}
		return out, nil
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			settled, err := settleValue(ctx, owner, item)
			if err != nil {
				return nil, err
			}
			out[key] = settled
		}
		return out, nil
	case *normalizedResponse:
		return v, nil
	default:
		return value, nil
	}
}

type promiseSnapshot struct {
	State  goja.PromiseState
	Result any
	Text   string
}

func waitForPromise(ctx context.Context, owner runtimeowner.Runner, promise *goja.Promise) (any, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		ret, err := owner.Call(ctx, "discord.bot.promise-state", func(_ context.Context, vm *goja.Runtime) (any, error) {
			result := promise.Result()
			return promiseSnapshot{
				State:  promise.State(),
				Result: exportSettledValue(result),
				Text:   describeSettledValue(vm, result),
			}, nil
		})
		if err != nil {
			return nil, err
		}
		snapshot, ok := ret.(promiseSnapshot)
		if !ok {
			return nil, fmt.Errorf("discord bot promise snapshot has unexpected type %T", ret)
		}
		switch snapshot.State {
		case goja.PromiseStatePending:
			time.Sleep(5 * time.Millisecond)
		case goja.PromiseStateRejected:
			message := strings.TrimSpace(snapshot.Text)
			if message == "" {
				message = fmt.Sprint(snapshot.Result)
			}
			return nil, fmt.Errorf("promise rejected: %s", message)
		case goja.PromiseStateFulfilled:
			return settleValue(ctx, owner, snapshot.Result)
		}
	}
}

func exportSettledValue(value goja.Value) any {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	return value.Export()
}

func describeSettledValue(vm *goja.Runtime, value goja.Value) string {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return ""
	}
	if obj := value.ToObject(vm); obj != nil {
		if stack := strings.TrimSpace(safeValueString(vm, obj.Get("stack"))); stack != "" {
			return stack
		}
	}
	if text := strings.TrimSpace(safeValueString(vm, value)); text != "" && text != "[object Object]" {
		return text
	}
	return strings.TrimSpace(fmt.Sprint(value.Export()))
}

func safeValueString(vm *goja.Runtime, value goja.Value) string {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return ""
	}
	if ex, ok := value.Export().(error); ok {
		return ex.Error()
	}
	if obj := value.ToObject(vm); obj != nil {
		if fn, ok := goja.AssertFunction(obj.Get("toString")); ok {
			if ret, err := fn(value); err == nil && !goja.IsUndefined(ret) && !goja.IsNull(ret) {
				return ret.String()
			}
		}
	}
	return value.String()
}
