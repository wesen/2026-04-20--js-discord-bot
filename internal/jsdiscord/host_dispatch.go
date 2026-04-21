package jsdiscord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func (h *Host) DispatchReady(ctx context.Context, session *discordgo.Session, ready *discordgo.Ready) error {
	_ = session
	if h == nil || h.handle == nil || ready == nil {
		return nil
	}
	_, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "ready",
		Me:       userMap(ready.User),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Command:  map[string]any{"event": "ready"},
		Interaction: map[string]any{
			"type": "ready",
		},
		Discord: buildDiscordOps(h.scriptPath, session),
	})
	return err
}

func (h *Host) DispatchGuildCreate(ctx context.Context, session *discordgo.Session, guild *discordgo.GuildCreate) error {
	_ = session
	if h == nil || h.handle == nil || guild == nil || guild.Guild == nil {
		return nil
	}
	_, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "guildCreate",
		Guild:    guildCreateMap(guild),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Command:  map[string]any{"event": "guildCreate"},
		Discord:  buildDiscordOps(h.scriptPath, session),
	})
	return err
}

func (h *Host) DispatchGuildMemberAdd(ctx context.Context, session *discordgo.Session, member *discordgo.GuildMemberAdd) error {
	if h == nil || h.handle == nil || member == nil || member.Member == nil {
		return nil
	}
	_, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "guildMemberAdd",
		Member:   memberMap(member.Member),
		User:     userMap(member.User),
		Guild:    guildMap(member.GuildID),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Command:  map[string]any{"event": "guildMemberAdd"},
		Discord:  buildDiscordOps(h.scriptPath, session),
	})
	return err
}

func (h *Host) DispatchGuildMemberUpdate(ctx context.Context, session *discordgo.Session, member *discordgo.GuildMemberUpdate) error {
	if h == nil || h.handle == nil || member == nil || member.Member == nil {
		return nil
	}
	_, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "guildMemberUpdate",
		Member:   memberMap(member.Member),
		Before:   memberMap(member.BeforeUpdate),
		User:     userMap(member.User),
		Guild:    guildMap(member.GuildID),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Command:  map[string]any{"event": "guildMemberUpdate"},
		Discord:  buildDiscordOps(h.scriptPath, session),
	})
	return err
}

func (h *Host) DispatchGuildMemberRemove(ctx context.Context, session *discordgo.Session, member *discordgo.GuildMemberRemove) error {
	if h == nil || h.handle == nil || member == nil || member.Member == nil {
		return nil
	}
	_, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "guildMemberRemove",
		Member:   memberMap(member.Member),
		User:     userMap(member.User),
		Guild:    guildMap(member.GuildID),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Command:  map[string]any{"event": "guildMemberRemove"},
		Discord:  buildDiscordOps(h.scriptPath, session),
	})
	return err
}

func (h *Host) DispatchMessageCreate(ctx context.Context, session *discordgo.Session, message *discordgo.MessageCreate) error {
	if h == nil || h.handle == nil || message == nil || message.Message == nil {
		return nil
	}
	responder := newChannelResponder(session, message.ChannelID, h.scriptPath)
	result, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "messageCreate",
		Message:  messageMap(message.Message),
		User:     userMap(message.Author),
		Guild:    guildMap(message.GuildID),
		Channel:  channelMap(message.ChannelID),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Discord:  buildDiscordOps(h.scriptPath, session),
		Reply:    responder.Reply,
		FollowUp: responder.FollowUp,
		Edit:     responder.Edit,
		Defer:    responder.Defer,
	})
	if err != nil {
		return err
	}
	if !responder.Acknowledged() {
		return emitEventResult(ctx, responder.Reply, result)
	}
	return nil
}

func (h *Host) DispatchMessageUpdate(ctx context.Context, session *discordgo.Session, message *discordgo.MessageUpdate) error {
	if h == nil || h.handle == nil || message == nil {
		return nil
	}
	responder := newChannelResponder(session, message.ChannelID, h.scriptPath)
	result, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "messageUpdate",
		Message:  messageMap(message.Message),
		Before:   messageMap(message.BeforeUpdate),
		User:     userMap(message.Author),
		Guild:    guildMap(message.GuildID),
		Channel:  channelMap(message.ChannelID),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Discord:  buildDiscordOps(h.scriptPath, session),
		Reply:    responder.Reply,
		FollowUp: responder.FollowUp,
		Edit:     responder.Edit,
		Defer:    responder.Defer,
	})
	if err != nil {
		return err
	}
	if !responder.Acknowledged() {
		return emitEventResult(ctx, responder.Reply, result)
	}
	return nil
}

