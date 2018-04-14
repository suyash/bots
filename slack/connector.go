package slack

import (
	"suy.io/bots/slack/connector"
)

// Connector defines the client interface to a service that manages WebSocket
// connections for this application.
type Connector interface {
	// add a new team and socket url to manage
	Add(team, url string) error

	// get incoming messages
	Messages() <-chan *connector.MessagePayload

	// send typing indicator for a team
	Typing(team, channel string) error

	// close the connector
	Close()
}

// internalConnector is an in-memory implementation that manages WebSocket connections
// for an app on the same service.
type internalConnector struct {
	conn *connector.Connector
	msgs chan *connector.MessagePayload
}

// newInternalConnector creates a new internalConnector
func newInternalConnector() *internalConnector {
	c := &internalConnector{
		conn: connector.NewConnector(),
		msgs: make(chan *connector.MessagePayload),
	}

	c.conn.SetMessageHandler(c.handleMessage)
	return c
}

// Add opens a new connection
func (c *internalConnector) Add(team, url string) error {
	return c.conn.Open(team, url)
}

// handleMessage handles a message from a connection
func (c *internalConnector) handleMessage(msg []byte, team string) {
	go func() { c.msgs <- &connector.MessagePayload{Team: team, Message: msg} }()
}

// Messages returns a receive only channel that gets messages.
func (c *internalConnector) Messages() <-chan *connector.MessagePayload {
	return c.msgs
}

// Close closes the connector, and stops listening to messages.
func (c *internalConnector) Close() {
	close(c.msgs)
	c.msgs = nil
}

// Typing sends a typing indicatior.
func (c *internalConnector) Typing(team, channel string) error {
	return c.conn.Typing(team, channel)
}

var _ Connector = &internalConnector{}
