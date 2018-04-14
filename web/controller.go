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

// BotIDCreator is a callback that gets a http request before it is hijacked and upgraded
// to WebSocket. Its sole job is to identify the bot ID from the request and return it.
type BotIDCreator func(*http.Request) (BotID, error)

// defaultBotIDCreator is used to generate ids by default, just generates a random integer
// for each connection.
func defaultBotIDCreator(*http.Request) (BotID, error) {
	return BotID(rand.Int63()), nil // ¯\_(ツ)_/¯
}

// ErrorHandler is a function that can be specified to intercept different errors occuring
// during operation that are not a result of a user action
type ErrorHandler func(err error)

// defaultErrorHandler is the default function used if nothing is specified,
// it simply logs and exits
func defaultErrorHandler(err error) {
	log.Fatal(err)
}

// Controller is the main interface to manage and control bots at a particular endpoint.
// Essentially, you create a Controller object, and set the ConnectionHandler as the request
// handler for a particular endpoint, and clients connecting to that endpoint will be managed
// by this object.
type Controller struct {
	botAdded  chan *Bot
	sanitizer *bluemonday.Policy

	cs ControllerStore

	conversations ConversationRegistry
	convs         ConversationStore

	messages chan *MessagePair

	idCreator  BotIDCreator
	errHandler ErrorHandler
}

// NewController creates a new Controller object. It can take options for specifying
// the ControllerStore, ConversationStore, BotIDCreator and ErrorHandler.
func NewController(options ...func(*Controller) error) (*Controller, error) {
	c := &Controller{
		sanitizer:     bluemonday.UGCPolicy(),
		botAdded:      make(chan *Bot),
		messages:      make(chan *MessagePair),
		conversations: NewConversationRegistry(),
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

// WithControllerStore can be passed as an option to NewController with the
// desired implementation of ControllerStore.
func WithControllerStore(store ControllerStore) func(*Controller) error {
	return func(c *Controller) error {
		if store == nil {
			return ErrNilControllerStore
		}

		c.cs = store
		return nil
	}
}

// WithConversationStore can be passed as an option to NewController with the
// desired implementation of ConversationStore.
func WithConversationStore(store ConversationStore) func(*Controller) error {
	return func(c *Controller) error {
		if store == nil {
			return ErrNilConversationStore
		}

		c.convs = store
		return nil
	}
}

// WithBotIDCreator can be passed as an option to NewController with the
// desired implementation of BotIDCreator.
func WithBotIDCreator(f BotIDCreator) func(*Controller) error {
	return func(c *Controller) error {
		if f == nil {
			return ErrNilIDCreator
		}

		c.idCreator = f
		return nil
	}
}

// WithErrorHandler can be passed as an option to NewController with the
// desired implementation of ErrorHandler.
func WithErrorHandler(f ErrorHandler) func(*Controller) error {
	return func(c *Controller) error {
		if f == nil {
			return ErrNilErrorHandler
		}

		c.errHandler = f
		return nil
	}
}

// ConnectionHandler returns a http.HandlerFunc that can be used to intercept
// and handle connections from clients. It essentially creates a new Bot for each connection,
// and upgrades the connection to WebSocket.
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
			// TODO: maybe use errhandler also here, the error may be informative
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

// addBot adds a new Bot from a *websocket.Conn
func (c *Controller) addBot(botID BotID, conn *websocket.Conn, req *http.Request) (*Bot, error) {
	if err := c.cs.Add(botID); err != nil {
		return nil, errors.Wrap(err, "Could not Add bot")
	}

	store, err := c.cs.Get(botID)
	if err != nil {
		return nil, errors.Wrap(err, "Could not Get bot")
	}

	bot := newBot(botID, conn, c.sanitizer, c.messages, store, c.conversations, c.convs, c.errHandler)

	bot.remove = func() {
		if err := c.removeBot(bot); err != nil {
			c.errHandler(err)
		}
	}

	go func() { c.botAdded <- bot }()
	return bot, nil
}

// removeBot removes a bot from a closed webSocket.Conn
func (c *Controller) removeBot(bot *Bot) error {
	close(bot.outgoingMessages)

	err1, err2 := c.cs.Remove(bot.id), bot.conn.Close()
	bot.conn = nil

	if err2 != nil {
		return err2
	}

	return err1
}

// BotAdded gets a receive only channel that'll get a bot payload whenever a new connection
// is obtained. You can use this to send messages like "Hi, I'm Online" to new users.
func (c *Controller) BotAdded() <-chan *Bot {
	return c.botAdded
}

// Messages returns a receive only channel that will get a bot-message pair every time a new message
// comes over.
func (c *Controller) Messages() <-chan *MessagePair {
	return c.messages
}

// RegisterConversation registers a new Conversation with the Controller, so it can
// be used with bot.StartConversation(name)
func (c *Controller) RegisterConversation(name string, conv *Conversation) error {
	return c.conversations.Add(name, conv)
}
