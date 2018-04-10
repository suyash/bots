package chat // import "suy.io/bots/slack/api/chat"

type Message struct {
	Channel        string        `json:"channel,omitempty" url:"channel,omitempty"`
	Text           string        `json:"text,omitempty" url:"text,omitempty"`
	Attachments    []*Attachment `json:"attachments,omitempty" url:"attachments,omitempty"`
	IconEmoji      string        `json:"icon_emoji,omitempty" url:"icon_emoji,omitempty"`
	IconURL        string        `json:"icon_url,omitempty" url:"icon_url,omitempty"`
	LinkNames      bool          `json:"link_names,omitempty" url:"link_names,omitempty"`
	Parse          string        `json:"parse,omitempty" url:"parse,omitempty"`
	ReplyBroadcast bool          `json:"reply_broadcast,omitempty" url:"reply_broadcast,omitempty"`
	ThreadTs       string        `json:"thread_ts,omitempty" url:"thread_ts,omitempty"`
	UnfurlLinks    bool          `json:"unfurl_links,omitempty" url:"unfurl_links,omitempty"`
	UnfurlMedia    bool          `json:"unfurl_media,omitempty" url:"unfurl_media,omitempty"`
	Username       string        `json:"username,omitempty" url:"username,omitempty"`
	Ts             string        `json:"ts,omitempty" url:"ts,omitempty"`
}

func TextMessage(text string) *Message {
	return &Message{Text: text}
}

type Attachment struct {
	Actions    []*Action `json:"actions,omitempty" url:"actions,omitempty"`
	AuthorIcon string    `json:"author_icon,omitempty" url:"author_icon,omitempty"`
	AuthorLink string    `json:"author_link,omitempty" url:"author_link,omitempty"`
	AuthorName string    `json:"author_name,omitempty" url:"author_name,omitempty"`
	CallbackID string    `json:"callback_id" url:"callback_id"`
	Color      string    `json:"color,omitempty" url:"color,omitempty"`
	Fallback   string    `json:"fallback,omitempty" url:"fallback,omitempty"`
	Fields     []*Field  `json:"fields,omitempty" url:"fields,omitempty"`
	Footer     string    `json:"footer,omitempty" url:"footer,omitempty"`
	FooterIcon string    `json:"footer_icon,omitempty" url:"footer_icon,omitempty"`
	ImageURL   string    `json:"image_url,omitempty" url:"image_url,omitempty"`
	MrkdwnIn   []string  `json:"mrkdwn_in,omitempty" url:"mrkdwn_in,omitempty"`
	Pretext    string    `json:"pretext,omitempty" url:"pretext,omitempty"`
	Text       string    `json:"text,omitempty" url:"text,omitempty"`
	ThumbURL   string    `json:"thumb_url,omitempty" url:"thumb_url,omitempty"`
	Title      string    `json:"title,omitempty" url:"title,omitempty"`
	TitleLink  string    `json:"title_link,omitempty" url:"title_link,omitempty"`
	Ts         int       `json:"ts,omitempty" url:"ts,omitempty"`
}

type ActionType string

const (
	ButtonActionType ActionType = "button"
	SelectActionType ActionType = "select"
)

type Action struct {
	Name         string         `json:"name" url:"name"`
	DataSource   string         `json:"data_source,omitempty" url:"data_source,omitempty"`
	Options      []*Option      `json:"options,omitempty" url:"options,omitempty"`
	OptionGroups []*OptionGroup `json:"option_groups,omitempty" url:"option_groups,omitempty"`
	Style        string         `json:"style,omitempty" url:"style,omitempty"`
	Text         string         `json:"text" url:"text"`
	Type         ActionType     `json:"type" url:"type"`
	Value        string         `json:"value,omitempty" url:"value,omitempty"`
	Confirm      *ActionConfirm `json:"confirm,omitempty" url:"confirm,omitempty"`
}

type Field struct {
	Title string `json:"title" url:"title"`
	Value string `json:"value" url:"value"`
	Short bool   `json:"short,omitempty" url:"short,omitempty"`
}

type Option struct {
	Text        string `json:"text" url:"text"`
	Value       string `json:"value" url:"value"`
	Description string `json:"description,omitempty" url:"description,omitempty"`
}

type OptionGroup struct {
	Text    string    `json:"text" url:"text"`
	Options []*Option `json:"options,omitempty" url:"options,omitempty"`
}

type ActionConfirm struct {
	Title       string `json:"title" url:"title"`
	Text        string `json:"text" url:"text"`
	OkText      string `json:"ok_text" url:"ok_text"`
	DismissText string `json:"dismiss_text" url:"dismiss_text"`
}

type EphemeralMessage struct {
	Channel        string       `json:"channel,omitempty" url:"channel,omitempty"`
	Text           string       `json:"text,omitempty" url:"text,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty" url:"attachments,omitempty"`
	IconEmoji      string       `json:"icon_emoji,omitempty" url:"icon_emoji,omitempty"`
	IconURL        string       `json:"icon_url,omitempty" url:"icon_url,omitempty"`
	LinkNames      bool         `json:"link_names,omitempty" url:"link_names,omitempty"`
	Parse          string       `json:"parse,omitempty" url:"parse,omitempty"`
	ReplyBroadcast bool         `json:"reply_broadcast,omitempty" url:"reply_broadcast,omitempty"`
	ThreadTs       string       `json:"thread_ts,omitempty" url:"thread_ts,omitempty"`
	UnfurlLinks    bool         `json:"unfurl_links,omitempty" url:"unfurl_links,omitempty"`
	UnfurlMedia    bool         `json:"unfurl_media,omitempty" url:"unfurl_media,omitempty"`
	Username       string       `json:"username,omitempty" url:"username,omitempty"`
	User           string       `json:"user,omitempty" url:"user,omitempty"`
	Ts             string       `json:"ts,omitempty" url:"ts,omitempty"`
}

//go:generate ffjson $GOFILE
