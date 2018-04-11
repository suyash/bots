> Happy chatbot developers are all alike; every unhappy chatbot developer is unhappy in his own way.

# bots

[![GoDoc](https://godoc.org/suy.io/bots?status.svg)](https://godoc.org/suy.io/bots) [![Build Status](https://travis-ci.org/suyash/bots.svg?branch=master)](https://travis-ci.org/suyash/bots)

A little library to write chatbots in go.

# Slack

> more examples in [examples/slack](examples/slack)

```go
package main

import (
	"log"

	"suy.io/bots/slack"
	"suy.io/bots/slack/api/chat"
)

func main() {
	c, err := slack.NewController()
	if err != nil {
		log.Fatal(err)
	}

	b, err := c.CreateBot("BOT_TOKEN") // https://my.slack.com/services/new/bot
	if err != nil {
		log.Fatal(err)
	}

	if err := b.Start(); err != nil {
		log.Fatal(err)
	}

	for msg := range c.DirectMessages() {
		_, err := msg.Reply(chat.TextMessage(msg.Text)) // echoes the same message back
		if err != nil {
			log.Fatal("error:", err)
		}
	}
}
```

### build

The project uses [ffjson](https://github.com/pquerna/ffjson) to optimize JSON encoding/decoding.

To regenerate

```
go generate ./...
```
