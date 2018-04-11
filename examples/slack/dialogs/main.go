package main

import (
	"flag"
	"log"
	"net/http"

	"suy.io/bots/slack"
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/dialog"
)

var token, port string

func main() {
	flag.StringVar(&token, "t", "", "bot token")
	flag.StringVar(&port, "p", "8080", "server port")
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

	http.HandleFunc("/slack/interaction", c.InteractionHandler())

	go handle(c)

	log.Println("starting on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handle(c *slack.Controller) {
	m := &chat.Message{
		Attachments: []*chat.Attachment{
			{
				Actions: []*chat.Action{
					{
						Name: "open",
						Text: "Open Dialog",
						Type: chat.ButtonActionType,
					},
				},
				CallbackID: "dialog",
				Text:       "Open",
			},
		},
	}

	var rep *chat.Message

	for {
		select {
		case msg := <-c.DirectMessages():
			log.Println("received message")
			rep, _ = msg.Reply(m)

		case iact := <-c.Interactions():
			log.Println("received interaction")
			switch iact.CallbackID {
			case "dialog":
				iact.OpenDialog(&dialog.Dialog{
					CallbackID: "dialogData",
					Title:      "Dialog Test",
					Elements: []*dialog.Element{
						{
							Name:  "text",
							Type:  dialog.TextElementType,
							Label: "text",
						},
						{
							Label: "textarea",
							Name:  "textarea",
							Type:  dialog.TextAreaElementType,
							Hint:  "text area hint",
						},
						{
							Type:        dialog.SelectElementType,
							Name:        "select",
							Label:       "select",
							Placeholder: "select option",
							Options: []*dialog.SelectOption{
								{
									Label: "Option 1",
									Value: "1",
								},
								{
									Label: "Option 2",
									Value: "2",
								},
								{
									Label: "Option 3",
									Value: "3",
								},
							},
						},
					},
				})
			case "dialogData":
				log.Println("received dialog data")
				iact.RespondWithEmptyBody()

				m.Text = "Submitted: " + string(iact.Submission)
				log.Println(m.Text)
				if rep != nil {
					log.Println("Updating")

					_, err := iact.Update(rep.Ts, m)
					if err != nil {
						log.Println(err)
					}
				} else {
					log.Println("Saying")

					m.Channel = iact.Channel.ID
					var err error
					rep, err = iact.Say(m)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}
