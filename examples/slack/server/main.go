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

var appID, clientID, clientSecret, scopes, redirect, state, port string

func main() {
	flag.StringVar(&appID, "a", "", "App ID")
	flag.StringVar(&clientID, "c", "", "Client ID")
	flag.StringVar(&clientSecret, "cs", "", "Client Secret")
	flag.StringVar(&scopes, "S", "", "Client Scopes")
	flag.StringVar(&redirect, "r", "", "Client Redirect")
	flag.StringVar(&state, "s", "", "Client State")
	flag.StringVar(&port, "p", "8080", "Server Port")
	flag.Parse()

	log.Println(scopes)

	c, err := slack.NewController(
		slack.WithClientID(clientID),
		slack.WithClientSecret(clientSecret),
	)

	if err != nil {
		log.Fatal(err)
	}

	url, err := c.CreateAddToSlackURL(strings.Split(scopes, ","), redirect, state)
	if err != nil {
		log.Fatal(err)
	}

	button := createAddToSlackButton(url)

	log.Println("Adding / Handler")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("/")
		fmt.Fprintln(w, button)
	})

	log.Println("Adding /slack/oauth Handler")
	http.HandleFunc("/slack/oauth", c.OAuthHandler(
		redirect,
		state,
		func(p *oauth.AccessResponse, w http.ResponseWriter, r *http.Request) {
			log.Println("/slack/oauth")
			http.Redirect(w, r, "https://slack.com/app_redirect?app="+appID, http.StatusFound)
		},
	))

	go handleBots(c)
	go handleMessages(c)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleBots(c *slack.Controller) {
	for bot := range c.BotAdded() {
		log.Println("Got bot, starting")
		if err := bot.Start(); err != nil {
			log.Fatal(err)
		}
	}
}

func handleMessages(c *slack.Controller) {
	for msg := range c.DirectMessages() {
		log.Println("Got", msg.Text)
		msg.Reply(&chat.Message{Text: msg.Text})
	}
}

func createAddToSlackButton(url string) string {
	return `<a href="` + url + `"><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcset="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a>`
}