func (h *Host) DispatchMessageDelete(ctx context.Context, session *discordgo.Session, message *discordgo.MessageDelete) error {
	if h == nil || h.handle == nil || message == nil {
		return nil
	}
	responder := newChannelResponder(session, message.ChannelID, h.scriptPath)
	result, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "messageDelete",
		Message:  messageDeleteMap(message),
		Before:   messageMap(message.BeforeDelete),
		Guild:    guildMap(message.GuildID),
		Channel:  channelMap(message.ChannelID),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Discord:  buildDiscordOps(h.scriptPath, session),
		Reply:    responder.Reply,
		FollowUp: responder.FollowUp,
		Edit:     responder.Edit,
		Defer:    responder.Defer,
	})
	if err != nil {
		return err
	}
	if !responder.Acknowledged() {
		return emitEventResult(ctx, responder.Reply, result)
	}
	return nil
}

func (h *Host) DispatchReactionAdd(ctx context.Context, session *discordgo.Session, reaction *discordgo.MessageReactionAdd) error {
	if h == nil || h.handle == nil || reaction == nil || reaction.MessageReaction == nil {
		return nil
	}
	responder := newChannelResponder(session, reaction.ChannelID, h.scriptPath)
	result, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "reactionAdd",
		Message:  map[string]any{"id": reaction.MessageID, "channelID": reaction.ChannelID, "guildID": reaction.GuildID},
		User:     userRefMap(reaction.UserID),
		Guild:    guildMap(reaction.GuildID),
		Channel:  channelMap(reaction.ChannelID),
		Member:   memberMap(reaction.Member),
		Reaction: reactionMap(reaction.MessageReaction),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Discord:  buildDiscordOps(h.scriptPath, session),
		Reply:    responder.Reply,
		FollowUp: responder.FollowUp,
		Edit:     responder.Edit,
		Defer:    responder.Defer,
	})
	if err != nil {
		return err
	}
	if !responder.Acknowledged() {
		return emitEventResult(ctx, responder.Reply, result)
	}
	return nil
}

func (h *Host) DispatchReactionRemove(ctx context.Context, session *discordgo.Session, reaction *discordgo.MessageReactionRemove) error {
	if h == nil || h.handle == nil || reaction == nil || reaction.MessageReaction == nil {
		return nil
	}
	responder := newChannelResponder(session, reaction.ChannelID, h.scriptPath)
	result, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "reactionRemove",
		Message:  map[string]any{"id": reaction.MessageID, "channelID": reaction.ChannelID, "guildID": reaction.GuildID},
		User:     userRefMap(reaction.UserID),
		Guild:    guildMap(reaction.GuildID),
		Channel:  channelMap(reaction.ChannelID),
		Reaction: reactionMap(reaction.MessageReaction),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Discord:  buildDiscordOps(h.scriptPath, session),
		Reply:    responder.Reply,
		FollowUp: responder.FollowUp,
		Edit:     responder.Edit,
		Defer:    responder.Defer,
	})
	if err != nil {
		return err
	}
	if !responder.Acknowledged() {
		return emitEventResult(ctx, responder.Reply, result)
	}
	return nil
}

