package web

// The types here need to have 1:1 correspondence with message.ts

type BotID int64

type ItemID int64

type ItemType string

const (
	MessageItemType ItemType = "message"
	ThreadItemType  ItemType = "thread"
	UpdateItemType  ItemType = "update"
)

// Item defines common elements for messages and threads
type Item interface {
	ItemID() ItemID
	ThreadItemID() ItemID
	ItemType() ItemType
}

// Cursor objects point to the previous and next item for both messages and threads
type Cursor struct {
	Type ItemType `json:"type"`
	ID   ItemID   `json:"id"`
}

// ItemSource field lets a client know if a message is a bot message or a user message
type ItemSource string

const (
	BotItemSource  ItemSource = "bot"
	UserItemSource ItemSource = "user"
)

// Message defines the contents of a single message sent by the bot
type Message struct {
	Attachments []Attachment `json:"attachments"`
	Source      ItemSource   `json:"source"`
	Type        ItemType     `json:"type"`
	Text        string       `json:"text"`
	ThreadID    ItemID       `json:"threadId"`
	ReplyID     ItemID       `json:"replyId"`
	ID          ItemID       `json:"id"`
	Prev        *Cursor      `json:"prev"`
	Next        *Cursor      `json:"next"`
}

func (msg *Message) clone() *Message {
	ans := &Message{
		msg.Attachments,
		msg.Source,
		msg.Type,
		msg.Text,
		msg.ThreadID,
		msg.ReplyID,
		msg.ID,
		nil,
		nil,
	}

	if msg.Prev != nil {
		ans.Prev = &Cursor{msg.Prev.Type, msg.Prev.ID}
	}

	if msg.Next != nil {
		ans.Next = &Cursor{msg.Next.Type, msg.Next.ID}
	}

	return ans
}

// ItemID gets the ID of the message
func (msg *Message) ItemID() ItemID {
	return msg.ID
}

// ThreadItemID gets the thread ID of the message
func (msg *Message) ThreadItemID() ItemID {
	return msg.ThreadID
}

// ItemType gets the type for the message i.e. MessageItemType
func (msg *Message) ItemType() ItemType {
	return MessageItemType
}

var _ Item = &Message{}

// TextMessage is a shorthand function that creates a message from the passed text
func TextMessage(text string) *Message {
	return &Message{Text: text}
}

// Thread defines a single messaging thread, which contains messages in itself
type Thread struct {
	Type     ItemType `json:"type"`
	ThreadID ItemID   `json:"threadId"`
	ID       ItemID   `json:"id"`
	Prev     *Cursor  `json:"prev"`
	Next     *Cursor  `json:"next"`
}

func newThread(src *Message) *Thread {
	return &Thread{ThreadID: src.ThreadID, ID: src.ID, Type: ThreadItemType}
}

// ItemID returns the id of the thread
func (t *Thread) ItemID() ItemID {
	return t.ID
}

// ThreadItemID returns the id of the parent thread for this thread
func (t *Thread) ThreadItemID() ItemID {
	return t.ThreadID
}

// ItemType returns ThreadItemType for a thread
func (t *Thread) ItemType() ItemType {
	return ThreadItemType
}

var _ Item = &Thread{}

// AttachmentType defines a string identifying the type of an message attachment
type AttachmentType string

const (
	ImageType        AttachmentType = "image"
	AudioType                       = "audio"
	VideoType                       = "video"
	LocationType                    = "location"
	FileDownloadType                = "file_download"
)

// Attachment defines the interface for all message attachments
type Attachment interface {
	Type() AttachmentType
}

type attachment struct {
	AttachmentType AttachmentType `json:"type,omitempty"`
	URL            string         `json:"url,omitempty"`
	Title          string         `json:"title,omitempty"`
	Text           string         `json:"text,omitempty"`
	Alt            string         `json:"alt,omitempty"`
	Lat            float64        `json:"lat,omitempty"`
	Long           float64        `json:"long,omitempty"`
}

// Type gets the type of the attachmnet
func (a *attachment) Type() AttachmentType {
	return ImageType
}

var _ Attachment = &attachment{}

// WithTitle is a shorthand that can add a title to any message attachment
func WithTitle(title string) func(Attachment) {
	return func(a Attachment) {
		a.(*attachment).Title = title
	}
}

// WithText is a shorthand that can add text to any message attachment
func WithText(text string) func(Attachment) {
	return func(a Attachment) {
		a.(*attachment).Text = text
	}
}

// Image defines a function to create an image attachment
func Image(url, alt string, options ...func(Attachment)) Attachment {
	ans := &attachment{
		AttachmentType: ImageType,
		URL:            url,
		Alt:            alt,
	}

	for _, opt := range options {
		opt(ans)
	}

	return ans
}

// Audio defines a function to create an audio attachment
func Audio(url string, options ...func(Attachment)) Attachment {
	ans := &attachment{
		AttachmentType: AudioType,
		URL:            url,
	}

	for _, opt := range options {
		opt(ans)
	}

	return ans
}

// Video defines a function to create a video attachment
func Video(url string, options ...func(Attachment)) Attachment {
	ans := &attachment{
		AttachmentType: VideoType,
		URL:            url,
	}

	for _, opt := range options {
		opt(ans)
	}

	return ans
}

// Location defines a function to create a location attachment
func Location(lat, long float64, options ...func(Attachment)) Attachment {
	ans := &attachment{
		AttachmentType: LocationType,
		Lat:            lat,
		Long:           long,
	}

	for _, opt := range options {
		opt(ans)
	}

	return ans
}

// FileDownload defines a function to create a download attachment
func FileDownload(url string, options ...func(Attachment)) Attachment {
	ans := &attachment{
		AttachmentType: FileDownloadType,
		URL:            url,
	}

	for _, opt := range options {
		opt(ans)
	}

	return ans
}

// MessagePair is the payload sent for DirectMessage Event.
//
// ffjson: skip
type MessagePair struct {
	*Message
	*Bot
}

// Reply uses the pair's bot to reply with the passed message
func (mp *MessagePair) Reply(msg *Message) {
	mp.Bot.Reply(mp.Message, msg)
}

// ReplyInThread uses the pair's bot to reply in a new thread
func (mp *MessagePair) ReplyInThread(msg *Message) {
	mp.Bot.ReplyInThread(mp.Message, msg)
}

// Update uses the pair's bot to update an existing message.
func (mp *MessagePair) Update() {
	mp.Bot.Update(mp.Message)
}

//go:generate ffjson $GOFILE
