package slack

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"suy.io/bots/slack/api"
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/oauth"
	"suy.io/bots/slack/api/rtm"
)

func TestNewController(t *testing.T) {
	type args struct {
		options []func(*Controller) error
	}

	c := newInternalConnector()

	tests := []struct {
		name    string
		args    args
		want    *Controller
		wantErr bool
	}{
		{"", args{[]func(*Controller) error{WithConnector(c)}}, &Controller{
			conversations: make(map[string]*Conversation),
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

			bots:      NewMemoryBotStore(),
			cs:        NewMemoryConversationStore(),
			connector: c,
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewController(tt.args.options...)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewController() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.want.botAdded = nil
			got.botAdded = nil

			tt.want.directMessages = nil
			got.directMessages = nil

			tt.want.directMentions = nil
			got.directMentions = nil

			tt.want.mentions = nil
			got.mentions = nil

			tt.want.ambientMessages = nil
			got.ambientMessages = nil

			tt.want.channelJoin = nil
			got.channelJoin = nil

			tt.want.userChannelJoin = nil
			got.userChannelJoin = nil

			tt.want.groupJoin = nil
			got.groupJoin = nil

			tt.want.interactions = nil
			got.interactions = nil

			tt.want.interactionOptions = nil
			got.interactionOptions = nil

			tt.want.commands = nil
			got.commands = nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewController() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: this is done wrong, see TestWithConnector for the proper way to test
func TestWithClientID(t *testing.T) {
	type args struct {
		id string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{"2222.4444"}, false},
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithClientID(tt.args.id)(c); (got != nil) != tt.wantErr || c.clientID != tt.args.id {
				t.Errorf("WithClientID() = %v, want %v", tt.args.id, c.clientID)
			}
		})
	}
}

// TODO: this is done wrong, see TestWithConnector for the proper way to test
func TestWithClientSecret(t *testing.T) {
	type args struct {
		secret string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{"abcddcba"}, false},
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithClientSecret(tt.args.secret)(c); (got != nil) != tt.wantErr || c.clientSecret != tt.args.secret {
				t.Errorf("WithClientSecret() = %v, want %v", c.clientSecret, tt.args.secret)
			}
		})
	}
}

// TODO: this is done wrong, see TestWithConnector for the proper way to test
func TestWithVerification(t *testing.T) {
	type args struct {
		secret string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{"qwertyui"}, false},
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithVerification(tt.args.secret)(c); (got != nil) != tt.wantErr || tt.args.secret != c.verification {
				t.Errorf("WithVerification() = %v, want %v", tt.args.secret, c.verification)
			}
		})
	}
}

func TestWithConnector(t *testing.T) {
	type args struct {
		conn Connector
	}

	conn := newInternalConnector()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{conn}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewController(WithConnector(tt.args.conn))
			if err != nil {
				t.Fatal(err)
			}

			if (err != nil) != tt.wantErr || !reflect.DeepEqual(tt.args.conn, c.connector) {
				t.Errorf("WithConnector() = %v, want %v", 0, 0)
			}
		})
	}
}

// TODO: this is done wrong, see TestWithConnector for the proper way to test
func TestWithBotStore(t *testing.T) {
	type args struct {
		b BotStore
	}

	s := NewMemoryBotStore()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{s}, false},
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithBotStore(tt.args.b)(c); (got != nil) != tt.wantErr || !reflect.DeepEqual(tt.args.b, c.bots) {
				t.Errorf("WithBotStore() = %v, want %v", 0, 0)
			}
		})
	}
}

// TODO: this is done wrong, see TestWithConnector for the proper way to test
func TestWithConversationStore(t *testing.T) {
	type args struct {
		cs ConversationStore
	}

	cs := NewMemoryConversationStore()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{cs}, false},
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithConversationStore(tt.args.cs)(c); (got != nil) != tt.wantErr || !reflect.DeepEqual(tt.args.cs, c.cs) {
				t.Errorf("WithConversationStore() = %v, want %v", 0, 0)
			}
		})
	}
}

