package jsdiscord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func (h *Host) baseDispatchRequest(session *discordgo.Session) DispatchRequest {
	return DispatchRequest{
		Metadata: map[string]any{"scriptPath": h.scriptPath},
		Config:   cloneMap(h.runtimeConfig),
		Discord:  buildDiscordOps(h.scriptPath, session),
		Me:       newCurrentUserSnapshot(session),
	}
}

func withChannelResponder(req DispatchRequest, r *channelResponder) DispatchRequest {
	req.Reply = r.Reply
	req.FollowUp = r.FollowUp
	req.Edit = r.Edit
	req.Defer = r.Defer
	return req
}

func withInteractionResponder(req DispatchRequest, r *interactionResponder) DispatchRequest {
	req.Reply = r.Reply
	req.FollowUp = r.FollowUp
	req.Edit = r.Edit
	req.Defer = r.Defer
	req.ShowModal = r.ShowModal
	return req
}

func (h *Host) DispatchReady(ctx context.Context, session *discordgo.Session, ready *discordgo.Ready) error {
	_ = session
	if h == nil || h.handle == nil || ready == nil {
		return nil
	}
	req := h.baseDispatchRequest(session)
	req.Name = "ready"
	req.Me = newUserSnapshot(ready.User)
	req.Command = map[string]any{"event": "ready"}
	req.Interaction = InteractionSnapshot{Type: "ready"}
	_, err := h.handle.DispatchEvent(ctx, req)
	return err
}

func (h *Host) DispatchGuildCreate(ctx context.Context, session *discordgo.Session, guild *discordgo.GuildCreate) error {
	_ = session
	if h == nil || h.handle == nil || guild == nil || guild.Guild == nil {
		return nil
	}
	req := h.baseDispatchRequest(session)
	req.Name = "guildCreate"
	req.Guild = guildCreateMap(guild)
	req.Command = map[string]any{"event": "guildCreate"}
	_, err := h.handle.DispatchEvent(ctx, req)
	return err
}

func (h *Host) DispatchGuildMemberAdd(ctx context.Context, session *discordgo.Session, member *discordgo.GuildMemberAdd) error {
	if h == nil || h.handle == nil || member == nil || member.Member == nil {
		return nil
	}
	req := h.baseDispatchRequest(session)
	req.Name = "guildMemberAdd"
	req.Member = newMemberSnapshot(member.Member)
	req.User = newUserSnapshot(member.User)
	req.Guild = guildMap(member.GuildID)
	req.Command = map[string]any{"event": "guildMemberAdd"}
	_, err := h.handle.DispatchEvent(ctx, req)
	return err
}

func (h *Host) DispatchGuildMemberUpdate(ctx context.Context, session *discordgo.Session, member *discordgo.GuildMemberUpdate) error {
	if h == nil || h.handle == nil || member == nil || member.Member == nil {
		return nil
	}
	req := h.baseDispatchRequest(session)
	req.Name = "guildMemberUpdate"
	req.Member = newMemberSnapshot(member.Member)
	req.Before = memberMap(member.BeforeUpdate)
	req.User = newUserSnapshot(member.User)
	req.Guild = guildMap(member.GuildID)
	req.Command = map[string]any{"event": "guildMemberUpdate"}
	_, err := h.handle.DispatchEvent(ctx, req)
	return err
}

func (h *Host) DispatchGuildMemberRemove(ctx context.Context, session *discordgo.Session, member *discordgo.GuildMemberRemove) error {
	if h == nil || h.handle == nil || member == nil || member.Member == nil {
		return nil
	}
	req := h.baseDispatchRequest(session)
	req.Name = "guildMemberRemove"
	req.Member = newMemberSnapshot(member.Member)
	req.User = newUserSnapshot(member.User)
	req.Guild = guildMap(member.GuildID)
	req.Command = map[string]any{"event": "guildMemberRemove"}
	_, err := h.handle.DispatchEvent(ctx, req)
	return err
}

