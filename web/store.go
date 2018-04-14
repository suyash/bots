package web

import (
	"sort"
	"sync"
)

// ControllerStore defines the interface for storing all the bots connected to the app.
// By default an [in memory implementation](https://godoc.org/suy.io/bots/web#MemoryControllerStore)
// is used. Implement your own and pass it inside [WithControllerStore](https://godoc.org/suy.io/bots/web#WithControllerStore)
// when initializing a Controller.
type ControllerStore interface {
	// Add is supposed to add a new BotID and initialize the ItemStore for the bot.
	// Ideally returns an error if a bot with same ID is already there, or if ItemStore could not be initialized.
	Add(BotID) error

	// Get gets the store for a bot with id BotID, error if no such bot.
	Get(BotID) (ItemStore, error)

	// Remove will remove a bot with specified ID, error if no such bot.
	// Also clears the ItemStore associated with the Bot.
	Remove(BotID) error
}

// MemoryControllerStore is an in-memory ControllerStore implementation.
type MemoryControllerStore struct {
	stores map[BotID]ItemStore
}

// NewMemoryControllerStore creates a new MemoryControllerStore.
func NewMemoryControllerStore() *MemoryControllerStore {
	return &MemoryControllerStore{
		stores: make(map[BotID]ItemStore),
	}
}

// Add adds a new BotID to the store, and initializes ItemStore for it.
func (s *MemoryControllerStore) Add(id BotID) error {
	_, ok := s.stores[id]
	if ok {
		return ErrBotAlreadyAdded
	}

	s.stores[id] = NewMemoryItemStore()
	return nil
}

// Get gets a new BotID to the store
func (s *MemoryControllerStore) Get(id BotID) (ItemStore, error) {
	is, ok := s.stores[id]
	if !ok {
		return nil, ErrBotNotFound
	}

	return is, nil
}

// Remove removes a bot from the store.
func (s *MemoryControllerStore) Remove(id BotID) error {
	if _, ok := s.stores[id]; !ok {
		return ErrBotNotFound
	}

	delete(s.stores, id)
	return nil
}

var _ ControllerStore = &MemoryControllerStore{}

// ItemStore stores data associated with a single bot.
type ItemStore interface {
	// Add a message to a thread, defaults to 0.
	// Once added, the message should have cursors set.
	// You can't add a thread with ID zero, that is reserved to
	// represent messages in global namespace.
	Add(Item) error

	// Get gets a specific item in a specific thread
	Get(ItemID, ItemID) (Item, error)

	// Update updates an existing item
	// returns an error if no item with the id was present
	Update(Item) error
}

// MemoryItemStore provides a highly inefficient, pretty much incompetent in-memory implementation
// for an ItemStore.
type MemoryItemStore struct {
	mu       sync.Mutex
	messages map[ItemID]*itemSet
}

// NewMemoryItemStore creates a new MemoryItemStore
func NewMemoryItemStore() *MemoryItemStore {
	return &MemoryItemStore{
		messages: make(map[ItemID]*itemSet),
	}
}

// Add adds a new item to the current store
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
		store.messages[item.ThreadItemID()] = newItemSet()
	}

	store.messages[item.ThreadItemID()].Add(item)

	return nil
}

// Update updates an already stored item's data.
func (store *MemoryItemStore) Update(item Item) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if item == nil {
		return ErrInvalidItem
	}

	return store.messages[item.ThreadItemID()].Set(item)
}

// Get gets a specific item in a specific thread.
func (store *MemoryItemStore) Get(id, thread ItemID) (Item, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	s, ok := store.messages[thread]
	if !ok {
		return nil, ErrThreadNotFound
	}

	return s.Get(id)
}

// All gets all items in a particular thread in this store.
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

// itemSet stores items in sorted order and is used to set cursors (prev, next) for added messages.
type itemSet struct {
	keys  []ItemID
	items map[ItemID]Item
}

// newItemSet creates a new ItemSet.
func newItemSet() *itemSet {
	return &itemSet{
		items: make(map[ItemID]Item),
	}
}

// Add adds a new message and sets prev and next for it based on timestamps.
//
// This is a highly inefficient implementation that pushes an item at the back and then sorts the whole
// thing to get it to the right position. For n messages, the worst case time will be O(n^2).
//
// TODO: consider reducing to O(nlog(n))
func (set *itemSet) Add(item Item) {
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

// Get gets item with the given id
func (set *itemSet) Get(id ItemID) (Item, error) {
	i, ok := set.items[id]
	if !ok {
		return nil, ErrItemNotFound
	}

	return i, nil
}

// Set adds item to the set.
func (set *itemSet) Set(i Item) error {
	_, ok := set.items[i.ItemID()]
	if !ok {
		return ErrItemNotFound
	}

	set.items[i.ItemID()] = i
	return nil
}

// Len gets total number of items in the set.
func (set *itemSet) Len() int {
	return len(set.keys)
}

// ConversationStore stores current active conversation data for bots, as well as state data.
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

// MemoryConversationStore implements an in memory ConversationStore.
type MemoryConversationStore struct {
	active map[BotID]*convdata
	data   map[BotID]map[string]string
}

// NewMemoryConversationStore creates a new MemoryConversationStore.
func NewMemoryConversationStore() *MemoryConversationStore {
	return &MemoryConversationStore{make(map[BotID]*convdata), make(map[BotID]map[string]string)}
}

// Start starts a conversation of specified ID with the specified Bot.
func (s *MemoryConversationStore) Start(bot BotID, id string) error {
	if _, ok := s.active[bot]; ok {
		return ErrConversationExists
	}

	s.active[bot] = &convdata{id, "start"}
	s.data[bot] = make(map[string]string)
	return nil
}

// IsActive returns true of the specified has an active conversation.
func (s *MemoryConversationStore) IsActive(bot BotID) bool {
	_, ok := s.active[bot]
	return ok
}

// Active returns conversation id and state for the bot with specified ID.
func (s *MemoryConversationStore) Active(bot BotID) (id, state string, err error) {
	c, ok := s.active[bot]

	if !ok {
		err = ErrConversationNotFound
		return
	}

	id, state = c.id, c.state
	return
}

// SetState sets the conversation state for the specified ID.
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

// SetData sets a key-value pair for the current conversation.
func (s *MemoryConversationStore) SetData(bot BotID, key, value string) error {
	d, ok := s.data[bot]

	if !ok {
		return ErrConversationNotFound
	}

	d[key] = value
	return nil
}

// GetData gets data with specified key for the current conversation.
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

// End ends the active conversation for the current bot.
func (s *MemoryConversationStore) End(bot BotID) error {
	if _, ok := s.active[bot]; !ok {
		return ErrConversationNotFound
	}

	delete(s.active, bot)
	delete(s.data, bot)

	return nil
}

var _ ConversationStore = &MemoryConversationStore{}