// TODO: figure this out
//
// func TestController_listen(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		c    *Controller
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.c.listen()
// 		})
// 	}
// }

// TODO: make these work for arbitrary order of query parameters
//
// func TestController_CreateAddToSlackURL(t *testing.T) {
// 	type args struct {
// 		scopes   []string
// 		redirect string
// 		state    string
// 	}

// 	c, err := NewController(WithClientID("2222.2222"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	tests := []struct {
// 		name    string
// 		c       *Controller
// 		args    args
// 		want    string
// 		wantErr bool
// 	}{
// 		{"", c, args{[]string{"bot", "commands"}, "https://a.a", "test"}, "https://slack.com/oauth/authorize?client_id=2222.2222&redirect_uri=https%3A%2F%2Fa.a&response_type=code&scope=bot+commands&state=test", false},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.c.CreateAddToSlackURL(tt.args.scopes, tt.args.redirect, tt.args.state)

// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Controller.CreateAddToSlackURL() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if got != tt.want {
// 				t.Errorf("Controller.CreateAddToSlackURL() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestController_CreateAddToSlackButton(t *testing.T) {
// 	type args struct {
// 		scopes   []string
// 		redirect string
// 		state    string
// 	}

// 	c, err := NewController(WithClientID("2222.2222"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	tests := []struct {
// 		name    string
// 		c       *Controller
// 		args    args
// 		want    string
// 		wantErr bool
// 	}{
// 		{"", c, args{[]string{"bot", "commands"}, "https://a.a", "test"}, `<a href="https://slack.com/oauth/authorize?client_id=2222.2222&redirect_uri=https%3A%2F%2Fa.a&response_type=code&scope=bot+commands&state=test"><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcset="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a>`, false},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.c.CreateAddToSlackButton(tt.args.scopes, tt.args.redirect, tt.args.state)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Controller.CreateAddToSlackButton() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("Controller.CreateAddToSlackButton() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestController_OAuthHandler(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		d, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		t.Log(string(d))

		c, err := url.ParseQuery(string(d))
		if err != nil {
			t.Fatal(err)
		}

		if c.Get("code") == "" {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		fmt.Fprint(res, `{"ok":true,"access_token":"90d64460d14870c08c81352a05dedd3465940a7c","team_id":"T1234","bot":{"bot_access_token":"xoxb-bob-lob-law","bot_user_id":"U1234"}}`)
	}))

	api.SLACK_API_ROOT = s.URL

	type args struct {
		redirect      string
		expectedState string
		onSuccess     func(*oauth.AccessResponse, http.ResponseWriter, *http.Request)
	}

	c, err := NewController(
		WithClientID("2222.2222"),
		WithClientSecret("aaaaaaaa"),
	)

	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		c          *Controller
		args       args
		req        *http.Request
		wantStatus int
	}{
		{"", c, args{"https://redirect.com", "bob-lob-law", nil}, httptest.NewRequest("GET", "/", nil), http.StatusUnauthorized},
		{"", c, args{"https://redirect.com", "bob-lob-law", nil}, httptest.NewRequest("GET", "/?state=bob-lob-law", nil), http.StatusInternalServerError},
		{"", c, args{"https://redirect.com", "bob-lob-law", nil}, httptest.NewRequest("GET", "/?state=bob-lob-law&code=asddsa", nil), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			tt.c.OAuthHandler(tt.args.redirect, tt.args.expectedState, tt.args.onSuccess)(res, tt.req)

			if res.Code != tt.wantStatus {
				t.Errorf("Controller.OAuthHandler() = %v, want %v", res.Code, tt.wantStatus)
			}
		})
	}
}