func (h *Host) DispatchMessageCreate(ctx context.Context, session *discordgo.Session, message *discordgo.MessageCreate) error {
	if h == nil || h.handle == nil || message == nil || message.Message == nil {
		return nil
	}
	responder := newChannelResponder(session, message.ChannelID, h.scriptPath)
	req := h.baseDispatchRequest(session)
	req.Name = "messageCreate"
	req.Message = newMessageSnapshot(message.Message)
	req.User = newUserSnapshot(message.Author)
	req.Guild = guildMap(message.GuildID)
	req.Channel = channelMap(message.ChannelID)
	req = withChannelResponder(req, responder)
	result, err := h.handle.DispatchEvent(ctx, req)
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
	req := h.baseDispatchRequest(session)
	req.Name = "messageUpdate"
	req.Message = newMessageSnapshot(message.Message)
	req.Before = newMessageSnapshot(message.BeforeUpdate).ToMap()
	req.User = newUserSnapshot(message.Author)
	req.Guild = guildMap(message.GuildID)
	req.Channel = channelMap(message.ChannelID)
	req = withChannelResponder(req, responder)
	result, err := h.handle.DispatchEvent(ctx, req)
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
	req := h.baseDispatchRequest(session)
	req.Name = "messageDelete"
	req.Message = newMessageDeleteSnapshot(message)
	req.Before = newMessageSnapshot(message.BeforeDelete).ToMap()
	req.Guild = guildMap(message.GuildID)
	req.Channel = channelMap(message.ChannelID)
	req = withChannelResponder(req, responder)
	result, err := h.handle.DispatchEvent(ctx, req)
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
	req := h.baseDispatchRequest(session)
	req.Name = "reactionAdd"
	req.Message = &MessageSnapshot{ID: reaction.MessageID, ChannelID: reaction.ChannelID, GuildID: reaction.GuildID}
	req.User = newUserRefSnapshot(reaction.UserID)
	req.Guild = guildMap(reaction.GuildID)
	req.Channel = channelMap(reaction.ChannelID)
	req.Member = newMemberSnapshot(reaction.Member)
	req.Reaction = newReactionSnapshot(reaction.MessageReaction)
	req = withChannelResponder(req, responder)
	result, err := h.handle.DispatchEvent(ctx, req)
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
	req := h.baseDispatchRequest(session)
	req.Name = "reactionRemove"
	req.Message = &MessageSnapshot{ID: reaction.MessageID, ChannelID: reaction.ChannelID, GuildID: reaction.GuildID}
	req.User = newUserRefSnapshot(reaction.UserID)
	req.Guild = guildMap(reaction.GuildID)
	req.Channel = channelMap(reaction.ChannelID)
	req.Reaction = newReactionSnapshot(reaction.MessageReaction)
	req = withChannelResponder(req, responder)
	result, err := h.handle.DispatchEvent(ctx, req)
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
		return h.dispatchApplicationCommandInteraction(ctx, session, interaction)
	case discordgo.InteractionMessageComponent:
		return h.dispatchMessageComponentInteraction(ctx, session, interaction)
	case discordgo.InteractionModalSubmit:
		return h.dispatchModalSubmitInteraction(ctx, session, interaction)
	case discordgo.InteractionApplicationCommandAutocomplete:
		return h.dispatchAutocompleteInteraction(ctx, session, interaction)
	default:
		return nil
	}
}

