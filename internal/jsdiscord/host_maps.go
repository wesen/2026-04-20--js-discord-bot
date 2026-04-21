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

func guildSnapshotMap(guild *discordgo.Guild) map[string]any {
	if guild == nil {
		return map[string]any{}
	}
	ret := map[string]any{
		"id":                guild.ID,
		"name":              guild.Name,
		"ownerID":           guild.OwnerID,
		"description":       guild.Description,
		"memberCount":       guild.MemberCount,
		"large":             guild.Large,
		"verificationLevel": fmt.Sprint(guild.VerificationLevel),
	}
	if len(guild.Features) > 0 {
		features := make([]string, 0, len(guild.Features))
		for _, feature := range guild.Features {
			features = append(features, string(feature))
		}
		ret["features"] = features
	}
	if guild.Icon != "" {
		ret["icon"] = guild.Icon
	}
	if guild.AfkChannelID != "" {
		ret["afkChannelID"] = guild.AfkChannelID
	}
	if guild.WidgetChannelID != "" {
		ret["widgetChannelID"] = guild.WidgetChannelID
	}
	return ret
}

func roleMap(guildID string, role *discordgo.Role) map[string]any {
	if role == nil {
		return map[string]any{}
	}
	ret := map[string]any{
		"id":           role.ID,
		"guildID":      strings.TrimSpace(guildID),
		"name":         role.Name,
		"color":        role.Color,
		"position":     role.Position,
		"permissions":  fmt.Sprint(role.Permissions),
		"managed":      role.Managed,
		"mentionable":  role.Mentionable,
		"hoist":        role.Hoist,
		"unicodeEmoji": role.UnicodeEmoji,
	}
	if role.Icon != "" {
		ret["icon"] = role.Icon
	}
	if role.Flags != 0 {
		ret["flags"] = int(role.Flags)
	}
	return ret
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
	if channel.OwnerID != "" {
		ret["ownerID"] = channel.OwnerID
	}
	if channel.MessageCount != 0 {
		ret["messageCount"] = channel.MessageCount
	}
	if channel.MemberCount != 0 {
		ret["memberCount"] = channel.MemberCount
	}
	if channel.IsThread() {
		ret["thread"] = true
	}
	if channel.ThreadMetadata != nil {
		ret["archived"] = channel.ThreadMetadata.Archived
		ret["autoArchiveDuration"] = channel.ThreadMetadata.AutoArchiveDuration
		ret["locked"] = channel.ThreadMetadata.Locked
		ret["invitable"] = channel.ThreadMetadata.Invitable
		ret["archiveTimestamp"] = channel.ThreadMetadata.ArchiveTimestamp.Format(time.RFC3339)
	}
	if channel.Member != nil {
		ret["currentMember"] = map[string]any{
			"id":            channel.Member.ID,
			"userID":        channel.Member.UserID,
			"joinTimestamp": channel.Member.JoinTimestamp.Format(time.RFC3339),
			"flags":         channel.Member.Flags,
		}
	}
	return ret
}

func messageMap(message *discordgo.Message) map[string]any {
	if message == nil {
		return map[string]any{}
	}
	ret := map[string]any{
		"id":        message.ID,
		"content":   message.Content,
		"guildID":   message.GuildID,
		"channelID": message.ChannelID,
		"author":    userMap(message.Author),
		"type":      int(message.Type),
	}
	if !message.Timestamp.IsZero() {
		ret["timestamp"] = message.Timestamp.Format(time.RFC3339)
	}
	if message.EditedTimestamp != nil && !message.EditedTimestamp.IsZero() {
		ret["editedTimestamp"] = message.EditedTimestamp.Format(time.RFC3339)
	}
	if message.MessageReference != nil {
		ret["messageReference"] = map[string]any{
			"messageID":       message.MessageReference.MessageID,
			"channelID":       message.MessageReference.ChannelID,
			"guildID":         message.MessageReference.GuildID,
			"failIfNotExists": message.MessageReference.FailIfNotExists,
		}
	}
	if len(message.Attachments) > 0 {
		attachments := make([]map[string]any, 0, len(message.Attachments))
		for _, att := range message.Attachments {
			attachments = append(attachments, map[string]any{
				"id":       att.ID,
				"filename": att.Filename,
				"size":     att.Size,
				"url":      att.URL,
				"proxyURL": att.ProxyURL,
				"width":    att.Width,
				"height":   att.Height,
				"contentType": att.ContentType,
			})
		}
		ret["attachments"] = attachments
	}
	if len(message.Embeds) > 0 {
		embeds := make([]map[string]any, 0, len(message.Embeds))
		for _, embed := range message.Embeds {
			embeds = append(embeds, map[string]any{
				"title":       embed.Title,
				"description": embed.Description,
				"url":         embed.URL,
				"color":       embed.Color,
			})
		}
		ret["embeds"] = embeds
	}
	if len(message.Mentions) > 0 {
		mentions := make([]map[string]any, 0, len(message.Mentions))
		for _, user := range message.Mentions {
			mentions = append(mentions, userMap(user))
		}
		ret["mentions"] = mentions
	}
	if message.ReferencedMessage != nil {
		ret["referencedMessage"] = messageMap(message.ReferencedMessage)
	}
	return ret
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
