package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"suy.io/bots/slack"
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/oauth"
	"suy.io/bots/slack/contrib/connector/httpclient"
	"suy.io/bots/slack/contrib/redis"
)

func main() {
	conn, err := httpclient.NewConnector(os.Getenv("CONNECTOR"))
	if err != nil {
		log.Fatal(err)
	}

	c, err := slack.NewController(
		slack.WithClientID(os.Getenv("CLIENT_ID")),
		slack.WithClientSecret(os.Getenv("CLIENT_SECRET")),
		slack.WithConnector(conn),
		slack.WithBotStore(redis.NewRedisBotStore(os.Getenv("REDIS"))),
	)
	if err != nil {
		log.Fatal(err)
	}

	button, err := c.CreateAddToSlackButton(strings.Split(os.Getenv("SCOPES"), ","), os.Getenv("REDIRECT"), os.Getenv("STATE"))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Adding /")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("/")
		fmt.Fprintln(w, button)
	})

	log.Println("Adding /slack/oauth")
	http.HandleFunc("/slack/oauth", c.OAuthHandler(
		os.Getenv("REDIRECT"),
		os.Getenv("STATE"),
		func(p *oauth.AccessResponse, w http.ResponseWriter, r *http.Request) {
			log.Println("/slack/oauth")
			http.Redirect(w, r, "https://slack.com/app_redirect?app="+os.Getenv("APPID"), http.StatusFound)
		},
	))

	log.Println("Adding /slack/message")
	http.Handle("/slack/message", conn)

	go handleBots(c)
	go handleMessages(c)

	log.Println("Starting on", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func handleBots(c *slack.Controller) {
	for bot := range c.BotAdded() {
		log.Println("Added a bot for team", bot.Team())
		if err := bot.Start(); err != nil {
			log.Fatal(err)
		}
	}
}

func handleMessages(c *slack.Controller) {
	for msg := range c.DirectMessages() {
		msg.Reply(chat.TextMessage(msg.Text))
	}
}
