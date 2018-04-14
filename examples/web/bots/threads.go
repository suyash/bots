package bots

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"suy.io/bots/web"
)

type ThreadBots struct {
	c  *web.Controller
	t  *template.Template
	pc PageContext
}

func NewThreadBots(t *template.Template, pc PageContext) (*ThreadBots, error) {
	c, err := web.NewController()
	if err != nil {
		return nil, err
	}

	b := &ThreadBots{c, t, pc}
	go b.handleBots()
	go b.handleMessages()

	return b, nil
}

func (b *ThreadBots) handleBots() {
	for b := range b.c.BotAdded() {
		log.Println("threads: New Bot Added")
		b.Say(&web.Message{Text: "Online"})
	}
}

func (b *ThreadBots) handleMessages() {
	for msg := range b.c.Messages() {
		log.Println("threads: Got", msg.Message)

		if strings.Index(msg.Text, "thread") != -1 {
			msg.ReplyInThread(msg.Message)
		} else {
			msg.Reply(msg.Message)
		}
	}
}

func (b *ThreadBots) ConnectionHandler() http.HandlerFunc {
	return b.c.ConnectionHandler()
}

func (b *ThreadBots) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	b.t.ExecuteTemplate(res, Template, b.pc)
}
