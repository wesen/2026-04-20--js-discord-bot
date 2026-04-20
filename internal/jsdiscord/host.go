package jsdiscord

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/rs/zerolog/log"
)

type Host struct {
	scriptPath    string
	runtime       *engine.Runtime
	handle        *BotHandle
	runtimeConfig map[string]any
}

func NewHost(ctx context.Context, scriptPath string) (*Host, error) {
	if strings.TrimSpace(scriptPath) == "" {
		return nil, fmt.Errorf("discord bot script path is empty")
	}
	absScript, err := filepath.Abs(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("resolve script path: %w", err)
	}
	factory, err := engine.NewBuilder(
		engine.WithModuleRootsFromScript(absScript, engine.DefaultModuleRootsOptions()),
	).WithModules(engine.DefaultRegistryModules()).
		WithRuntimeModuleRegistrars(NewRegistrar(Config{})).
		WithRequireOptions(require.WithGlobalFolders(filepath.Dir(absScript), filepath.Join(filepath.Dir(absScript), "node_modules"))).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build js runtime: %w", err)
	}
	rt, err := factory.NewRuntime(ctx)
	if err != nil {
		return nil, fmt.Errorf("create js runtime: %w", err)
	}
	value, err := rt.Require.Require(absScript)
	if err != nil {
		_ = rt.Close(context.Background())
		return nil, fmt.Errorf("load js bot script: %w", err)
	}
	handle, err := CompileBot(rt.VM, value)
	if err != nil {
		_ = rt.Close(context.Background())
		return nil, fmt.Errorf("compile js bot: %w", err)
	}
	return &Host{scriptPath: absScript, runtime: rt, handle: handle, runtimeConfig: map[string]any{}}, nil
}

func (h *Host) SetRuntimeConfig(config map[string]any) {
	if h == nil {
		return
	}
	h.runtimeConfig = cloneMap(config)
}

func (h *Host) Close(ctx context.Context) error {
	if h == nil || h.runtime == nil {
		return nil
	}
	return h.runtime.Close(ctx)
}

func (h *Host) Describe(ctx context.Context) (map[string]any, error) {
	if h == nil || h.handle == nil {
		return nil, fmt.Errorf("discord js host is nil")
	}
	return h.handle.Describe(ctx)
}

func (h *Host) ApplicationCommands(ctx context.Context) ([]*discordgo.ApplicationCommand, error) {
	desc, err := h.Describe(ctx)
	if err != nil {
		return nil, err
	}
	rawCommands := commandSnapshots(desc["commands"])
	commands := make([]*discordgo.ApplicationCommand, 0, len(rawCommands))
	for _, raw := range rawCommands {
		snapshot, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		command, err := applicationCommandFromSnapshot(snapshot)
		if err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	return commands, nil
}

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

func payloadLogFields(payload any) map[string]any {
	fields := map[string]any{
		"payloadType": fmt.Sprintf("%T", payload),
	}
	switch v := payload.(type) {
	case nil:
		return fields
	case string:
		fields["contentPreview"] = truncateLogText(v, 120)
		return fields
	case map[string]any:
		if content, ok := v["content"]; ok {
			fields["contentPreview"] = truncateLogText(fmt.Sprint(content), 120)
		}
		if ephemeral, ok := v["ephemeral"].(bool); ok {
			fields["ephemeral"] = ephemeral
		}
		if customID := strings.TrimSpace(fmt.Sprint(v["customId"])); customID != "" {
			fields["customId"] = customID
		}
		if title := strings.TrimSpace(fmt.Sprint(v["title"])); title != "" {
			fields["title"] = title
		}
		if embeds, ok := sliceLen(v["embeds"]); ok {
			fields["embedCount"] = embeds
		}
		if components, ok := sliceLen(v["components"]); ok {
			fields["componentCount"] = components
		}
		if files, ok := sliceLen(v["files"]); ok {
			fields["fileCount"] = files
		}
	}
	return fields
}

func sliceLen(value any) (int, bool) {
	switch v := value.(type) {
	case []any:
		return len(v), true
	case []map[string]any:
		return len(v), true
	case []string:
		return len(v), true
	default:
		return 0, false
	}
}

func truncateLogText(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func mergeLogFields(base map[string]any, extras ...map[string]any) map[string]any {
	ret := map[string]any{}
	for key, value := range base {
		ret[key] = value
	}
	for _, extra := range extras {
		for key, value := range extra {
			ret[key] = value
		}
	}
	return ret
}

func logLifecycleDebug(message string, fields map[string]any) {
	e := log.Debug()
	applyFields(e, fields)
	e.Msg(message)
}

func emitEventResult(ctx context.Context, reply func(context.Context, any) error, result any) error {
	if reply == nil || result == nil {
		return nil
	}
	switch v := result.(type) {
	case []any:
		for _, item := range v {
			if item == nil {
				continue
			}
			if err := reply(ctx, item); err != nil {
				return err
			}
		}
		return nil
	default:
		return reply(ctx, result)
	}
}

type interactionResponder struct {
	session     *discordgo.Session
	interaction *discordgo.InteractionCreate
	scriptPath  string
	mu          sync.Mutex
	acked       bool
}

func newInteractionResponder(session *discordgo.Session, interaction *discordgo.InteractionCreate, scriptPath string) *interactionResponder {
	return &interactionResponder{session: session, interaction: interaction, scriptPath: scriptPath}
}

func (r *interactionResponder) Acknowledged() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.acked
}

func (r *interactionResponder) markAcked() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.acked {
		return false
	}
	r.acked = true
	return true
}

