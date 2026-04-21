package jsdiscord

import (
	"context"

	"github.com/dop251/goja"
)

// op1 wraps a 1-argument Discord operation with nil-guard and zero-value fallback.
func op1[T any](fn func(context.Context, string) (T, error), ctx context.Context, zero T) func(string) (any, error) {
	if fn == nil {
		return func(string) (any, error) { return zero, nil }
	}
	return func(a string) (any, error) { return fn(ctx, a) }
}

// op1E wraps a 1-argument Discord operation that returns only an error.
func op1E(fn func(context.Context, string) error, ctx context.Context) func(string) error {
	if fn == nil {
		return func(string) error { return nil }
	}
	return func(a string) error { return fn(ctx, a) }
}

// op2 wraps a 2-argument Discord operation with nil-guard and zero-value fallback.
func op2[T any](fn func(context.Context, string, string) (T, error), ctx context.Context, zero T) func(string, string) (any, error) {
	if fn == nil {
		return func(string, string) (any, error) { return zero, nil }
	}
	return func(a, b string) (any, error) { return fn(ctx, a, b) }
}

// op2E wraps a 2-argument Discord operation that returns only an error.
func op2E(fn func(context.Context, string, string) error, ctx context.Context) func(string, string) error {
	if fn == nil {
		return func(string, string) error { return nil }
	}
	return func(a, b string) error { return fn(ctx, a, b) }
}

// op3E wraps a 3-argument Discord operation that returns only an error.
func op3E(fn func(context.Context, string, string, string) error, ctx context.Context) func(string, string, string) error {
	if fn == nil {
		return func(string, string, string) error { return nil }
	}
	return func(a, b, c string) error { return fn(ctx, a, b, c) }
}

// op1A wraps a 1-string+1-any Discord operation with nil-guard.
func op1A[T any](fn func(context.Context, string, any) (T, error), ctx context.Context, zero T) func(string, any) (any, error) {
	if fn == nil {
		return func(string, any) (any, error) { return zero, nil }
	}
	return func(a string, b any) (any, error) { return fn(ctx, a, b) }
}

// op1AErr wraps a 1-string+1-any Discord operation that returns only an error.
func op1AErr(fn func(context.Context, string, any) error, ctx context.Context) func(string, any) error {
	if fn == nil {
		return func(string, any) error { return nil }
	}
	return func(a string, b any) error { return fn(ctx, a, b) }
}

// op1ASlice wraps a 1-string+1-any Discord operation that returns a slice.
func op1ASlice(fn func(context.Context, string, any) ([]map[string]any, error), ctx context.Context) func(string, any) (any, error) {
	if fn == nil {
		return func(string, any) (any, error) { return []map[string]any{}, nil }
	}
	return func(a string, b any) (any, error) { return fn(ctx, a, b) }
}

// op2A wraps a 2-string+1-any Discord operation that returns only an error.
func op2A(fn func(context.Context, string, string, any) error, ctx context.Context) func(string, string, any) error {
	if fn == nil {
		return func(string, string, any) error { return nil }
	}
	return func(a, b string, c any) error { return fn(ctx, a, b, c) }
}

