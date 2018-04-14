package slack

import "suy.io/bots/slack/api/oauth"

// BotStore is an interface used to store bot data.
type BotStore interface {
	// AddBot adds a new bot
	AddBot(*oauth.AccessResponse) error

	// GetBot gets the bot payload for a team id
	GetBot(string) (*oauth.AccessResponse, error)

	// RemoveBot removes a bot given the team ID
	RemoveBot(string) error

	// AllBots gets all the bots stored ever
	AllBots() ([]*oauth.AccessResponse, error)
}

// MemoryBotStore is an in-memory BotStore implementation.
//
// ffjson: skip
type MemoryBotStore struct {
	bots map[string]*oauth.AccessResponse
}

// NewMemoryBotStore creates a new MemoryBotStore object.
func NewMemoryBotStore() *MemoryBotStore {
	return &MemoryBotStore{
		bots: make(map[string]*oauth.AccessResponse),
	}
}

// AddBot adds a new paylaod to the store.
func (bs *MemoryBotStore) AddBot(p *oauth.AccessResponse) error {
	if _, ok := bs.bots[p.TeamID]; ok {
		return ErrBotAlreadyAdded
	}

	bs.bots[p.TeamID] = p
	return nil
}

// GetBot gets a stored payload given a team.
func (bs *MemoryBotStore) GetBot(team string) (*oauth.AccessResponse, error) {
	b, ok := bs.bots[team]
	if !ok {
		return nil, ErrBotNotFound
	}

	return b, nil
}

// RemoveBot removes a stored payload given the team.
func (bs *MemoryBotStore) RemoveBot(team string) error {
	if _, ok := bs.bots[team]; !ok {
		return ErrBotNotFound
	}

	delete(bs.bots, team)
	return nil
}

// AllBots gets all stored payloads.
func (bs *MemoryBotStore) AllBots() ([]*oauth.AccessResponse, error) {
	bots := make([]*oauth.AccessResponse, 0, len(bs.bots))
	for _, b := range bs.bots {
		bots = append(bots, b)
	}

	return bots, nil
}

var _ BotStore = &MemoryBotStore{}

// ConversationStore defines the interface for storing conversation data.
type ConversationStore interface {
	// Start starts a new conversation given a user, channel and team
	Start(user, channel, team, id string) error

	// IsActive checks if a conversation is active
	IsActive(user, channel, team string) bool

	// Active returns the active conversation id and state
	Active(user, channel, team string) (id, state string, err error)

	// SetState sets the state for the current conversation
	SetState(user, channel, team, state string) error

	// SetData sets a key-value pair for the current conversation
	SetData(user, channel, team, key, value string) error

	// GetData gets the value stored for a key for the current conversation
	GetData(user, channel, team, key string) (string, error)

	// End ends the current conversation.
	End(user, channel, team string) error
}

type convdata struct {
	id, state string
}

// MemoryConversationStore is an in-memory implementation of ConversationStore.
//
// ffjson: skip
type MemoryConversationStore struct {
	active map[string]*convdata
	data   map[string]map[string]string
}

// NewMemoryConversationStore creates a new MemoryConversationStore object.
func NewMemoryConversationStore() *MemoryConversationStore {
	return &MemoryConversationStore{make(map[string]*convdata), make(map[string]map[string]string)}
}

// Start starts a conversation
func (s *MemoryConversationStore) Start(user, channel, team, id string) error {
	i := user + "_" + channel + "_" + team
	if _, ok := s.active[i]; ok {
		return ErrConversationExists
	}

	s.active[i] = &convdata{id, "start"}
	s.data[i] = make(map[string]string)
	return nil
}

// IsActive checks if a conversation is active.
func (s *MemoryConversationStore) IsActive(user, channel, team string) bool {
	i := user + "_" + channel + "_" + team
	_, ok := s.active[i]
	return ok
}

// Active gets the active conversation.
func (s *MemoryConversationStore) Active(user, channel, team string) (id, state string, err error) {
	i := user + "_" + channel + "_" + team
	c, ok := s.active[i]

	if !ok {
		err = ErrConversationNotFound
		return
	}

	id, state = c.id, c.state
	return
}

// SetState sets the state for the active conversation.
func (s *MemoryConversationStore) SetState(user, channel, team, state string) error {
	i := user + "_" + channel + "_" + team
	c, ok := s.active[i]

	if !ok {
		return ErrConversationNotFound
	}

	c.state = state
	return nil
}

// SetData sets the value for a key in the current conversation.
func (s *MemoryConversationStore) SetData(user, channel, team, key, value string) error {
	i := user + "_" + channel + "_" + team
	d, ok := s.data[i]

	if !ok {
		return ErrConversationNotFound
	}

	d[key] = value
	return nil
}

// GetData gets the stored value for a key for a conversation.
func (s *MemoryConversationStore) GetData(user, channel, team, key string) (string, error) {
	i := user + "_" + channel + "_" + team
	d, ok := s.data[i]

	if !ok {
		return "", ErrConversationNotFound
	}

	ans, ok := d[key]
	if !ok {
		return "", ErrItemNotFound
	}

	return ans, nil
}

// End ends the conversation.
func (s *MemoryConversationStore) End(user, channel, team string) error {
	i := user + "_" + channel + "_" + team
	if _, ok := s.active[i]; !ok {
		return ErrConversationNotFound
	}

	delete(s.active, i)
	delete(s.data, i)

	return nil
}

var _ ConversationStore = &MemoryConversationStore{}
