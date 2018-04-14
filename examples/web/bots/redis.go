package bots

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"suy.io/bots/web"
	"suy.io/bots/web/contrib/redis"
)

type RedisBots struct {
	c     *web.Controller
	store *redis.RedisControllerStore
	t     *template.Template
	pc    PageContext
}

func NewRedisBots(redisHost string, t *template.Template, pc PageContext) (*RedisBots, error) {
	store, err := redis.NewRedisControllerStore(redisHost)
	if err != nil {
		return nil, err
	}

	c, err := web.NewController(
		web.WithControllerStore(store),
		web.WithBotIDCreator(func(req *http.Request) (web.BotID, error) {
			c, err := req.Cookie("botID")
			if err != nil {
				return 0, err
			}

			v, err := strconv.ParseInt(c.Value, 10, 64)
			if err != nil {
				log.Fatal(errors.Wrap(err, "Could not create BotID"))
			}

			return web.BotID(v), nil
		}),
	)

	if err != nil {
		return nil, err
	}

	b := &RedisBots{c, store, t, pc}
	go b.handleBots()
	go b.handleMessages()

	return b, nil
}

func (b *RedisBots) handleBots() {
	for bot := range b.c.BotAdded() {
		log.Println("redis: New Bot Added")
		go b.sendHistory(bot, 10)
	}
}

func (b *RedisBots) sendHistory(bot *web.Bot, n int64) {
	s, err := b.store.Get(bot.ID())
	if err != nil {
		log.Println(errors.Wrap(err, "Could not get bot item store"))
		return
	}

	store := s.(*redis.RedisItemStore)

	items, err := store.Last(n, 0)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("got items", items)
	for _, item := range items {
		log.Println("Sending saved item", item)
		bot.Send(item)
	}
}

func (b *RedisBots) handleMessages() {
	for msg := range b.c.DirectMessages() {
		log.Println("redis: Got", msg.Message)
		msg.Reply(msg.Message)
	}
}

func (b *RedisBots) ConnectionHandler() http.HandlerFunc {
	return b.c.ConnectionHandler()
}

func (b *RedisBots) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	_, err := req.Cookie("botID")
	if err != nil {
		nid := time.Now().UnixNano()
		http.SetCookie(res, &http.Cookie{
			Name:  "botID",
			Value: strconv.FormatInt(nid, 10),
		})
	}

	b.t.ExecuteTemplate(res, Template, b.pc)
}
