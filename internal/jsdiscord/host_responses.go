package jsdiscord

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func emitEventResult(ctx context.Context, reply func(context.Context, any) error, result any) error {
	if reply == nil || result == nil {
		return nil
	}
	log.Info().Str("resultType", fmt.Sprintf("%T", result)).Msg("emitting event result")
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

// responseType determines the correct Discord interaction response type.
// Component interactions default to UPDATE_MESSAGE (type 7, edit in-place).
// Everything else uses CHANNEL_MESSAGE_WITH_SOURCE (type 4, new message).
// If the payload is a *normalizedResponse with FollowUp=true, force type 4.
func (r *interactionResponder) responseType(payload any) discordgo.InteractionResponseType {
	if nr, ok := payload.(*normalizedResponse); ok && nr.FollowUp {
		return discordgo.InteractionResponseChannelMessageWithSource
	}
	if r.interaction.Type == discordgo.InteractionMessageComponent {
		return discordgo.InteractionResponseUpdateMessage
	}
	return discordgo.InteractionResponseChannelMessageWithSource
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
	// Component interactions use DEFERRED_UPDATE_MESSAGE (type 6) — no loading state shown.
	// Others use DEFERRED_CHANNEL_MESSAGE_WITH_SOURCE (type 5) — shows "thinking...".
	deferType := discordgo.InteractionResponseDeferredChannelMessageWithSource
	if r.interaction.Type == discordgo.InteractionMessageComponent {
		deferType = discordgo.InteractionResponseDeferredMessageUpdate
	}
	err = r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
		Type: deferType,
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
			log.Error().Err(err).Str("payloadType", fmt.Sprintf("%T", payload)).Fields(payloadLogFields(payload)).Msg("failed to normalize javascript interaction response")
			return err
		}

		// Determine response type:
		// - Component interactions (button/select clicks) → UPDATE_MESSAGE (type 7) = edit in-place
		// - Slash commands and others → CHANNEL_MESSAGE_WITH_SOURCE (type 4) = new message
		// - Unless the payload explicitly requests a followUp (new message)
		responseType := r.responseType(payload)

		log.Info().Str("payloadType", fmt.Sprintf("%T", payload)).Int("responseType", int(responseType)).Fields(payloadLogFields(payload)).Msg("sending interaction response to discord")

		err = r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
			Type: responseType,
			Data: data,
		})
		if err == nil {
			logLifecycleDebug("replied to javascript interaction", mergeLogFields(interactionLogFields(r.scriptPath, r.interaction), payloadLogFields(payload), map[string]any{"action": "reply"}))
		} else {
			log.Error().Err(err).Fields(payloadLogFields(payload)).Msg("discord rejected interaction response")
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
