package main

import (
	"log"
	"net/http"
	"os"

	"suy.io/bots/slack/connector/contrib/httpserver"
)

func main() {
	c := httpserver.NewConnector(os.Getenv("MESSAGE_URL"))

	ah, th := c.AddHandler(), c.TypingHandler()

	log.Println("Adding /slack/add")
	http.HandleFunc("/slack/add", func(res http.ResponseWriter, req *http.Request) {
		log.Println("Adding A Bot")
		ah(res, req)
	})

	log.Println("Adding /slack/typing")
	http.HandleFunc("/slack/typing", func(res http.ResponseWriter, req *http.Request) {
		log.Println("Sending Typing For A Bot")
		th(res, req)
	})

	log.Println("Starting on", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
