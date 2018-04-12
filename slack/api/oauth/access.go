package oauth

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type AccessRequest struct {
	ClientID     string `json:"client_id" url:"client_id"`
	ClientSecret string `json:"client_secret" url:"client_secret"`
	Code         string `json:"code" url:"code"`
	Redirect     string `json:"redirect_uri,omitempty" url:"redirect_uri,omitempty"`
}

type AccessResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	UserID      string `json:"user_id"`
	TeamName    string `json:"team_name"`
	TeamID      string `json:"team_id"`
	Bot         *Bot   `json:"bot"`
}

type Bot struct {
	BotUserID      string `json:"bot_user_id"`
	BotAccessToken string `json:"bot_access_token"`
}

func Access(req *AccessRequest) (*AccessResponse, error) {
	res := &AccessResponse{}
	if err := api.Request("oauth.access", req, false, res, ""); err != nil {
		return nil, errors.Wrap(err, "oauth.access failed")
	}

	return res, nil
}

//go:generate ffjson $GOFILE
