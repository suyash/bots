package auth // import "suy.io/bots/slack/api/auth"

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type TestRequest struct {
	Token string `json:"token" url:"token"`
}

// ffjson: noencoder
type TestResponse struct {
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

func Test(req *TestRequest) (*TestResponse, error) {
	res := &TestResponse{}
	if err := api.Request("auth.test", req, false, res, ""); err != nil {
		return nil, errors.Wrap(err, "auth.test failed")
	}

	return res, nil
}

//go:generate ffjson $GOFILE
