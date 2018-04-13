package web

import (
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type BotIDCreator func(*http.Request) (BotID, error)

func defaultBotIDCreator(*http.Request) (BotID, error) {
	return BotID(rand.Int63()), nil // ¯\_(ツ)_/¯
}

type ErrorHandler func(err error)

func defaultErrorHandler(err error) {
	log.Fatal(err)
}

type Controller struct {
	botAdded  chan *Bot
	sanitizer *bluemonday.Policy

	cs ControllerStore

	conversations ConversationRegistry
	convs         ConversationStore

	directMessages chan *MessagePair

	idCreator  BotIDCreator
	errHandler ErrorHandler
}

func NewController(options ...func(*Controller) error) (*Controller, error) {
	c := &Controller{
		sanitizer:      bluemonday.UGCPolicy(),
		botAdded:       make(chan *Bot),
		directMessages: make(chan *MessagePair),
		conversations:  NewConversationRegistry(),
	}

	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, errors.Wrap(err, "NewController Failed")
		}
	}

	if c.cs == nil {
		c.cs = NewMemoryControllerStore()
	}

	if c.convs == nil {
		c.convs = NewMemoryConversationStore()
	}

	if c.idCreator == nil {
		c.idCreator = defaultBotIDCreator
	}

	if c.errHandler == nil {
		c.errHandler = defaultErrorHandler
	}

	return c, nil
}

func WithControllerStore(store ControllerStore) func(*Controller) error {
	return func(c *Controller) error {
		if store == nil {
			return ErrNilControllerStore
		}

		c.cs = store
		return nil
	}
}

func WithConversationStore(store ConversationStore) func(*Controller) error {
	return func(c *Controller) error {
		if store == nil {
			return ErrNilConversationStore
		}

		c.convs = store
		return nil
	}
}

func WithBotIDCreator(f BotIDCreator) func(*Controller) error {
	return func(c *Controller) error {
		if f == nil {
			return ErrNilIDCreator
		}

		c.idCreator = f
		return nil
	}
}

func WithErrorHandler(f ErrorHandler) func(*Controller) error {
	return func(c *Controller) error {
		if f == nil {
			return ErrNilErrorHandler
		}

		c.errHandler = f
		return nil
	}
}

func (c *Controller) ConnectionHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		botID, err := c.idCreator(req)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		conn, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			// NOTE: this is not needed, the upgrader sends http.StatusBadRequest on its own
			// http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		bot, err := c.addBot(botID, conn, req)
		if err != nil {
			// TODO: maybe use errhandler also here, the error may be informative
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		go bot.write()
		go bot.read()

		res.WriteHeader(http.StatusOK)
	}
}

func (c *Controller) addBot(botID BotID, conn *websocket.Conn, req *http.Request) (*Bot, error) {
	if err := c.cs.Add(botID); err != nil {
		return nil, errors.Wrap(err, "Could not Add bot")
	}

	store, err := c.cs.Get(botID)
	if err != nil {
		return nil, errors.Wrap(err, "Could not Get bot")
	}

	bot := newBot(botID, conn, c.sanitizer, c.directMessages, store, c.conversations, c.convs, c.errHandler)

	bot.remove = func() {
		if err := c.removeBot(bot); err != nil {
			c.errHandler(err)
		}
	}

	go func() { c.botAdded <- bot }()
	return bot, nil
}

func (c *Controller) removeBot(bot *Bot) error {
	close(bot.outgoingMessages)

	err1, err2 := c.cs.Remove(bot.id), bot.conn.Close()
	bot.conn = nil

	if err2 != nil {
		return err2
	}

	return err1
}

func (c *Controller) BotAdded() <-chan *Bot {
	return c.botAdded
}

func (c *Controller) DirectMessages() <-chan *MessagePair {
	return c.directMessages
}

func (c *Controller) RegisterConversation(name string, conv *Conversation) error {
	return c.conversations.Add(name, conv)
}
