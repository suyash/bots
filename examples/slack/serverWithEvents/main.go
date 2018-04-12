package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"suy.io/bots/slack"
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/oauth"
)

var appID, clientID, clientSecret, verification, scopes, redirect, state, port string

func main() {
	flag.StringVar(&appID, "a", "", "App ID")
	flag.StringVar(&clientID, "c", "", "Client ID")
	flag.StringVar(&clientSecret, "cs", "", "Client Secret")
	flag.StringVar(&verification, "v", "", "Verification Token")
	flag.StringVar(&scopes, "S", "", "Client Scopes")
	flag.StringVar(&redirect, "r", "", "Client Redirect")
	flag.StringVar(&state, "s", "", "Client State")
	flag.StringVar(&port, "p", "8080", "Server Port")
	flag.Parse()

	c, err := slack.NewController(
		slack.WithClientID(clientID),
		slack.WithClientSecret(clientSecret),
		slack.WithVerification(verification),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/event", c.EventHandler())

	button, err := c.CreateAddToSlackButton(strings.Split(scopes, ","), redirect, state)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, button)
	})

	http.HandleFunc("/slack/oauth", c.OAuthHandler(
		redirect,
		state,
		func(p *oauth.AccessResponse, w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://slack.com/app_redirect?app="+appID, http.StatusFound)
		},
	))

	go handleBots(c)
	go handleMessages(c)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleBots(c *slack.Controller) {
	for b := range c.BotAdded() {
		log.Println("Added Bot For Team", b.Team())
		// DO NOT START THE BOT, OR BE PREPARED TO SEE DOUBLE
	}
}

func handleMessages(c *slack.Controller) {
	for msg := range c.DirectMessages() {
		log.Println("Got", msg.Text)
		msg.Reply(&chat.Message{Text: msg.Text})
	}
}
