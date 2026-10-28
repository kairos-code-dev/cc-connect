package discord

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestBotPolicy_OnlyExplicitlyAllowedBotsCanMention(t *testing.T) {
	p := &Platform{allowFrom: "human", allowBotsFrom: "trusted-bot, second-bot", groupReplyAllGuilds: []string{"*"}}
	message := func(id string, bot bool) *discordgo.MessageCreate {
		return &discordgo.MessageCreate{Message: &discordgo.Message{
			ID: "message", GuildID: "guild", Author: &discordgo.User{ID: id, Bot: bot},
		}}
	}
	for _, tc := range []struct {
		id        string
		bot, want bool
	}{
		{"human", false, true}, {"stranger", false, false},
		{"trusted-bot", true, true}, {"second-bot", true, true},
		{"unknown-bot", true, false}, {"self", true, false},
	} {
		if got := p.allowMessageSender(message(tc.id, tc.bot), "self"); got != tc.want {
			t.Errorf("sender %q: allowed=%v, want %v", tc.id, got, tc.want)
		}
	}
	bot := message("trusted-bot", true)
	if !p.requiresMessageMention(bot) || isDiscordBotMention(bot, "self", "role", false) {
		t.Fatal("reply-all must not let unmentioned bot replies wake other agents")
	}
	bot.Mentions = []*discordgo.User{{ID: "self"}}
	if !isDiscordBotMention(bot, "self", "role", false) {
		t.Fatal("allowlisted bot's direct mention must reach the agent")
	}
	bot.Mentions = nil
	bot.MentionRoles = []string{"role"}
	if !isDiscordBotMention(bot, "self", "role", false) {
		t.Fatal("allowlisted bot's managed-role mention must reach the agent")
	}
	bot.MentionRoles = nil
	bot.MentionEveryone = true
	if isDiscordBotMention(bot, "self", "role", false) {
		t.Fatal("a bot's mass mention must not trigger agent collaboration")
	}
	bot.GuildID = ""
	if p.allowMessageSender(bot, "self") {
		t.Fatal("bot DMs must not bypass the explicit guild mention policy")
	}
	bot.Content = "/shell whoami"
	if got := botRequestContent(bot); !strings.Contains(got, "trusted-bot") || strings.HasPrefix(got, "/") {
		t.Fatalf("bot provenance or command isolation lost: %q", got)
	}
	human := message("human", false)
	human.Content = "/status"
	if got := botRequestContent(human); got != "/status" {
		t.Fatalf("human command changed: %q", got)
	}
}

func TestBotPolicy_DefaultAndHumanWildcardStillRejectBots(t *testing.T) {
	m := &discordgo.MessageCreate{Message: &discordgo.Message{GuildID: "guild", Author: &discordgo.User{ID: "other", Bot: true}}}
	if (&Platform{allowFrom: "*"}).allowMessageSender(m, "self") {
		t.Fatal("a wildcard human allowlist must not enable bot messages")
	}
	if _, err := New(map[string]any{"token": "test-token", "allow_bots_from": "trusted,*"}); err == nil {
		t.Fatal("bot wildcard must be rejected")
	}
	if (&Platform{}).allowMessageSender(nil, "self") {
		t.Fatal("nil message must be rejected")
	}
}

func TestBotPolicy_ReplyAllChannelsOnlyAffectsHumansInNamedChannel(t *testing.T) {
	platform, err := New(map[string]any{"token": "test-token", "progress_style": "legacy", "reply_all_channels": " own-channel,second-channel "})
	if err != nil {
		t.Fatal(err)
	}
	p := platform.(*Platform)
	m := &discordgo.MessageCreate{Message: &discordgo.Message{GuildID: "guild", ChannelID: "own-channel", Author: &discordgo.User{ID: "human"}}}
	if p.requiresMessageMention(m) {
		t.Fatal("human message in dedicated channel must not need mention")
	}
	m.ChannelID = "general"
	if !p.requiresMessageMention(m) {
		t.Fatal("other guild channels must still require mention")
	}
	m.ChannelID = "second-channel"
	m.Author.Bot = true
	if !p.requiresMessageMention(m) {
		t.Fatal("dedicated channel must not enable unmentioned bot loops")
	}
	if _, err := New(map[string]any{"token": "test-token", "reply_all_channels": "*"}); err == nil {
		t.Fatal("channel wildcard must not silently enable unsolicited guild replies")
	}
}
