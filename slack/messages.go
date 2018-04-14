package slack

import (
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/rtm"
)

// MessagePair is sent as the payload when a new message is received.
// ffjson: skip
type MessagePair struct {
	*rtm.Message
	*Bot
}

// Reply replies with the passed message to the original message with the pair's bot.
func (mp *MessagePair) Reply(msg *chat.Message) (*chat.Message, error) {
	return mp.Bot.Reply(mp.Message, msg)
}

// StartConversation starts a conversation with the pair's bot.
func (mp *MessagePair) StartConversation(name string) error {
	return mp.Bot.StartConversation(mp.Message.User, mp.Message.Channel, name)
}

// Typing sends a typing indicator.
func (mp *MessagePair) Typing() error {
	return mp.Bot.Typing(mp.Channel)
}

// ChannelJoinMessagePair is sent for channel join events.
//
// ffjson: skip
type ChannelJoinMessagePair struct {
	*rtm.ChannelJoinMessage
	*Bot
}

// UserChannelJoinMessagePair is sent for user channel join events.
//
// ffjson: skip
type UserChannelJoinMessagePair struct {
	*rtm.UserChannelJoinMessage
	*Bot
}

// GroupJoinMessagePair is sent for group join events.
//
// ffjson: skip
type GroupJoinMessagePair struct {
	*rtm.GroupJoinMessage
	*Bot
}

// InteractionPair is sent for interactions
//
// ffjson: skip
type InteractionPair struct {
	*Interaction
	*Bot
}

// InteractionOptionsPair is sent for interaction options.
//
// ffjson: skip
type InteractionOptionsPair struct {
	*InteractionOptions
	*Bot
}
