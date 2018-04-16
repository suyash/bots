package bots

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"suy.io/bots/web"
)

type ConversationBots struct {
	c  *web.Controller
	t  *template.Template
	pc PageContext
}

func NewConversationBots(t *template.Template, pc PageContext) (*ConversationBots, error) {
	c, err := web.NewController()
	if err != nil {
		return nil, err
	}

	b := &ConversationBots{c, t, pc}

	password := web.NewConversation()

	password.On("start", func(msg *web.Message, controls *web.Controls) {
		msg.Text = "Please specify a length"
		controls.Bot().Say(msg)
		controls.To("length")
	})

	password.On("length", func(msg *web.Message, controls *web.Controls) {
		_, err := strconv.ParseInt(msg.Text, 10, 64)
		if err != nil {
			// NOTE: we send a message and stay in this state, instead of transitioning
			// anywhere. This is how "repeat" works. This will repeat indefinitely until
			// a valid value is obtained, or can set a state variable to repeat a fixed
			// number of times before calling 'controls.End()'.
			controls.Bot().Say(&web.Message{Text: "Invalid Value, Please try again"})
			return
		}

		controls.Set("length", msg.Text)

		msg.Text = "Do you want numbers"
		controls.Bot().Say(msg)

		controls.To("numbers")
	})

	password.On("numbers", func(msg *web.Message, controls *web.Controls) {
		if lt := strings.ToLower(msg.Text); lt == "no" || lt == "nope" {
			controls.Bot().Say(&web.Message{Text: "Not Using Numbers"})
			controls.Set("numbers", "false")
		} else {
			controls.Bot().Say(&web.Message{Text: "Using Numbers"})
			controls.Set("numbers", "true")
		}

		controls.Bot().Say(&web.Message{Text: "Do you want special characters"})
		controls.To("characters")
	})

	password.On("characters", func(msg *web.Message, controls *web.Controls) {
		characters := true

		if lt := strings.ToLower(msg.Text); lt == "no" || lt == "nope" {
			controls.Bot().Say(&web.Message{Text: "Not Using Special Characters"})
			characters = false
		} else {
			controls.Bot().Say(&web.Message{Text: "Using Special Characters"})
		}

		l, err := controls.Get("length")
		if err != nil {
			log.Fatal(errors.Wrap(err, "Did not get length from conversation state"))
		}

		length, _ := strconv.ParseInt(l, 10, 64)

		numbers := true

		n, err := controls.Get("numbers")
		if err != nil {
			log.Fatal(errors.Wrap(err, "Did not get numbers from conversation state"))
		}

		if n == "false" {
			numbers = false
		}

		ans := generate(length, numbers, characters)

		controls.Bot().Say(&web.Message{Text: "Your Password is '" + ans + "'"})
		controls.End()
	})

	c.RegisterConversation("password", password)

	go b.handleBots()
	go b.handleMessages()

	return b, nil
}

func (b *ConversationBots) handleBots() {
	for b := range b.c.BotAdded() {
		log.Println("conversation: New Bot Added")
		b.Say(&web.Message{Text: `Online.

This is a conversation example, with the source at https://github.com/suyash/bots/blob/master/examples/web/bots/conversations.go.
Essentially typing anything starts a conversation, which ends with a password being generated.			
`})
	}
}

func (b *ConversationBots) handleMessages() {
	for msg := range b.c.Messages() {
		log.Println("conversation: Got", msg.Message)
		msg.StartConversation("password")
	}
}

func (b *ConversationBots) ConnectionHandler() http.HandlerFunc {
	return b.c.ConnectionHandler()
}

func (b *ConversationBots) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	b.t.ExecuteTemplate(res, Template, b.pc)
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
