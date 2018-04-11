> Happy chatbot developers are all alike; every unhappy chatbot developer is unhappy in his own way.

# bots

[![GoDoc](https://godoc.org/suy.io/bots?status.svg)](https://godoc.org/suy.io/bots) [![Build Status](https://travis-ci.org/suyash/bots.svg?branch=master)](https://travis-ci.org/suyash/bots)

A little library to write chatbots in go.

## Slack

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

### Conversations

The following example is a password generation conversation. ([full example](examples/slack/conversations/main.go))

```go
password := slack.NewConversation()

password.On("start", func(msg *chat.Message, controls *slack.Controls) {
	controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Please specify a length"))
	controls.To("length")
})

password.On("length", func(msg *chat.Message, controls *slack.Controls) {
	_, err := strconv.ParseInt(msg.Text, 10, 64)
	if err != nil {
		// NOTE: we send a message and stay in this state, instead of transitioning
		// anywhere. This is how "repeat" works. This will repeat indefinitely until
		// a valid value is obtained, or can set a state variable to repeat a fixed
		// number of times before calling 'controls.End()'.
		controls.Bot().Reply(chat.RTMMessage(msg), &chat.Message{Text: "Invalid Value, Please try again"})
		return
	}

	controls.Set("length", msg.Text)
	controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Do you want numbers"))
	controls.To("numbers")
})

password.On("numbers", func(msg *chat.Message, controls *slack.Controls) {
	if lt := strings.ToLower(msg.Text); lt == "no" || lt == "nope" {
		controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Not Using Numbers"))
		controls.Set("numbers", "false")
	} else {
		controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Using Numbers"))
		controls.Set("numbers", "true")
	}

	controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Do you want special characters"))
	controls.To("characters")
})

password.On("characters", func(msg *chat.Message, controls *slack.Controls) {
	...

	controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Your Password is '"+ans+"'"))
	controls.End()
})
```

## build

The project uses [ffjson](https://github.com/pquerna/ffjson) to optimize JSON encoding/decoding.

To regenerate

```
go generate ./...
```
