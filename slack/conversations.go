package slack

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api/chat"
)

var (
	ErrStateAlreadyExists = errors.New("State Already Defined")
)

// ffjson: skip
type Controls struct {
	b             *Bot
	user, channel string
}

func (c *Controls) Get(key string) (string, error) {
	return c.b.cs.GetData(c.user, c.channel, c.b.teamID, key)
}

func (c *Controls) Set(key string, value string) error {
	return c.b.cs.SetData(c.user, c.channel, c.b.teamID, key, value)
}

func (c *Controls) To(state string) error {
	return c.b.cs.SetState(c.user, c.channel, c.b.teamID, state)
}

func (c *Controls) End() error {
	return c.b.cs.End(c.user, c.channel, c.b.teamID)
}

func (c *Controls) Bot() *Bot { return c.b }

type ConversationHandler func(msg *chat.Message, controls *Controls)

// ffjson: skip
type Conversation struct {
	mp map[string]ConversationHandler
}

func NewConversation() *Conversation {
	return &Conversation{make(map[string]ConversationHandler)}
}

func (s *Conversation) On(state string, handler ConversationHandler) error {
	if _, ok := s.mp[state]; ok {
		return ErrStateAlreadyExists
	}

	s.mp[state] = handler
	return nil
}

type ConversationRegistry map[string]*Conversation

func NewConversationRegistry() ConversationRegistry {
	return make(ConversationRegistry)
}

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

func (c ConversationRegistry) Get(name string) (*Conversation, error) {
	conv, ok := c[name]

	if !ok {
		return nil, ErrConversationNotFound
	}

	return conv, nil
}