func (r *interactionResponder) Defer(ctx context.Context, payload any) error {
	_ = ctx
	if r == nil || r.session == nil || r.interaction == nil {
		return nil
	}
	if !r.markAcked() {
		return nil
	}
	data, err := normalizeResponsePayload(payload)
	if err != nil {
		return err
	}
	err = r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: data.Flags},
	})
	if err == nil {
		logLifecycleDebug("deferred javascript interaction response", mergeLogFields(interactionLogFields(r.scriptPath, r.interaction), payloadLogFields(payload), map[string]any{"action": "defer"}))
	}
	return err
}

func (r *interactionResponder) Reply(ctx context.Context, payload any) error {
	if r == nil || r.session == nil || r.interaction == nil {
		return nil
	}
	if !r.Acknowledged() {
		if !r.markAcked() {
			return nil
		}
		data, err := normalizeResponsePayload(payload)
		if err != nil {
			return err
		}
		err = r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		})
		if err == nil {
			logLifecycleDebug("replied to javascript interaction", mergeLogFields(interactionLogFields(r.scriptPath, r.interaction), payloadLogFields(payload), map[string]any{"action": "reply"}))
		}
		return err
	}
	return r.FollowUp(ctx, payload)
}

func (r *interactionResponder) FollowUp(ctx context.Context, payload any) error {
	_ = ctx
	if r == nil || r.session == nil || r.interaction == nil {
		return nil
	}
	params, err := normalizeWebhookParams(payload)
	if err != nil {
		return err
	}
	_, err = r.session.FollowupMessageCreate(r.interaction.Interaction, false, params)
	if err == nil {
		logLifecycleDebug("sent javascript interaction follow-up", mergeLogFields(interactionLogFields(r.scriptPath, r.interaction), payloadLogFields(payload), map[string]any{"action": "followUp"}))
	}
	return err
}

func (r *interactionResponder) Edit(ctx context.Context, payload any) error {
	_ = ctx
	if r == nil || r.session == nil || r.interaction == nil {
		return nil
	}
	if !r.Acknowledged() {
		return fmt.Errorf("cannot edit an interaction response before replying or deferring")
	}
	edit, err := normalizeWebhookEdit(payload)
	if err != nil {
		return err
	}
	_, err = r.session.InteractionResponseEdit(r.interaction.Interaction, edit)
	if err == nil {
		logLifecycleDebug("edited javascript interaction response", mergeLogFields(interactionLogFields(r.scriptPath, r.interaction), payloadLogFields(payload), map[string]any{"action": "edit"}))
	}
	return err
}

func (r *interactionResponder) ShowModal(ctx context.Context, payload any) error {
	_ = ctx
	if r == nil || r.session == nil || r.interaction == nil {
		return nil
	}
	if !r.markAcked() {
		return fmt.Errorf("cannot show a modal after the interaction has already been acknowledged")
	}
	data, err := normalizeModalPayload(payload)
	if err != nil {
		return err
	}
	err = r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: data,
	})
	if err == nil {
		logLifecycleDebug("showed javascript interaction modal", mergeLogFields(interactionLogFields(r.scriptPath, r.interaction), payloadLogFields(payload), map[string]any{"action": "showModal"}))
	}
	return err
}

type channelResponder struct {
	session    *discordgo.Session
	channelID  string
	scriptPath string
	mu         sync.Mutex
	replied    bool
}

func newChannelResponder(session *discordgo.Session, channelID, scriptPath string) *channelResponder {
	return &channelResponder{session: session, channelID: channelID, scriptPath: scriptPath}
}

func (r *channelResponder) Acknowledged() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.replied
}

func (r *channelResponder) markReplied() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.replied = true
}

func (r *channelResponder) Reply(ctx context.Context, payload any) error {
	_ = ctx
	if r == nil || r.session == nil || strings.TrimSpace(r.channelID) == "" {
		return nil
	}
	message, err := normalizeMessageSend(payload)
	if err != nil {
		return err
	}
	_, err = r.session.ChannelMessageSendComplex(r.channelID, message)
	if err == nil {
		r.markReplied()
		logLifecycleDebug("sent javascript channel reply", mergeLogFields(map[string]any{"script": r.scriptPath, "channelId": r.channelID, "action": "reply"}, payloadLogFields(payload)))
	}
	return err
}

func (r *channelResponder) FollowUp(ctx context.Context, payload any) error {
	return r.Reply(ctx, payload)
}

func (r *channelResponder) Edit(ctx context.Context, payload any) error {
	_ = ctx
	return fmt.Errorf("messageCreate event does not support editing the original triggering user message")
}

