package jsdiscord

import (
	"fmt"

	"github.com/go-go-golems/go-go-goja/engine"
)

// HostOption customizes how a JavaScript bot host is created.
type HostOption func(*hostOptions) error

type hostOptions struct {
	runtimeModuleRegistrars []engine.RuntimeModuleRegistrar
}

func applyHostOptions(opts ...HostOption) (hostOptions, error) {
	cfg := hostOptions{}
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return hostOptions{}, fmt.Errorf("apply host option %d: %w", i, err)
		}
	}
	return cfg, nil
}

// WithRuntimeModuleRegistrars appends per-runtime native module registrars.
func WithRuntimeModuleRegistrars(registrars ...engine.RuntimeModuleRegistrar) HostOption {
	return func(cfg *hostOptions) error {
		for i, registrar := range registrars {
			if registrar == nil {
				return fmt.Errorf("runtime module registrar at index %d is nil", i)
			}
		}
		cfg.runtimeModuleRegistrars = append(cfg.runtimeModuleRegistrars, registrars...)
		return nil
	}
}
