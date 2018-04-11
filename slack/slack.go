// Package slack provides utilities for implementing slack bots in go.
package slack // import "suy.io/bots/slack"

import "github.com/pkg/errors"

var (
	ErrInvalidClientID            = errors.New("Invalid Client ID")
	ErrInvalidClientSecret        = errors.New("Invalid Client Secret")
	ErrInvalidVerification        = errors.New("Invalid Verification Token")
	ErrInvalidConnector           = errors.New("Invalid Connector")
	ErrInvalidBotStorage          = errors.New("Invalid Bot Storage")
	ErrInvalidConversationStorage = errors.New("Invalid Conversation Storage")

	ErrConversationExists        = errors.New("Conversation Already Exists")
	ErrConversationNotFound      = errors.New("Conversation Not Found")
	ErrConversationAlreadyActive = errors.New("Conversation Already Active")
	ErrNoStartState              = errors.New("Conversation Has no start state")

	ErrInvalidMessage = errors.New("invalid message")
	ErrChannelUnset   = errors.New("channel is not set")

	ErrExceededResponseCommand = errors.New("Can only respond upto 5 times for a command")

	ErrExceededResponseInteraction = errors.New("Can only respond upto 5 times for an interaction")
	ErrEmptyResponseURLInteraction = errors.New("Empty Response URL")

	ErrStateAlreadyExists = errors.New("State Already Defined")

	ErrBotNotFound     = errors.New("Bot Not Found")
	ErrBotAlreadyAdded = errors.New("Bot Already Added")
	ErrItemNotFound    = errors.New("Item Not Found")
)