func (r *channelResponder) Defer(ctx context.Context, payload any) error {
	_ = ctx
	_ = payload
	return nil
}

func buildDiscordOps(scriptPath string, session *discordgo.Session) *DiscordOps {
	if session == nil {
		return nil
	}
	return &DiscordOps{
		ChannelSend: func(ctx context.Context, channelID string, payload any) error {
			_ = ctx
			channelID = strings.TrimSpace(channelID)
			if channelID == "" {
				return fmt.Errorf("channel send requires channel ID")
			}
			message, err := normalizeMessageSend(payload)
			if err != nil {
				return err
			}
			_, err = session.ChannelMessageSendComplex(channelID, message)
			if err == nil {
				logLifecycleDebug("sent discord channel message from javascript", mergeLogFields(map[string]any{"script": scriptPath, "channelId": channelID, "action": "discord.channels.send"}, payloadLogFields(payload)))
			}
			return err
		},
		MessageEdit: func(ctx context.Context, channelID, messageID string, payload any) error {
			_ = ctx
			channelID = strings.TrimSpace(channelID)
			messageID = strings.TrimSpace(messageID)
			if channelID == "" || messageID == "" {
				return fmt.Errorf("message edit requires channel ID and message ID")
			}
			edit, err := normalizeChannelMessageEdit(channelID, messageID, payload)
			if err != nil {
				return err
			}
			_, err = session.ChannelMessageEditComplex(edit)
			if err == nil {
				logLifecycleDebug("edited discord channel message from javascript", mergeLogFields(map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "action": "discord.messages.edit"}, payloadLogFields(payload)))
			}
			return err
		},
		MessageDelete: func(ctx context.Context, channelID, messageID string) error {
			_ = ctx
			channelID = strings.TrimSpace(channelID)
			messageID = strings.TrimSpace(messageID)
			if channelID == "" || messageID == "" {
				return fmt.Errorf("message delete requires channel ID and message ID")
			}
			err := session.ChannelMessageDelete(channelID, messageID)
			if err == nil {
				logLifecycleDebug("deleted discord channel message from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "action": "discord.messages.delete"})
			}
			return err
		},
		MessageReact: func(ctx context.Context, channelID, messageID, emoji string) error {
			_ = ctx
			channelID = strings.TrimSpace(channelID)
			messageID = strings.TrimSpace(messageID)
			emoji = strings.TrimSpace(emoji)
			if channelID == "" || messageID == "" || emoji == "" {
				return fmt.Errorf("message react requires channel ID, message ID, and emoji")
			}
			err := session.MessageReactionAdd(channelID, messageID, emoji)
			if err == nil {
				logLifecycleDebug("reacted to discord channel message from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "emoji": emoji, "action": "discord.messages.react"})
			}
			return err
		},
	}
}

type normalizedResponse struct {
	Content         string
	Embeds          []*discordgo.MessageEmbed
	Components      []discordgo.MessageComponent
	AllowedMentions *discordgo.MessageAllowedMentions
	Files           []*discordgo.File
	Reference       *discordgo.MessageReference
	TTS             bool
	Ephemeral       bool
}

func normalizeResponsePayload(payload any) (*discordgo.InteractionResponseData, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	data := &discordgo.InteractionResponseData{
		Content:         normalized.Content,
		Embeds:          normalized.Embeds,
		Components:      normalized.Components,
		AllowedMentions: normalized.AllowedMentions,
		Files:           normalized.Files,
		TTS:             normalized.TTS,
	}
	if normalized.Ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}
	return data, nil
}

func normalizeModalPayload(payload any) (*discordgo.InteractionResponseData, error) {
	mapping, _ := payload.(map[string]any)
	if len(mapping) == 0 {
		return nil, fmt.Errorf("modal payload must be an object")
	}
	customID := strings.TrimSpace(fmt.Sprint(mapping["customId"]))
	if customID == "" {
		return nil, fmt.Errorf("modal payload missing customId")
	}
	title := strings.TrimSpace(fmt.Sprint(mapping["title"]))
	if title == "" {
		return nil, fmt.Errorf("modal payload missing title")
	}
	components, err := normalizeComponents(mapping["components"])
	if err != nil {
		return nil, err
	}
	if len(components) == 0 {
		return nil, fmt.Errorf("modal payload must include at least one component row")
	}
	return &discordgo.InteractionResponseData{CustomID: customID, Title: title, Components: components}, nil
}

func normalizeAutocompleteChoices(payload any) ([]*discordgo.ApplicationCommandOptionChoice, error) {
	switch v := payload.(type) {
	case nil:
		return nil, nil
	case []*discordgo.ApplicationCommandOptionChoice:
		return v, nil
	case []any:
		choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(v))
		for _, item := range v {
			choice, err := normalizeAutocompleteChoice(item)
			if err != nil {
				return nil, err
			}
			if choice != nil {
				choices = append(choices, choice)
			}
		}
		if len(choices) > 25 {
			choices = choices[:25]
		}
		return choices, nil
	default:
		return nil, fmt.Errorf("unsupported autocomplete result type %T", payload)
	}
}

