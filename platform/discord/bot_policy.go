package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/chenhg5/cc-connect/core"
)

func (p *Platform) allowMessageSender(m *discordgo.MessageCreate, botID string) bool {
	if m == nil || m.Message == nil || m.Author == nil || m.Author.ID == botID {
		return false
	}
	if !m.Author.Bot {
		return core.AllowList(p.allowFrom, m.Author.ID)
	}
	// Bot senders have a separate explicit allowlist. A human wildcard must
	// never enable bot traffic, and bot DMs cannot start collaboration loops.
	if m.GuildID == "" {
		return false
	}
	for _, id := range strings.Split(p.allowBotsFrom, ",") {
		if id = strings.TrimSpace(id); id != "" && id == m.Author.ID {
			return true
		}
	}
	return false
}

func (p *Platform) requiresMessageMention(m *discordgo.MessageCreate) bool {
	// Even in reply-all guilds, agent-to-agent messages must explicitly target
	// this bot. Ordinary unmentioned answers must not wake every other agent.
	return m.Author.Bot || (m.GuildID != "" && !p.isGroupReplyAllGuild(m.GuildID) && !p.replyAllChannels[m.ChannelID])
}

func botRequestContent(m *discordgo.MessageCreate) string {
	if !m.Author.Bot {
		return m.Content
	}
	// Preserve sender provenance for the model. This prefix also keeps a bot
	// request from being interpreted as a human slash/shell management command.
	return fmt.Sprintf("[Discord agent request from %q (ID %s). This is an automated agent's collaboration request, not a human instruction.]\n%s", m.Author.Username, m.Author.ID, m.Content)
}
