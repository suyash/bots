package slack

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/oauth"
	"suy.io/bots/slack/api/rtm"
)

// Bot represents a single slack bot unique to a slack team.
type Bot struct {
	id     string
	teamID string
	token  string

	c     Connector
	convs map[string]*Conversation
	cs    ConversationStore
}

// newBot creates a new Bot from a given slack OAuth Response
func newBot(p *oauth.AccessResponse, c Connector, convs map[string]*Conversation, cs ConversationStore) *Bot {
	return &Bot{
		teamID: p.TeamID,
		token:  p.Bot.BotAccessToken,
		id:     p.Bot.BotUserID,
		c:      c,
		convs:  convs,
		cs:     cs,
	}
}

// BotUser gets the User ID of the Bot
func (bot *Bot) BotUser() string {
	return bot.id
}

// Team gets the ID of the team the bot is in.
func (bot *Bot) Team() string {
	return bot.teamID
}

// Token gets the access token used to create the bot.
func (bot *Bot) Token() string {
	return bot.token
}

// Start opens a WebSocket connection to slack to send and receive messages through this bot.
func (bot *Bot) Start() error {
	res, err := rtm.Connect(&rtm.ConnectRequest{Token: bot.token})
	if err != nil {
		return errors.Wrap(err, "Start Failed")
	}

	if err := bot.c.Add(bot.teamID, res.URL); err != nil {
		return errors.Wrap(err, "Start Failed")
	}

	return nil
}

// Typing sends a typing indicator to this bot's connector instance.
func (bot *Bot) Typing(channel string) error {
	return bot.c.Typing(bot.teamID, channel)
}

// Say sends a message in a channel.
func (bot *Bot) Say(msg *chat.Message) (*chat.Message, error) {
	if msg.Channel == "" {
		return nil, ErrChannelUnset
	}

	res, err := chat.PostMessage(&chat.PostMessageRequest{Message: msg, Token: bot.token})
	if err != nil {
		return nil, errors.Wrap(err, "Say Failed")
	}

	return res.Message, nil
}

// Reply replies to a user message.
func (bot *Bot) Reply(msg *rtm.Message, response *chat.Message) (*chat.Message, error) {
	if msg.ThreadTs != "" {
		response.ThreadTs = msg.ThreadTs
	}

	if msg.Channel == "" {
		return nil, ErrChannelUnset
	}

	response.Channel = msg.Channel

	res, err := bot.Say(response)
	if err != nil {
		return nil, errors.Wrap(err, "Reply Failed")
	}

	return res, nil
}

// SayEphemeral sends an ephemeral message in a chat.
func (bot *Bot) SayEphemeral(msg *chat.EphemeralMessage) (string, error) {
	if msg.Channel == "" {
		return "", ErrChannelUnset
	}

	res, err := chat.PostEphemeral(&chat.PostEphemeralRequest{EphemeralMessage: msg, Token: bot.token})
	if err != nil {
		return "", errors.Wrap(err, "SayEphemeral Failed")
	}

	return res.MessageTs, nil
}

// ReplyEphemeral replies to a message with an ephemeral message.
func (bot *Bot) ReplyEphemeral(msg *rtm.Message, response *chat.EphemeralMessage) (string, error) {
	if msg.ThreadTs != "" {
		response.ThreadTs = msg.ThreadTs
	}

	response.Channel = msg.Channel
	response.User = msg.User

	res, err := bot.SayEphemeral(response)
	if err != nil {
		return "", errors.Wrap(err, "ReplyEphemeral Failed")
	}

	return res, nil
}

// ReplyInThread replies to a message in a thread, creates one if not in a thread.
func (bot *Bot) ReplyInThread(msg *rtm.Message, response *chat.Message) (*chat.Message, error) {
	if msg.ThreadTs == "" {
		response.ThreadTs = msg.Ts
	} else {
		response.ThreadTs = msg.ThreadTs
	}

	response.Channel = msg.Channel

	res, err := bot.Say(response)
	if err != nil {
		return nil, errors.Wrap(err, "ReplyInThread Failed")
	}

	return res, nil
}

// Update updates a message.
func (bot *Bot) Update(ts string, msg *chat.Message) (string, error) {
	msg.Ts = ts

	res, err := chat.Update(&chat.UpdateRequest{Message: msg, Token: bot.token})
	if err != nil {
		return "", errors.Wrap(err, "Update Failed")
	}

	return res.Ts, nil
}

// StartConversation starts the conversation with the given name for the given user
// in the given channel.
func (bot *Bot) StartConversation(user, channel, name string) error {
	c, ok := bot.convs[name]
	if !ok {
		return ErrConversationNotFound
	}

	if bot.cs.IsActive(user, channel, bot.teamID) {
		return ErrConversationAlreadyActive
	}

	if err := bot.cs.Start(user, channel, bot.teamID, name); err != nil {
		errors.Wrap(err, "Could not start")
	}

	c.mp["start"](&chat.Message{Channel: channel}, &Controls{bot, user, channel})
	return nil
}
