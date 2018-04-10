package httpserver // import "suy.io/bots/slack/connector/contrib/httpserver"

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"suy.io/bots/slack/connector"
)

type Connector struct {
	*connector.Connector
}

func NewConnector(messageURL string) *Connector {
	c := connector.NewConnector()

	c.SetMessageHandler(func(msg []byte, team string) {
		c.Typing(team, "")

		p := &connector.MessagePayload{Message: msg, Team: team}
		data, err := json.Marshal(p)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Could not create Message Payload"))
		}

		res, err := http.Post(messageURL, "application/json", bytes.NewReader(data))
		if err != nil {
			log.Fatal(errors.Wrap(err, "Could not Post Message"))
		}

		if res.StatusCode != http.StatusOK {
			log.Fatal(errors.Wrap(err, "Post Message was not successful"))
		}
	})

	return &Connector{c}
}

type AddPayload struct {
	Team string `json:"team"`
	URL  string `json:"url"`
}

func (c *Connector) AddHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		data, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := &AddPayload{}
		if err := json.Unmarshal(data, p); err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := c.Open(p.Team, p.URL); err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

type TypingPayload struct {
	Team    string `json:"team"`
	Channel string `json:"channel"`
}

func (c *Connector) TypingHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		data, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := &TypingPayload{}
		if err := json.Unmarshal(data, p); err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := c.Typing(p.Team, p.Channel); err != nil {
			if err == connector.ErrBotNotFound {
				http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			} else {
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			return
		}

		res.WriteHeader(http.StatusOK)
	}
}
