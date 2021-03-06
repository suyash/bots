package connector // import "suy.io/bots/slack/connector"

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var ErrBotNotFound = errors.New("No Such registered team")

// MessageHandler is a callback to attach that is invoked whenever a new message is received
type MessageHandler func(msg []byte, team string)

// ffjson: skip
type connection struct {
	conn *websocket.Conn
	url  string
}

// Connector internally manages connections to slack teams.
//
// ffjson: skip
type Connector struct {
	bots          map[string]*connection
	handleMessage MessageHandler
}

// NewConnector creates a new Connector object
func NewConnector() *Connector {
	return &Connector{
		bots: make(map[string]*connection),
	}
}

// Open opens a websocket connection for the passed URL and assigns it to the given team.
func (c *Connector) Open(team, url string) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return errors.Wrap(err, "Start Failed")
	}

	c.bots[team] = &connection{conn, url}
	go c.readConn(conn, team)
	return nil
}

// SetMessageHandler sets the message handler for the connector.
func (c *Connector) SetMessageHandler(messageHandler MessageHandler) {
	c.handleMessage = messageHandler
}

// ffjson: nodecoder
type typingPayload struct {
	ID      int    `json:"id"`
	Channel string `json:"channel"`
	Type    string `json:"type"`
}

// Typing sends a typing payload.
func (c *Connector) Typing(team, channel string) error {
	co, ok := c.bots[team]
	if !ok {
		return ErrBotNotFound
	}

	if err := co.conn.WriteJSON(&typingPayload{1, channel, "typing"}); err != nil {
		return errors.Wrap(err, "Typing Failed")
	}

	return nil
}

// ffjson: noencoder
type reconnect struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// TODO: set up time.Ticker to refresh connection
func (c *Connector) readConn(conn *websocket.Conn, team string) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatal(errors.Wrap(err, "Unexpected error while reading message"))
		}

		r := &reconnect{}
		if err := json.Unmarshal(msg, r); err != nil {
			log.Fatal(errors.Wrap(err, "Unexpected error while parsing message"))
		}

		if r.Type == "reconnect_url" {
			c.bots[team].url = r.URL
			continue
		}

		c.handleMessage(msg, team)
	}
}

// MessagePayload is the payload received from slack over a connection.
type MessagePayload struct {
	Message json.RawMessage `json:"message"`
	Team    string          `json:"team"`
}

//go:generate ffjson $GOFILE
