package web

import (
	"encoding/json"
	"net/http"
)

// The types here need to have 1:1 correspondence with message.ts

type BotID int64

type ItemID int64

type ItemType string

const (
	MessageItemType ItemType = "message"
	ThreadItemType  ItemType = "thread"
	UpdateItemType  ItemType = "update"
)

type Item interface {
	ItemID() ItemID
	ThreadItemID() ItemID
	ItemType() ItemType
}

type Cursor struct {
	Type ItemType `json:"type"`
	ID   ItemID   `json:"id"`
}

type ItemSource string

const (
	BotItemSource  ItemSource = "bot"
	UserItemSource ItemSource = "user"
)

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

func (msg *Message) ItemID() ItemID {
	return msg.ID
}

func (msg *Message) ThreadItemID() ItemID {
	return msg.ThreadID
}

func (msg *Message) ItemType() ItemType {
	return MessageItemType
}

var _ Item = &Message{}

func TextMessage(text string) *Message {
	return &Message{Text: text}
}

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

func (t *Thread) ItemID() ItemID {
	return t.ID
}

func (t *Thread) ThreadItemID() ItemID {
	return t.ThreadID
}

func (t *Thread) ItemType() ItemType {
	return ThreadItemType
}

var _ Item = &Thread{}

func UnmarshalJSONItem(js []byte) (Item, error) {
	s := &struct {
		Type ItemType `json:"type"`
	}{}

	if err := json.Unmarshal(js, s); err != nil {
		return nil, err
	}

	if s.Type == ThreadItemType {
		t := &Thread{}
		if err := json.Unmarshal(js, t); err != nil {
			return nil, err
		}

		return t, nil
	} else if s.Type == MessageItemType {
		msg := &Message{}
		if err := json.Unmarshal(js, msg); err != nil {
			return nil, err
		}

		return msg, nil
	}

	return nil, ErrInvalidItem
}

type AttachmentType string

const (
	ImageType        AttachmentType = "image"
	AudioType                       = "audio"
	VideoType                       = "video"
	LocationType                    = "location"
	FileDownloadType                = "file_download"
)

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

func (a *attachment) Type() AttachmentType {
	return ImageType
}

var _ Attachment = &attachment{}

func WithTitle(title string) func(Attachment) {
	return func(a Attachment) {
		a.(*attachment).Title = title
	}
}

func WithText(text string) func(Attachment) {
	return func(a Attachment) {
		a.(*attachment).Text = text
	}
}

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

// ffjson: skip
type BotPair struct {
	*Bot
	*http.Request
}

// ffjson: skip
type MessagePair struct {
	*Message
	*Bot
}

func (mp *MessagePair) Reply(msg *Message) {
	mp.Bot.Reply(mp.Message, msg)
}

func (mp *MessagePair) ReplyInThread(msg *Message) {
	mp.Bot.ReplyInThread(mp.Message, msg)
}

func (mp *MessagePair) Update() {
	mp.Bot.Update(mp.Message)
}

//go:generate ffjson $GOFILE