func normalizeAutocompleteChoice(raw any) (*discordgo.ApplicationCommandOptionChoice, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case *discordgo.ApplicationCommandOptionChoice:
		return v, nil
	case map[string]any:
		name := strings.TrimSpace(fmt.Sprint(v["name"]))
		if name == "" {
			return nil, fmt.Errorf("autocomplete choice missing name")
		}
		value, ok := v["value"]
		if !ok {
			return nil, fmt.Errorf("autocomplete choice %q missing value", name)
		}
		return &discordgo.ApplicationCommandOptionChoice{Name: name, Value: value}, nil
	default:
		return nil, fmt.Errorf("unsupported autocomplete choice type %T", raw)
	}
}

func normalizeWebhookParams(payload any) (*discordgo.WebhookParams, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	params := &discordgo.WebhookParams{
		Content:         normalized.Content,
		Embeds:          normalized.Embeds,
		Components:      normalized.Components,
		AllowedMentions: normalized.AllowedMentions,
		Files:           normalized.Files,
		TTS:             normalized.TTS,
	}
	if normalized.Ephemeral {
		params.Flags = discordgo.MessageFlagsEphemeral
	}
	return params, nil
}

func normalizeWebhookEdit(payload any) (*discordgo.WebhookEdit, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	content := normalized.Content
	components := normalized.Components
	embeds := normalized.Embeds
	return &discordgo.WebhookEdit{
		Content:         &content,
		Components:      &components,
		Embeds:          &embeds,
		Files:           normalized.Files,
		AllowedMentions: normalized.AllowedMentions,
	}, nil
}

func normalizeMessageSend(payload any) (*discordgo.MessageSend, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	message := &discordgo.MessageSend{
		Content:         normalized.Content,
		Embeds:          normalized.Embeds,
		Components:      normalized.Components,
		Files:           normalized.Files,
		AllowedMentions: normalized.AllowedMentions,
		Reference:       normalized.Reference,
		TTS:             normalized.TTS,
	}
	if normalized.Ephemeral {
		message.Flags = discordgo.MessageFlagsEphemeral
	}
	return message, nil
}

func normalizeChannelMessageEdit(channelID, messageID string, payload any) (*discordgo.MessageEdit, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	edit := discordgo.NewMessageEdit(channelID, messageID)
	content := normalized.Content
	components := normalized.Components
	embeds := normalized.Embeds
	edit.Content = &content
	edit.Components = &components
	edit.Embeds = &embeds
	edit.Files = normalized.Files
	edit.AllowedMentions = normalized.AllowedMentions
	return edit, nil
}

func normalizePayload(payload any) (*normalizedResponse, error) {
	switch v := payload.(type) {
	case nil:
		return &normalizedResponse{}, nil
	case string:
		return &normalizedResponse{Content: v}, nil
	case map[string]any:
		ret := &normalizedResponse{}
		if content, ok := v["content"]; ok {
			ret.Content = fmt.Sprint(content)
		}
		if tts, ok := v["tts"].(bool); ok {
			ret.TTS = tts
		}
		if ephemeral, ok := v["ephemeral"].(bool); ok {
			ret.Ephemeral = ephemeral
		}
		embeds, err := normalizeEmbeds(v)
		if err != nil {
			return nil, err
		}
		ret.Embeds = embeds
		components, err := normalizeComponents(v["components"])
		if err != nil {
			return nil, err
		}
		ret.Components = components
		mentions, err := normalizeAllowedMentions(v["allowedMentions"])
		if err != nil {
			return nil, err
		}
		ret.AllowedMentions = mentions
		files, err := normalizeFiles(v["files"])
		if err != nil {
			return nil, err
		}
		ret.Files = files
		reference, err := normalizeMessageReference(v["replyTo"])
		if err != nil {
			return nil, err
		}
		ret.Reference = reference
		return ret, nil
	default:
		return &normalizedResponse{Content: fmt.Sprint(payload)}, nil
	}
}

func normalizeEmbeds(payload map[string]any) ([]*discordgo.MessageEmbed, error) {
	if payload == nil {
		return nil, nil
	}
	if raw, ok := payload["embeds"]; ok {
		return normalizeEmbedArray(raw)
	}
	if raw, ok := payload["embed"]; ok {
		embeds, err := normalizeEmbedArray([]any{raw})
		if err != nil {
			return nil, err
		}
		return embeds, nil
	}
	return nil, nil
}

func normalizeEmbedArray(raw any) ([]*discordgo.MessageEmbed, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []*discordgo.MessageEmbed:
		return v, nil
	case []any:
		embeds := make([]*discordgo.MessageEmbed, 0, len(v))
		for _, item := range v {
			embed, err := normalizeEmbed(item)
			if err != nil {
				return nil, err
			}
			if embed != nil {
				embeds = append(embeds, embed)
			}
		}
		return embeds, nil
	case map[string]any:
		embed, err := normalizeEmbed(v)
		if err != nil {
			return nil, err
		}
		if embed == nil {
			return nil, nil
		}
		return []*discordgo.MessageEmbed{embed}, nil
	default:
		return nil, fmt.Errorf("unsupported embeds payload type %T", raw)
	}
}

