package slack

import (
	"github.com/pkg/errors"

	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/oauth"
	"suy.io/bots/slack/api/rtm"
)

// ffjson: skip
type Bot struct {
	id     string
	teamID string
	token  string

	c     Connector
	convs map[string]*Conversation
	cs    ConversationStore
}

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

func (bot *Bot) BotUser() string {
	return bot.id
}

func (bot *Bot) Team() string {
	return bot.teamID
}

func (bot *Bot) Token() string {
	return bot.token
}

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

func (bot *Bot) Typing(channel string) error {
	return bot.c.Typing(bot.teamID, channel)
}

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

func (bot *Bot) Update(ts string, msg *chat.Message) (string, error) {
	msg.Ts = ts

	res, err := chat.Update(&chat.UpdateRequest{Message: msg, Token: bot.token})
	if err != nil {
		return "", errors.Wrap(err, "Update Failed")
	}

	return res.Ts, nil
}

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
