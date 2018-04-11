package main

import (
	"flag"
	"log"

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

	b, err := c.CreateBot(token)
	if err != nil {
		log.Fatal(err)
	}

	if err := b.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("Running")
	for msg := range c.DirectMessages() {
		log.Println("Received", msg.Text)
		_, err := msg.Reply(&chat.Message{Text: msg.Text})
		if err != nil {
			log.Println("error:", err)
		}
	}
}