func normalizeEmbed(raw any) (*discordgo.MessageEmbed, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case *discordgo.MessageEmbed:
		return v, nil
	case map[string]any:
		embed := &discordgo.MessageEmbed{}
		if title, ok := v["title"]; ok {
			embed.Title = fmt.Sprint(title)
		}
		if desc, ok := v["description"]; ok {
			embed.Description = fmt.Sprint(desc)
		}
		if url, ok := v["url"]; ok {
			embed.URL = fmt.Sprint(url)
		}
		if timestamp, ok := v["timestamp"]; ok {
			embed.Timestamp = fmt.Sprint(timestamp)
		}
		if color, ok := intValue(v["color"]); ok {
			embed.Color = color
		}
		if footer, ok := v["footer"].(map[string]any); ok {
			embed.Footer = &discordgo.MessageEmbedFooter{}
			if text, ok := footer["text"]; ok {
				embed.Footer.Text = fmt.Sprint(text)
			}
			if iconURL, ok := footer["iconURL"]; ok {
				embed.Footer.IconURL = fmt.Sprint(iconURL)
			}
		}
		if author, ok := v["author"].(map[string]any); ok {
			embed.Author = &discordgo.MessageEmbedAuthor{}
			if name, ok := author["name"]; ok {
				embed.Author.Name = fmt.Sprint(name)
			}
			if url, ok := author["url"]; ok {
				embed.Author.URL = fmt.Sprint(url)
			}
			if iconURL, ok := author["iconURL"]; ok {
				embed.Author.IconURL = fmt.Sprint(iconURL)
			}
		}
		if image, ok := v["image"].(map[string]any); ok {
			embed.Image = &discordgo.MessageEmbedImage{}
			if url, ok := image["url"]; ok {
				embed.Image.URL = fmt.Sprint(url)
			}
		}
		if thumbnail, ok := v["thumbnail"].(map[string]any); ok {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{}
			if url, ok := thumbnail["url"]; ok {
				embed.Thumbnail.URL = fmt.Sprint(url)
			}
		}
		if fieldsRaw, ok := v["fields"].([]any); ok {
			fields := make([]*discordgo.MessageEmbedField, 0, len(fieldsRaw))
			for _, rawField := range fieldsRaw {
				fieldMap, _ := rawField.(map[string]any)
				if len(fieldMap) == 0 {
					continue
				}
				field := &discordgo.MessageEmbedField{}
				if name, ok := fieldMap["name"]; ok {
					field.Name = fmt.Sprint(name)
				}
				if value, ok := fieldMap["value"]; ok {
					field.Value = fmt.Sprint(value)
				}
				if inline, ok := fieldMap["inline"].(bool); ok {
					field.Inline = inline
				}
				fields = append(fields, field)
			}
			embed.Fields = fields
		}
		return embed, nil
	default:
		return nil, fmt.Errorf("unsupported embed payload type %T", raw)
	}
}

func normalizeComponents(raw any) ([]discordgo.MessageComponent, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []discordgo.MessageComponent:
		return v, nil
	case []any:
		components := make([]discordgo.MessageComponent, 0, len(v))
		for _, item := range v {
			component, err := normalizeComponent(item)
			if err != nil {
				return nil, err
			}
			if component != nil {
				components = append(components, component)
			}
		}
		return components, nil
	default:
		return nil, fmt.Errorf("unsupported components payload type %T", raw)
	}
}

func normalizeComponent(raw any) (discordgo.MessageComponent, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(mapping["type"]))) {
	case "", "actionrow", "action-row", "row":
		rawChildren, _ := mapping["components"].([]any)
		children := make([]discordgo.MessageComponent, 0, len(rawChildren))
		for _, child := range rawChildren {
			component, err := normalizeLeafComponent(child)
			if err != nil {
				return nil, err
			}
			if component != nil {
				children = append(children, component)
			}
		}
		return discordgo.ActionsRow{Components: children}, nil
	default:
		return normalizeLeafComponent(mapping)
	}
}

func normalizeLeafComponent(raw any) (discordgo.MessageComponent, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(mapping["type"]))) {
	case "button":
		style, err := buttonStyleFromValue(mapping["style"])
		if err != nil {
			return nil, err
		}
		button := discordgo.Button{Style: style}
		if label, ok := mapping["label"]; ok {
			button.Label = fmt.Sprint(label)
		}
		if disabled, ok := mapping["disabled"].(bool); ok {
			button.Disabled = disabled
		}
		if customID, ok := mapping["customId"]; ok {
			button.CustomID = fmt.Sprint(customID)
		}
		if url, ok := mapping["url"]; ok {
			button.URL = fmt.Sprint(url)
		}
		return button, nil
	case "select", "stringselect", "string-select":
		return normalizeSelectMenu(mapping, discordgo.StringSelectMenu)
	case "userselect", "user-select":
		return normalizeSelectMenu(mapping, discordgo.UserSelectMenu)
	case "roleselect", "role-select":
		return normalizeSelectMenu(mapping, discordgo.RoleSelectMenu)
	case "mentionableselect", "mentionable-select":
		return normalizeSelectMenu(mapping, discordgo.MentionableSelectMenu)
	case "channelselect", "channel-select":
		return normalizeSelectMenu(mapping, discordgo.ChannelSelectMenu)
	case "textinput", "text-input":
		return normalizeTextInput(mapping)
	case "", "actionrow", "action-row", "row":
		return normalizeComponent(mapping)
	default:
		return nil, fmt.Errorf("unsupported component type %q", mapping["type"])
	}
}