func TestController_CreateBot(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/team.info" {
			t.Fatalf("expected path to be %v, got %v", "team.info", req.URL.Path)
		}

		d, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		// empty
		if string(d) == "token=" {
			fmt.Fprint(res, `{"ok":false}`)
		} else {
			fmt.Fprint(res, `{"ok":true,"team":{"id":"T12345678"}}`)
		}
	}))

	type args struct {
		token string
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		c        *Controller
		args     args
		want     *Bot
		wantErr  bool
		override bool
	}{
		{"", c, args{"xoxb-bob-lob-law"}, nil, true, false},
		{
			"",
			c,
			args{"xoxb-bob-lob-law"},
			&Bot{
				token:  "xoxb-bob-lob-law",
				teamID: "T12345678",
				c:      c.connector,
				convs:  c.conversations,
				cs:     c.cs,
			},
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.override {
				api.SLACK_API_ROOT = s.URL
			}

			got, err := tt.c.CreateBot(tt.args.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("Controller.CreateBot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Controller.CreateBot() = %v, want %v", got, tt.want)
			}

			if !tt.wantErr {
				b := <-c.botAdded

				if !reflect.DeepEqual(b, tt.want) {
					t.Errorf("Controller.CreateBot() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TODO: kinda already done in above test case, for now.
//
// func TestController_BotAdded(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		c    *Controller
// 		want <-chan *Bot
// 	}{
// 		// TODO: Add test cases.
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := tt.c.BotAdded(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Controller.BotAdded() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestController_EventHandler(t *testing.T) {
	c, err := NewController(WithVerification("bob-lob-law"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantBody   []byte
	}{
		{"", httptest.NewRequest(http.MethodGet, "/", nil), http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", nil), http.StatusInternalServerError, []byte("Internal Server Error\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"token":"foo-bar-baz"}`)), http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"token":"bob-lob-law","type":"url_verification","challenge":"bob-lob-law"}`)), http.StatusOK, []byte("bob-lob-law")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"token":"bob-lob-law","type":"event_callback","challenge":"bob-lob-law"}`)), http.StatusInternalServerError, []byte("Internal Server Error\n")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			c.EventHandler()(res, tt.req)

			if res.Code != tt.wantStatus {
				t.Errorf("Controller.EventHandler() = %v, want %v", res.Code, tt.wantStatus)
			}

			if !reflect.DeepEqual(res.Body.Bytes(), tt.wantBody) {
				t.Errorf("Controller.CreateBot() = %v, want %v", string(res.Body.Bytes()), string(tt.wantBody))
			}
		})
	}
}

func TestController_handleEvent(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/team.info" {
			t.Fatalf("expected path to be %v, got %v", "team.info", req.URL.Path)
		}

		fmt.Fprint(res, `{"ok":true,"team":{"id":"T12345678"}}`)
	}))

	api.SLACK_API_ROOT = s.URL

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.CreateBot("xoxb-bob-lob-law")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		data []byte
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{[]byte("")}, true},
		{"", c, args{[]byte(`{"team_id":"T87654321"}`)}, true},
		{"", c, args{[]byte(`{"team_id":"T12345678","event":{"text":"foo bar baz","type":"message"}}`)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.handleEvent(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Controller.handleEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_InteractionHandler(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/team.info" {
			t.Fatalf("expected path to be %v, got %v", "team.info", req.URL.Path)
		}

		fmt.Fprint(res, `{"ok":true,"team":{"id":"T12345678"}}`)
	}))

	api.SLACK_API_ROOT = s.URL

	c, err := NewController(WithVerification("bob-lob-law"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.CreateBot("xoxb-bob-lob-law")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantBody   []byte
	}{
		{"", httptest.NewRequest(http.MethodGet, "/", nil), http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", nil), http.StatusInternalServerError, []byte("Internal Server Error\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`payload=%7B%22token%22%3A%22aaa%22%7D`)), http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`payload=%7B%22token%22%3A%22bob-lob-law%22%2C%22team%22%3A%7B%22id%22%3A%22T87654321%22%7D%7D`)), http.StatusInternalServerError, []byte("Internal Server Error\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`payload=%7B%22token%22%3A%22bob-lob-law%22%2C%22team%22%3A%7B%22id%22%3A%22T12345678%22%7D%7D`)), http.StatusOK, []byte("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			c.InteractionHandler()(res, tt.req)

			if res.Code != tt.wantStatus {
				t.Errorf("Controller.InteractionHandler() = %v, want %v", res.Code, tt.wantStatus)
			}

			if res.Code != 200 && !reflect.DeepEqual(res.Body.Bytes(), tt.wantBody) {
				t.Errorf("Controller.InteractionHandler() = %v, want %v", string(res.Body.Bytes()), string(tt.wantBody))
			}
		})
	}
}

func TestController_InteractionOptionsHandler(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/team.info" {
			t.Fatalf("expected path to be %v, got %v", "team.info", req.URL.Path)
		}

		fmt.Fprint(res, `{"ok":true,"team":{"id":"T12345678"}}`)
	}))

	api.SLACK_API_ROOT = s.URL

	c, err := NewController(WithVerification("bob-lob-law"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.CreateBot("xoxb-bob-lob-law")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantBody   []byte
	}{
		{"", httptest.NewRequest(http.MethodGet, "/", nil), http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", nil), http.StatusInternalServerError, []byte("Internal Server Error\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`ssl_check=1`)), http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`ssl_check=1&token=2`)), http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`ssl_check=1&token=bob-lob-law`)), http.StatusOK, []byte("")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`payload=%7B%22token%22%3A%22bob-lob-law%22%2C%22team%22%3A%7B%22id%22%3A%22T87654321%22%7D%7D`)), http.StatusInternalServerError, []byte("Internal Server Error\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`payload=%7B%22token%22%3A%22bob-lob-law%22%2C%22team%22%3A%7B%22id%22%3A%22T12345678%22%7D%7D`)), http.StatusOK, []byte("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			if tt.wantStatus == http.StatusOK {
				go func() {
					iactopt := <-c.interactionOptions
					iactopt.Respond(&InteractionOptionsResponse{})
				}()
			}

			c.InteractionOptionsHandler()(res, tt.req)

			if res.Code != tt.wantStatus {
				t.Errorf("Controller.InteractionHandler() = %v, want %v", res.Code, tt.wantStatus)
			}

			if res.Code != 200 && !reflect.DeepEqual(res.Body.Bytes(), tt.wantBody) {
				t.Errorf("Controller.InteractionHandler() = %v, want %v", string(res.Body.Bytes()), string(tt.wantBody))
			}

			t.Log(string(res.Body.Bytes()))
		})
	}
}

func TestController_CommandHandler(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			fmt.Fprint(res, ``)
		} else if req.URL.Path == "/team.info" {
			fmt.Fprint(res, `{"ok":true,"team":{"id":"T12345678"}}`)
		} else {
			t.Fatalf("expected path to be %v, got %v", "team.info", req.URL.Path)
		}
	}))

	api.SLACK_API_ROOT = s.URL

	c, err := NewController(WithVerification("bob-lob-law"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.CreateBot("xoxb-bob-lob-law")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name               string
		req                *http.Request
		respondImmediately bool
		wantStatus         int
		wantBody           []byte
	}{
		{"", httptest.NewRequest(http.MethodGet, "/", nil), false, http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", nil), false, http.StatusUnauthorized, []byte("Unauthorized\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader("token=bob-lob-law")), false, http.StatusBadRequest, []byte("Bad Request\n")},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader("token=bob-lob-law&response_url="+s.URL)), true, http.StatusOK, []byte(`{"response_type":"in_channel","text":"text"}`)},
		{"", httptest.NewRequest(http.MethodPost, "/", strings.NewReader("token=bob-lob-law&response_url="+s.URL)), false, http.StatusOK, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			if tt.respondImmediately {
				go func() {
					comm := <-c.Commands()
					comm.RespondImmediately(chat.TextMessage("text"), true)
				}()
			}

			c.CommandHandler()(res, tt.req)

			if res.Code != tt.wantStatus {
				t.Errorf("Controller.InteractionHandler() = %v, want %v", res.Code, tt.wantStatus)
			}

			if !reflect.DeepEqual(res.Body.Bytes(), tt.wantBody) {
				t.Errorf("Controller.InteractionHandler() = %v, want %v", string(res.Body.Bytes()), string(tt.wantBody))
			}

			t.Log(res.Body.Bytes())
		})
	}
}

func TestController_handleMessage(t *testing.T) {
	type args struct {
		msg  []byte
		team string
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{nil, "T123"}, true},
		{"", c, args{[]byte(`{"type":"message","user":"U12345678"}`), "T123"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.handleMessage(tt.args.msg, tt.args.team); (err != nil) != tt.wantErr {
				t.Errorf("Controller.handleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_handleMessageType(t *testing.T) {
	type args struct {
		msg     []byte
		subtype string
		user    string
		team    string
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{[]byte(`{"type":"message","user":"U12345678"}`), "", "U12345678", "T123"}, true},
		{"", c, args{[]byte(`{"type":"message","user":"U12345678"}`), "channel_join", "U12345678", "T123"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.handleMessageType(tt.args.msg, tt.args.subtype, tt.args.user, tt.args.team); (err != nil) != tt.wantErr {
				t.Errorf("Controller.handleMessageType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_handleNormalMessage(t *testing.T) {
	type args struct {
		msg *rtm.Message
		bot *Bot
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{&rtm.Message{Text: "test", User: "U123456"}, &Bot{id: "U123456"}}, false},
		{"", c, args{&rtm.Message{Text: "test", User: "U123456"}, &Bot{id: "U654321"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.handleNormalMessage(tt.args.msg, tt.args.bot); (err != nil) != tt.wantErr {
				t.Errorf("Controller.handleNormalMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_handleDirectMessage(t *testing.T) {
	type args struct {
		m *rtm.Message
		b *Bot
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{&rtm.Message{Text: "test"}, nil}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.handleDirectMessage(tt.args.m, tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Controller.handleDirectMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_handleDirectMention(t *testing.T) {
	type args struct {
		m *rtm.Message
		b *Bot
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{&rtm.Message{Text: "test"}, nil}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.handleDirectMention(tt.args.m, tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Controller.handleDirectMention() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_handleMention(t *testing.T) {
	type args struct {
		m *rtm.Message
		b *Bot
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{&rtm.Message{Text: "test"}, nil}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.handleMention(tt.args.m, tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Controller.handleMention() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_handleAmbientMessage(t *testing.T) {
	type args struct {
		m *rtm.Message
		b *Bot
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{&rtm.Message{Text: "test"}, nil}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.handleAmbientMessage(tt.args.m, tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Controller.handleAmbientMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_handleChannelJoin(t *testing.T) {
	type args struct {
		msg []byte
		b   *Bot
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		c    *Controller
		args args
	}{
		{"", c, args{[]byte(`{}`), nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.handleChannelJoin(tt.args.msg, tt.args.b)
		})
	}
}

func TestController_handleUserChannelJoin(t *testing.T) {
	type args struct {
		msg []byte
		bot *Bot
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		c    *Controller
		args args
	}{
		{"", c, args{[]byte(`{}`), nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.handleUserChannelJoin(tt.args.msg, tt.args.bot)
		})
	}
}

func TestController_handleGroupJoin(t *testing.T) {
	type args struct {
		msg []byte
		bot *Bot
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		c    *Controller
		args args
	}{
		{"", c, args{[]byte(`{}`), nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.handleGroupJoin(tt.args.msg, tt.args.bot)
		})
	}
}

func TestController_RegisterConversation(t *testing.T) {
	type args struct {
		name string
		conv *Conversation
	}

	c, err := NewController()
	if err != nil {
		t.Fatal(err)
	}

	conv := NewConversation()

	conv2 := NewConversation()
	conv2.On("start", func(msg *chat.Message, controls *Controls) {})

	tests := []struct {
		name    string
		c       *Controller
		args    args
		wantErr bool
	}{
		{"", c, args{"test", conv}, true},
		{"", c, args{"test", conv2}, false},
		{"", c, args{"test", conv2}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.RegisterConversation(tt.args.name, tt.args.conv); (err != nil) != tt.wantErr {
				t.Errorf("Controller.RegisterConversation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