func (h *Host) DispatchInteraction(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	if h == nil || h.handle == nil || interaction == nil {
		return nil
	}

	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		data := interaction.ApplicationCommandData()
		log.Debug().
			Str("script", h.scriptPath).
			Str("interactionType", "applicationCommand").
			Str("command", data.Name).
			Str("guildId", interaction.GuildID).
			Str("channelId", interaction.ChannelID).
			Str("userId", interactionUserID(interaction)).
			Msg("dispatching javascript interaction")
		responder := newInteractionResponder(session, interaction, h.scriptPath)
		result, err := h.handle.DispatchCommand(ctx, DispatchRequest{
			Name:        data.Name,
			Args:        optionMap(data.Options),
			Command:     map[string]any{"name": data.Name, "id": data.ID},
			Interaction: interactionMap(interaction),
			Message:     messageMap(interaction.Message),
			User:        interactionUserMap(interaction),
			Guild:       guildMap(interaction.GuildID),
			Channel:     channelMap(interaction.ChannelID),
			Me:          currentUserMap(session),
			Metadata:    map[string]any{"scriptPath": h.scriptPath},
			Config:      cloneMap(h.runtimeConfig),
			Discord:     buildDiscordOps(h.scriptPath, session),
			Reply:       responder.Reply,
			FollowUp:    responder.FollowUp,
			Edit:        responder.Edit,
			Defer:       responder.Defer,
			ShowModal:   responder.ShowModal,
		})
		if err != nil {
			if !responder.Acknowledged() {
				_ = responder.Reply(ctx, map[string]any{"content": "command failed: " + err.Error(), "ephemeral": true})
			}
			return fmt.Errorf("dispatch command %q for script %q: %w", data.Name, h.scriptPath, err)
		}
		if !responder.Acknowledged() {
			return emitEventResult(ctx, responder.Reply, result)
		}
		return nil
	case discordgo.InteractionMessageComponent:
		data := interaction.MessageComponentData()
		log.Debug().
			Str("script", h.scriptPath).
			Str("interactionType", "component").
			Str("customId", data.CustomID).
			Str("guildId", interaction.GuildID).
			Str("channelId", interaction.ChannelID).
			Str("userId", interactionUserID(interaction)).
			Msg("dispatching javascript interaction")
		responder := newInteractionResponder(session, interaction, h.scriptPath)
		result, err := h.handle.DispatchComponent(ctx, DispatchRequest{
			Name:        data.CustomID,
			Values:      componentValues(data),
			Command:     map[string]any{"event": "component"},
			Interaction: interactionMap(interaction),
			Message:     messageMap(interaction.Message),
			User:        interactionUserMap(interaction),
			Guild:       guildMap(interaction.GuildID),
			Channel:     channelMap(interaction.ChannelID),
			Me:          currentUserMap(session),
			Metadata:    map[string]any{"scriptPath": h.scriptPath},
			Config:      cloneMap(h.runtimeConfig),
			Discord:     buildDiscordOps(h.scriptPath, session),
			Component:   componentMap(data),
			Reply:       responder.Reply,
			FollowUp:    responder.FollowUp,
			Edit:        responder.Edit,
			Defer:       responder.Defer,
			ShowModal:   responder.ShowModal,
		})
		if err != nil {
			if !responder.Acknowledged() {
				_ = responder.Reply(ctx, map[string]any{"content": "component failed: " + err.Error(), "ephemeral": true})
			}
			return fmt.Errorf("dispatch component %q for script %q: %w", data.CustomID, h.scriptPath, err)
		}
		if !responder.Acknowledged() {
			return emitEventResult(ctx, responder.Reply, result)
		}
		return nil
	case discordgo.InteractionModalSubmit:
		data := interaction.ModalSubmitData()
		log.Debug().
			Str("script", h.scriptPath).
			Str("interactionType", "modal").
			Str("customId", data.CustomID).
			Str("guildId", interaction.GuildID).
			Str("channelId", interaction.ChannelID).
			Str("userId", interactionUserID(interaction)).
			Msg("dispatching javascript interaction")
		responder := newInteractionResponder(session, interaction, h.scriptPath)
		result, err := h.handle.DispatchModal(ctx, DispatchRequest{
			Name:        data.CustomID,
			Values:      modalValues(data.Components),
			Command:     map[string]any{"event": "modal"},
			Interaction: interactionMap(interaction),
			Message:     messageMap(interaction.Message),
			User:        interactionUserMap(interaction),
			Guild:       guildMap(interaction.GuildID),
			Channel:     channelMap(interaction.ChannelID),
			Me:          currentUserMap(session),
			Metadata:    map[string]any{"scriptPath": h.scriptPath},
			Config:      cloneMap(h.runtimeConfig),
			Discord:     buildDiscordOps(h.scriptPath, session),
			Modal:       map[string]any{"customId": data.CustomID},
			Reply:       responder.Reply,
			FollowUp:    responder.FollowUp,
			Edit:        responder.Edit,
			Defer:       responder.Defer,
		})
		if err != nil {
			if !responder.Acknowledged() {
				_ = responder.Reply(ctx, map[string]any{"content": "modal failed: " + err.Error(), "ephemeral": true})
			}
			return fmt.Errorf("dispatch modal %q for script %q: %w", data.CustomID, h.scriptPath, err)
		}
		if !responder.Acknowledged() {
			return emitEventResult(ctx, responder.Reply, result)
		}
		return nil
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := interaction.ApplicationCommandData()
		log.Debug().
			Str("script", h.scriptPath).
			Str("interactionType", "autocomplete").
			Str("command", data.Name).
			Str("guildId", interaction.GuildID).
			Str("channelId", interaction.ChannelID).
			Str("userId", interactionUserID(interaction)).
			Msg("dispatching javascript interaction")
		focused := findFocusedOption(data.Options)
		if focused == nil {
			return fmt.Errorf("autocomplete interaction for %q did not include a focused option", data.Name)
		}
		result, err := h.handle.DispatchAutocomplete(ctx, DispatchRequest{
			Name:        data.Name,
			Args:        optionMap(data.Options),
			Command:     map[string]any{"name": data.Name, "id": data.ID},
			Interaction: interactionMap(interaction),
			User:        interactionUserMap(interaction),
			Guild:       guildMap(interaction.GuildID),
			Channel:     channelMap(interaction.ChannelID),
			Me:          currentUserMap(session),
			Metadata:    map[string]any{"scriptPath": h.scriptPath},
			Config:      cloneMap(h.runtimeConfig),
			Discord:     buildDiscordOps(h.scriptPath, session),
			Focused:     focusedOptionMap(focused),
		})
		if err != nil {
			return fmt.Errorf("dispatch autocomplete %q for script %q: %w", data.Name, h.scriptPath, err)
		}
		choices, err := normalizeAutocompleteChoices(result)
		if err != nil {
			return err
		}
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{Choices: choices},
		})
		if err == nil {
			logLifecycleDebug("replied to javascript autocomplete interaction", mergeLogFields(interactionLogFields(h.scriptPath, interaction), map[string]any{"action": "autocomplete.reply", "choiceCount": len(choices)}))
		}
		return err
	default:
		return nil
	}
}

