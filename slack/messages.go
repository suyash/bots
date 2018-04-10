package slack

import (
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/rtm"
)

// ffjson: skip
type MessagePair struct {
	*rtm.Message
	*Bot
}

func (mp *MessagePair) Reply(msg *chat.Message) (*chat.Message, error) {
	return mp.Bot.Reply(mp.Message, msg)
}

func (mp *MessagePair) StartConversation(name string) error {
	return mp.Bot.StartConversation(mp.Message.User, mp.Message.Channel, name)
}

func (mp *MessagePair) Typing() error {
	return mp.Bot.Typing(mp.Channel)
}

// ffjson: skip
type ChannelJoinMessagePair struct {
	*rtm.ChannelJoinMessage
	*Bot
}

// ffjson: skip
type UserChannelJoinMessagePair struct {
	*rtm.UserChannelJoinMessage
	*Bot
}

// ffjson: skip
type GroupJoinMessagePair struct {
	*rtm.GroupJoinMessage
	*Bot
}

// ffjson: skip
type InteractionPair struct {
	*Interaction
	*Bot
}

// ffjson: skip
type InteractionOptionsPair struct {
	*InteractionOptions
	*Bot
}
