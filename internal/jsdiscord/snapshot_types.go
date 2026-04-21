package jsdiscord

// UserSnapshot captures a Discord user for dispatch.
type UserSnapshot struct {
	ID            string
	Username      string
	Discriminator string
	Bot           bool
}

func (u UserSnapshot) ToMap() map[string]any {
	if u.ID == "" {
		return map[string]any{}
	}
	return map[string]any{
		"id":            u.ID,
		"username":      u.Username,
		"discriminator": u.Discriminator,
		"bot":           u.Bot,
	}
}

// MemberSnapshot captures a Discord guild member for dispatch.
type MemberSnapshot struct {
	GuildID  string
	Nick     string
	Roles    []string
	Pending  bool
	Deaf     bool
	Mute     bool
	User     *UserSnapshot
	ID       string
	JoinedAt string
}

func (m *MemberSnapshot) ToMap() map[string]any {
	if m == nil {
		return map[string]any{}
	}
	ret := map[string]any{
		"guildId": m.GuildID,
		"nick":    m.Nick,
		"roles":   append([]string(nil), m.Roles...),
		"pending": m.Pending,
		"deaf":    m.Deaf,
		"mute":    m.Mute,
	}
	if m.User != nil {
		ret["user"] = m.User.ToMap()
		ret["id"] = m.User.ID
	}
	if m.ID != "" {
		ret["id"] = m.ID
	}
	if m.JoinedAt != "" {
		ret["joinedAt"] = m.JoinedAt
	}
	return ret
}

// InteractionSnapshot captures a Discord interaction for dispatch.
type InteractionSnapshot struct {
	ID        string
	Type      string
	GuildID   string
	ChannelID string
}

func (i InteractionSnapshot) ToMap() map[string]any {
	if i.ID == "" {
		return map[string]any{}
	}
	return map[string]any{
		"id":        i.ID,
		"type":      i.Type,
		"guildID":   i.GuildID,
		"channelID": i.ChannelID,
	}
}

// EmojiSnapshot captures a Discord emoji for dispatch.
type EmojiSnapshot struct {
	ID       string
	Name     string
	Animated bool
}

func (e EmojiSnapshot) ToMap() map[string]any {
	return map[string]any{
		"id":       e.ID,
		"name":     e.Name,
		"animated": e.Animated,
	}
}

// ReactionSnapshot captures a Discord message reaction for dispatch.
type ReactionSnapshot struct {
	UserID    string
	MessageID string
	ChannelID string
	GuildID   string
	Emoji     EmojiSnapshot
}

func (r ReactionSnapshot) ToMap() map[string]any {
	ret := map[string]any{}
	if r.UserID != "" {
		ret["userId"] = r.UserID
	}
	if r.MessageID != "" {
		ret["messageId"] = r.MessageID
	}
	if r.ChannelID != "" {
		ret["channelId"] = r.ChannelID
	}
	if r.GuildID != "" {
		ret["guildId"] = r.GuildID
	}
	if r.Emoji.Name != "" || r.Emoji.ID != "" {
		ret["emoji"] = r.Emoji.ToMap()
	}
	return ret
}

// EmbedSnapshot captures a Discord message embed for dispatch.
type EmbedSnapshot struct {
	Title       string
	Description string
	URL         string
	Color       int
}

func (e EmbedSnapshot) ToMap() map[string]any {
	ret := map[string]any{}
	if e.Title != "" {
		ret["title"] = e.Title
	}
	if e.Description != "" {
		ret["description"] = e.Description
	}
	if e.URL != "" {
		ret["url"] = e.URL
	}
	if e.Color != 0 {
		ret["color"] = e.Color
	}
	return ret
}

// AttachmentSnapshot captures a Discord message attachment for dispatch.
type AttachmentSnapshot struct {
	ID          string
	Filename    string
	Size        int
	URL         string
	ProxyURL    string
	Width       int
	Height      int
	ContentType string
}