func interactionUserID(interaction *discordgo.InteractionCreate) string {
	if interaction == nil || interaction.Interaction == nil {
		return ""
	}
	if interaction.Member != nil && interaction.Member.User != nil {
		return interaction.Member.User.ID
	}
	if interaction.User != nil {
		return interaction.User.ID
	}
	return ""
}

func interactionTypeLabel(kind discordgo.InteractionType) string {
	switch kind {
	case discordgo.InteractionApplicationCommand:
		return "applicationCommand"
	case discordgo.InteractionMessageComponent:
		return "component"
	case discordgo.InteractionApplicationCommandAutocomplete:
		return "autocomplete"
	case discordgo.InteractionModalSubmit:
		return "modal"
	default:
		return fmt.Sprintf("interaction:%d", kind)
	}
}

func interactionLogFields(scriptPath string, interaction *discordgo.InteractionCreate) map[string]any {
	fields := map[string]any{
		"script": scriptPath,
	}
	if interaction == nil || interaction.Interaction == nil {
		return fields
	}
	fields["interactionType"] = interactionTypeLabel(interaction.Type)
	fields["interactionId"] = interaction.ID
	fields["guildId"] = interaction.GuildID
	fields["channelId"] = interaction.ChannelID
	fields["userId"] = interactionUserID(interaction)
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		data := interaction.ApplicationCommandData()
		fields["command"] = data.Name
	case discordgo.InteractionMessageComponent:
		fields["customId"] = interaction.MessageComponentData().CustomID
	case discordgo.InteractionModalSubmit:
		fields["customId"] = interaction.ModalSubmitData().CustomID
	}
	return fields
}
