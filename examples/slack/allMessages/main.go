package main

import (
	"flag"
	"log"

	"suy.io/bots/slack"
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
	for {
		select {
		case msg := <-c.DirectMessages():
			log.Println("DirectMessage:", msg)
		case msg := <-c.SelfMessages():
			log.Println("SelfMessage:", msg)
		case msg := <-c.DirectMentions():
			log.Println("DirectMention:", msg)
		case msg := <-c.Mentions():
			log.Println("Mention:", msg)
		case msg := <-c.AmbientMessages():
			log.Println("AmbientMessage:", msg)
		case msg := <-c.ChannelJoin():
			log.Println("ChannelJoin:", msg)
		case msg := <-c.UserChannelJoin():
			log.Println("UserChannelJoin:", msg)
		case msg := <-c.GroupJoin():
			log.Println("GroupJoin:", msg)
		}
	}
}
