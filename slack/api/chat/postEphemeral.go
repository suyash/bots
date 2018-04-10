package chat

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type PostEphemeralRequest struct {
	*EphemeralMessage
	Token string `json:"token" url:"token"`
}

// ffjson: noencoder
type PostEphemeralResponse struct {
	MessageTs string `json:"message_ts"`
}

func PostEphemeral(req *PostEphemeralRequest) (*PostEphemeralResponse, error) {
	res := &PostEphemeralResponse{}
	if err := api.Request("chat.postEphemeral", req, true, res, req.Token); err != nil {
		return nil, errors.Wrap(err, "chat.postEphemeral failed")
	}

	return res, nil
}

//go:generate ffjson $GOFILE
