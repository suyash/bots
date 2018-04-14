package web

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestNewController(t *testing.T) {
	type args struct {
		options []func(*Controller) error
	}

	tests := []struct {
		name    string
		args    args
		want    *Controller
		wantErr bool
	}{
		{"", args{}, &Controller{
			botAdded:      nil,
			sanitizer:     nil,
			cs:            nil,
			conversations: nil,
			convs:         nil,
			messages:      nil,
			idCreator:     nil,
			errHandler:    nil,
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewController(tt.args.options...)
			got.botAdded = nil
			got.sanitizer = nil
			got.cs = nil
			got.conversations = nil
			got.convs = nil
			got.messages = nil
			got.idCreator = nil
			got.errHandler = nil

			if (err != nil) != tt.wantErr {
				t.Errorf("NewController() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewController() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithControllerStore(t *testing.T) {
	type args struct {
		store ControllerStore
	}

	tests := []struct {
		name string
		args args
	}{
		{"", args{NewMemoryControllerStore()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewController(WithControllerStore(tt.args.store))
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(c.cs, tt.args.store) {
				t.Errorf("WithControllerStore() = %v, want %v", c.cs, tt.args.store)
			}
		})
	}
}

func TestWithConversationStore(t *testing.T) {
	type args struct {
		store ConversationStore
	}

	tests := []struct {
		name string
		args args
	}{
		{"", args{NewMemoryConversationStore()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewController(WithConversationStore(tt.args.store))
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(c.convs, tt.args.store) {
				t.Errorf("WithConversationStore() = %v, want %v", c.cs, tt.args.store)
			}
		})
	}
}

func TestWithBotIDCreator(t *testing.T) {
	type args struct {
		f BotIDCreator
	}

	tests := []struct {
		name string
		args args
		want BotID
	}{
		{"", args{func(*http.Request) (BotID, error) { return BotID(21), nil }}, 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewController(WithBotIDCreator(tt.args.f))
			if err != nil {
				t.Error(err)
			}

			if got, err := c.idCreator(nil); err != nil || got != tt.want {
				t.Errorf("WithBotIDCreator() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: figure this out
//
// func TestWithErrorHandler(t *testing.T) {
// 	type args struct {
// 		f ErrorHandler
// 	}

// 	tests := []struct {
// 		name string
// 		args args
// 	}{}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := WithErrorHandler(tt.args.f); !reflect.DeepEqual(got, tt.args) {
// 				t.Errorf("WithErrorHandler() = %v, want %v", 0, 0)
// 			}
// 		})
// 	}
// }

func TestController_ConnectionHandler(t *testing.T) {
	c, err := NewController(
		WithErrorHandler(func(err error) {
			if !strings.Contains(err.Error(), "WebSocket Unexpected Close") {
				t.Fatal(err)
			}
		}),
	)

	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		c          *Controller
		req        *http.Request
		wantStatus int
		wantBody   []byte
	}{
		{"", c, httptest.NewRequest(http.MethodPost, "/", nil), http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", c, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusBadRequest, []byte("Bad Request\n")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			tt.c.ConnectionHandler()(res, tt.req)

			if res.Code != tt.wantStatus {
				t.Errorf("Controller.ConnectionHandler() = %v, want %v", res.Code, tt.wantStatus)
			}

			if !reflect.DeepEqual(res.Body.Bytes(), tt.wantBody) {
				t.Errorf("Controller.ConnectionHandler() = %v, want %v", res.Body.Bytes(), tt.wantBody)
			}
		})
	}

	s := httptest.NewServer(c.ConnectionHandler())

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Connection", "Upgrade")
	req.Header.Add("Upgrade", "websocket")
	req.Header.Add("Sec-Websocket-Version", "13")
	req.Header.Add("Sec-Websocket-Key", "13")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusSwitchingProtocols {
		t.Fatal("not upgrading")
	}
}

func TestController_addBot(t *testing.T) {
	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		botID BotID
		conn  *websocket.Conn
		req   *http.Request
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		want    *Bot
		wantErr bool
	}{
		{"", c, args{42, &websocket.Conn{}, httptest.NewRequest(http.MethodGet, "/", nil)}, &Bot{
			id:   42,
			conn: &websocket.Conn{},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.addBot(tt.args.botID, tt.args.conn, tt.args.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.addBot() = %v, want %v", err, nil)
			}

			if !reflect.DeepEqual(got.id, tt.want.id) {
				t.Errorf("Controller.addBot() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(got.conn, tt.want.conn) {
				t.Errorf("Controller.addBot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_removeBot(t *testing.T) {
	s, _ := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	c.addBot(21, conn, httptest.NewRequest(http.MethodGet, "/", nil))

	type args struct {
		bot *Bot
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{&Bot{id: 21, conn: conn, outgoingMessages: make(chan Item)}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.removeBot(tt.args.bot); (err != nil) != tt.wantErr {
				t.Errorf("Controller.removeBot() = %v, want %v", err, nil)
			}
		})
	}
}

func TestController_RegisterConversation(t *testing.T) {
	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		name string
		conv *Conversation
	}

	conv := NewConversation()
	conv.On("start", func(*Message, *Controls) {})

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{"test", NewConversation()}, true},
		{"", c, args{"test", conv}, false},
		{"", c, args{"test", conv}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.RegisterConversation(tt.args.name, tt.args.conv); (err != nil) != tt.wantErr {
				t.Errorf("Controller.RegisterConversation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
