package web

import (
	"sort"
	"sync"
)

type ControllerStore interface {
	Add(BotID) error
	Get(BotID) (ItemStore, error)
	Remove(BotID) error
}

type MemoryControllerStore struct {
	stores map[BotID]ItemStore
}

func NewMemoryControllerStore() *MemoryControllerStore {
	return &MemoryControllerStore{
		stores: make(map[BotID]ItemStore),
	}
}

func (s *MemoryControllerStore) Add(id BotID) error {
	_, ok := s.stores[id]
	if ok {
		return ErrBotAlreadyAdded
	}

	s.stores[id] = NewMemoryItemStore()
	return nil
}

func (s *MemoryControllerStore) Get(id BotID) (ItemStore, error) {
	is, ok := s.stores[id]
	if !ok {
		return nil, ErrBotNotFound
	}

	return is, nil
}

func (s *MemoryControllerStore) Remove(id BotID) error {
	if _, ok := s.stores[id]; !ok {
		return ErrBotNotFound
	}

	delete(s.stores, id)
	return nil
}

var _ ControllerStore = &MemoryControllerStore{}

type ItemStore interface {
	// Add a message to a thread, defaults to 0
	// Once added, the message should have cursors set
	Add(Item) error

	// Get gets a specific item in a specific thread
	Get(ItemID, ItemID) (Item, error)

	// Update updates an existing item
	// returns an error if no item with the id was present
	Update(Item) error
}

type MemoryItemStore struct {
	mu       sync.Mutex
	messages map[ItemID]*ItemSet
}

func NewMemoryItemStore() *MemoryItemStore {
	return &MemoryItemStore{
		messages: make(map[ItemID]*ItemSet),
	}
}

func (store *MemoryItemStore) Add(item Item) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if item == nil {
		return ErrInvalidItem
	}

	if item.ItemType() == ThreadItemType && item.ItemID() == ItemID(0) {
		return ErrCannotAddThreadZero
	}

	_, ok := store.messages[item.ThreadItemID()]
	if !ok {
		store.messages[item.ThreadItemID()] = NewItemSet()
	}

	store.messages[item.ThreadItemID()].Add(item)

	return nil
}

func (store *MemoryItemStore) Update(item Item) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if item == nil {
		return ErrInvalidItem
	}

	return store.messages[item.ThreadItemID()].Set(item)
}

func (store *MemoryItemStore) Get(id, thread ItemID) (Item, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	s, ok := store.messages[thread]
	if !ok {
		return nil, ErrThreadNotFound
	}

	return s.Get(id)
}

func (store *MemoryItemStore) All(thread ItemID) ([]Item, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	s, ok := store.messages[thread]
	if !ok {
		return nil, ErrThreadNotFound
	}

	ans := make([]Item, 0, s.Len())

	for _, v := range s.items {
		ans = append(ans, v)
	}

	return ans, nil
}

var _ ItemStore = &MemoryItemStore{}

type ItemSet struct {
	keys  []ItemID
	items map[ItemID]Item
}

func NewItemSet() *ItemSet {
	return &ItemSet{
		items: make(map[ItemID]Item),
	}
}

