package core

import (
	"fmt"
	"log/slog"
	"strings"
)

func (e *Engine) trySteerBusyMessage(p Platform, msg *Message, key string, session *Session, sessions *SessionManager) bool {
	if strings.TrimSpace(msg.Content) == "" || len(msg.Images) > 0 || len(msg.Files) > 0 || msg.Audio != nil || msg.Location != nil || msg.ModeOverride != "" {
		return false
	}
	e.interactiveMu.Lock()
	state := e.interactiveStates[key]
	e.interactiveMu.Unlock()
	if state == nil {
		return false
	}
	state.liveInputMu.Lock()
	defer state.liveInputMu.Unlock()
	state.mu.Lock()
	as := state.agentSession
	if state.stopped || as == nil || !as.Alive() || state.pending != nil || state.pendingUserInput != nil || len(state.pendingMessages) > 0 {
		state.mu.Unlock()
		return false
	}
	sender, ok := as.(LiveInputSteerer)
	if !ok {
		state.mu.Unlock()
		return false
	}
	if e.isStaleUserMessageLocked(state, msg.UserMessageTimeMs) {
		state.mu.Unlock()
		return true
	}
	state.mu.Unlock()

	// Keep the existing event reader and turn owner. Steering adds input to the
	// same turn; it must not start a second reader or wait for a second result.
	prompt := e.buildSenderPrompt(msg.Content, msg.UserID, msg.UserName, msg.Platform, msg.SessionKey, msg.ChannelKey)
	steered, err := sender.SteerMessage(prompt)
	if err != nil {
		slog.Warn("live input delivery failed", "session", key, "error", err)
		e.reply(p, msg.ReplyCtx, fmt.Sprintf(e.i18n.T(MsgError), err))
		return true
	}
	if !steered {
		return false
	}
	session.TouchUserActivity()
	session.AddHistory("user", msg.Content)
	sessions.Save()
	e.noteUserMessageAccepted(key, msg.UserMessageTimeMs)
	runMessageAccepted(msg)
	slog.Debug("user message delivered to active turn", "session", key, "msg_id", msg.MessageID)
	return true
}
