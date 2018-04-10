package httpclient // import "suy.io/bots/slack/contrib/connector/httpclient"

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"suy.io/bots/slack"
	"suy.io/bots/slack/connector"
	"suy.io/bots/slack/connector/contrib/httpserver"
)

type Connector struct {
	url  *url.URL
	msgs chan *connector.MessagePayload
}

func NewConnector(host string) (*Connector, error) {
	u, err := url.Parse(host)
	if err != nil || u.Host == "" {
		return nil, errors.New("NewConnector failed, invalid URL")
	}

	return &Connector{
		url:  u,
		msgs: make(chan *connector.MessagePayload),
	}, nil
}

func sendExternalRequest(url *url.URL, data interface{}) error {
	// NOTE: cannot use anonymous struct here, as ffjson would not be able to optimize it.
	body, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Post Failed")
	}

	res, err := http.Post(url.String(), "application/json", bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "Post Failed")
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("Post Failed" + res.Status)
	}

	return nil
}

func (c *Connector) Add(team, socketURL string) error {
	a, err := url.Parse("./slack/add")
	if err != nil {
		return errors.Wrap(err, "Add Failed")
	}

	p := &httpserver.AddPayload{
		Team: team,
		URL:  socketURL,
	}

	if err := sendExternalRequest(c.url.ResolveReference(a), p); err != nil {
		return errors.Wrap(err, "Add Failed")
	}

	return nil
}

// Typing Typing
func (c *Connector) Typing(team, channel string) error {
	t, err := url.Parse("./slack/typing")
	if err != nil {
		return errors.Wrap(err, "Typing Failed")
	}

	p := &httpserver.TypingPayload{
		Team:    team,
		Channel: channel,
	}

	if err := sendExternalRequest(c.url.ResolveReference(t), p); err != nil {
		return errors.Wrap(err, "Typing Failed")
	}

	return nil
}

func (c *Connector) Messages() <-chan *connector.MessagePayload {
	return c.msgs
}

func (c *Connector) Close() {
	close(c.msgs)
	c.msgs = nil
}

func (c *Connector) ServeHTTP(res http.ResponseWriter, req *http.Request) {
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

	p := &connector.MessagePayload{}
	if err := json.Unmarshal(data, p); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	go func() { c.msgs <- p }()
	res.WriteHeader(http.StatusOK)
}

var _ slack.Connector = &Connector{}
var _ http.Handler = &Connector{}
