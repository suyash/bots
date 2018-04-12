package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"

	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/oauth"
	"suy.io/bots/slack/api/rtm"
	"suy.io/bots/slack/api/team"
)

// ffjson: skip
type Controller struct {
	clientID     string
	clientSecret string
	verification string

	connector     Connector
	bots          BotStore
	conversations ConversationRegistry
	cs            ConversationStore
	botAdded      chan *Bot

	directMessages  chan *MessagePair
	selfMessages    chan *MessagePair
	directMentions  chan *MessagePair
	mentions        chan *MessagePair
	ambientMessages chan *MessagePair
	channelJoin     chan *ChannelJoinMessagePair
	userChannelJoin chan *UserChannelJoinMessagePair
	groupJoin       chan *GroupJoinMessagePair

	interactions       chan *InteractionPair
	interactionOptions chan *InteractionOptionsPair

	commands chan *Command
}

func NewController(options ...func(*Controller) error) (*Controller, error) {
	controller := &Controller{
		conversations: NewConversationRegistry(),
		botAdded:      make(chan *Bot),

		directMessages:  make(chan *MessagePair),
		selfMessages:    make(chan *MessagePair),
		directMentions:  make(chan *MessagePair),
		mentions:        make(chan *MessagePair),
		ambientMessages: make(chan *MessagePair),
		channelJoin:     make(chan *ChannelJoinMessagePair),
		userChannelJoin: make(chan *UserChannelJoinMessagePair),
		groupJoin:       make(chan *GroupJoinMessagePair),

		interactions:       make(chan *InteractionPair),
		interactionOptions: make(chan *InteractionOptionsPair),

		commands: make(chan *Command),
	}

	for _, opt := range options {
		if err := opt(controller); err != nil {
			return nil, errors.Wrap(err, "NewController Failed")
		}
	}

	if controller.bots == nil {
		controller.bots = NewMemoryBotStore()
	}

	if controller.cs == nil {
		controller.cs = NewMemoryConversationStore()
	}

	if controller.connector == nil {
		controller.connector = newInternalConnector()
	}

	// load cached bots from storage

	bots, err := controller.bots.AllBots()
	if err != nil {
		return nil, errors.Wrap(err, "NewController Failed")
	}

	for _, b := range bots {
		bot := newBot(b, controller.connector, controller.conversations, controller.cs)
		if err := bot.Start(); err != nil {
			return nil, errors.Wrap(err, "NewController Failed")
		}
	}

	go controller.listen()
	return controller, nil
}

func WithClientID(id string) func(*Controller) error {
	return func(c *Controller) error {
		if id == "" {
			return ErrInvalidClientID
		}

		c.clientID = id
		return nil
	}
}

func WithClientSecret(secret string) func(*Controller) error {
	return func(c *Controller) error {
		if secret == "" {
			return ErrInvalidClientSecret
		}

		c.clientSecret = secret
		return nil
	}
}

func WithVerification(secret string) func(*Controller) error {
	return func(c *Controller) error {
		if secret == "" {
			return ErrInvalidVerification
		}

		c.verification = secret
		return nil
	}
}

func WithConnector(conn Connector) func(*Controller) error {
	return func(c *Controller) error {
		if conn == nil {
			return ErrInvalidConnector
		}

		c.connector = conn
		return nil
	}
}

func WithBotStore(b BotStore) func(*Controller) error {
	return func(c *Controller) error {
		if b == nil {
			return ErrInvalidBotStorage
		}

		c.bots = b
		return nil
	}
}

func WithConversationStore(cs ConversationStore) func(*Controller) error {
	return func(c *Controller) error {
		if cs == nil {
			return ErrInvalidConversationStorage
		}

		c.cs = cs
		return nil
	}
}

func (c *Controller) listen() {
	for msg := range c.connector.Messages() {
		c.handleMessage(msg.Message, msg.Team)
	}
}

func (c *Controller) CreateAddToSlackURL(scopes []string, redirect, state string) (string, error) {
	v := make(url.Values)
	v.Add("client_id", c.clientID)
	v.Add("scope", strings.Join(scopes, ","))
	v.Add("redirect_uri", redirect)
	v.Add("state", state)

	return "https://slack.com/oauth/authorize?" + v.Encode(), nil
}

