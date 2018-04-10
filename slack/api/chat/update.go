package chat

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type UpdateRequest struct {
	*Message
	Token string `json:"token" url:"token"`
}

// ffjson: noencoder
type UpdateResponse struct {
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
	Text    string `json:"text"`
}

func Update(req *UpdateRequest) (*UpdateResponse, error) {
	res := &UpdateResponse{}
	if err := api.Request("chat.update", req, true, res, req.Token); err != nil {
		return nil, errors.Wrap(err, "chat.update failed")
	}

	return res, nil
}

//go:generate ffjson $GOFILE
