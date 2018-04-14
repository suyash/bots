package slack

import (
	"suy.io/bots/slack/api/chat"
)

// Controls is an object passed to conversation handlers and allows setting, getting
// conversation state and data.
type Controls struct {
	b             *Bot
	user, channel string
}

// Get gets a value for a key in the current conversation state.
func (c *Controls) Get(key string) (string, error) {
	return c.b.cs.GetData(c.user, c.channel, c.b.teamID, key)
}

// Set sets the value for a key.
func (c *Controls) Set(key string, value string) error {
	return c.b.cs.SetData(c.user, c.channel, c.b.teamID, key, value)
}

// To makes a state transition.
func (c *Controls) To(state string) error {
	return c.b.cs.SetState(c.user, c.channel, c.b.teamID, state)
}

// End ends the conversation.
func (c *Controls) End() error {
	return c.b.cs.End(c.user, c.channel, c.b.teamID)
}

// Bot gets the bot associated with the current conversation.
func (c *Controls) Bot() *Bot { return c.b }

// ConversationHandler is a handler for a conversation state.
type ConversationHandler func(msg *chat.Message, controls *Controls)

// Conversation is a mapping of states to ConversationHandler functions.
//
// ffjson: skip
type Conversation struct {
	mp map[string]ConversationHandler
}

// NewConversation creates a new conversation.
func NewConversation() *Conversation {
	return &Conversation{make(map[string]ConversationHandler)}
}

// On adds a new handler for a particular state.
func (s *Conversation) On(state string, handler ConversationHandler) error {
	if _, ok := s.mp[state]; ok {
		return ErrStateAlreadyExists
	}

	s.mp[state] = handler
	return nil
}

// ConversationRegistry is a mapping of conversation names to implementations.
type ConversationRegistry map[string]*Conversation

// NewConversationRegistry creates a ConversationRegistry object.
func NewConversationRegistry() ConversationRegistry {
	return make(ConversationRegistry)
}

// Add adds a new name -> Conversation mapping.
func (c ConversationRegistry) Add(name string, conv *Conversation) error {
	if _, ok := c[name]; ok {
		return ErrConversationExists
	}

	if _, ok := conv.mp["start"]; !ok {
		return ErrNoStartState
	}

	c[name] = conv
	return nil
}

// Get gets the conversation associated with a particular name.
func (c ConversationRegistry) Get(name string) (*Conversation, error) {
	conv, ok := c[name]

	if !ok {
		return nil, ErrConversationNotFound
	}

	return conv, nil
}
