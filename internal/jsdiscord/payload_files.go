package jsdiscord

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
			contentText := fmt.Sprint(content)
			if encoding := strings.ToLower(strings.TrimSpace(fmt.Sprint(mapping["encoding"]))); encoding == "base64" {
				decoded, err := base64.StdEncoding.DecodeString(contentText)
				if err != nil {
					return nil, fmt.Errorf("file payload %q has invalid base64 content", name)
				}
				file := &discordgo.File{Name: name, Reader: bytes.NewReader(decoded)}
				if contentType, ok := mapping["contentType"]; ok {
					file.ContentType = fmt.Sprint(contentType)
				}
				files = append(files, file)
				continue
			}
			file := &discordgo.File{Name: name, Reader: bytes.NewReader([]byte(contentText))}
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
