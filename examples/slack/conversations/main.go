package main

import (
	"flag"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"suy.io/bots/slack"
	"suy.io/bots/slack/api/chat"
)

var token string

func main() {
	flag.StringVar(&token, "t", "", "bot token")
	flag.Parse()

	c, err := slack.NewController()
	if err != nil {
		log.Fatal(err)
	}

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
		characters := true

		if lt := strings.ToLower(msg.Text); lt == "no" || lt == "nope" {
			controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Not Using Special Characters"))
			characters = false
		} else {
			controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Using Special Characters"))
		}

		l, err := controls.Get("length")
		if err != nil {
			controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Internal Error"))
			log.Fatal(errors.Wrap(err, "Did not get length from conversation state"))
		}

		length, _ := strconv.ParseInt(l, 10, 64)

		numbers := true

		n, err := controls.Get("numbers")
		if err != nil {
			controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Internal Error"))
			log.Fatal(errors.Wrap(err, "Did not get numbers from conversation state"))
		}

		if n == "false" {
			numbers = false
		}

		ans := generate(length, numbers, characters)

		controls.Bot().Reply(chat.RTMMessage(msg), chat.TextMessage("Your Password is '"+ans+"'"))
		controls.End()
	})

	c.RegisterConversation("password", password)

	b, err := c.CreateBot(token)
	if err != nil {
		log.Fatal(err)
	}

	if err := b.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("Online")

	for msg := range c.DirectMessages() {
		if err := msg.StartConversation("password"); err != nil {
			log.Fatal(err)
		}
	}
}

func generate(l int64, numbers, characters bool) string {
	s := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	if numbers {
		s += "0123456789"
	}

	if characters {
		s += "!@#$%^&*()_+}{[]"
	}

	ans := make([]byte, l)

	for i := range ans {
		ans[i] = s[rand.Intn(len(s))]
	}

	return string(ans)
}
