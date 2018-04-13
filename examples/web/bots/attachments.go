package bots

import (
	"html/template"
	"log"
	"net/http"

	"suy.io/bots/web"
)

type AttachmentBots struct {
	c  *web.Controller
	t  *template.Template
	pc PageContext
}

func NewAttachmentBots(t *template.Template, pc PageContext) (*AttachmentBots, error) {
	c, err := web.NewController()
	if err != nil {
		return nil, err
	}

	b := &AttachmentBots{c, t, pc}
	go b.handleBots()
	go b.handleMessages()

	return b, nil
}

func (b *AttachmentBots) handleBots() {
	for b := range b.c.BotAdded() {
		log.Println("threads: New Bot Added")
		b.Say(&web.Message{Text: "Online"})
	}
}

func (b *AttachmentBots) handleMessages() {
	for msg := range b.c.DirectMessages() {
		log.Println("attachments: Got", msg.Message)

		switch msg.Text {
		case "image":
			msg.Attachments = []web.Attachment{
				web.Image(
					"https://images.unsplash.com/photo-1519176336903-04be58a477d2?ixlib=rb-0.3.5&ixid=eyJhcHBfaWQiOjEyMDd9&s=eda05ddcb3154f39fd8ce88fdd44f531&dpr=1&auto=format&fit=crop&w=1000&q=80&cs=tinysrgb",
					"https://unsplash.com/photos/VzjedZTySDk",
					web.WithTitle("Blake Wisz"),
					web.WithText("image attachment"),
				),
			}

			msg.Text = ""
			msg.Reply(msg.Message)
		case "audio":
			msg.Attachments = []web.Attachment{
				web.Audio(
					"https://freemusicarchive.org/file/music/no_curator/Tours/Enthusiast/Tours_-_01_-_Enthusiast.mp3",
					web.WithText("http://freemusicarchive.org/music/Tours/Enthusiast/Tours_-_Enthusiast"),
					web.WithTitle("Enthusiast by Tours"),
				),
			}

			msg.Text = ""
			msg.Reply(msg.Message)
		case "video":
			msg.Attachments = []web.Attachment{
				web.Video(
					"https://www.videvo.net/videvo_files/converted/2012_08/videos/Birds%20at%20a%20dock-H264%2075.mov37412.mp4",
					web.WithTitle("Birds at a Dock"),
					web.WithText("https://www.videvo.net/video/birds-at-a-dock/379/"),
				),
			}

			msg.Text = ""
			msg.Reply(msg.Message)
		case "location":
			msg.Attachments = []web.Attachment{
				web.Location(46.234035, 6.05245, web.WithTitle("New York City"), web.WithText("USA")),
			}

			msg.Text = ""
			msg.Reply(msg.Message)
		case "download":
			msg.Attachments = []web.Attachment{
				web.FileDownload(
					"https://unsplash.com/photos/qr4d407hSjo/download?force=true",
					web.WithText("https://unsplash.com/photos/qr4d407hSjo"),
				),
			}

			msg.Text = ""
			msg.Reply(msg.Message)
		default:
			msg.Reply(msg.Message)
		}
	}
}

func (b *AttachmentBots) ConnectionHandler() http.HandlerFunc {
	return b.c.ConnectionHandler()
}

func (b *AttachmentBots) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	b.t.ExecuteTemplate(res, Template, b.pc)
}