func discordOpsObject(vm *goja.Runtime, ctx context.Context, ops *DiscordOps) *goja.Object {
	root := vm.NewObject()
	guilds := vm.NewObject()
	roles := vm.NewObject()
	threads := vm.NewObject()
	channels := vm.NewObject()
	messages := vm.NewObject()
	members := vm.NewObject()
	if ops == nil {
		_ = guilds.Set("fetch", op1(nil, ctx, map[string]any{}))
		_ = roles.Set("list", op1(nil, ctx, []map[string]any{}))
		_ = roles.Set("fetch", op2(nil, ctx, map[string]any{}))
		_ = threads.Set("fetch", op1(nil, ctx, map[string]any{}))
		_ = threads.Set("join", op1E(nil, ctx))
		_ = threads.Set("leave", op1E(nil, ctx))
		_ = threads.Set("start", op1A(nil, ctx, map[string]any{}))
		_ = channels.Set("send", op1AErr(nil, ctx))
		_ = channels.Set("fetch", op1(nil, ctx, map[string]any{}))
		_ = channels.Set("setTopic", op2E(nil, ctx))
		_ = channels.Set("setSlowmode", func(channelID string, seconds int) error { return nil })
		_ = messages.Set("fetch", op2(nil, ctx, map[string]any{}))
		_ = messages.Set("list", op1ASlice(nil, ctx))
		_ = messages.Set("edit", op2A(nil, ctx))
		_ = messages.Set("delete", op2E(nil, ctx))
		_ = messages.Set("react", op3E(nil, ctx))
		_ = messages.Set("pin", op2E(nil, ctx))
		_ = messages.Set("unpin", op2E(nil, ctx))
		_ = messages.Set("listPinned", op1(nil, ctx, []map[string]any{}))
		_ = messages.Set("bulkDelete", op1AErr(nil, ctx))
		_ = members.Set("fetch", op2(nil, ctx, map[string]any{}))
		_ = members.Set("list", op1ASlice(nil, ctx))
		_ = members.Set("addRole", op3E(nil, ctx))
		_ = members.Set("removeRole", op3E(nil, ctx))
		_ = members.Set("timeout", op2A(nil, ctx))
		_ = members.Set("kick", op2A(nil, ctx))
		_ = members.Set("ban", op2A(nil, ctx))
		_ = members.Set("unban", op2E(nil, ctx))
	} else {
		_ = guilds.Set("fetch", op1(ops.GuildFetch, ctx, map[string]any{}))
		_ = roles.Set("list", op1(ops.RoleList, ctx, []map[string]any{}))
		_ = roles.Set("fetch", op2(ops.RoleFetch, ctx, map[string]any{}))
		_ = threads.Set("fetch", op1(ops.ThreadFetch, ctx, map[string]any{}))
		_ = threads.Set("join", op1E(ops.ThreadJoin, ctx))
		_ = threads.Set("leave", op1E(ops.ThreadLeave, ctx))
		_ = threads.Set("start", op1A(ops.ThreadStart, ctx, map[string]any{}))
		_ = channels.Set("send", op1AErr(ops.ChannelSend, ctx))
		_ = channels.Set("fetch", op1(ops.ChannelFetch, ctx, map[string]any{}))
		_ = channels.Set("setTopic", op2E(ops.ChannelSetTopic, ctx))
		_ = channels.Set("setSlowmode", func(channelID string, seconds int) error {
			if ops.ChannelSetSlowmode == nil {
				return nil
			}
			return ops.ChannelSetSlowmode(ctx, channelID, seconds)
		})
		_ = messages.Set("fetch", op2(ops.MessageFetch, ctx, map[string]any{}))
		_ = messages.Set("list", op1ASlice(ops.MessageList, ctx))
		_ = messages.Set("edit", op2A(ops.MessageEdit, ctx))
		_ = messages.Set("delete", op2E(ops.MessageDelete, ctx))
		_ = messages.Set("react", op3E(ops.MessageReact, ctx))
		_ = messages.Set("pin", op2E(ops.MessagePin, ctx))
		_ = messages.Set("unpin", op2E(ops.MessageUnpin, ctx))
		_ = messages.Set("listPinned", op1(ops.MessageListPinned, ctx, []map[string]any{}))
		_ = messages.Set("bulkDelete", op1AErr(ops.MessageBulkDelete, ctx))
		_ = members.Set("fetch", op2(ops.MemberFetch, ctx, map[string]any{}))
		_ = members.Set("list", op1ASlice(ops.MemberList, ctx))
		_ = members.Set("addRole", op3E(ops.MemberAddRole, ctx))
		_ = members.Set("removeRole", op3E(ops.MemberRemoveRole, ctx))
		_ = members.Set("timeout", op2A(ops.MemberSetTimeout, ctx))
		_ = members.Set("kick", op2A(ops.MemberKick, ctx))
		_ = members.Set("ban", op2A(ops.MemberBan, ctx))
		_ = members.Set("unban", op2E(ops.MemberUnban, ctx))
	}
	_ = root.Set("guilds", guilds)
	_ = root.Set("roles", roles)
	_ = root.Set("threads", threads)
	_ = root.Set("channels", channels)
	_ = root.Set("messages", messages)
	_ = root.Set("members", members)
	return root
}
