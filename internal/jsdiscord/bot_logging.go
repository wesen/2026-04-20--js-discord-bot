package jsdiscord

import (
	"sort"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func loggerObject(vm *goja.Runtime, kind, name string, metadata map[string]any) *goja.Object {
	obj := vm.NewObject()
	baseFields := map[string]any{"jsKind": kind, "jsName": name}
	for key, value := range metadata {
		baseFields["meta."+key] = value
	}
	setLogMethod := func(level string, fn func(string, map[string]any)) { _ = obj.Set(level, fn) }
	setLogMethod("info", func(msg string, fields map[string]any) {
		e := log.Info()
		applyFields(e, baseFields)
		applyFields(e, fields)
		e.Msg(msg)
	})
	setLogMethod("debug", func(msg string, fields map[string]any) {
		e := log.Debug()
		applyFields(e, baseFields)
		applyFields(e, fields)
		e.Msg(msg)
	})
	setLogMethod("warn", func(msg string, fields map[string]any) {
		e := log.Warn()
		applyFields(e, baseFields)
		applyFields(e, fields)
		e.Msg(msg)
	})
	setLogMethod("error", func(msg string, fields map[string]any) {
		e := log.Error()
		applyFields(e, baseFields)
		applyFields(e, fields)
		e.Msg(msg)
	})
	return obj
}

func applyFields(event *zerolog.Event, fields map[string]any) {
	if event == nil || len(fields) == 0 {
		return
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		event.Interface(key, fields[key])
	}
}
