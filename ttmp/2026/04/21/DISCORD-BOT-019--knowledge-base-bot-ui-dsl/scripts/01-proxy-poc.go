//go:build ignore
//
// 01-proxy-poc.go — Proof of concept: Goja ES6 Proxy with native Go trap handlers.
//
// Tests whether Goja's ProxyTrapConfig.Get can intercept property access on a builder
// object and dynamically return methods, enabling a Go-side DSL where JS never sees
// raw data properties — only the chain methods the Go side chooses to expose.
//
// Run:
//   go run ttmp/2026/04/21/DISCORD-BOT-019--knowledge-base-bot-ui-dsl/scripts/01-proxy-poc.go
//   (or: go run scripts/01-proxy-poc.go from the ticket dir)

package main

import (
	"fmt"
	"github.com/dop251/goja"
)

func main() {
	vm := goja.New()

	// Create a target object — this is the "real" data the Go side owns
	target := vm.NewObject()
	_ = target.Set("type", "select")
	_ = target.Set("customId", "test:sel")

	// Create a Proxy with a Get trap that intercepts ALL property access.
	// JS sees the proxy as a builder. The Go side decides what properties
	// are readable and what method calls are valid.
	proxy := vm.NewProxy(target, &goja.ProxyTrapConfig{
		Get: func(target *goja.Object, property string, receiver goja.Value) goja.Value {
			// If the target already has a real data property, return it
			if v := target.Get(property); v != nil && v != goja.Undefined() {
				return v
			}

			// Otherwise, return a chain method dynamically
			switch property {
			case "placeholder":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					text := call.Argument(0).String()
					_ = target.Set("placeholder", text)
					return receiver // return proxy for chaining
				})
			case "option":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					label := call.Argument(0).String()
					value := call.Argument(1).String()
					// Read existing options
					var opts []map[string]string
					if v := target.Get("options"); v != nil && v != goja.Undefined() {
						if exp, ok := v.Export().([]any); ok {
							for _, e := range exp {
								if m, ok := e.(map[string]string); ok {
									opts = append(opts, m)
								}
							}
						}
					}
					opts = append(opts, map[string]string{"label": label, "value": value})
					_ = target.Set("options", opts)
					return receiver
				})
			case "build":
				return vm.ToValue(func(call goja.FunctionCall) goja.Value {
					return target // unwrap proxy, return raw data
				})
			default:
				// Unknown property → return undefined (JS gets undefined, not an error)
				return goja.Undefined()
			}
		},
	})

	_ = vm.Set("sel", proxy)

	// Test: fluent chain from JS
	val, err := vm.RunString(`
		sel.placeholder("Pick one")
		sel.option("A", "a")
		sel.option("B", "b")
		sel.option("C", "c")
		built = sel.build()
		JSON.stringify(built)
	`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Result:", val.String())

	// Test: unknown method returns undefined, not an error
	val2, err := vm.RunString(`typeof sel.unknownMethod`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Unknown method type:", val2.String())

	// Test: data properties are readable through the proxy
	val3, err := vm.RunString(`sel.type + " / " + sel.customId`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Data props:", val3.String())

	fmt.Println("\n✓ Goja Proxy with Go-native Get trap works for builder DSL")
}
