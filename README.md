# bots

[![GoDoc](https://godoc.org/suy.io/bots?status.svg)](https://godoc.org/suy.io/bots) [![Build Status](https://travis-ci.org/suyash/bots.svg?branch=master)](https://travis-ci.org/suyash/bots)

A little library to write chatbots in go.

## web

```
go get suy.io/bots/web
```

> more examples in [examples/web/bots](examples/web/bots). Small Deployment example in [examples/web/app](examples/web/app). Also deployed online on https://app-ruhxhowvkv.now.sh.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"suy.io/bots/web"
)

func main() {
	c, err := web.NewController()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/chat", c.ConnectionHandler())

	go handleMessages(c)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMessages(c *web.Controller) {
	for msg := range c.DirectMessages() {
		msg.Reply(web.TextMessage(msg.Text))
	}
}

func index(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "%s", `
<!DOCTYPE html>
<html lang="en">
<head>
	...
</head>
<body>
	<template id="user-message">
		...
	</template>

	<template id="bot-message">
		...
	</template>

	<div id="chat">
		<div id="messages"></div>
		<form id="form" action="/" method="post">
			<input type="text" name="message" id="message" placeholder="enter" required="required">
			<button type="submit">Send</button>
		</form>
	</div>

	<script src="/lib/chat.js"></script>
</body>
</html>`,
	)
}
```

`lib/chat.js`

```js
import Chat from "@suy/bots-web-client";

window.addEventListener("DOMContentLoaded", loaded);
async function loaded() {
    const chat = new Chat(`ws://${window.location.host}/chat`);
    await chat.open();

    const form = document.querySelector("#form");
    const message = document.querySelector("#message");
    form.addEventListener("submit", (e) => {
        e.preventDefault();
	chat.say({ text: message.value });
	...
    });

    for await (const msg of chat.messages()) {
        log("received", msg);
	...
    }

    log("closed")
}
```

### Client

There is a very simple browser client implemented at [web/browser](web/browser) that provides incoming messages over an async iterator (for-await in the above example).

### Conversations

> [full example](examples/web/bots/conversations.go)

```go
password := web.NewConversation()

password.On("start", func(msg *web.Message, controls *web.Controls) {
	msg.Text = "Please specify a length"
	controls.Bot().Say(msg)
	controls.To("length")
})

password.On("length", func(msg *web.Message, controls *web.Controls) {
	_, err := strconv.ParseInt(msg.Text, 10, 64)
	if err != nil {
		// NOTE: we send a message and stay in this state, instead of transitioning
		// anywhere. This is how "repeat" works. This will repeat indefinitely until
		// a valid value is obtained, or can set a state variable to repeat a fixed
		// number of times before calling 'controls.End()'.
		controls.Bot().Say(&web.Message{Text: "Invalid Value, Please try again"})
		return
	}

	controls.Set("length", msg.Text)

	msg.Text = "Do you want numbers"
	controls.Bot().Say(msg)

	controls.To("numbers")
})

password.On("numbers", func(msg *web.Message, controls *web.Controls) {
	if lt := strings.ToLower(msg.Text); lt == "no" || lt == "nope" {
		controls.Bot().Say(&web.Message{Text: "Not Using Numbers"})
		controls.Set("numbers", "false")
	} else {
		controls.Bot().Say(&web.Message{Text: "Using Numbers"})
		controls.Set("numbers", "true")
	}

	controls.Bot().Say(&web.Message{Text: "Do you want special characters"})
	controls.To("characters")
})

password.On("characters", func(msg *web.Message, controls *web.Controls) {
	characters := true

	...

	controls.Bot().Say(&web.Message{Text: "Your Password is '" + ans + "'"})
	controls.End()
})
```

### Storage

There are 3 main storage interfaces,

- [ControllerStore](https://godoc.org/suy.io/bots/web#ControllerStore)

  A controller store essentially stores the current bots by ID. [Example Redis Implementation](https://godoc.org/suy.io/bots/web/contrib/redis#RedisControllerStore)

- [ItemStore](https://godoc.org/suy.io/bots/web#ItemStore)

  An ItemStore stores items (threads, messages) for a single bot. [Example Redis Implementation](https://godoc.org/suy.io/bots/web/contrib/redis#RedisItemStore)

- [ConversationStore](https://godoc.org/suy.io/bots/web#RedisConversationStore)

  A ConversationStore stores and manages conversation state. [Example Redis Implementation](https://godoc.org/suy.io/bots/web/contrib/redis#RedisItemStore)

### BotID Creation

[BotIDCreator](https://godoc.org/suy.io/bots/web#BotIDCreator) is a function that can be passed to [WithBotIDCreator](https://godoc.org/suy.io/bots/web#WithBotIDCreator) when initializing controller to generate bot IDs. The default one generates a timestamp to get an id, so every connection is unique. The actual request is passed along in the function so any request parameters can be used to detect identity. A very simple example using cookies can be seen at https://github.com/suyash/bots/blob/master/examples/web/bots/redis.go#L31-L43

### Issues

- [ ] support for sending more than text from client
- [ ] shell client.
- [ ] MemoryItemStore can be made O(nlogn).
- [ ] HTTP + SSE?

## Slack

```
go get suy.io/bots/slack
```

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

### Connector

A connector is a websocket connection pool defined at https://godoc.org/suy.io/bots/slack#Connector. The connector package provides a type that can manage connections. By default all connections are also a part of the same service, but if required, can be abstracted out and the two services can talk using any transport mechanism. Sample HTTP implementations are by [httpserver](https://godoc.org/suy.io/bots/slack/connector/contrib/httpserver) and [httpclient](https://godoc.org/suy.io/bots/slack/contrib/connector/httpclient) respectively. There is also [an example](examples/slack/compose).

### Storage

There are 3 main storage interfaces

- [BotStore](https://godoc.org/suy.io/bots/slack#BotStore)

  This essentially stores `oauth.AccessResponse`s of all bots that have been authenticated with the service. A custom implementation can be provided by passing it inside `WithBotStore` function when initializing a controller. An example [redis implementation](https://godoc.org/suy.io/bots/slack/contrib/redis#RedisBotStore).

- [ConversationStore](https://godoc.org/suy.io/bots/slack#ConversationStore)

  This stores and manages conversation data and state. A custom implementation can be provided at initialization by using `WithConversationStore` when initializing a controller. An example [redis implementation](https://godoc.org/suy.io/bots/slack/contrib/redis#RedisConversationStore).

### Issues

- [ ] `*websocket.Conn` does not `Close()` and throws an error.

- [ ] ffjson not generating fflib import for interactions

- [ ] does not work on appengine, [because of using http.DefaultClient](https://github.com/suyash/bots/blob/master/slack/api/request.go#L63). Figure out a way to switch out the client.

## build

The project uses [ffjson](https://github.com/pquerna/ffjson) to optimize JSON encoding/decoding.

To regenerate

```
go generate ./...
```