func (c *Controller) CreateAddToSlackButton(scopes []string, redirect, state string) (string, error) {
	url, err := c.CreateAddToSlackURL(scopes, redirect, state)
	if err != nil {
		return "", errors.Wrap(err, "CreateAddToSlackButton Failed")
	}

	return `<a href="` + url + `"><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcset="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a>`, nil
}

func (c *Controller) OAuthHandler(redirect, expectedState string, onSuccess func(*oauth.AccessResponse, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		currentState := req.URL.Query().Get("state")

		if currentState != expectedState {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		payload, err := oauth.Access(&oauth.AccessRequest{c.clientID, c.clientSecret, code, redirect})
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := c.bots.AddBot(payload); err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		b := newBot(payload, c.connector, c.conversations, c.cs)
		go func() { c.botAdded <- b }()

		if onSuccess != nil {
			onSuccess(payload, res, req)
		} else {
			res.WriteHeader(http.StatusOK)
		}
	}
}

func (c *Controller) CreateBot(token string) (*Bot, error) {
	info, err := team.Info(&team.InfoRequest{Token: token})
	if err != nil {
		return nil, errors.Wrap(err, "CreateBot Failed")
	}

	teamID := info.Team.ID
	payload := &oauth.AccessResponse{TeamID: teamID, Bot: &oauth.Bot{BotAccessToken: token}}
	if err := c.bots.AddBot(payload); err != nil {
		return nil, errors.Wrap(err, "CreateBot Failed")
	}

	b := newBot(payload, c.connector, c.conversations, c.cs)
	go func() { c.botAdded <- b }()
	return b, nil
}

func (c *Controller) BotAdded() <-chan *Bot {
	return c.botAdded
}

func (c *Controller) EventHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		t := &typ{}

		if err := json.Unmarshal(d, t); err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if c.verification != "" && t.Token != c.verification {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		switch t.Type {
		case "url_verification":
			fmt.Fprint(res, t.Challenge)
		case "event_callback":
			if err := c.handleEvent(d); err != nil {
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			res.WriteHeader(http.StatusOK)
		}
	}
}

// ffjson: noencoder
type Event struct {
	*rtm.Message
	Type    string `json:"type"`
	EventTs string `json:"event_ts"`
}

// ffjson: noencoder
type EventPayload struct {
	Token       string   `json:"token"`
	TeamID      string   `json:"team_id"`
	APIAppID    string   `json:"api_app_id"`
	Event       *Event   `json:"event"`
	Type        string   `json:"type"`
	EventID     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
	AuthedUsers []string `json:"authed_users"`
}

func (c *Controller) handleEvent(data []byte) error {
	p := &EventPayload{
		Event: &Event{
			Message: &rtm.Message{},
		},
	}

	if err := json.Unmarshal(data, p); err != nil {
		return errors.Wrap(err, "Could not handle event")
	}

	payload, err := c.bots.GetBot(p.TeamID)
	if err != nil {
		return errors.Wrap(err, "Could not handle event")
	}

	b := newBot(payload, c.connector, c.conversations, c.cs)
	switch p.Event.Type {
	case "message":
		if err := c.handleNormalMessage(p.Event.Message, b); err != nil {
			return errors.Wrap(err, "Could not handle event")
		}
	}

	return nil
}

// ffjson: noencoder
type interaction struct {
	*Interaction
	Token string `json:"token"`
}

func (c *Controller) InteractionHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		d, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		values, err := url.ParseQuery(string(d))
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		iact := &interaction{&Interaction{}, ""}

		if err := json.Unmarshal([]byte(values.Get("payload")), iact); err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if c.verification != "" && iact.Token != c.verification {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		payload, err := c.bots.GetBot(iact.Team.ID)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		iact.immediateResponse, iact.token = make(chan []byte), payload.AccessToken
		b := newBot(payload, c.connector, c.conversations, c.cs)
		go func() { c.interactions <- &InteractionPair{iact.Interaction, b} }()

		select {
		case m := <-iact.immediateResponse:
			if m != nil {
				close(iact.immediateResponse)
				res.Header().Set("Content-Type", "application/json")
				fmt.Fprint(res, string(m))
			} else {
				res.WriteHeader(http.StatusOK)
			}
		case <-time.After(InteractionImmediateResponseTimeout):
			res.WriteHeader(http.StatusOK)
		}
	}
}

