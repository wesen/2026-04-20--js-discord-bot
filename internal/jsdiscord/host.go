package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
)

type Host struct {
	scriptPath string
	runtime    *engine.Runtime
	handle     *BotHandle
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
	return &Host{scriptPath: absScript, runtime: rt, handle: handle}, nil
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
		Command:  map[string]any{"event": "ready"},
		Interaction: map[string]any{
			"type": "ready",
		},
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
		Command:  map[string]any{"event": "guildCreate"},
	})
	return err
}

func (h *Host) DispatchMessageCreate(ctx context.Context, session *discordgo.Session, message *discordgo.MessageCreate) error {
	if h == nil || h.handle == nil || message == nil || message.Message == nil {
		return nil
	}
	responder := newChannelResponder(session, message.ChannelID)
	result, err := h.handle.DispatchEvent(ctx, DispatchRequest{
		Name:     "messageCreate",
		Message:  messageMap(message.Message),
		User:     userMap(message.Author),
		Guild:    guildMap(message.GuildID),
		Channel:  channelMap(message.ChannelID),
		Me:       currentUserMap(session),
		Metadata: map[string]any{"scriptPath": h.scriptPath},
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
	if h == nil || h.handle == nil {
		return nil
	}
	if interaction == nil || interaction.Type != discordgo.InteractionApplicationCommand {
		return nil
	}
	data := interaction.ApplicationCommandData()
	responder := newInteractionResponder(session, interaction)
	result, err := h.handle.DispatchCommand(ctx, DispatchRequest{
		Name:        data.Name,
		Args:        optionMap(data.Options),
		Command:     map[string]any{"name": data.Name, "id": data.ID},
		Interaction: interactionMap(interaction),
		User:        interactionUserMap(interaction),
		Guild:       guildMap(interaction.GuildID),
		Channel:     channelMap(interaction.ChannelID),
		Me:          currentUserMap(session),
		Metadata:    map[string]any{"scriptPath": h.scriptPath},
		Reply:       responder.Reply,
		FollowUp:    responder.FollowUp,
		Edit:        responder.Edit,
		Defer:       responder.Defer,
	})
	if err != nil {
		if !responder.Acknowledged() {
			_ = responder.Reply(ctx, map[string]any{"content": "command failed: " + err.Error(), "ephemeral": true})
		}
		return err
	}
	if !responder.Acknowledged() {
		return emitEventResult(ctx, responder.Reply, result)
	}
	return nil
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
	mu          sync.Mutex
	acked       bool
}

func newInteractionResponder(session *discordgo.Session, interaction *discordgo.InteractionCreate) *interactionResponder {
	return &interactionResponder{session: session, interaction: interaction}
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
	return r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: data.Flags},
	})
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
		return r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		})
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
	return err
}

type channelResponder struct {
	session   *discordgo.Session
	channelID string
	mu        sync.Mutex
	replied   bool
}

func newChannelResponder(session *discordgo.Session, channelID string) *channelResponder {
	return &channelResponder{session: session, channelID: channelID}
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

type normalizedResponse struct {
	Content         string
	Embeds          []*discordgo.MessageEmbed
	Components      []discordgo.MessageComponent
	AllowedMentions *discordgo.MessageAllowedMentions
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
		TTS:             normalized.TTS,
	}
	if normalized.Ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}
	return data, nil
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
		AllowedMentions: normalized.AllowedMentions,
		TTS:             normalized.TTS,
	}
	if normalized.Ephemeral {
		message.Flags = discordgo.MessageFlagsEphemeral
	}
	return message, nil
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
	case "", "actionrow", "action-row", "row":
		return normalizeComponent(mapping)
	default:
		return nil, fmt.Errorf("unsupported component type %q", mapping["type"])
	}
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
	out := []*discordgo.ApplicationCommandOption{}
	switch v := rawOptions.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child, err := optionSpecToDiscord(key, v[key])
			if err != nil {
				return nil, err
			}
			out = append(out, child)
		}
	case []any:
		for _, raw := range v {
			mapping, _ := raw.(map[string]any)
			name := strings.TrimSpace(fmt.Sprint(mapping["name"]))
			child, err := optionSpecToDiscord(name, mapping)
			if err != nil {
				return nil, err
			}
			out = append(out, child)
		}
	}
	return out, nil
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
	return ret, nil
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
		ret[option.Name] = option.Value
	}
	return ret
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
