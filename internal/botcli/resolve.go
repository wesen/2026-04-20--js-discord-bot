package botcli

import (
	"fmt"
	"sort"
	"strings"
)

func ResolveBot(selector string, discovered []DiscoveredBot) (DiscoveredBot, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return DiscoveredBot{}, fmt.Errorf("bot selector is empty")
	}
	matches := make([]DiscoveredBot, 0, len(discovered))
	for _, bot := range discovered {
		if bot.Name() == selector || bot.SourceLabel() == selector {
			matches = append(matches, bot)
		}
	}
	switch len(matches) {
	case 0:
		return DiscoveredBot{}, fmt.Errorf("bot %q not found", selector)
	case 1:
		return matches[0], nil
	default:
		names := make([]string, 0, len(matches))
		for _, bot := range matches {
			names = append(names, bot.Name())
		}
		sort.Strings(names)
		return DiscoveredBot{}, fmt.Errorf("bot selector %q is ambiguous: %s", selector, strings.Join(names, ", "))
	}
}

func ResolveBots(selectors []string, discovered []DiscoveredBot) ([]DiscoveredBot, error) {
	ret := make([]DiscoveredBot, 0, len(selectors))
	seen := map[string]struct{}{}
	for _, selector := range selectors {
		bot, err := ResolveBot(selector, discovered)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[bot.Name()]; ok {
			continue
		}
		seen[bot.Name()] = struct{}{}
		ret = append(ret, bot)
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("no bots selected")
	}
	return ret, nil
}