// ffjson: noencoder
type interactionOptions struct {
	*InteractionOptions
	Token string `json:"token"`
}

// NOTE: this blocks calling thread
func (c *Controller) InteractionOptionsHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		values, err := url.ParseQuery(string(d))
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// ssl_check=1&token=YslVmQnOJtw9lSwayF6mQKLn
		if values.Get("ssl_check") != "" {
			if c.verification != "" && values.Get("token") != c.verification {
				http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			} else {
				res.WriteHeader(http.StatusOK)
			}

			return
		}

		iactopt := &interactionOptions{&InteractionOptions{}, ""}

		if err := json.Unmarshal([]byte(values.Get("payload")), iactopt); err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if c.verification != "" && iactopt.Token != c.verification {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		payload, err := c.bots.GetBot(iactopt.Team.ID)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		iactopt.immediateResponse = make(chan []byte)
		b := newBot(payload, c.connector, c.conversations, c.cs)
		go func() { c.interactionOptions <- &InteractionOptionsPair{iactopt.InteractionOptions, b} }()

		m := <-iactopt.immediateResponse

		close(iactopt.immediateResponse)
		res.Header().Set("Content-Type", "application/json")
		fmt.Fprint(res, string(m))
	}
}

func (c *Controller) CommandHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer req.Body.Close()

		values, err := url.ParseQuery(string(d))
		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if c.verification != "" && values.Get("token") != c.verification {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if values.Get("response_url") == "" {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		command := newCommand(values)
		go func() { c.commands <- command }()

		select {
		case m := <-command.immediateResponse:
			if m != nil {
				close(command.immediateResponse)
				res.Header().Set("Content-Type", "application/json")
				fmt.Fprint(res, string(m))
			} else {
				res.WriteHeader(http.StatusOK)
			}
		case <-time.After(CommandImmediateResponseTimeout):
			res.WriteHeader(http.StatusOK)
		}
	}
}

func (c *Controller) Commands() <-chan *Command {
	if c.commands == nil {
		c.commands = make(chan *Command)
	}

	return c.commands
}

//
// Messages
//

// ffjson: noencoder
type typ struct {
	SubType   string `json:"subtype"`
	Type      string `json:"type"`
	User      string `json:"user"`
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
}

func (c *Controller) handleMessage(msg []byte, team string) error {
	t := &typ{}
	if err := json.Unmarshal(msg, t); err != nil {
		return errors.Wrap(err, "Could not handle message")
	}

	switch t.Type {
	case "message":
		return c.handleMessageType(msg, t.SubType, t.User, team)
	default:
		return nil
	}
}

// handleMessageType handles the case where the type field of an RTM command is 'message'
// It categorizes the message into 8 types, each of which have their respective channels
// The categories are
//
// - DirectMessage
// - SelfMessage
// - DirectMention
// - Mention
// - Ambient
// - ChannelJoin
// - UserChannelJoin
// - GroupJoin
func (c *Controller) handleMessageType(msg []byte, subtype, user, team string) error {
	payload, err := c.bots.GetBot(team)
	if err != nil {
		return errors.Wrap(err, "Could not handle Message Type")
	}

	bot := newBot(payload, c.connector, c.conversations, c.cs)

	if subtype != "" {
		if subtype == "channel_join" {
			if user != bot.id {
				c.handleUserChannelJoin(msg, bot)
			} else {
				c.handleChannelJoin(msg, bot)
			}
		} else if subtype == "group_join" {
			c.handleGroupJoin(msg, bot)
		}

		// NOTE: this helps us skip handling a message when subtype is
		// - "message_changed"
		return nil
	}

	m := &rtm.Message{}
	if err := json.Unmarshal(msg, m); err != nil {
		return errors.Wrap(err, "Could not handle Message Type")
	}

	c.handleNormalMessage(m, bot)
	return nil
}

