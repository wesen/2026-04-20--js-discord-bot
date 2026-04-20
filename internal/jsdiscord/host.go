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
		Defer:       responder.Defer,
	})
	if err != nil {
		if !responder.Acknowledged() {
			_ = responder.Reply(ctx, map[string]any{"content": "command failed: " + err.Error(), "ephemeral": true})
		}
		return err
	}
	if !responder.Acknowledged() && result != nil {
		return responder.Reply(ctx, result)
	}
	return nil
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

func (r *interactionResponder) Defer(ctx context.Context) error {
	_ = ctx
	if r == nil || r.session == nil || r.interaction == nil {
		return nil
	}
	if !r.markAcked() {
		return nil
	}
	return r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func (r *interactionResponder) Reply(ctx context.Context, payload any) error {
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
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

func normalizeResponsePayload(payload any) (*discordgo.InteractionResponseData, error) {
	switch v := payload.(type) {
	case nil:
		return &discordgo.InteractionResponseData{}, nil
	case string:
		return &discordgo.InteractionResponseData{Content: v}, nil
	case map[string]any:
		data := &discordgo.InteractionResponseData{}
		if content, ok := v["content"]; ok {
			data.Content = fmt.Sprint(content)
		}
		if ephemeral, ok := v["ephemeral"].(bool); ok && ephemeral {
			data.Flags = discordgo.MessageFlagsEphemeral
		}
		return data, nil
	default:
		return &discordgo.InteractionResponseData{Content: fmt.Sprint(payload)}, nil
	}
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

func channelMap(channelID string) map[string]any {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return map[string]any{}
	}
	return map[string]any{"id": channelID}
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
