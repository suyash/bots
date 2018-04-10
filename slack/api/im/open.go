package im // import "suy.io/bots/slack/api/im"

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type OpenRequest struct {
	Token         string `json:"token" url:"token"`
	User          string `json:"user" url:"user"`
	IncludeLocale bool   `json:"include_locale,omitempty" url:"include_locale,omitempty"`
	ReturnIm      bool   `json:"return_im,omitempty" url:"return_im,omitempty"`
}

// ffjson: noencoder
type OpenResponse struct {
	NoOp        bool `json:"no_op"`
	AlreadyOpen bool `json:"already_open"`
	Channel     struct {
		ID string `json:"id"`
	} `json:"channel"`
}

func Open(req *OpenRequest) (*OpenResponse, error) {
	res := &OpenResponse{}
	if err := api.Request("im.open", req, true, res, req.Token); err != nil {
		return nil, errors.Wrap(err, "im.open failed")
	}

	return res, nil
}

//go:generate ffjson $GOFILE
