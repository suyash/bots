package redis // import "suy.io/bots/web/contrib/redis"

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/go-redis/redis"

	"suy.io/bots/web"
)

// TODO: improve pagination capabilities

type RedisControllerStore struct {
	client *redis.Client
	bots   map[web.BotID]*RedisItemStore
}

func NewRedisControllerStore(host string) (*RedisControllerStore, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0,
	})

	if _, err := c.Ping().Result(); err != nil {
		return nil, err
	}

	return &RedisControllerStore{c, make(map[web.BotID]*RedisItemStore)}, nil
}

func (rcs *RedisControllerStore) Add(botID web.BotID) error {
	return rcs.client.SAdd("bots", int64(botID)).Err()
}

func (rcs *RedisControllerStore) Get(botID web.BotID) (web.ItemStore, error) {
	is, ok := rcs.bots[botID]
	if ok {
		return is, nil
	}

	if !rcs.client.SIsMember("bots", int64(botID)).Val() {
		return nil, web.ErrBotNotFound
	}

	is = newRedisItemStore(rcs.client, botID)
	rcs.bots[botID] = is
	return is, nil
}

func (rcs *RedisControllerStore) Remove(botID web.BotID) error {
	delete(rcs.bots, botID)

	if err := rcs.client.SRem("bots", int64(botID)).Err(); err != nil {
		return err
	}

	return nil
}

var _ web.ControllerStore = &RedisControllerStore{}

type RedisItemStore struct {
	mu     sync.Mutex
	client *redis.Client
	botid  web.BotID
}

func newRedisItemStore(c *redis.Client, b web.BotID) *RedisItemStore {
	return &RedisItemStore{
		client: c,
		botid:  b,
	}
}

func (is *RedisItemStore) Add(item web.Item) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	botidstr, threadidstr, itemidstr := strconv.FormatInt(int64(is.botid), 10), strconv.FormatInt(int64(item.ThreadItemID()), 10), strconv.FormatInt(int64(item.ItemID()), 10)

	// key = botID:threadID
	key := botidstr + ":" + threadidstr

	// itemstr = itemType:itemID
	itemstr := string(item.ItemType()) + ":" + itemidstr

	// add itemstr to sorted set at key `key` ranked by itemID
	if err := is.client.ZAdd(key, redis.Z{Score: float64(item.ItemID()), Member: itemstr}).Err(); err != nil {
		return err
	}

	// get the rank of the item that was just added
	r, err := is.client.ZRank(key, itemstr).Result()
	if err != nil {
		return err
	}

	// get the total number of items in the sorted set
	l, err := is.client.ZCard(key).Result()
	if err != nil {
		return err
	}

	// if there is a previous item
	prev, err := is.prev(r, key, threadidstr, item.ItemType(), item.ItemID())
	if err != nil {
		return err
	}

	// if there is a next item
	next, err := is.next(r, l, key, threadidstr, item.ItemType(), item.ItemID())
	if err != nil {
		return err
	}

	switch i := item.(type) {
	case *web.Message:
		i.Prev, i.Next = prev, next
	case *web.Thread:
		i.Prev, i.Next = prev, next
	}

	// save the item to database and return result
	return is.save(item, key)
}

func (is *RedisItemStore) save(item web.Item, keyPrefix string) error {
	s, err := json.Marshal(item)
	if err != nil {
		return err
	}

	// itemKey = botID:threadID:itemID
	itemKey := keyPrefix + ":" + strconv.FormatInt(int64(item.ItemID()), 10)
	log.Println("[REDIS] saving item with key", itemKey)

	// store entire data
	log.Println("[REDIS] adding message", string(s))

	return is.client.Set(itemKey, s, 0).Err()
}

// also sets next of prev
func (is *RedisItemStore) prev(r int64, keyPrefix, threadstr string, typ web.ItemType, id web.ItemID) (*web.Cursor, error) {
	if r == 0 {
		return nil, nil
	}

	ans := &web.Cursor{}

	// get one item before r
	ra, err := is.client.ZRange(keyPrefix, r-1, r).Result()
	if err != nil {
		return nil, err
	}

	// split at ":"
	s := strings.Split(ra[0], ":")

	// set Type
	ans.Type = web.ItemType(s[0])

	cid, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return nil, err
	}

	// set id
	ans.ID = web.ItemID(cid)

	// get prev object
	previ, err := is.get(s[1], threadstr)
	if err != nil {
		return nil, err
	}

	switch i := previ.(type) {
	case *web.Message:
		i.Next = &web.Cursor{Type: typ, ID: id}
	case *web.Thread:
		i.Next = &web.Cursor{Type: typ, ID: id}
	}

	if err := is.save(previ, keyPrefix); err != nil {
		return nil, err
	}

	return ans, nil
}

