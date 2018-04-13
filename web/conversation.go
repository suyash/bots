package web

import (
	"github.com/pkg/errors"
)

var (
	ErrConversationExists        = errors.New("Conversation Already Exists")
	ErrConversationNotFound      = errors.New("Conversation Not Found")
	ErrConversationAlreadyActive = errors.New("Conversation Already Active")
	ErrNoStartState              = errors.New("Conversation Has no start state")

	ErrStateAlreadyExists = errors.New("State Already Defined")

	ErrNilHandler = errors.New("Nil Handler")
)

// ffjson: skip
type Controls struct {
	b *Bot
}

func (c *Controls) Get(key string) (string, error) {
	return c.b.convs.GetData(c.b.id, key)
}

func (c *Controls) Set(key string, value string) error {
	return c.b.convs.SetData(c.b.id, key, value)
}

func (c *Controls) To(state string) error {
	return c.b.convs.SetState(c.b.id, state)
}

func (c *Controls) End() error {
	return c.b.convs.End(c.b.id)
}

func (c *Controls) Bot() *Bot { return c.b }

type ConversationHandler func(msg *Message, controls *Controls)

// ffjson: skip
type Conversation struct {
	mp map[string]ConversationHandler
}

func NewConversation() *Conversation {
	return &Conversation{make(map[string]ConversationHandler)}
}

func (s *Conversation) On(state string, handler ConversationHandler) error {
	if handler == nil {
		return ErrNilHandler
	}

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
