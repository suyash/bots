package api // import "suy.io/bots/slack/api"

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"
)

// ffjson: noencoder
type status struct {
	Ok bool `json:"ok"`
}

type Error struct {
	Description string `json:"error" url:"error"`
}

func (e *Error) Error() string {
	return e.Description
}

var SLACK_API_ROOT = "https://slack.com/api"

// Request can be used to make a raw slack API request
func Request(method string, body interface{}, isJSON bool, res interface{}, token string) error {
	var data []byte
	var err error

	if isJSON {
		data, err = json.Marshal(body)
		if err != nil {
			return errors.Wrap(err, "Request Failed, could not create JSON payload")
		}
	} else {
		v, err := query.Values(body)
		if err != nil {
			return errors.Wrap(err, "Request Failed, could not create URLEncoded payload")
		}

		data = []byte(v.Encode())
	}

	url := SLACK_API_ROOT + "/" + method
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return errors.Wrap(err, "Request Failed, could not create request")
	}

	if isJSON {
		req.Header.Add("Content-Type", "application/json;charset=utf8")
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	apires, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Request Failed, could not send request")
	}

	defer apires.Body.Close()

	rawres, err := ioutil.ReadAll(apires.Body)
	if err != nil {
		return errors.Wrap(err, "Request Failed, could not read response")
	}

	stat := &status{}

	if err := json.Unmarshal(rawres, stat); err != nil {
		return errors.Wrap(err, "Request Failed, could not unmarshal received response type")
	}

	if !stat.Ok {
		se := &Error{}

		if err := json.Unmarshal(rawres, se); err != nil {
			return errors.Wrap(err, "Request Failed, received response was not OK")
		}

		return se
	}

	if res != nil {
		if err := json.Unmarshal(rawres, res); err != nil {
			return errors.Wrap(err, "Request Failed, could not unmarshal received response")
		}
	}

	return nil
}

//go:generate ffjson $GOFILE
