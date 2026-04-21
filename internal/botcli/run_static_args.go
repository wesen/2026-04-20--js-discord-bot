package botcli

import (
	"fmt"
	"os"
	"strings"
)

// staticRunnerArgs holds the statically-known flags parsed before dynamic schema parsing.
type staticRunnerArgs struct {
	Selector          string
	DynamicArgs       []string
	BotRepositories   []string
	BotToken          string
	ApplicationID     string
	GuildID           string
	PublicKey         string
	ClientID          string
	ClientSecret      string
	SyncOnStart       bool
	PrintParsedValues bool
	ShowHelp          bool
}

func defaultStaticRunnerArgs() staticRunnerArgs {
	return staticRunnerArgs{
		BotToken:      os.Getenv("DISCORD_BOT_TOKEN"),
		ApplicationID: os.Getenv("DISCORD_APPLICATION_ID"),
		GuildID:       os.Getenv("DISCORD_GUILD_ID"),
		PublicKey:     os.Getenv("DISCORD_PUBLIC_KEY"),
		ClientID:      os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret:  os.Getenv("DISCORD_CLIENT_SECRET"),
	}
}

func parseStaticRunnerArgs(args []string, defaults staticRunnerArgs) (staticRunnerArgs, error) {
	ret := defaults
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			ret.DynamicArgs = append(ret.DynamicArgs, args[i+1:]...)
			break
		}
		if arg == "--help" || arg == "-h" {
			ret.ShowHelp = true
			continue
		}
		if strings.HasPrefix(arg, "--") {
			name, inlineValue, hasInline := splitLongFlag(arg)
			switch name {
			case "bot-repository":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.BotRepositories = append(ret.BotRepositories, value)
				i = next
			case "bot-token":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.BotToken = value
				i = next
			case "application-id":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.ApplicationID = value
				i = next
			case "guild-id":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.GuildID = value
				i = next
			case "public-key":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.PublicKey = value
				i = next
			case "client-id":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.ClientID = value
				i = next
			case "client-secret":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.ClientSecret = value
				i = next
			case "sync-on-start":
				value, next, err := consumeBoolFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.SyncOnStart = value
				i = next
			case "print-parsed-values":
				value, next, err := consumeBoolFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return staticRunnerArgs{}, err
				}
				ret.PrintParsedValues = value
				i = next
			default:
				ret.DynamicArgs = appendUnknownDynamicArg(ret.DynamicArgs, args, &i)
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			ret.DynamicArgs = appendUnknownDynamicArg(ret.DynamicArgs, args, &i)
			continue
		}
		if strings.TrimSpace(ret.Selector) == "" {
			ret.Selector = strings.TrimSpace(arg)
			continue
		}
		return staticRunnerArgs{}, fmt.Errorf("bots run accepts exactly one bot selector; unexpected argument %q", arg)
	}
	if strings.TrimSpace(ret.Selector) == "" && !ret.ShowHelp {
		return staticRunnerArgs{}, fmt.Errorf("bot selector is required")
	}
	return ret, nil
}

func splitLongFlag(arg string) (string, string, bool) {
	trimmed := strings.TrimPrefix(arg, "--")
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return trimmed, "", false
}

func consumeStringFlagValue(name, inlineValue string, hasInline bool, args []string, index int) (string, int, error) {
	if hasInline {
		return inlineValue, index, nil
	}
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("missing value for flag --%s", name)
	}
	return args[index+1], index + 1, nil
}

func consumeBoolFlagValue(name, inlineValue string, hasInline bool, args []string, index int) (bool, int, error) {
	if hasInline {
		switch strings.ToLower(strings.TrimSpace(inlineValue)) {
		case "true", "1", "yes", "on":
			return true, index, nil
		case "false", "0", "no", "off":
			return false, index, nil
		default:
			return false, index, fmt.Errorf("invalid boolean value %q for flag --%s", inlineValue, name)
		}
	}
	if index+1 < len(args) && isBoolLiteral(args[index+1]) {
		return consumeBoolFlagValue(name, args[index+1], true, args, index+1)
	}
	return true, index, nil
}

func isBoolLiteral(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on", "false", "0", "no", "off":
		return true
	default:
		return false
	}
}

func appendUnknownDynamicArg(dynamic []string, args []string, index *int) []string {
	dynamic = append(dynamic, args[*index])
	if *index+1 < len(args) && !strings.HasPrefix(args[*index+1], "-") {
		dynamic = append(dynamic, args[*index+1])
		*index = *index + 1
	}
	return dynamic
}
