package team // import "suy.io/bots/slack/api/team"

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
)

// ffjson: nodecoder
type InfoRequest struct {
	Token string `json:"token" url:"token"`
}

// ffjson: noencoder
type InfoResponse struct {
	Team struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Domain      string `json:"domain"`
		EmailDomain string `json:"email_domain"`
		Icon        struct {
			Image34      string `json:"image_34"`
			Image44      string `json:"image_44"`
			Image68      string `json:"image_68"`
			Image88      string `json:"image_88"`
			Image102     string `json:"image_102"`
			Image132     string `json:"image_132"`
			ImageDefault bool   `json:"image_default"`
		} `json:"icon"`
		EnterpriseID   string `json:"enterprise_id"`
		EnterpriseName string `json:"enterprise_name"`
	} `json:"team"`
}

func Info(req *InfoRequest) (*InfoResponse, error) {
	res := &InfoResponse{}
	if err := api.Request("team.info", req, false, res, ""); err != nil {
		return nil, errors.Wrap(err, "team.info failed")
	}

	return res, nil
}

//go:generate ffjson $GOFILE
