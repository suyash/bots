package slack

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/dialog"
)

const InteractionImmediateResponseTimeout = 2000 * time.Millisecond

// ffjson: skip
type Interaction struct {
	responded         int8
	immediateResponse chan []byte
	token             string

	Actions []*struct {
		Name            string `json:"name"`
		Value           string `json:"value,omitempty"`
		Text            string `json:"text,omitempty"`
		Type            string `json:"type"`
		SelectedOptions []struct {
			Value string `json:"value"`
		} `json:"selected_options,omitempty"`
	} `json:"actions"`
	Team *struct {
		ID     string `json:"id"`
		Domain string `json:"domain"`
	} `json:"team"`
	Channel *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	User *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	ActionTs        string          `json:"action_ts"`
	MessageTs       string          `json:"message_ts"`
	AttachmentID    string          `json:"attachment_id"`
	CallbackID      string          `json:"callback_id"`
	IsAppUnfurl     bool            `json:"is_app_unfurl"`
	OriginalMessage *chat.Message   `json:"original_message"`
	ResponseURL     string          `json:"response_url"`
	TriggerID       string          `json:"trigger_id"`
	Type            string          `json:"type"`
	Submission      json.RawMessage `json:"submission"`
}

func (iact *Interaction) OpenDialog(d *dialog.Dialog) error {
	if err := dialog.Open(&dialog.OpenRequest{TriggerID: iact.TriggerID, Token: iact.token, Dialog: d}); err != nil {
		return errors.Wrap(err, "OpenDialog Failed")
	}

	return nil
}

// TODO: the channel is closed after the first time, what then?
func (iact *Interaction) RespondImmediately(msg *chat.Message) error {
	if msg == nil {
		return ErrInvalidMessage
	}

	d, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "RespondImmediately Failed")
	}

	iact.immediateResponse <- d
	return nil
}

func (iact *Interaction) Respond(msg *chat.Message) error {
	if iact.responded == 5 {
		// TODO: also check for 30 minutes condition
		return ErrExceededResponseInteraction
	}

	if iact.ResponseURL == "" {
		return ErrEmptyResponseURLInteraction
	}

	d, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "Respond Failed")
	}

	res, err := http.Post(iact.ResponseURL, "application/json", bytes.NewReader(d))
	if err != nil {
		return errors.Wrap(err, "Respond Failed")
	}
	defer res.Body.Close()

	d, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "Respond Failed")
	}

	iact.responded++
	return nil
}

func (iact *Interaction) RespondWithEmptyBody() error {
	close(iact.immediateResponse)
	return nil
}

func (iact *Interaction) RespondWithErrors(errs []*dialog.Error) error {
	derrs := struct {
		Items []*dialog.Error `json:"errors"`
	}{Items: errs}

	d, err := json.Marshal(derrs)
	if err != nil {
		return errors.Wrap(err, "RespondWithErrors Failed")
	}

	iact.immediateResponse <- d
	return nil
}

// ffjson: skip
type InteractionOptions struct {
	immediateResponse chan []byte

	Name       string `json:"name"`
	Value      string `json:"value"`
	CallbackID string `json:"callback_id"`
	Team       struct {
		ID     string `json:"id"`
		Domain string `json:"domain"`
	} `json:"team"`
	Channel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	ActionTs     string `json:"action_ts"`
	MessageTs    string `json:"message_ts"`
	AttachmentID string `json:"attachment_id"`
}

// ffjson: nodecoder
type InteractionOptionsResponse struct {
	Options      []*chat.Option      `json:"options,omitempty"`
	OptionGroups []*chat.OptionGroup `json:"option_groups,omitempty"`
}

func (iactopt *InteractionOptions) Respond(msg *InteractionOptionsResponse) error {
	d, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "Respond Failed")
	}

	iactopt.immediateResponse <- d
	return nil
}

//go:generate ffjson $GOFILE