func normalizeSelectMenu(mapping map[string]any, menuType discordgo.SelectMenuType) (discordgo.MessageComponent, error) {
	menu := discordgo.SelectMenu{MenuType: menuType}
	if customID, ok := mapping["customId"]; ok {
		menu.CustomID = fmt.Sprint(customID)
	}
	if placeholder, ok := mapping["placeholder"]; ok {
		menu.Placeholder = fmt.Sprint(placeholder)
	}
	if disabled, ok := mapping["disabled"].(bool); ok {
		menu.Disabled = disabled
	}
	if minValues, ok := intPointer(mapping["minValues"]); ok {
		menu.MinValues = minValues
	}
	if maxValues, ok := intValue(mapping["maxValues"]); ok {
		menu.MaxValues = maxValues
	}
	if options, err := normalizeSelectMenuOptions(mapping["options"]); err != nil {
		return nil, err
	} else if len(options) > 0 {
		menu.Options = options
	}
	return menu, nil
}

func normalizeSelectMenuOptions(raw any) ([]discordgo.SelectMenuOption, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []any:
		ret := make([]discordgo.SelectMenuOption, 0, len(v))
		for _, item := range v {
			mapping, _ := item.(map[string]any)
			if len(mapping) == 0 {
				continue
			}
			label := strings.TrimSpace(fmt.Sprint(mapping["label"]))
			value := strings.TrimSpace(fmt.Sprint(mapping["value"]))
			if label == "" || value == "" {
				return nil, fmt.Errorf("select option requires label and value")
			}
			option := discordgo.SelectMenuOption{Label: label, Value: value}
			if description, ok := mapping["description"]; ok {
				option.Description = fmt.Sprint(description)
			}
			if emoji, ok := mapping["emoji"].(map[string]any); ok {
				normalized := normalizeComponentEmoji(emoji)
				option.Emoji = &normalized
			}
			if defaultValue, ok := mapping["default"].(bool); ok {
				option.Default = defaultValue
			}
			ret = append(ret, option)
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("unsupported select menu options payload type %T", raw)
	}
}

func normalizeComponentEmoji(mapping map[string]any) discordgo.ComponentEmoji {
	return discordgo.ComponentEmoji{
		ID:       strings.TrimSpace(fmt.Sprint(mapping["id"])),
		Name:     strings.TrimSpace(fmt.Sprint(mapping["name"])),
		Animated: boolValue(mapping["animated"]),
	}
}

func normalizeTextInput(mapping map[string]any) (discordgo.MessageComponent, error) {
	customID := strings.TrimSpace(fmt.Sprint(mapping["customId"]))
	if customID == "" {
		return nil, fmt.Errorf("text input requires customId")
	}
	label := strings.TrimSpace(fmt.Sprint(mapping["label"]))
	if label == "" {
		return nil, fmt.Errorf("text input %q requires label", customID)
	}
	style, err := textInputStyleFromValue(mapping["style"])
	if err != nil {
		return nil, err
	}
	input := discordgo.TextInput{CustomID: customID, Label: label, Style: style}
	if placeholder, ok := mapping["placeholder"]; ok {
		input.Placeholder = fmt.Sprint(placeholder)
	}
	if value, ok := mapping["value"]; ok {
		input.Value = fmt.Sprint(value)
	}
	if required, ok := mapping["required"].(bool); ok {
		input.Required = required
	}
	if minLength, ok := intValue(mapping["minLength"]); ok {
		input.MinLength = minLength
	}
	if maxLength, ok := intValue(mapping["maxLength"]); ok {
		input.MaxLength = maxLength
	}
	return input, nil
}

func buttonStyleFromValue(raw any) (discordgo.ButtonStyle, error) {
	if raw == nil {
		return discordgo.PrimaryButton, nil
	}
	if value, ok := intValue(raw); ok {
		return discordgo.ButtonStyle(value), nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(raw))) {
	case "", "primary":
		return discordgo.PrimaryButton, nil
	case "secondary":
		return discordgo.SecondaryButton, nil
	case "success", "green":
		return discordgo.SuccessButton, nil
	case "danger", "red":
		return discordgo.DangerButton, nil
	case "link":
		return discordgo.LinkButton, nil
	default:
		return discordgo.PrimaryButton, fmt.Errorf("unsupported button style %q", raw)
	}
}

