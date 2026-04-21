package jsdiscord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
