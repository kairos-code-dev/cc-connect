package core

import (
	"errors"
	"testing"
)

type liveInputTestSession struct{ *asyncInputSession }

func (s *liveInputTestSession) SteerMessage(prompt string) (bool, error) {
	accepted, err := s.SteerUserInput(prompt)
	if accepted {
		s.events <- Event{Type: EventText, Content: "Immediate answer: " + prompt + "\n"}
	}
	return accepted, err
}

func TestLiveInput_FallsBackWithoutASecondReaderOrTurn(t *testing.T) {
	p := &stubPlatformEngine{n: "test"}
	s := &liveInputTestSession{&asyncInputSession{events: make(chan Event, 4), active: false}}
	e := NewEngine("test", &controllableAgent{nextSession: s}, []Platform{p}, "", LangEnglish)
	key := "test:user"
	t.Cleanup(func() { _ = e.Stop() })
	state := &interactiveState{agentSession: s}
	e.interactiveStates[key] = state
	msg := &Message{SessionKey: key, Content: "late input", ReplyCtx: "ctx"}
	session := e.sessions.GetOrCreateActive(key)
	if e.trySteerBusyMessage(p, msg, key, session, e.sessions) {
		t.Fatal("ended turn must use ordinary queue")
	}
	s.active = true
	msg.Files = []FileAttachment{{}}
	if e.trySteerBusyMessage(p, msg, key, session, e.sessions) {
		t.Fatal("attachments must use ordinary Send staging")
	}
	msg.Files = nil
	msg.ModeOverride = "suggest"
	if e.trySteerBusyMessage(p, msg, key, session, e.sessions) {
		t.Fatal("per-message permission mode must use ordinary Send lifecycle")
	}
	msg.ModeOverride = ""
	state.pendingMessages = []queuedMessage{{content: "older message"}}
	if e.trySteerBusyMessage(p, msg, key, session, e.sessions) {
		t.Fatal("live input must not overtake older queued messages")
	}
	state.pendingMessages = nil
	s.err = errors.New("delivery unknown")
	if !e.trySteerBusyMessage(p, msg, key, session, e.sessions) {
		t.Fatal("uncertain delivery must not be retried as another turn")
	}
	if len(p.getSent()) == 0 {
		t.Fatal("delivery failure must be visible to the user")
	}
}
