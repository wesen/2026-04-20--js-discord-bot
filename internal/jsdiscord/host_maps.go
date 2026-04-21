package jsdiscord

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func optionMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]any {
	ret := map[string]any{}
	for _, option := range options {
		if option == nil {
			continue
		}
		if len(option.Options) > 0 {
			for key, value := range optionMap(option.Options) {
				ret[key] = value
			}
			continue
		}
		ret[option.Name] = option.Value
	}
	return ret
}

func findFocusedOption(options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option == nil {
			continue
		}
		if option.Focused {
			return option
		}
		if focused := findFocusedOption(option.Options); focused != nil {
			return focused
		}
	}
	return nil
}

func focusedOptionMap(option *discordgo.ApplicationCommandInteractionDataOption) map[string]any {
	if option == nil {
		return map[string]any{}
	}
	return map[string]any{
		"name":  option.Name,
		"type":  option.Type.String(),
		"value": option.Value,
	}
}

func componentMap(data discordgo.MessageComponentInteractionData) map[string]any {
	return map[string]any{
		"customId": data.CustomID,
		"type":     componentTypeLabel(data.ComponentType),
	}
}

func componentValues(data discordgo.MessageComponentInteractionData) any {
	if len(data.Values) == 0 {
		return []string{}
	}
	return append([]string(nil), data.Values...)
}

func modalValues(components []discordgo.MessageComponent) map[string]any {
	values := map[string]any{}
	for _, component := range components {
		row, ok := component.(*discordgo.ActionsRow)
		if !ok || row == nil {
			continue
		}
		for _, child := range row.Components {
			input, ok := child.(*discordgo.TextInput)
			if !ok || input == nil {
				continue
			}
			values[input.CustomID] = input.Value
		}
	}
	return values
}

func componentTypeLabel(componentType discordgo.ComponentType) string {
	switch componentType {
	case discordgo.ActionsRowComponent:
		return "actionRow"
	case discordgo.ButtonComponent:
		return "button"
	case discordgo.SelectMenuComponent:
		return "select"
	case discordgo.UserSelectMenuComponent:
		return "userSelect"
	case discordgo.RoleSelectMenuComponent:
		return "roleSelect"
	case discordgo.MentionableSelectMenuComponent:
		return "mentionableSelect"
	case discordgo.ChannelSelectMenuComponent:
		return "channelSelect"
	case discordgo.TextInputComponent:
		return "textInput"
	default:
		return fmt.Sprint(componentType)
	}
}

func interactionMap(interaction *discordgo.InteractionCreate) map[string]any {
	if interaction == nil || interaction.Interaction == nil {
		return map[string]any{}
	}
	return map[string]any{"id": interaction.ID, "type": fmt.Sprint(interaction.Type), "guildID": interaction.GuildID, "channelID": interaction.ChannelID}
}

func userMap(user *discordgo.User) map[string]any {
	if user == nil {
		return map[string]any{}
	}
	return map[string]any{"id": user.ID, "username": user.Username, "discriminator": user.Discriminator, "bot": user.Bot}
}

func userRefMap(userID string) map[string]any {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return map[string]any{}
	}
	return map[string]any{"id": userID}
}

func interactionUserMap(interaction *discordgo.InteractionCreate) map[string]any {
	if interaction == nil {
		return map[string]any{}
	}
	if interaction.Member != nil && interaction.Member.User != nil {
		return userMap(interaction.Member.User)
	}
	return userMap(interaction.User)
}

func currentUserMap(session *discordgo.Session) map[string]any {
	if session == nil || session.State == nil {
		return map[string]any{}
	}
	return userMap(session.State.User)
}

func memberMap(member *discordgo.Member) map[string]any {
	if member == nil {
		return map[string]any{}
	}
	ret := map[string]any{
		"guildId": member.GuildID,
		"nick":    member.Nick,
		"roles":   append([]string(nil), member.Roles...),
		"pending": member.Pending,
		"deaf":    member.Deaf,
		"mute":    member.Mute,
	}
	if member.User != nil {
		ret["user"] = userMap(member.User)
		ret["id"] = member.User.ID
	}
	if !member.JoinedAt.IsZero() {
		ret["joinedAt"] = member.JoinedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return ret
}

func emojiMap(emoji discordgo.Emoji) map[string]any {
	return map[string]any{
		"id":       emoji.ID,
		"name":     emoji.Name,
		"animated": emoji.Animated,
	}
}

func reactionMap(reaction *discordgo.MessageReaction) map[string]any {
	if reaction == nil {
		return map[string]any{}
	}
	return map[string]any{
		"userId":    reaction.UserID,
		"messageId": reaction.MessageID,
		"channelId": reaction.ChannelID,
		"guildId":   reaction.GuildID,
		"emoji":     emojiMap(reaction.Emoji),
	}
}

func guildMap(guildID string) map[string]any {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return map[string]any{}
	}
	return map[string]any{"id": guildID}
}

func guildCreateMap(guild *discordgo.GuildCreate) map[string]any {
	if guild == nil || guild.Guild == nil {
		return map[string]any{}
	}
	return map[string]any{
		"id":          guild.ID,
		"name":        guild.Name,
		"memberCount": guild.MemberCount,
	}
}

func channelMap(channelID string) map[string]any {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return map[string]any{}
	}
	return map[string]any{"id": channelID}
}

func channelSnapshotMap(channel *discordgo.Channel) map[string]any {
	if channel == nil {
		return map[string]any{}
	}
	ret := map[string]any{
		"id":               channel.ID,
		"guildID":          channel.GuildID,
		"parentID":         channel.ParentID,
		"name":             channel.Name,
		"type":             fmt.Sprint(channel.Type),
		"topic":            channel.Topic,
		"nsfw":             channel.NSFW,
		"position":         channel.Position,
		"rateLimitPerUser": channel.RateLimitPerUser,
	}
	if channel.LastMessageID != "" {
		ret["lastMessageID"] = channel.LastMessageID
	}
	if channel.LastPinTimestamp != nil {
		ret["lastPinTimestamp"] = channel.LastPinTimestamp.Format(time.RFC3339)
	}
	return ret
}

func messageMap(message *discordgo.Message) map[string]any {
	if message == nil {
		return map[string]any{}
	}
	return map[string]any{
		"id":        message.ID,
		"content":   message.Content,
		"guildID":   message.GuildID,
		"channelID": message.ChannelID,
		"author":    userMap(message.Author),
	}
}

func messageDeleteMap(message *discordgo.MessageDelete) map[string]any {
	if message == nil {
		return map[string]any{}
	}
	if message.Message != nil {
		ret := messageMap(message.Message)
		ret["deleted"] = true
		return ret
	}
	return map[string]any{
		"id":        message.ID,
		"guildID":   message.GuildID,
		"channelID": message.ChannelID,
		"deleted":   true,
	}
}

func commandSnapshots(value any) []any {
	switch v := value.(type) {
	case []any:
		return v
	case []map[string]any:
		ret := make([]any, 0, len(v))
		for _, item := range v {
			ret = append(ret, item)
		}
		return ret
	default:
		return nil
	}
}
