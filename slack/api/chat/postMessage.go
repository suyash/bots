package chat

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type PostMessageRequest struct {
	*Message
	Token string `json:"token" url:"token"`
}

// ffjson: noencoder
type PostMessageResponse struct {
	Message *Message `json:"message"`
}

func PostMessage(req *PostMessageRequest) (*PostMessageResponse, error) {
	res := &PostMessageResponse{}
	if err := api.Request("chat.postMessage", req, true, res, req.Token); err != nil {
		return nil, errors.Wrap(err, "chat.postMessage failed")
	}

	return res, nil
}

//go:generate ffjson $GOFILE
