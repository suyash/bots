package dialog // import "suy.io/bots/slack/api/dialog"

// ffjson: nodecoder
type Dialog struct {
	CallbackID  string     `json:"callback_id" url:"callback_id"`
	Title       string     `json:"title" url:"title"`
	SubmitLabel string     `json:"submit_label,omitempty" url:"submit_label,omitempty"`
	Elements    []*Element `json:"elements" url:"elements"`
}

type ElementType string

const (
	SelectElementType   ElementType = "select"
	TextElementType     ElementType = "text"
	TextAreaElementType ElementType = "textarea"
)

type ElementSubType string

const (
	EmailElementSubType  ElementSubType = "email"
	TelElementSubType    ElementSubType = "tel"
	NumberElementSubType ElementSubType = "number"
	URLElementSubType    ElementSubType = "url"
)

// ffjson: nodecoder
type Element struct {
	Label       string          `json:"label" url:"label"`
	Name        string          `json:"name" url:"name"`
	Type        ElementType     `json:"type" url:"type"`
	SubType     ElementSubType  `json:"subtype,omitempty" url:"subtype,omitempty"`
	Optional    bool            `json:"optional,omitempty" url:"optional,omitempty"`
	Placeholder string          `json:"placeholder,omitempty" url:"placeholder,omitempty"`
	Value       string          `json:"value,omitempty" url:"value,omitempty"`
	Options     []*SelectOption `json:"options,omitempty" url:"options,omitempty"`
	Hint        string          `json:"hint,omitempty" url:"hint,omitempty"`
	MaxLength   int             `json:"max_length,omitempty" url:"max_length,omitempty"`
	MinLength   int             `json:"min_length,omitempty" url:"min_length,omitempty"`
}

// ffjson: nodecoder
type SelectOption struct {
	Label string `json:"label" url:"label"`
	Value string `json:"value" url:"value"`
}

// ffjson: nodecoder
type Error struct {
	Name  string `json:"name" url:"name"`
	Error string `json:"error" url:"error"`
}

//go:generate ffjson $GOFILE