func (h *Host) dispatchApplicationCommandInteraction(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	data := interaction.ApplicationCommandData()
	log.Debug().
		Str("script", h.scriptPath).
		Str("interactionType", "applicationCommand").
		Str("command", data.Name).
		Str("commandType", fmt.Sprint(data.CommandType)).
		Str("guildId", interaction.GuildID).
		Str("channelId", interaction.ChannelID).
		Str("userId", interactionUserID(interaction)).
		Msg("dispatching javascript interaction")
	responder := newInteractionResponder(session, interaction, h.scriptPath)

	switch data.CommandType {
	case discordgo.UserApplicationCommand:
		args := map[string]any{}
		if data.Resolved != nil && data.Resolved.Users != nil {
			if targetUser, ok := data.Resolved.Users[data.TargetID]; ok {
				args["target"] = userMap(targetUser)
			}
		}
		req := h.baseDispatchRequest(session)
		req.Name = data.Name
		req.Args = args
		req.Command = map[string]any{"name": data.Name, "id": data.ID, "type": "user"}
		req.Interaction = newInteractionSnapshot(interaction)
		req.Message = newMessageSnapshot(interaction.Message)
		req.User = newInteractionUserSnapshot(interaction)
		req.Guild = guildMap(interaction.GuildID)
		req.Channel = channelMap(interaction.ChannelID)
		req = withInteractionResponder(req, responder)
		result, err := h.handle.DispatchCommand(ctx, req)
		if err != nil {
			if !responder.Acknowledged() {
				_ = responder.Reply(ctx, map[string]any{"content": "user command failed: " + err.Error(), "ephemeral": true})
			}
			return fmt.Errorf("dispatch user command %q for script %q: %w", data.Name, h.scriptPath, err)
		}
		if !responder.Acknowledged() {
			return emitEventResult(ctx, responder.Reply, result)
		}
		return nil

	case discordgo.MessageApplicationCommand:
		args := map[string]any{}
		if data.Resolved != nil && data.Resolved.Messages != nil {
			if targetMessage, ok := data.Resolved.Messages[data.TargetID]; ok {
				args["target"] = messageMap(targetMessage)
			}
		}
		req := h.baseDispatchRequest(session)
		req.Name = data.Name
		req.Args = args
		req.Command = map[string]any{"name": data.Name, "id": data.ID, "type": "message"}
		req.Interaction = newInteractionSnapshot(interaction)
		req.Message = newMessageSnapshot(interaction.Message)
		req.User = newInteractionUserSnapshot(interaction)
		req.Guild = guildMap(interaction.GuildID)
		req.Channel = channelMap(interaction.ChannelID)
		req = withInteractionResponder(req, responder)
		result, err := h.handle.DispatchCommand(ctx, req)
		if err != nil {
			if !responder.Acknowledged() {
				_ = responder.Reply(ctx, map[string]any{"content": "message command failed: " + err.Error(), "ephemeral": true})
			}
			return fmt.Errorf("dispatch message command %q for script %q: %w", data.Name, h.scriptPath, err)
		}
		if !responder.Acknowledged() {
			return emitEventResult(ctx, responder.Reply, result)
		}
		return nil

	default:
		// Chat input command (slash command) — check for subcommands
		args := optionMap(data.Options)
		if len(data.Options) > 0 && data.Options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
			subName := data.Options[0].Name
			subArgs := optionMap(data.Options[0].Options)
			req := h.baseDispatchRequest(session)
			req.Name = data.Name + "/" + subName
			req.RootName = data.Name
			req.SubName = subName
			req.Args = subArgs
			req.Command = map[string]any{"name": data.Name, "id": data.ID, "subName": subName}
			req.Interaction = newInteractionSnapshot(interaction)
			req.Message = newMessageSnapshot(interaction.Message)
			req.User = newInteractionUserSnapshot(interaction)
			req.Guild = guildMap(interaction.GuildID)
			req.Channel = channelMap(interaction.ChannelID)
			req = withInteractionResponder(req, responder)
			result, err := h.handle.DispatchSubcommand(ctx, req)
			if err != nil {
				if !responder.Acknowledged() {
					_ = responder.Reply(ctx, map[string]any{"content": "subcommand failed: " + err.Error(), "ephemeral": true})
				}
				return fmt.Errorf("dispatch subcommand %q/%q for script %q: %w", data.Name, subName, h.scriptPath, err)
			}
			if !responder.Acknowledged() {
				return emitEventResult(ctx, responder.Reply, result)
			}
			return nil
		}

		req := h.baseDispatchRequest(session)
		req.Name = data.Name
		req.Args = args
		req.Command = map[string]any{"name": data.Name, "id": data.ID}
		req.Interaction = newInteractionSnapshot(interaction)
		req.Message = newMessageSnapshot(interaction.Message)
		req.User = newInteractionUserSnapshot(interaction)
		req.Guild = guildMap(interaction.GuildID)
		req.Channel = channelMap(interaction.ChannelID)
		req = withInteractionResponder(req, responder)
		result, err := h.handle.DispatchCommand(ctx, req)
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
	}
}

func (h *Host) dispatchMessageComponentInteraction(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
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
	req := h.baseDispatchRequest(session)
	req.Name = data.CustomID
	req.Values = componentValues(data)
	req.Command = map[string]any{"event": "component"}
	req.Interaction = newInteractionSnapshot(interaction)
	req.Message = newMessageSnapshot(interaction.Message)
	req.User = newInteractionUserSnapshot(interaction)
	req.Guild = guildMap(interaction.GuildID)
	req.Channel = channelMap(interaction.ChannelID)
	req.Component = newComponentSnapshot(data)
	req = withInteractionResponder(req, responder)
	result, err := h.handle.DispatchComponent(ctx, req)
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
}

func (h *Host) dispatchModalSubmitInteraction(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
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
	req := h.baseDispatchRequest(session)
	req.Name = data.CustomID
	req.Values = modalValues(data.Components)
	req.Command = map[string]any{"event": "modal"}
	req.Interaction = newInteractionSnapshot(interaction)
	req.Message = newMessageSnapshot(interaction.Message)
	req.User = newInteractionUserSnapshot(interaction)
	req.Guild = guildMap(interaction.GuildID)
	req.Channel = channelMap(interaction.ChannelID)
	req.Modal = map[string]any{"customId": data.CustomID}
	req.Reply = responder.Reply
	req.FollowUp = responder.FollowUp
	req.Edit = responder.Edit
	req.Defer = responder.Defer
	result, err := h.handle.DispatchModal(ctx, req)
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
}

func (h *Host) dispatchAutocompleteInteraction(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
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
	req := h.baseDispatchRequest(session)
	req.Name = data.Name
	req.Args = optionMap(data.Options)
	req.Command = map[string]any{"name": data.Name, "id": data.ID}
	req.Interaction = newInteractionSnapshot(interaction)
	req.User = newInteractionUserSnapshot(interaction)
	req.Guild = guildMap(interaction.GuildID)
	req.Channel = channelMap(interaction.ChannelID)
	req.Focused = newFocusedOptionSnapshot(focused)
	result, err := h.handle.DispatchAutocomplete(ctx, req)
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
