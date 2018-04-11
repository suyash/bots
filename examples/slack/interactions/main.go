package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"suy.io/bots/slack"
	"suy.io/bots/slack/api/chat"
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

	log.Println("Running")

	http.HandleFunc("/slack/interaction", c.InteractionHandler())
	http.HandleFunc("/slack/interaction/options", c.InteractionOptionsHandler())

	go handle(c)

	log.Println("starting on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handle(c *slack.Controller) {
	for {
		select {
		case msg := <-c.DirectMessages():
			log.Println("Got", msg.Text)

			if strings.Index(msg.Text, "menu") != -1 {
				respondWithMenu(msg)
			} else if strings.Index(msg.Text, "button") != -1 {
				respondWithButton(msg)
			} else {
				msg.Reply(&chat.Message{Text: "type menu for menu and button for button"})
			}

		case iact := <-c.Interactions():
			log.Println("Interaction", iact.Actions[0].Name)

			if iact.Actions[0].Type == "select" {
				onMenu(iact)
			} else {
				onButton(iact)
			}

		case iactopt := <-c.InteractionOptions():
			log.Println("InteractionOption", iactopt.Name)

			iactopt.Respond(&slack.InteractionOptionsResponse{
				OptionGroups: []*chat.OptionGroup{
					{
						Text: "Group 1",
						Options: []*chat.Option{
							{Text: "Option 1", Value: "11"},
							{Text: "Option 2", Value: "12"},
							{Text: "Option 3", Value: "13"},
						},
					},
					{
						Text: "Group 2",
						Options: []*chat.Option{
							{Text: "Option 1", Value: "21"},
							{Text: "Option 2", Value: "22"},
							{Text: "Option 3", Value: "23"},
						},
					},
				},
			})
		}
	}
}

func respondWithMenu(msg *slack.MessagePair) error {
	_, err := msg.Reply(&chat.Message{
		Attachments: []*chat.Attachment{
			{
				Title:      "Interactive Message Menu Test",
				Text:       "Nothing Selected",
				CallbackID: "interaction",
				Actions: []*chat.Action{
					{
						Type: "select",
						Text: "Select an item",
						Name: "select",
						Options: []*chat.Option{
							{
								Text:  "Option 1",
								Value: "1",
							},
							{
								Text:  "Option 2",
								Value: "2",
							},
							{
								Text:  "Option 3",
								Value: "3",
							},
						},
					},
					{
						Type:       "select",
						Text:       "Select a User",
						Name:       "user",
						DataSource: "users",
					},
					{
						Type:       "select",
						Text:       "Select a conversation",
						Name:       "conversation",
						DataSource: "conversations",
					},
					{
						Type:       "select",
						Text:       "Select an external",
						Name:       "external",
						DataSource: "external",
					},
				},
				MrkdwnIn: []string{"text", "pretext", "fields"},
			},
		},
	})

	return err
}

func respondWithButton(msg *slack.MessagePair) error {
	_, err := msg.Reply(&chat.Message{
		Attachments: []*chat.Attachment{
			{
				Title:      "Interactive Message Button Test",
				Text:       "This button has been clicked *0* times",
				CallbackID: "interaction",
				Actions: []*chat.Action{
					{
						Type:  "button",
						Text:  "Click",
						Name:  "click",
						Value: "0",
					},
				},
				MrkdwnIn: []string{"text", "pretext", "fields"},
			},
		},
	})

	return err
}

func onMenu(iact *slack.InteractionPair) {
	res := iact.OriginalMessage
	act := iact.Actions[0]
	selected := act.SelectedOptions[0].Value

	switch act.Name {
	case "select":
		res.Attachments[0].Text = fmt.Sprintf("Selected: *%v*", selected)
	case "user":
		res.Attachments[0].Text = fmt.Sprintf("Selected: *<@%v>*", selected)
	case "conversation":
		res.Attachments[0].Text = fmt.Sprintf("Selected: *<#%v>*", selected)
	case "external":
		res.Attachments[0].Text = fmt.Sprintf("Selected: *%v*", selected)
	}

	if err := iact.RespondImmediately(res); err != nil {
		log.Println(err)
	}
}

func onButton(iact *slack.InteractionPair) {
	res := iact.OriginalMessage

	count, err := strconv.Atoi(res.Attachments[0].Actions[0].Value)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("onButton with count", count)

	res.Attachments[0].Actions[0].Value = strconv.Itoa(count + 1)
	res.Attachments[0].Text = fmt.Sprintf("This button has been clicked *%d* times", count+1)

	if err := iact.RespondImmediately(res); err != nil {
		log.Println(err)
	}
}
