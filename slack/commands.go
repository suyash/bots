package slack

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/dialog"
)

// CommandImmediateResponseTimeout is the maximum time a command invocation waits
// before assuming you'll reply using the ResponseURL
const CommandImmediateResponseTimeout = 2000 * time.Millisecond

// Command is the payload sent by slack for a command.
// ffjson: skip
type Command struct {
	immediateResponse chan []byte
	responded         byte

	TeamID         string `json:"team_id"`
	TeamDomain     string `json:"team_domain"`
	EnterpriseID   string `json:"enterprise_id"`
	EnterpriseName string `json:"enterprise_name"`
	ChannelID      string `json:"channel_id"`
	ChannelName    string `json:"channel_name"`
	UserID         string `json:"user_id"`
	Command        string `json:"command"`
	Text           string `json:"text"`
	ResponseURL    string `json:"response_url"`
	TriggerID      string `json:"trigger_id"`
}

// newCommand creates a new command from query parameters.
func newCommand(vals url.Values) *Command {
	ans := &Command{}

	ans.immediateResponse = make(chan []byte)
	ans.responded = 0

	ans.TeamID = vals.Get("team_id")
	ans.TeamDomain = vals.Get("team_domain")
	ans.EnterpriseID = vals.Get("enterprise_id")
	ans.EnterpriseName = vals.Get("enterprise_name")
	ans.ChannelID = vals.Get("channel_id")
	ans.ChannelName = vals.Get("channel_name")
	ans.UserID = vals.Get("user_id")
	ans.Command = vals.Get("command")
	ans.Text = vals.Get("text")
	ans.ResponseURL = vals.Get("response_url")
	ans.TriggerID = vals.Get("trigger_id")

	return ans
}

// commandResponse represents the payload sent as a response to a command.
// ffjson: nodecoder
type commandResponse struct {
	*chat.Message
	ResponseType string `json:"response_type,omitempty"`
}

// RespondImmediately sends an immediate response to a command
func (command *Command) RespondImmediately(msg *chat.Message, inChannel bool) error {
	resType := "ephemeral"
	if inChannel {
		resType = "in_channel"
	}

	res := &commandResponse{
		Message:      msg,
		ResponseType: resType,
	}

	d, err := json.Marshal(res)
	if err != nil {
		return errors.Wrap(err, "RespondImmediately Failed")
	}

	command.immediateResponse <- d
	return nil
}

// Respond responds to a command using the response URL
func (command *Command) Respond(msg *chat.Message, inChannel bool) error {
	if command.responded == 5 {
		// TODO: also check for 30 minutes condition
		return ErrExceededResponseCommand
	}

	resType := "ephemeral"
	if inChannel {
		resType = "in_channel"
	}

	resMsg := &commandResponse{
		Message:      msg,
		ResponseType: resType,
	}

	d, err := json.Marshal(resMsg)
	if err != nil {
		return errors.Wrap(err, "Respond Failed")
	}

	res, err := http.Post(command.ResponseURL, "application/json", bytes.NewReader(d))
	defer res.Body.Close()
	if err != nil {
		return errors.Wrap(err, "Respond Failed")
	}

	d, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "Respond Failed")
	}

	command.responded++
	return nil
}

// OpenDialog opens a dialog as a response to a command.
func (command *Command) OpenDialog(d *dialog.Dialog, token string) error {
	if err := dialog.Open(&dialog.OpenRequest{TriggerID: command.TriggerID, Token: token, Dialog: d}); err != nil {
		return errors.Wrap(err, "OpenDialog Failed")
	}

	return nil
}

// RespondWithEmptyBody responds to a command with an empty response.
func (command *Command) RespondWithEmptyBody() error {
	close(command.immediateResponse)
	return nil
}

// RespondWithErrors sends errors as a command response.
func (command *Command) RespondWithErrors(errs []*dialog.Error) error {
	derrs := struct {
		Items []*dialog.Error `json:"errors"`
	}{Items: errs}

	d, err := json.Marshal(derrs)
	if err != nil {
		return errors.Wrap(err, "RespondWithErrors Failed")
	}

	command.immediateResponse <- d
	return nil
}

//go:generate ffjson $GOFILE