// This is a highly inefficient implementation that pushes an item at the back and then sorts the whole
// thing to get it to the right position. For n messages, the worst case time will be O(n^2).
//
// TODO: consider reducing to O(nlog(n))
func (set *ItemSet) Add(item Item) {
	set.keys = append(set.keys, item.ItemID())
	set.items[item.ItemID()] = item

	sort.Slice(set.keys, func(i, j int) bool { return set.keys[i] < set.keys[j] })
	index := sort.Search(len(set.keys), func(i int) bool { return set.keys[i] >= item.ItemID() })

	var pItem, nItem Item

	if index > 0 {
		pItem = set.items[set.keys[index-1]]

		switch pi := pItem.(type) {
		case *Message:
			pi.Next = &Cursor{item.ItemType(), item.ItemID()}
		case *Thread:
			pi.Next = &Cursor{item.ItemType(), item.ItemID()}
		}
	}

	if index < len(set.keys)-1 {
		nItem = set.items[set.keys[index+1]]

		switch ni := nItem.(type) {
		case *Message:
			ni.Prev = &Cursor{item.ItemType(), item.ItemID()}
		case *Thread:
			ni.Prev = &Cursor{item.ItemType(), item.ItemID()}
		}
	}

	switch i := item.(type) {
	case *Message:
		if index > 0 {
			i.Prev = &Cursor{pItem.ItemType(), pItem.ItemID()}
		}

		if index < len(set.keys)-1 {
			i.Next = &Cursor{nItem.ItemType(), nItem.ItemID()}
		}
	case *Thread:
		if index > 0 {
			i.Prev = &Cursor{pItem.ItemType(), pItem.ItemID()}
		}

		if index < len(set.keys)-1 {
			i.Next = &Cursor{nItem.ItemType(), nItem.ItemID()}
		}
	}
}

func (set *ItemSet) Get(id ItemID) (Item, error) {
	i, ok := set.items[id]
	if !ok {
		return nil, ErrItemNotFound
	}

	return i, nil
}

func (set *ItemSet) Set(i Item) error {
	_, ok := set.items[i.ItemID()]
	if !ok {
		return ErrItemNotFound
	}

	set.items[i.ItemID()] = i
	return nil
}

func (set *ItemSet) Len() int {
	return len(set.keys)
}

type ConversationStore interface {
	Start(bot BotID, id string) error
	IsActive(bot BotID) bool
	Active(bot BotID) (id, state string, err error)
	SetState(bot BotID, state string) error
	SetData(bot BotID, key, value string) error
	GetData(bot BotID, key string) (string, error)
	End(bot BotID) error
}

type convdata struct {
	id, state string
}

// MemoryConversationStore implements a highly inefficient in memory cache for
// storing conversation data
//
// ffjson: skip
type MemoryConversationStore struct {
	active map[BotID]*convdata
	data   map[BotID]map[string]string
}

func NewMemoryConversationStore() *MemoryConversationStore {
	return &MemoryConversationStore{make(map[BotID]*convdata), make(map[BotID]map[string]string)}
}

func (s *MemoryConversationStore) Start(bot BotID, id string) error {
	if _, ok := s.active[bot]; ok {
		return ErrConversationExists
	}

	s.active[bot] = &convdata{id, "start"}
	s.data[bot] = make(map[string]string)
	return nil
}

func (s *MemoryConversationStore) IsActive(bot BotID) bool {
	_, ok := s.active[bot]
	return ok
}

func (s *MemoryConversationStore) Active(bot BotID) (id, state string, err error) {
	c, ok := s.active[bot]

	if !ok {
		err = ErrConversationNotFound
		return
	}

	id, state = c.id, c.state
	return
}

func (s *MemoryConversationStore) SetState(bot BotID, state string) error {
	c, ok := s.active[bot]

	if !ok {
		return ErrConversationNotFound
	}

	if state != c.state {
		c.state = state
	}

	return nil
}

func (s *MemoryConversationStore) SetData(bot BotID, key, value string) error {
	d, ok := s.data[bot]

	if !ok {
		return ErrConversationNotFound
	}

	d[key] = value
	return nil
}

func (s *MemoryConversationStore) GetData(bot BotID, key string) (string, error) {
	d, ok := s.data[bot]

	if !ok {
		return "", ErrConversationNotFound
	}

	ans, ok := d[key]
	if !ok {
		return "", ErrItemNotFound
	}

	return ans, nil
}

func (s *MemoryConversationStore) End(bot BotID) error {
	if _, ok := s.active[bot]; !ok {
		return ErrConversationNotFound
	}

	delete(s.active, bot)
	delete(s.data, bot)

	return nil
}

var _ ConversationStore = &MemoryConversationStore{}