// also sets prev of next
func (is *RedisItemStore) next(r, l int64, keyPrefix, threadstr string, typ web.ItemType, id web.ItemID) (*web.Cursor, error) {
	if r == l-1 {
		return nil, nil
	}

	ans := &web.Cursor{}

	// get one item after r
	ra, err := is.client.ZRange(keyPrefix, r+1, r+2).Result()
	if err != nil {
		return nil, err
	}

	// split at ":"
	s := strings.Split(ra[0], ":")

	// set Type
	ans.Type = web.ItemType(s[0])

	cid, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return nil, err
	}

	// set Ts
	ans.ID = web.ItemID(cid)

	// get next object
	nexti, err := is.get(s[1], threadstr)
	if err != nil {
		return nil, err
	}

	switch i := nexti.(type) {
	case *web.Message:
		i.Prev = &web.Cursor{Type: typ, ID: id}
	case *web.Thread:
		i.Prev = &web.Cursor{Type: typ, ID: id}
	}

	if err := is.save(nexti, keyPrefix); err != nil {
		return nil, err
	}

	return ans, nil
}

func (is *RedisItemStore) Get(id, threadID web.ItemID) (web.Item, error) {
	return is.get(strconv.FormatInt(int64(id), 10), strconv.FormatInt(int64(threadID), 10))
}

func (is *RedisItemStore) get(id, threadID string) (web.Item, error) {
	key := strconv.FormatInt(int64(is.botid), 10) + ":" + threadID + ":" + id

	res, err := is.client.Get(key).Result()
	if err != nil {
		return nil, err
	}

	return UnmarshalJSONItem([]byte(res))
}

func (is *RedisItemStore) Last(n int64, thread web.ItemID) (items []web.Item, err error) {
	threadstr := strconv.FormatInt(int64(thread), 10)

	// key = botID:threadID
	key := strconv.FormatInt(int64(is.botid), 10) + ":" + threadstr

	r, err := is.client.ZRange(key, -1-n, -1).Result()
	if err != nil {
		return nil, err
	}

	log.Println("[REDIS] last", n, "items  ", r)

	itemc := make(chan web.Item)
	getItem := func(id string) {
		a := strings.Split(id, ":")

		i, e := is.get(a[1], threadstr)
		if e != nil {
			err = e
		}

		itemc <- i
	}

	for _, id := range r {
		go getItem(id)
	}

	items = make([]web.Item, 0, len(r))
	for i := 0; i < len(r); i++ {
		items = append(items, <-itemc)
	}

	close(itemc)
	return
}

func (is *RedisItemStore) Update(item web.Item) error {
	botidstr, threadidstr := strconv.FormatInt(int64(is.botid), 10), strconv.FormatInt(int64(item.ThreadItemID()), 10)
	return is.save(item, botidstr+":"+threadidstr)
}

var _ web.ItemStore = &RedisItemStore{}

type RedisConversationStore struct {
	client *redis.Client
}

func NewRedisConversationStore(host string) (*RedisConversationStore, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0,
	})

	if _, err := c.Ping().Result(); err != nil {
		return nil, err
	}

	return &RedisConversationStore{c}, nil
}

func (cs *RedisConversationStore) Start(botid web.BotID, id string) error {
	if err := cs.client.Set(strconv.Itoa(int(botid)), id, 0).Err(); err != nil {
		return err
	}

	return cs.client.Set(strconv.Itoa(int(botid))+":state", "start", 0).Err()
}

func (cs *RedisConversationStore) IsActive(botid web.BotID) bool {
	return cs.client.Get(strconv.Itoa(int(botid))).Err() == nil
}

func (cs *RedisConversationStore) Active(botid web.BotID) (id, state string, err error) {
	id, err = cs.client.Get(strconv.Itoa(int(botid))).Result()
	if err != nil {
		return "", "", err
	}

	state = cs.client.Get(strconv.Itoa(int(botid)) + ":state").String()
	return
}

func (cs *RedisConversationStore) SetState(botid web.BotID, state string) error {
	return cs.client.Set(strconv.Itoa(int(botid))+":state", state, 0).Err()
}

func (cs *RedisConversationStore) SetData(botid web.BotID, key, value string) error {
	return cs.client.HSet(strconv.Itoa(int(botid))+":data", key, value).Err()
}

func (cs *RedisConversationStore) GetData(botid web.BotID, key string) (string, error) {
	return cs.client.HGet(strconv.Itoa(int(botid))+":data", key).Result()
}

func (cs *RedisConversationStore) End(botid web.BotID) error {
	if err := cs.client.HDel(strconv.Itoa(int(botid)) + ":data").Err(); err != nil {
		return err
	}

	if err := cs.client.Del(strconv.Itoa(int(botid)) + ":state").Err(); err != nil {
		return err
	}

	return cs.client.Del(strconv.Itoa(int(botid))).Err()
}

var _ web.ConversationStore = &RedisConversationStore{}

// UnmarshalJSONItem unmarshals a JSON payload into an item depending on whether the item
// is a message or a thread
func UnmarshalJSONItem(js []byte) (web.Item, error) {
	s := &struct {
		Type web.ItemType `json:"type"`
	}{}

	if err := json.Unmarshal(js, s); err != nil {
		return nil, err
	}

	if s.Type == web.ThreadItemType {
		t := &web.Thread{}
		if err := json.Unmarshal(js, t); err != nil {
			return nil, err
		}

		return t, nil
	} else if s.Type == web.MessageItemType {
		msg := &web.Message{}
		if err := json.Unmarshal(js, msg); err != nil {
			return nil, err
		}

		return msg, nil
	}

	return nil, web.ErrInvalidItem
}