func textInputStyleFromValue(raw any) (discordgo.TextInputStyle, error) {
	if raw == nil {
		return discordgo.TextInputShort, nil
	}
	if value, ok := intValue(raw); ok {
		return discordgo.TextInputStyle(value), nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(raw))) {
	case "", "short", "single-line":
		return discordgo.TextInputShort, nil
	case "paragraph", "long", "multi-line":
		return discordgo.TextInputParagraph, nil
	default:
		return discordgo.TextInputShort, fmt.Errorf("unsupported text input style %q", raw)
	}
}

func normalizeAllowedMentions(raw any) (*discordgo.MessageAllowedMentions, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	mentions := &discordgo.MessageAllowedMentions{}
	if parseRaw, ok := mapping["parse"].([]any); ok {
		for _, item := range parseRaw {
			switch strings.ToLower(strings.TrimSpace(fmt.Sprint(item))) {
			case "users":
				mentions.Parse = append(mentions.Parse, discordgo.AllowedMentionTypeUsers)
			case "roles":
				mentions.Parse = append(mentions.Parse, discordgo.AllowedMentionTypeRoles)
			case "everyone":
				mentions.Parse = append(mentions.Parse, discordgo.AllowedMentionTypeEveryone)
			}
		}
	}
	if repliedUser, ok := mapping["repliedUser"].(bool); ok {
		mentions.RepliedUser = repliedUser
	}
	if usersRaw, ok := mapping["users"].([]any); ok {
		mentions.Users = stringSlice(usersRaw)
	}
	if rolesRaw, ok := mapping["roles"].([]any); ok {
		mentions.Roles = stringSlice(rolesRaw)
	}
	return mentions, nil
}

func normalizeFiles(raw any) ([]*discordgo.File, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []any:
		files := make([]*discordgo.File, 0, len(v))
		for _, item := range v {
			mapping, _ := item.(map[string]any)
			if len(mapping) == 0 {
				continue
			}
			name := strings.TrimSpace(fmt.Sprint(mapping["name"]))
			if name == "" {
				return nil, fmt.Errorf("file payload requires name")
			}
			content, ok := mapping["content"]
			if !ok {
				return nil, fmt.Errorf("file payload %q requires content", name)
			}
			file := &discordgo.File{Name: name, Reader: bytes.NewReader([]byte(fmt.Sprint(content)))}
			if contentType, ok := mapping["contentType"]; ok {
				file.ContentType = fmt.Sprint(contentType)
			}
			files = append(files, file)
		}
		return files, nil
	default:
		return nil, fmt.Errorf("unsupported files payload type %T", raw)
	}
}

func normalizeMessageReference(raw any) (*discordgo.MessageReference, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	messageID := strings.TrimSpace(fmt.Sprint(mapping["messageId"]))
	if messageID == "" {
		return nil, fmt.Errorf("replyTo.messageId is required")
	}
	reference := &discordgo.MessageReference{MessageID: messageID}
	if channelID := strings.TrimSpace(fmt.Sprint(mapping["channelId"])); channelID != "" {
		reference.ChannelID = channelID
	}
	if guildID := strings.TrimSpace(fmt.Sprint(mapping["guildId"])); guildID != "" {
		reference.GuildID = guildID
	}
	return reference, nil
}

func intValue(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}

func boolValue(value any) bool {
	ret, _ := value.(bool)
	return ret
}

func intPointer(value any) (*int, bool) {
	if v, ok := intValue(value); ok {
		ret := v
		return &ret, true
	}
	return nil, false
}

func floatValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func floatPointer(value any) (*float64, bool) {
	if v, ok := floatValue(value); ok {
		ret := v
		return &ret, true
	}
	return nil, false
}

func stringSlice(values []any) []string {
	ret := make([]string, 0, len(values))
	for _, value := range values {
		ret = append(ret, fmt.Sprint(value))
	}
	return ret
}

func applicationCommandFromSnapshot(snapshot map[string]any) (*discordgo.ApplicationCommand, error) {
	name := strings.TrimSpace(fmt.Sprint(snapshot["name"]))
	if name == "" {
		return nil, fmt.Errorf("discord command snapshot missing name")
	}
	spec, _ := snapshot["spec"].(map[string]any)
	description := "JavaScript command"
	if spec != nil {
		if raw, ok := spec["description"]; ok && strings.TrimSpace(fmt.Sprint(raw)) != "" {
			description = strings.TrimSpace(fmt.Sprint(raw))
		}
	}
	options, err := applicationCommandOptions(spec)
	if err != nil {
		return nil, fmt.Errorf("discord command %s: %w", name, err)
	}
	return &discordgo.ApplicationCommand{Name: name, Description: description, Options: options}, nil
}

