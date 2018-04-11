package slack

type BotStore interface {
	AddBot(*OAuthPayload) error
	GetBot(string) (*OAuthPayload, error)
	RemoveBot(string) error
	AllBots() ([]*OAuthPayload, error)
}

// ffjson: skip
type MemoryBotStore struct {
	bots map[string]*OAuthPayload
}

func NewMemoryBotStore() *MemoryBotStore {
	return &MemoryBotStore{
		bots: make(map[string]*OAuthPayload),
	}
}

func (bs *MemoryBotStore) AddBot(p *OAuthPayload) error {
	if _, ok := bs.bots[p.Team]; ok {
		return ErrBotAlreadyAdded
	}

	bs.bots[p.Team] = p
	return nil
}

func (bs *MemoryBotStore) GetBot(team string) (*OAuthPayload, error) {
	b, ok := bs.bots[team]
	if !ok {
		return nil, ErrBotNotFound
	}

	return b, nil
}

func (bs *MemoryBotStore) RemoveBot(team string) error {
	if _, ok := bs.bots[team]; !ok {
		return ErrBotNotFound
	}

	delete(bs.bots, team)
	return nil
}

func (bs *MemoryBotStore) AllBots() ([]*OAuthPayload, error) {
	bots := make([]*OAuthPayload, 0, len(bs.bots))
	for _, b := range bs.bots {
		bots = append(bots, b)
	}

	return bots, nil
}

var _ BotStore = &MemoryBotStore{}

type ConversationStore interface {
	Start(user, channel, team, id string) error
	IsActive(user, channel, team string) bool
	Active(user, channel, team string) (id, state string, err error)
	SetState(user, channel, team, state string) error
	SetData(user, channel, team, key, value string) error
	GetData(user, channel, team, key string) (string, error)
	End(user, channel, team string) error
}

type convdata struct {
	id, state string
}

//
// ffjson: skip
type MemoryConversationStore struct {
	active map[string]*convdata
	data   map[string]map[string]string
}

func NewMemoryConversationStore() *MemoryConversationStore {
	return &MemoryConversationStore{make(map[string]*convdata), make(map[string]map[string]string)}
}

func (s *MemoryConversationStore) Start(user, channel, team, id string) error {
	i := user + "_" + channel + "_" + team
	if _, ok := s.active[i]; ok {
		return ErrConversationExists
	}

	s.active[i] = &convdata{id, "start"}
	s.data[i] = make(map[string]string)
	return nil
}

func (s *MemoryConversationStore) IsActive(user, channel, team string) bool {
	i := user + "_" + channel + "_" + team
	_, ok := s.active[i]
	return ok
}

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

func (s *MemoryConversationStore) SetState(user, channel, team, state string) error {
	i := user + "_" + channel + "_" + team
	c, ok := s.active[i]

	if !ok {
		return ErrConversationNotFound
	}

	c.state = state
	return nil
}

func (s *MemoryConversationStore) SetData(user, channel, team, key, value string) error {
	i := user + "_" + channel + "_" + team
	d, ok := s.data[i]

	if !ok {
		return ErrConversationNotFound
	}

	d[key] = value
	return nil
}

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
