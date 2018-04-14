package web

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
)

// Bot is essentially managing a single WebSocket connection for a controller.
// It can also start conversations registered with a Controller.
type Bot struct {
	id   BotID
	conn *websocket.Conn

	sanitizer *bluemonday.Policy

	conversations ConversationRegistry

	is    ItemStore
	convs ConversationStore

	incomingMessages chan *MessagePair
	outgoingMessages chan Item
	remover          sync.Once

	remove     func()
	errHandler ErrorHandler
}

// newBot creates a new Bot
func newBot(id BotID, conn *websocket.Conn, sanitizer *bluemonday.Policy, incomingMessages chan *MessagePair, is ItemStore, conversations ConversationRegistry, convs ConversationStore, errhandler ErrorHandler) *Bot {
	bot := &Bot{
		id:   id,
		conn: conn,

		sanitizer: sanitizer,

		is:    is,
		convs: convs,

		conversations: conversations,

		incomingMessages: incomingMessages,
		outgoingMessages: make(chan Item),

		errHandler: errhandler,
	}

	return bot
}

// ID returns the current bot's ID
func (bot *Bot) ID() BotID { return bot.id }

// read starts a read loop over the WebSocket connection that stops if and only if
// the connection itself closes.
func (bot *Bot) read() {
	defer bot.remover.Do(bot.remove)

	for {
		_, rawmsg, err := bot.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				bot.errHandler(errors.Wrap(err, "WebSocket Unexpected Close"))
			}
			break
		}

		msg := &Message{}
		if err := json.Unmarshal(rawmsg, msg); err != nil {
			bot.errHandler(errors.Wrap(err, "read failed"))
		}

		msg.ID = ItemID(time.Now().UnixNano())
		bot.handleMessage(msg)
	}
}

// handleMessage handles new messages from clients
func (bot *Bot) handleMessage(msg *Message) {
	msg.Text = bot.sanitizer.Sanitize(msg.Text)

	go bot.say(msg.clone())

	if id, state, err := bot.convs.Active(bot.id); err == nil {
		conv, err := bot.conversations.Get(id)
		if err != nil {
			bot.errHandler(err)
			return
		}

		conv.mp[state](msg, &Controls{bot})
	} else {
		go func() { bot.incomingMessages <- &MessagePair{msg, bot} }()
	}
}

// write implements a write loop that writes whenever something is available in the
// outgoingMessages channel.
func (bot *Bot) write() {
	defer bot.remover.Do(bot.remove)

	for msg := range bot.outgoingMessages {
		rawmsg, err := json.Marshal(msg)

		if err != nil {
			bot.errHandler(errors.Wrap(err, "Could Not marshal"))
		}

		if err := bot.conn.WriteMessage(websocket.BinaryMessage, rawmsg); err != nil {
			bot.errHandler(errors.Wrap(err, "Could not create writer"))
		}
	}
}

// send directly adds a new message to the outgoingChannel
func (bot *Bot) send(i Item) {
	bot.outgoingMessages <- i
}

// Send adds an item to be sent without setting ID or cursors. This is typically
// used to send old messages from history that already have ID and cursors set, or cached items.
func (bot *Bot) Send(item Item) error {
	switch i := item.(type) {
	case *Message:
		if i.ID == 0 {
			return ErrIDNotSet
		}

		if i.Source == "" {
			return ErrSourceNotSet
		}

		if i.Type == "" {
			return ErrTypeNotSet
		}

	case *Thread:
		if i.ID == 0 {
			return ErrIDNotSet
		}

		if i.Type == "" {
			return ErrTypeNotSet
		}
	}

	go bot.send(item)
	return nil
}

// say sends an item after adding it to the Bot's ItemStore
func (bot *Bot) say(i Item) {
	if err := bot.is.Add(i); err != nil {
		bot.errHandler(errors.Wrap(err, "Could Not Add"))
	}

	bot.send(i)
}

// Say directly sends a message to the client.
func (bot *Bot) Say(msg *Message) {
	msg.Source = BotItemSource
	msg.Type = MessageItemType
	msg.ID = ItemID(time.Now().UnixNano())
	go bot.say(msg)
}

// Reply replies to a client message, the outgoing message has the ID of the message being
// replied to set.
func (bot *Bot) Reply(original, reply *Message) {
	reply.Prev, reply.Next, reply.Source, reply.Type = nil, nil, BotItemSource, MessageItemType
	reply.ReplyID = original.ID
	reply.ID = ItemID(time.Now().UnixNano())

	if original.ThreadID != 0 {
		reply.ThreadID = original.ThreadID
	}

	go bot.say(reply)
}

// ReplyInThread replies to a client message in the same thread or creates a new thread, if the
// original message wasn't in a thread.
func (bot *Bot) ReplyInThread(original, reply *Message) {
	reply.Prev, reply.Next, reply.Source, reply.Type = nil, nil, BotItemSource, MessageItemType
	reply.ReplyID = original.ID
	reply.ID = ItemID(time.Now().UnixNano())

	if original.ThreadID != 0 {
		reply.ThreadID = original.ThreadID
		go bot.say(reply)
	} else {
		t := newThread(original)
		reply.ThreadID = t.ID
		reply.Source = BotItemSource

		go func() {
			// NOTE: this needs to be sequential as if reply goes before the thread,
			// it'll try to add a message in a thread that doesn't exist
			bot.say(t)
			bot.say(reply)
		}()
	}
}

// Update updates an existing message already sent to the client.
func (bot *Bot) Update(msg *Message) {
	msg.Prev, msg.Next, msg.Source, msg.Type = nil, nil, BotItemSource, UpdateItemType
	go bot.say(msg)
}

// StartConversation starts a conversation with the given name with the client.
func (bot *Bot) StartConversation(name string) error {
	c, ok := bot.conversations[name]
	if !ok {
		return ErrConversationNotFound
	}

	if bot.convs.IsActive(bot.id) {
		return ErrConversationAlreadyActive
	}

	if err := bot.convs.Start(bot.id, name); err != nil {
		errors.Wrap(err, "Could not start")
	}

	c.mp["start"](&Message{}, &Controls{bot})
	return nil
}
