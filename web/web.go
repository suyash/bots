// Package web implements custom websocket bots
package web // import "suy.io/bots/web"

import "github.com/pkg/errors"

var (
	ErrNilControllerStore   = errors.New("Controller Store cannot be nil")
	ErrNilConversationStore = errors.New("Conversation Store cannot be nil")
	ErrNilErrorHandler      = errors.New("ErrorHandler cannot be nil")
	ErrNilIDCreator         = errors.New("ID Creator cannot be nil")

	ErrConversationExists        = errors.New("Conversation Already Exists")
	ErrConversationNotFound      = errors.New("Conversation Not Found")
	ErrConversationAlreadyActive = errors.New("Conversation Already Active")
	ErrNoStartState              = errors.New("Conversation Has no start state")
	ErrStateAlreadyExists        = errors.New("State Already Defined")
	ErrNilHandler                = errors.New("Nil Handler")

	ErrIDNotSet       = errors.New("cannot send a message without ID set")
	ErrPrevNextNotSet = errors.New("cannot send a message without either prev or next set")
	ErrSourceNotSet   = errors.New("cannot send a message without source set")
	ErrTypeNotSet     = errors.New("cannot send a message without type set")

	ErrItemNotFound        = errors.New("Item Not Found")
	ErrThreadNotFound      = errors.New("Thread Not Found")
	ErrBotNotFound         = errors.New("Bot Not Found")
	ErrBotAlreadyAdded     = errors.New("Bot Already Added")
	ErrCannotAddThreadZero = errors.New("Cannot explicitly add a thread with ID zero")
)
