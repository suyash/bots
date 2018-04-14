package main

import (
	"html/template"
	"log"
	"net/http"

	"suy.io/bots/examples/web/bots"
)

var templates *template.Template

var examples = []struct{ Name, Link string }{
	{"echo", "/echo"},
	{"threads", "/threads"},
	{"attachments", "/attachments"},
	{"conversations", "/conversations"},
}

func main() {
	var err error

	templates, err = template.ParseGlob("./tmpl/*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	templates, err = templates.ParseGlob("./tmpl/partials/*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	e, err := bots.NewEchoBots(templates, bots.PageContext{"echo", "/static/js/lib/echo.js"})
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/echo", e)
	http.HandleFunc("/echo_chat", e.ConnectionHandler())

	c, err := bots.NewConversationBots(templates, bots.PageContext{"echo", "/static/js/lib/conversations.js"})
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/conversations", c)
	http.HandleFunc("/conversations_chat", c.ConnectionHandler())

	t, err := bots.NewThreadBots(templates, bots.PageContext{"echo", "/static/js/lib/threads.js"})
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/threads", t)
	http.HandleFunc("/threads_chat", t.ConnectionHandler())

	a, err := bots.NewAttachmentBots(templates, bots.PageContext{"echo", "/static/js/lib/attachments.js"})
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/attachments", a)
	http.HandleFunc("/attachments_chat", a.ConnectionHandler())

	r, err := bots.NewRedisBots("redis:6379", templates, bots.PageContext{"redis", "/static/js/lib/redis.js"})
	if err != nil {
		log.Println("Not adding redis bots", err)
	} else {
		http.Handle("/redis", r)
		http.HandleFunc("/redis_chat", r.ConnectionHandler())

		examples = append(examples, struct{ Name, Link string }{"redis", "/redis"})
	}

	http.HandleFunc("/", index)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(res http.ResponseWriter, req *http.Request) {
	templates.ExecuteTemplate(res, "index.tmpl", struct {
		Items []struct{ Name, Link string }
	}{
		Items: examples,
	})
}
