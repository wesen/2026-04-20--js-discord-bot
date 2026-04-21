package jsdiscord

import "github.com/dop251/goja"

func storeObject(vm *goja.Runtime, store *MemoryStore) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("get", func(key string, defaultValue any) any {
		if store == nil {
			return defaultValue
		}
		return store.Get(key, defaultValue)
	})
	_ = obj.Set("set", func(key string, value any) {
		if store != nil {
			store.Set(key, value)
		}
	})
	_ = obj.Set("delete", func(key string) bool {
		if store == nil {
			return false
		}
		return store.Delete(key)
	})
	_ = obj.Set("keys", func(prefix string) []string {
		if store == nil {
			return nil
		}
		return store.Keys(prefix)
	})
	_ = obj.Set("namespace", func(parts ...string) any {
		if store == nil {
			return storeObject(vm, NewMemoryStore().Namespace(parts...))
		}
		return storeObject(vm, store.Namespace(parts...))
	})
	return obj
}
