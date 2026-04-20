package config

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

// Settings holds the Discord credentials and deployment scope for the bot.
type Settings struct {
	BotToken      string `glazed:"bot-token"`
	ApplicationID string `glazed:"application-id"`
	GuildID       string `glazed:"guild-id"`
	PublicKey     string `glazed:"public-key"`
	ClientID      string `glazed:"client-id"`
	ClientSecret  string `glazed:"client-secret"`
	BotScript     string `glazed:"bot-script"`
}

// FromValues decodes the default Glazed section into bot settings.
func FromValues(vals *values.Values) (Settings, error) {
	settings := Settings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, &settings); err != nil {
		return Settings{}, fmt.Errorf("decode discord settings: %w", err)
	}
	return settings, nil
}

// Validate checks the minimum required values for a gateway-based bot.
func (s Settings) Validate() error {
	var missing []string
	if strings.TrimSpace(s.BotToken) == "" {
		missing = append(missing, "DISCORD_BOT_TOKEN")
	}
	if strings.TrimSpace(s.ApplicationID) == "" {
		missing = append(missing, "DISCORD_APPLICATION_ID")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// HasGuild reports whether a guild ID is present.
func (s Settings) HasGuild() bool {
	return strings.TrimSpace(s.GuildID) != ""
}

// RedactedToken returns a token marker suitable for logs and rows.
func (s Settings) RedactedToken() string {
	token := strings.TrimSpace(s.BotToken)
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "…" + token[len(token)-4:]
}