func (c *Controller) handleNormalMessage(msg *rtm.Message, bot *Bot) error {
	if id, state, err := c.cs.Active(msg.User, msg.Channel, msg.Team); err == nil {
		conv, err := c.conversations.Get(id)
		if err != nil {
			return err
		}

		conv.mp[state](&chat.Message{
			Channel:  msg.Channel,
			Text:     msg.Text,
			Ts:       msg.Ts,
			ThreadTs: msg.ThreadTs,
		}, &Controls{
			bot, msg.User, msg.Channel,
		})

		return nil
	}

	if msg.User == bot.id {
		return c.handleSelfMessage(msg, bot)
	} else if strings.HasPrefix(msg.Channel, "D") {
		return c.handleDirectMessage(msg, bot)
	} else if strings.HasPrefix(msg.Text, "<@"+bot.id) {
		return c.handleDirectMention(msg, bot)
	} else if strings.Contains(msg.Text, "<@"+bot.id) {
		return c.handleMention(msg, bot)
	} else {
		return c.handleAmbientMessage(msg, bot)
	}
}

func (c *Controller) handleDirectMessage(m *rtm.Message, b *Bot) error {
	go func() { c.directMessages <- &MessagePair{m, b} }()
	return nil
}

func (c *Controller) DirectMessages() <-chan *MessagePair {
	return c.directMessages
}

func (c *Controller) handleSelfMessage(m *rtm.Message, b *Bot) error {
	go func() { c.selfMessages <- &MessagePair{m, b} }()
	return nil
}

func (c *Controller) SelfMessages() <-chan *MessagePair {
	return c.selfMessages
}

func (c *Controller) handleDirectMention(m *rtm.Message, b *Bot) error {
	go func() { c.directMentions <- &MessagePair{m, b} }()
	return nil
}

func (c *Controller) DirectMentions() <-chan *MessagePair {
	return c.directMentions
}

func (c *Controller) handleMention(m *rtm.Message, b *Bot) error {
	go func() { c.mentions <- &MessagePair{m, b} }()
	return nil
}

func (c *Controller) Mentions() <-chan *MessagePair {
	return c.mentions
}

func (c *Controller) handleAmbientMessage(m *rtm.Message, b *Bot) error {
	go func() { c.ambientMessages <- &MessagePair{m, b} }()
	return nil
}

func (c *Controller) AmbientMessages() <-chan *MessagePair {
	return c.ambientMessages
}

func (c *Controller) handleChannelJoin(msg []byte, b *Bot) {
	m := &rtm.ChannelJoinMessage{}
	if err := json.Unmarshal(msg, m); err != nil {
		return
	}

	go func() { c.channelJoin <- &ChannelJoinMessagePair{m, b} }()
}

func (c *Controller) ChannelJoin() <-chan *ChannelJoinMessagePair {
	return c.channelJoin
}

func (c *Controller) handleUserChannelJoin(msg []byte, bot *Bot) {
	m := &rtm.UserChannelJoinMessage{}
	if err := json.Unmarshal(msg, m); err != nil {
		return
	}

	go func() { c.userChannelJoin <- &UserChannelJoinMessagePair{m, bot} }()
}

func (c *Controller) UserChannelJoin() <-chan *UserChannelJoinMessagePair {
	return c.userChannelJoin
}

func (c *Controller) handleGroupJoin(msg []byte, bot *Bot) {
	m := &rtm.GroupJoinMessage{}
	if err := json.Unmarshal(msg, m); err != nil {
		return
	}

	go func() { c.groupJoin <- &GroupJoinMessagePair{m, bot} }()
}

func (c *Controller) GroupJoin() <-chan *GroupJoinMessagePair {
	return c.groupJoin
}

//
// Interactions
//

func (c *Controller) Interactions() <-chan *InteractionPair {
	return c.interactions
}

func (c *Controller) InteractionOptions() <-chan *InteractionOptionsPair {
	return c.interactionOptions
}

func (c *Controller) RegisterConversation(name string, conv *Conversation) error {
	return c.conversations.Add(name, conv)
}

//go:generate ffjson $GOFILE
