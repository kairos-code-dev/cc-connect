package discord

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/chenhg5/cc-connect/core"
)

func TestRegisterCommands_NamespacedSkillsDoNotBreakAllSlashCommands(t *testing.T) {
	session, err := discordgo.New("Bot test-token")
	if err != nil {
		t.Fatal(err)
	}
	var submitted []*discordgo.ApplicationCommand
	session.Client = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if err := json.NewDecoder(req.Body).Decode(&submitted); err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("[]"))}, nil
	})}
	p := &Platform{session: session, appID: "app", guildID: "guild", readyCh: make(chan struct{})}
	close(p.readyCh)
	if err := p.RegisterCommands([]core.BotCommandInfo{
		{Command: "status", Description: "Show status"},
		{Command: "pages:write-page", Description: "Namespaced skill"},
		{Command: "help", Description: "Show help"},
	}); err != nil {
		t.Fatal(err)
	}
	if len(submitted) != 2 || submitted[0].Name != "status" || submitted[1].Name != "help" {
		t.Fatalf("submitted commands = %#v; want valid commands without namespaced skill", submitted)
	}
}
