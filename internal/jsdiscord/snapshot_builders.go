package jsdiscord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func newUserSnapshot(user *discordgo.User) UserSnapshot {
	if user == nil {
		return UserSnapshot{}
	}
	return UserSnapshot{
		ID:            user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		Bot:           user.Bot,
	}
}

func newMemberSnapshot(member *discordgo.Member) *MemberSnapshot {
	if member == nil {
		return nil
	}
	ret := &MemberSnapshot{
		GuildID: member.GuildID,
		Nick:    member.Nick,
		Roles:   append([]string(nil), member.Roles...),
		Pending: member.Pending,
		Deaf:    member.Deaf,
		Mute:    member.Mute,
		User:    nil,
		ID:      "",
	}
	if member.User != nil {
		ret.User = ptr(newUserSnapshot(member.User))
		ret.ID = member.User.ID
	}
	// JoinedAt is a discordgo.Timestamp (string); keep as-is or format
	if !member.JoinedAt.IsZero() {
		ret.JoinedAt = member.JoinedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return ret
}

func newInteractionSnapshot(interaction *discordgo.InteractionCreate) InteractionSnapshot {
	if interaction == nil || interaction.Interaction == nil {
		return InteractionSnapshot{}
	}
	return InteractionSnapshot{
		ID:        interaction.ID,
		Type:      fmt.Sprint(interaction.Type),
		GuildID:   interaction.GuildID,
		ChannelID: interaction.ChannelID,
	}
}

func newEmojiSnapshot(emoji discordgo.Emoji) EmojiSnapshot {
	return EmojiSnapshot{
		ID:       emoji.ID,
		Name:     emoji.Name,
		Animated: emoji.Animated,
	}
}

func newReactionSnapshot(reaction *discordgo.MessageReaction) ReactionSnapshot {
	if reaction == nil {
		return ReactionSnapshot{}
	}
	return ReactionSnapshot{
		UserID:    reaction.UserID,
		MessageID: reaction.MessageID,
		ChannelID: reaction.ChannelID,
		GuildID:   reaction.GuildID,
		Emoji:     newEmojiSnapshot(reaction.Emoji),
	}
}

func newAttachmentSnapshot(att *discordgo.MessageAttachment) AttachmentSnapshot {
	if att == nil {
		return AttachmentSnapshot{}
	}
	return AttachmentSnapshot{
		ID:          att.ID,
		Filename:    att.Filename,
		Size:        att.Size,
		URL:         att.URL,
		ProxyURL:    att.ProxyURL,
		Width:       att.Width,
		Height:      att.Height,
		ContentType: att.ContentType,
	}
}

func newEmbedSnapshot(embed *discordgo.MessageEmbed) EmbedSnapshot {
	if embed == nil {
		return EmbedSnapshot{}
	}
	return EmbedSnapshot{
		Title:       embed.Title,
		Description: embed.Description,
		URL:         embed.URL,
		Color:       embed.Color,
	}
}

func newMessageReferenceSnapshot(ref *discordgo.MessageReference) *MessageReferenceSnapshot {
	if ref == nil {
		return nil
	}
	return &MessageReferenceSnapshot{
		MessageID:       ref.MessageID,
		ChannelID:       ref.ChannelID,
		GuildID:         ref.GuildID,
		FailIfNotExists: ref.FailIfNotExists != nil && *ref.FailIfNotExists,
	}
}

func newMessageSnapshot(message *discordgo.Message) *MessageSnapshot {
	if message == nil {
		return nil
	}
	ret := &MessageSnapshot{
		ID:        message.ID,
		Content:   message.Content,
		GuildID:   message.GuildID,
		ChannelID: message.ChannelID,
		Author:    ptr(newUserSnapshot(message.Author)),
		Type:      int(message.Type),
	}
	if !message.Timestamp.IsZero() {
		ret.Timestamp = message.Timestamp.Format("2006-01-02T15:04:05Z07:00")
	}
	if message.EditedTimestamp != nil && !message.EditedTimestamp.IsZero() {
		ret.EditedTimestamp = message.EditedTimestamp.Format("2006-01-02T15:04:05Z07:00")
	}
	if message.MessageReference != nil {
		ret.MessageReference = newMessageReferenceSnapshot(message.MessageReference)
	}
	if len(message.Attachments) > 0 {
		ret.Attachments = make([]AttachmentSnapshot, 0, len(message.Attachments))
		for _, att := range message.Attachments {
			ret.Attachments = append(ret.Attachments, newAttachmentSnapshot(att))
		}
	}
	if len(message.Embeds) > 0 {
		ret.Embeds = make([]EmbedSnapshot, 0, len(message.Embeds))
		for _, embed := range message.Embeds {
			ret.Embeds = append(ret.Embeds, newEmbedSnapshot(embed))
		}
	}
	if len(message.Mentions) > 0 {
		ret.Mentions = make([]UserSnapshot, 0, len(message.Mentions))
		for _, user := range message.Mentions {
			ret.Mentions = append(ret.Mentions, newUserSnapshot(user))
		}
	}
	if message.ReferencedMessage != nil {
		ret.ReferencedMessage = newMessageSnapshot(message.ReferencedMessage)
	}
	return ret
}

func newMessageDeleteSnapshot(message *discordgo.MessageDelete) *MessageSnapshot {
	if message == nil {
		return nil
	}
	if message.Message != nil {
		ret := newMessageSnapshot(message.Message)
		ret.Deleted = true
		return ret
	}
	return &MessageSnapshot{
		ID:        message.ID,
		GuildID:   message.GuildID,
		ChannelID: message.ChannelID,
		Deleted:   true,
	}
}

func newComponentSnapshot(data discordgo.MessageComponentInteractionData) ComponentSnapshot {
	return ComponentSnapshot{
		CustomID: data.CustomID,
		Type:     componentTypeLabel(data.ComponentType),
	}
}

func newFocusedOptionSnapshot(option *discordgo.ApplicationCommandInteractionDataOption) FocusedOptionSnapshot {
	if option == nil {
		return FocusedOptionSnapshot{}
	}
	return FocusedOptionSnapshot{
		Name:  option.Name,
		Type:  option.Type.String(),
		Value: option.Value,
	}
}

func newCurrentUserSnapshot(session *discordgo.Session) UserSnapshot {
	if session == nil || session.State == nil || session.State.User == nil {
		return UserSnapshot{}
	}
	return newUserSnapshot(session.State.User)
}

func newUserRefSnapshot(userID string) UserSnapshot {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return UserSnapshot{}
	}
	return UserSnapshot{ID: userID}
}

func newInteractionUserSnapshot(interaction *discordgo.InteractionCreate) UserSnapshot {
	if interaction == nil {
		return UserSnapshot{}
	}
	if interaction.Member != nil && interaction.Member.User != nil {
		return newUserSnapshot(interaction.Member.User)
	}
	return newUserSnapshot(interaction.User)
}

// ptr returns a pointer to v.
func ptr[T any](v T) *T {
	return &v
}
