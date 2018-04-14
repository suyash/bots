package web

// Controls is a data structure passed to conversation handlers to control the current
// conversational flow, change states, store data, or end the conversation
type Controls struct {
	b *Bot
}

// Get gets the value for a key in the current conversation
func (c *Controls) Get(key string) (string, error) {
	return c.b.convs.GetData(c.b.id, key)
}

// Set sets the value for a key in the current conversation
func (c *Controls) Set(key string, value string) error {
	return c.b.convs.SetData(c.b.id, key, value)
}

// To makes a state transition
func (c *Controls) To(state string) error {
	return c.b.convs.SetState(c.b.id, state)
}

// End ends the conversation
func (c *Controls) End() error {
	return c.b.convs.End(c.b.id)
}

// Bot gets the bot in the current conversation
func (c *Controls) Bot() *Bot { return c.b }

// ConversationHandler is the callback that is invoked for each state in a conversation
type ConversationHandler func(msg *Message, controls *Controls)

// Conversation stores a set of states and ConversationHandler mappings
type Conversation struct {
	mp map[string]ConversationHandler
}

// NewConversation creates a new conversation
func NewConversation() *Conversation {
	return &Conversation{make(map[string]ConversationHandler)}
}

// On adds a new state to the conversation
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

// ConversationRegistry stores a mapping of string ids and conversations
type ConversationRegistry map[string]*Conversation

// NewConversationRegistry creates a new registry
func NewConversationRegistry() ConversationRegistry {
	return make(ConversationRegistry)
}

// Add adds a new conversation with the specified name
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

// Get gets the conversation with the specified name
func (c ConversationRegistry) Get(name string) (*Conversation, error) {
	conv, ok := c[name]

	if !ok {
		return nil, ErrConversationNotFound
	}

	return conv, nil
}
