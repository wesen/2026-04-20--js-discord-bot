package jsdiscord

import (
	"context"

	"github.com/dop251/goja"
)

func discordOpsObject(vm *goja.Runtime, ctx context.Context, ops *DiscordOps) *goja.Object {
	root := vm.NewObject()
	guilds := vm.NewObject()
	roles := vm.NewObject()
	threads := vm.NewObject()
	channels := vm.NewObject()
	messages := vm.NewObject()
	members := vm.NewObject()
	if ops == nil {
		_ = guilds.Set("fetch", func(string) any { return map[string]any{} })
		_ = roles.Set("list", func(string) any { return []map[string]any{} })
		_ = roles.Set("fetch", func(string, string) any { return map[string]any{} })
		_ = threads.Set("fetch", func(string) any { return map[string]any{} })
		_ = threads.Set("join", func(string) error { return nil })
		_ = threads.Set("leave", func(string) error { return nil })
		_ = threads.Set("start", func(string, any) any { return map[string]any{} })
		_ = channels.Set("send", func(string, any) error { return nil })
		_ = channels.Set("fetch", func(string) any { return map[string]any{} })
		_ = channels.Set("setTopic", func(string, string) error { return nil })
		_ = channels.Set("setSlowmode", func(string, int) error { return nil })
		_ = messages.Set("fetch", func(string, string) any { return map[string]any{} })
		_ = messages.Set("list", func(string, any) any { return []map[string]any{} })
		_ = messages.Set("edit", func(string, string, any) error { return nil })
		_ = messages.Set("delete", func(string, string) error { return nil })
		_ = messages.Set("react", func(string, string, string) error { return nil })
		_ = messages.Set("pin", func(string, string) error { return nil })
		_ = messages.Set("unpin", func(string, string) error { return nil })
		_ = messages.Set("listPinned", func(string) any { return []map[string]any{} })
		_ = messages.Set("bulkDelete", func(string, any) error { return nil })
		_ = members.Set("fetch", func(string, string) any { return map[string]any{} })
		_ = members.Set("list", func(string, any) any { return []map[string]any{} })
		_ = members.Set("addRole", func(string, string, string) error { return nil })
		_ = members.Set("removeRole", func(string, string, string) error { return nil })
		_ = members.Set("timeout", func(string, string, any) error { return nil })
		_ = members.Set("kick", func(string, string, any) error { return nil })
		_ = members.Set("ban", func(string, string, any) error { return nil })
		_ = members.Set("unban", func(string, string) error { return nil })
	} else {
		_ = guilds.Set("fetch", func(guildID string) (any, error) {
			if ops.GuildFetch == nil {
				return map[string]any{}, nil
			}
			return ops.GuildFetch(ctx, guildID)
		})
		_ = roles.Set("list", func(guildID string) (any, error) {
			if ops.RoleList == nil {
				return []map[string]any{}, nil
			}
			return ops.RoleList(ctx, guildID)
		})
		_ = roles.Set("fetch", func(guildID, roleID string) (any, error) {
			if ops.RoleFetch == nil {
				return map[string]any{}, nil
			}
			return ops.RoleFetch(ctx, guildID, roleID)
		})
		_ = threads.Set("fetch", func(threadID string) (any, error) {
			if ops.ThreadFetch == nil {
				return map[string]any{}, nil
			}
			return ops.ThreadFetch(ctx, threadID)
		})
		_ = threads.Set("join", func(threadID string) error {
			if ops.ThreadJoin == nil {
				return nil
			}
			return ops.ThreadJoin(ctx, threadID)
		})
		_ = threads.Set("leave", func(threadID string) error {
			if ops.ThreadLeave == nil {
				return nil
			}
			return ops.ThreadLeave(ctx, threadID)
		})
		_ = threads.Set("start", func(channelID string, payload any) (any, error) {
			if ops.ThreadStart == nil {
				return map[string]any{}, nil
			}
			return ops.ThreadStart(ctx, channelID, payload)
		})
		_ = channels.Set("send", func(channelID string, payload any) error {
			if ops.ChannelSend == nil {
				return nil
			}
			return ops.ChannelSend(ctx, channelID, payload)
		})
		_ = channels.Set("fetch", func(channelID string) (any, error) {
			if ops.ChannelFetch == nil {
				return map[string]any{}, nil
			}
			return ops.ChannelFetch(ctx, channelID)
		})
		_ = channels.Set("setTopic", func(channelID, topic string) error {
			if ops.ChannelSetTopic == nil {
				return nil
			}
			return ops.ChannelSetTopic(ctx, channelID, topic)
		})
		_ = channels.Set("setSlowmode", func(channelID string, seconds int) error {
			if ops.ChannelSetSlowmode == nil {
				return nil
			}
			return ops.ChannelSetSlowmode(ctx, channelID, seconds)
		})
		_ = messages.Set("fetch", func(channelID, messageID string) (any, error) {
			if ops.MessageFetch == nil {
				return map[string]any{}, nil
			}
			return ops.MessageFetch(ctx, channelID, messageID)
		})
		_ = messages.Set("list", func(channelID string, payload any) (any, error) {
			if ops.MessageList == nil {
				return []map[string]any{}, nil
			}
			return ops.MessageList(ctx, channelID, payload)
		})
		_ = messages.Set("edit", func(channelID, messageID string, payload any) error {
			if ops.MessageEdit == nil {
				return nil
			}
			return ops.MessageEdit(ctx, channelID, messageID, payload)
		})
		_ = messages.Set("delete", func(channelID, messageID string) error {
			if ops.MessageDelete == nil {
				return nil
			}
			return ops.MessageDelete(ctx, channelID, messageID)
		})
		_ = messages.Set("react", func(channelID, messageID, emoji string) error {
			if ops.MessageReact == nil {
				return nil
			}
			return ops.MessageReact(ctx, channelID, messageID, emoji)
		})
		_ = messages.Set("pin", func(channelID, messageID string) error {
			if ops.MessagePin == nil {
				return nil
			}
			return ops.MessagePin(ctx, channelID, messageID)
		})
		_ = messages.Set("unpin", func(channelID, messageID string) error {
			if ops.MessageUnpin == nil {
				return nil
			}
			return ops.MessageUnpin(ctx, channelID, messageID)
		})
		_ = messages.Set("listPinned", func(channelID string) (any, error) {
			if ops.MessageListPinned == nil {
				return []map[string]any{}, nil
			}
			return ops.MessageListPinned(ctx, channelID)
		})
		_ = messages.Set("bulkDelete", func(channelID string, messageIDs any) error {
			if ops.MessageBulkDelete == nil {
				return nil
			}
			return ops.MessageBulkDelete(ctx, channelID, messageIDs)
		})
		_ = members.Set("fetch", func(guildID, userID string) (any, error) {
			if ops.MemberFetch == nil {
				return map[string]any{}, nil
			}
			return ops.MemberFetch(ctx, guildID, userID)
		})
		_ = members.Set("list", func(guildID string, payload any) (any, error) {
			if ops.MemberList == nil {
				return []map[string]any{}, nil
			}
			return ops.MemberList(ctx, guildID, payload)
		})
		_ = members.Set("addRole", func(guildID, userID, roleID string) error {
			if ops.MemberAddRole == nil {
				return nil
			}
			return ops.MemberAddRole(ctx, guildID, userID, roleID)
		})
		_ = members.Set("removeRole", func(guildID, userID, roleID string) error {
			if ops.MemberRemoveRole == nil {
				return nil
			}
			return ops.MemberRemoveRole(ctx, guildID, userID, roleID)
		})
		_ = members.Set("timeout", func(guildID, userID string, payload any) error {
			if ops.MemberSetTimeout == nil {
				return nil
			}
			return ops.MemberSetTimeout(ctx, guildID, userID, payload)
		})
		_ = members.Set("kick", func(guildID, userID string, payload any) error {
			if ops.MemberKick == nil {
				return nil
			}
			return ops.MemberKick(ctx, guildID, userID, payload)
		})
		_ = members.Set("ban", func(guildID, userID string, payload any) error {
			if ops.MemberBan == nil {
				return nil
			}
			return ops.MemberBan(ctx, guildID, userID, payload)
		})
		_ = members.Set("unban", func(guildID, userID string) error {
			if ops.MemberUnban == nil {
				return nil
			}
			return ops.MemberUnban(ctx, guildID, userID)
		})
	}
	_ = root.Set("guilds", guilds)
	_ = root.Set("roles", roles)
	_ = root.Set("threads", threads)
	_ = root.Set("channels", channels)
	_ = root.Set("messages", messages)
	_ = root.Set("members", members)
	return root
}
