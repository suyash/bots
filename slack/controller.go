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

// Controller is essentially a manager for a single slack App.
//
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

// NewController creates a new Controller using the provided functional arguments.
func NewController(options ...func(*Controller) error) (*Controller, error) {
	controller := &Controller{
		conversations: NewConversationRegistry(),
		botAdded:      make(chan *Bot),

		directMessages:  make(chan *MessagePair),
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

// WithClientID can be passed to NewController to set the Slack Client ID.
func WithClientID(id string) func(*Controller) error {
	return func(c *Controller) error {
		if id == "" {
			return ErrInvalidClientID
		}

		c.clientID = id
		return nil
	}
}

// WithClientSecret can be passed to NewController to set the Slack Client Secret.
func WithClientSecret(secret string) func(*Controller) error {
	return func(c *Controller) error {
		if secret == "" {
			return ErrInvalidClientSecret
		}

		c.clientSecret = secret
		return nil
	}
}

// WithVerification can be passed to NewController to set the slack Verification token.
func WithVerification(secret string) func(*Controller) error {
	return func(c *Controller) error {
		if secret == "" {
			return ErrInvalidVerification
		}

		c.verification = secret
		return nil
	}
}

// WithConnector can set a custom Connector instance to manage WebSocket connections.
func WithConnector(conn Connector) func(*Controller) error {
	return func(c *Controller) error {
		if conn == nil {
			return ErrInvalidConnector
		}

		c.connector = conn
		return nil
	}
}

// WithBotStore sets a custom BotStore implementation for storing bot data.
func WithBotStore(b BotStore) func(*Controller) error {
	return func(c *Controller) error {
		if b == nil {
			return ErrInvalidBotStorage
		}

		c.bots = b
		return nil
	}
}

// WithConversationStore sets a custom ConversationStore implementation for storing conversation data.
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

// CreateAddToSlackURL creates a slack OAuth authorize URL.
func (c *Controller) CreateAddToSlackURL(scopes []string, redirect, state string) (string, error) {
	v := make(url.Values)
	v.Add("client_id", c.clientID)
	v.Add("scope", strings.Join(scopes, ","))
	v.Add("redirect_uri", redirect)
	v.Add("state", state)

	return "https://slack.com/oauth/authorize?" + v.Encode(), nil
}

// OAuthHandler returns a http.HandlerFunc that can complete OAuth handshake with slack, creating new Bots.
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
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

// CreateBot adds a new Bot given a slack access token.
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

// BotAdded returns a receive only channel that gets a payload each time a new bot is added.
func (c *Controller) BotAdded() <-chan *Bot {
	return c.botAdded
}

// EventHandler returns a http.HandlerFunc that can listen to slack events.
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

// Event is a placeholder for event key payload in the Event payload
//
// ffjson: noencoder
type Event struct {
	*rtm.Message
	Type    string `json:"type"`
	EventTs string `json:"event_ts"`
}

// EventPayload represents the entire payload received for a slack event.
//
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

// InteractionHandler returns a http.HandlerFunc that can be used to handle interactions from slack.
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

// InteractionOptionsHandler returns a http.HandlerFunc that can handle options requests.
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

// CommandHandler returns a http.HandlerFunc that can handle slack Command requests.
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

// Commands returns a receive only channel that'll send a value each time a new command
// invocation is made.
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

	if strings.HasPrefix(msg.Channel, "D") {
		return c.handleDirectMessage(msg, bot)
	} else if strings.HasPrefix(msg.Text, "<@"+bot.id) {
		return c.handleDirectMention(msg, bot)
	} else if strings.Contains(msg.Text, "<@"+bot.id) {
		return c.handleMention(msg, bot)
	}

	return c.handleAmbientMessage(msg, bot)
}

func (c *Controller) handleDirectMessage(m *rtm.Message, b *Bot) error {
	go func() { c.directMessages <- &MessagePair{m, b} }()
	return nil
}

// DirectMessages returns a receive only channel to get direct messages.
func (c *Controller) DirectMessages() <-chan *MessagePair {
	return c.directMessages
}

func (c *Controller) handleDirectMention(m *rtm.Message, b *Bot) error {
	go func() { c.directMentions <- &MessagePair{m, b} }()
	return nil
}

// DirectMentions returns messages which are sent by mentioning the bot as the first term.
func (c *Controller) DirectMentions() <-chan *MessagePair {
	return c.directMentions
}

func (c *Controller) handleMention(m *rtm.Message, b *Bot) error {
	go func() { c.mentions <- &MessagePair{m, b} }()
	return nil
}

// Mentions are messages where the bot is mentioned somewhere in the middle.
func (c *Controller) Mentions() <-chan *MessagePair {
	return c.mentions
}

func (c *Controller) handleAmbientMessage(m *rtm.Message, b *Bot) error {
	go func() { c.ambientMessages <- &MessagePair{m, b} }()
	return nil
}

// AmbientMessages are messages in conversations that the bot is in, but not mentioned in the message.
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

// ChannelJoin sends a payload each time the bot is added to a new channel.
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

// UserChannelJoin sends a payload each time a user joins a new channel.
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

// GroupJoin sends a payload each time a user joins a group chat.
func (c *Controller) GroupJoin() <-chan *GroupJoinMessagePair {
	return c.groupJoin
}

//
// Interactions
//

// Interactions returns a payload for each new interaction
func (c *Controller) Interactions() <-chan *InteractionPair {
	return c.interactions
}

// InteractionOptions returns a payload each time a new option is added.
func (c *Controller) InteractionOptions() <-chan *InteractionOptionsPair {
	return c.interactionOptions
}

// RegisterConversation registers a new conversation with the bot.
func (c *Controller) RegisterConversation(name string, conv *Conversation) error {
	return c.conversations.Add(name, conv)
}

//go:generate ffjson $GOFILE
