package main

import (
	"flag"
	"log"
	"net/http"

	"suy.io/bots/slack"
	"suy.io/bots/slack/api/chat"
)

var verification, port string

func main() {
	flag.StringVar(&verification, "v", "", "verification token")
	flag.StringVar(&port, "p", "8080", "server port")
	flag.Parse()

	c, err := slack.NewController(slack.WithVerification(verification))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Adding Handler")
	http.HandleFunc("/slack/command", c.CommandHandler())

	go handleCommands(c)

	log.Println("starting on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleCommands(c *slack.Controller) {
	for command := range c.Commands() {
		log.Println("Got Command", command.Command)
		if err := command.RespondImmediately(&chat.Message{Text: command.Command + ": " + command.Text}, false); err != nil {
			log.Fatal(err)
		}
	}
}
