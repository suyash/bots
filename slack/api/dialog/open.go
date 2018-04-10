package dialog

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type OpenRequest struct {
	Token     string  `json:"token" url:"token"`
	TriggerID string  `json:"trigger_id" url:"trigger_id"`
	Dialog    *Dialog `json:"dialog" url:"dialog"`
}

func Open(req *OpenRequest) error {
	if err := api.Request("dialog.open", req, true, nil, req.Token); err != nil {
		return errors.Wrap(err, "dialog.open failed")
	}

	return nil
}

//go:generate ffjson $GOFILE