func applicationCommandOptions(spec map[string]any) ([]*discordgo.ApplicationCommandOption, error) {
	if len(spec) == 0 {
		return nil, nil
	}
	rawOptions, ok := spec["options"]
	if !ok || rawOptions == nil {
		return nil, nil
	}
	type optionDraft struct {
		name     string
		raw      any
		required bool
		order    int
	}
	drafts := make([]optionDraft, 0)
	appendDraft := func(name string, raw any, order int) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("option missing name")
		}
		drafts = append(drafts, optionDraft{
			name:     name,
			raw:      raw,
			required: optionRequired(raw),
			order:    order,
		})
		return nil
	}
	switch v := rawOptions.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for i, key := range keys {
			if err := appendDraft(key, v[key], i); err != nil {
				return nil, err
			}
		}
	case []any:
		for i, raw := range v {
			mapping, _ := raw.(map[string]any)
			if err := appendDraft(strings.TrimSpace(fmt.Sprint(mapping["name"])), mapping, i); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported options payload type %T", rawOptions)
	}
	required := make([]optionDraft, 0, len(drafts))
	optional := make([]optionDraft, 0, len(drafts))
	for _, draft := range drafts {
		if draft.required {
			required = append(required, draft)
		} else {
			optional = append(optional, draft)
		}
	}
	sort.SliceStable(required, func(i, j int) bool {
		if required[i].name != required[j].name {
			return required[i].name < required[j].name
		}
		return required[i].order < required[j].order
	})
	sort.SliceStable(optional, func(i, j int) bool {
		if optional[i].name != optional[j].name {
			return optional[i].name < optional[j].name
		}
		return optional[i].order < optional[j].order
	})
	out := make([]*discordgo.ApplicationCommandOption, 0, len(drafts))
	for _, draft := range append(required, optional...) {
		child, err := optionSpecToDiscord(draft.name, draft.raw)
		if err != nil {
			return nil, err
		}
		out = append(out, child)
	}
	return out, nil
}

func optionRequired(raw any) bool {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return false
	}
	required, _ := mapping["required"].(bool)
	return required
}

func optionSpecToDiscord(name string, raw any) (*discordgo.ApplicationCommandOption, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("option missing name")
	}
	mapping, _ := raw.(map[string]any)
	description := "Option for JavaScript command"
	if mapping != nil {
		if rawDesc, ok := mapping["description"]; ok && strings.TrimSpace(fmt.Sprint(rawDesc)) != "" {
			description = strings.TrimSpace(fmt.Sprint(rawDesc))
		}
	}
	optionType, err := optionTypeFromSpec(mapping)
	if err != nil {
		return nil, fmt.Errorf("option %s: %w", name, err)
	}
	ret := &discordgo.ApplicationCommandOption{Name: name, Description: description, Type: optionType}
	if required, ok := mapping["required"].(bool); ok {
		ret.Required = required
	}
	if autocomplete, ok := mapping["autocomplete"].(bool); ok {
		ret.Autocomplete = autocomplete
	}
	choices, err := optionChoicesFromSpec(mapping)
	if err != nil {
		return nil, fmt.Errorf("option %s: %w", name, err)
	}
	ret.Choices = choices
	if ret.Autocomplete && len(ret.Choices) > 0 {
		return nil, fmt.Errorf("option %s cannot define both autocomplete and static choices", name)
	}
	if minLength, ok := intPointer(mapping["minLength"]); ok {
		ret.MinLength = minLength
	}
	if maxLength, ok := intValue(mapping["maxLength"]); ok {
		ret.MaxLength = maxLength
	}
	if minValue, ok := floatPointer(mapping["minValue"]); ok {
		ret.MinValue = minValue
	}
	if maxValue, ok := floatValue(mapping["maxValue"]); ok {
		ret.MaxValue = maxValue
	}
	return ret, nil
}

func optionChoicesFromSpec(mapping map[string]any) ([]*discordgo.ApplicationCommandOptionChoice, error) {
	if len(mapping) == 0 {
		return nil, nil
	}
	raw, ok := mapping["choices"]
	if !ok || raw == nil {
		return nil, nil
	}
	values, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("choices must be an array")
	}
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(values))
	for _, item := range values {
		choice, err := normalizeAutocompleteChoice(item)
		if err != nil {
			return nil, err
		}
		if choice != nil {
			choices = append(choices, choice)
		}
	}
	return choices, nil
}

func optionTypeFromSpec(mapping map[string]any) (discordgo.ApplicationCommandOptionType, error) {
	if mapping == nil {
		return discordgo.ApplicationCommandOptionString, nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(mapping["type"]))) {
	case "", "string":
		return discordgo.ApplicationCommandOptionString, nil
	case "int", "integer":
		return discordgo.ApplicationCommandOptionInteger, nil
	case "bool", "boolean":
		return discordgo.ApplicationCommandOptionBoolean, nil
	case "number", "float":
		return discordgo.ApplicationCommandOptionNumber, nil
	case "user":
		return discordgo.ApplicationCommandOptionUser, nil
	case "channel":
		return discordgo.ApplicationCommandOptionChannel, nil
	case "role":
		return discordgo.ApplicationCommandOptionRole, nil
	case "mentionable":
		return discordgo.ApplicationCommandOptionMentionable, nil
	default:
		return discordgo.ApplicationCommandOptionString, fmt.Errorf("unsupported option type %q", mapping["type"])
	}
}

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
