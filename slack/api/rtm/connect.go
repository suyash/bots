package rtm

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type ConnectRequest struct {
	Token string `json:"token" url:"token"`
}

// ffjson: noencoder
type ConnectResponse struct {
	Self struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"self"`
	Team struct {
		Domain string `json:"domain"`
		ID     string `json:"id"`
		Name   string `json:"name"`
	} `json:"team"`
	URL string `json:"url"`
}

func Connect(req *ConnectRequest) (*ConnectResponse, error) {
	res := &ConnectResponse{}
	if err := api.Request("rtm.connect", req, false, res, ""); err != nil {
		return nil, errors.Wrap(err, "rtm.connect failed")
	}

	return res, nil
}

//go:generate ffjson $GOFILE
