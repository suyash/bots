package bots

import (
	"html/template"
	"log"
	"net/http"

	"suy.io/bots/web"
)

const Template = "chat.tmpl"

type PageContext struct {
	Title  string
	Script string
}

type EchoBots struct {
	c  *web.Controller
	t  *template.Template
	pc PageContext
}

func NewEchoBots(t *template.Template, pc PageContext) (*EchoBots, error) {
	c, err := web.NewController()
	if err != nil {
		return nil, err
	}

	b := &EchoBots{c, t, pc}
	go b.handleBots()
	go b.handleMessages()

	return b, nil
}

func (b *EchoBots) handleBots() {
	for b := range b.c.BotAdded() {
		log.Println("echo: New Bot Added")
		b.Say(&web.Message{Text: "Online"})
	}
}

func (b *EchoBots) handleMessages() {
	for msg := range b.c.Messages() {
		log.Println("echo: Got", msg.Message)
		msg.Reply(msg.Message)
	}
}

func (b *EchoBots) ConnectionHandler() http.HandlerFunc {
	return b.c.ConnectionHandler()
}

func (b *EchoBots) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	b.t.ExecuteTemplate(res, Template, b.pc)
}
