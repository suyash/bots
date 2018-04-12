package redis

import (
	"encoding/json"
	"log"

	"github.com/go-redis/redis"

	"suy.io/bots/slack"
	"suy.io/bots/slack/api/oauth"
)

type RedisBotStore struct {
	client *redis.Client
}

func NewRedisBotStore(host string) *RedisBotStore {
	c := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0,
	})

	if _, err := c.Ping().Result(); err != nil {
		log.Fatal(err)
	}

	return &RedisBotStore{c}
}

func (bs *RedisBotStore) AddBot(p *oauth.AccessResponse) error {
	log.Println("Adding Bot For team", p.TeamID)

	d, err := json.Marshal(p)
	if err != nil {
		return err
	}

	if err := bs.client.Set(p.TeamID, d, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (bs *RedisBotStore) GetBot(team string) (*oauth.AccessResponse, error) {
	log.Println("Getting Bot For team", team)

	d, err := bs.client.Get(team).Result()
	if err != nil {
		return nil, err
	}

	p := &oauth.AccessResponse{
		Bot: &oauth.Bot{},
	}

	if err := json.Unmarshal([]byte(d), p); err != nil {
		return nil, err
	}

	return p, nil
}

func (bs *RedisBotStore) RemoveBot(team string) error {
	log.Println("Removing Bot For team", team)

	if err := bs.client.Del(team).Err(); err != nil {
		return err
	}

	return nil
}

func (bs *RedisBotStore) AllBots() ([]*oauth.AccessResponse, error) {
	log.Println("Getting All Bots")

	keys, err := bs.client.Keys("*").Result()
	if err != nil {
		return nil, err
	}

	bots := make([]*oauth.AccessResponse, 0, len(keys))
	for _, k := range keys {
		p, err := bs.GetBot(k)
		if err != nil {
			return nil, err
		}

		bots = append(bots, p)
	}

	return bots, nil
}

var _ slack.BotStore = &RedisBotStore{}

type RedisConversationStore struct {
	client *redis.Client
}

func NewRedisConversationStore(host string) *RedisConversationStore {
	c := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0,
	})

	if _, err := c.Ping().Result(); err != nil {
		log.Fatal(err)
	}

	return &RedisConversationStore{c}
}

func (cs *RedisConversationStore) Start(user, channel, team, id string) error {
	if err := cs.client.Set(team+"/"+channel+"/"+user, id, 0).Err(); err != nil {
		return err
	}

	return cs.client.Set(team+"/"+channel+"/"+user+":state", "start", 0).Err()
}

func (cs *RedisConversationStore) IsActive(user, channel, team string) bool {
	return cs.client.Get(team+"/"+channel+"/"+user).Err() == nil
}

func (cs *RedisConversationStore) Active(user, channel, team string) (id, state string, err error) {
	id, err = cs.client.Get(team + "/" + channel + "/" + user).Result()
	if err != nil {
		return "", "", err
	}

	state = cs.client.Get(team + "/" + channel + "/" + user + ":state").String()
	return
}

func (cs *RedisConversationStore) SetState(user, channel, team, state string) error {
	return cs.client.Set(team+"/"+channel+"/"+user+":state", state, 0).Err()
}

func (cs *RedisConversationStore) SetData(user, channel, team, key, value string) error {
	return cs.client.HSet(team+"/"+channel+"/"+user+":data", key, value).Err()
}

func (cs *RedisConversationStore) GetData(user, channel, team, key string) (string, error) {
	return cs.client.HGet(team+"/"+channel+"/"+user+":data", key).Result()
}

func (cs *RedisConversationStore) End(user, channel, team string) error {
	if err := cs.client.HDel(team + "/" + channel + "/" + user + ":data").Err(); err != nil {
		return err
	}

	if err := cs.client.Del(team + "/" + channel + "/" + user + ":state").Err(); err != nil {
		return err
	}

	return cs.client.Del(team + "/" + channel + "/" + user).Err()
}

var _ slack.ConversationStore = &RedisConversationStore{}
