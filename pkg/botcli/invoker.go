package botcli

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
)

func makeBotVerbInvoker(cfg commandOptions) func(context.Context, *jsverbs.Registry, *jsverbs.VerbSpec, *values.Values) (interface{}, error) {
	factory := cfg.runtimeFactory
	if factory == nil {
		factory = defaultRuntimeFactory(cfg)
	}
	return func(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec, parsedValues *values.Values) (interface{}, error) {
		runtime, err := factory.NewRuntimeForVerb(ctx, registry, verb)
		if err != nil {
			return nil, err
		}
		if runtime == nil {
			return nil, fmt.Errorf("runtime factory returned nil runtime")
		}
		defer func() {
			_ = runtime.Close(context.Background())
		}()
		return registry.InvokeInRuntime(ctx, runtime, verb, parsedValues)
	}
}
