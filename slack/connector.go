package slack

import (
	"suy.io/bots/slack/connector"
)

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

// ffjson: skip
type internalConnector struct {
	conn *connector.Connector
	msgs chan *connector.MessagePayload
}

func newInternalConnector() *internalConnector {
	c := &internalConnector{
		conn: connector.NewConnector(),
		msgs: make(chan *connector.MessagePayload),
	}

	c.conn.SetMessageHandler(c.handleMessage)
	return c
}

func (c *internalConnector) Add(team, url string) error {
	return c.conn.Open(team, url)
}

func (c *internalConnector) handleMessage(msg []byte, team string) {
	go func() { c.msgs <- &connector.MessagePayload{Team: team, Message: msg} }()
}

func (c *internalConnector) Messages() <-chan *connector.MessagePayload {
	return c.msgs
}

func (c *internalConnector) Close() {
	close(c.msgs)
	c.msgs = nil
}

func (c *internalConnector) Typing(team, channel string) error {
	return c.conn.Typing(team, channel)
}

var _ Connector = &internalConnector{}
