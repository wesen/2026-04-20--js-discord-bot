package jsdiscord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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