func (a AttachmentSnapshot) ToMap() map[string]any {
	ret := map[string]any{
		"id":       a.ID,
		"filename": a.Filename,
		"size":     a.Size,
		"url":      a.URL,
		"proxyURL": a.ProxyURL,
	}
	if a.Width != 0 {
		ret["width"] = a.Width
	}
	if a.Height != 0 {
		ret["height"] = a.Height
	}
	if a.ContentType != "" {
		ret["contentType"] = a.ContentType
	}
	return ret
}

// MessageReferenceSnapshot captures a Discord message reference for dispatch.
type MessageReferenceSnapshot struct {
	MessageID       string
	ChannelID       string
	GuildID         string
	FailIfNotExists bool
}

func (m MessageReferenceSnapshot) ToMap() map[string]any {
	ret := map[string]any{}
	if m.MessageID != "" {
		ret["messageID"] = m.MessageID
	}
	if m.ChannelID != "" {
		ret["channelID"] = m.ChannelID
	}
	if m.GuildID != "" {
		ret["guildID"] = m.GuildID
	}
	ret["failIfNotExists"] = m.FailIfNotExists
	return ret
}

// MessageSnapshot captures a Discord message for dispatch.
type MessageSnapshot struct {
	ID                string
	Content           string
	GuildID           string
	ChannelID         string
	Author            *UserSnapshot
	Type              int
	Timestamp         string
	EditedTimestamp   string
	Attachments       []AttachmentSnapshot
	Embeds            []EmbedSnapshot
	Mentions          []UserSnapshot
	ReferencedMessage *MessageSnapshot
	MessageReference  *MessageReferenceSnapshot
	Deleted           bool
}

func (m *MessageSnapshot) ToMap() map[string]any {
	if m == nil {
		return map[string]any{}
	}
	ret := map[string]any{
		"id":        m.ID,
		"content":   m.Content,
		"guildID":   m.GuildID,
		"channelID": m.ChannelID,
		"type":      m.Type,
	}
	if m.Author != nil {
		ret["author"] = m.Author.ToMap()
	}
	if m.Timestamp != "" {
		ret["timestamp"] = m.Timestamp
	}
	if m.EditedTimestamp != "" {
		ret["editedTimestamp"] = m.EditedTimestamp
	}
	if len(m.Attachments) > 0 {
		atts := make([]map[string]any, 0, len(m.Attachments))
		for _, a := range m.Attachments {
			atts = append(atts, a.ToMap())
		}
		ret["attachments"] = atts
	}
	if len(m.Embeds) > 0 {
		embeds := make([]map[string]any, 0, len(m.Embeds))
		for _, e := range m.Embeds {
			embeds = append(embeds, e.ToMap())
		}
		ret["embeds"] = embeds
	}
	if len(m.Mentions) > 0 {
		mentions := make([]map[string]any, 0, len(m.Mentions))
		for _, u := range m.Mentions {
			mentions = append(mentions, u.ToMap())
		}
		ret["mentions"] = mentions
	}
	if m.ReferencedMessage != nil {
		ret["referencedMessage"] = m.ReferencedMessage.ToMap()
	}
	if m.MessageReference != nil {
		ret["messageReference"] = m.MessageReference.ToMap()
	}
	if m.Deleted {
		ret["deleted"] = true
	}
	return ret
}

// ComponentSnapshot captures a Discord component interaction for dispatch.
type ComponentSnapshot struct {
	CustomID string
	Type     string
}

func (c ComponentSnapshot) ToMap() map[string]any {
	return map[string]any{
		"customId": c.CustomID,
		"type":     c.Type,
	}
}

// FocusedOptionSnapshot captures a focused autocomplete option for dispatch.
type FocusedOptionSnapshot struct {
	Name  string
	Type  string
	Value any
}

func (f FocusedOptionSnapshot) ToMap() map[string]any {
	return map[string]any{
		"name":  f.Name,
		"type":  f.Type,
		"value": f.Value,
	}
}
