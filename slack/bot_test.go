package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/pkg/errors"

	"suy.io/bots/slack/api"
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/oauth"
	"suy.io/bots/slack/api/rtm"
	"suy.io/bots/slack/connector"
)

func Test_newBot(t *testing.T) {
	type args struct {
		p     *oauth.AccessResponse
		c     Connector
		convs map[string]*Conversation
		cs    ConversationStore
	}

	tests := []struct {
		name string
		args args
		want *Bot
	}{
		{"", args{p: &oauth.AccessResponse{Bot: &oauth.Bot{}}}, &Bot{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newBot(tt.args.p, tt.args.c, tt.args.convs, tt.args.cs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newBot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_BotUser(t *testing.T) {
	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"", fields{id: "a"}, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			if got := bot.BotUser(); got != tt.want {
				t.Errorf("Bot.BotUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_Team(t *testing.T) {
	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"", fields{teamID: "a"}, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			if got := bot.Team(); got != tt.want {
				t.Errorf("Bot.Team() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_Token(t *testing.T) {
	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"", fields{token: "a"}, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			if got := bot.Token(); got != tt.want {
				t.Errorf("Bot.Token() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_Start(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}

		if string(d) != "token=xoxb-bob-lob-law" {
			t.Fatal("invalid token in connect request")
		}

		fmt.Fprint(res, "{\"team\":{\"id\":\"T12345\"},\"url\":\"wss://chat.slack.com\",\"ok\":true}")
	}))

	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	c := &testConnector{make(map[string]string)}

	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		override bool
	}{
		{
			"",
			fields{
				teamID: "T12345",
				token:  "xoxb-bob-lob-law",
				c:      c,
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

			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			if err := bot.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Bot.Start() error = %v, wantErr %v", err, tt.wantErr)
			}

			if c.connections["T12345"] != "wss://chat.slack.com" {
				t.Errorf("expected team to be added to connector")
			}
		})
	}
}

func TestBot_Typing(t *testing.T) {
	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	type args struct {
		channel string
	}

	c := &testConnector{make(map[string]string)}
	c.Add("T12345", "")

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"", fields{teamID: "T12345", c: c}, args{"C12345"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			if err := bot.Typing(tt.args.channel); (err != nil) != tt.wantErr {
				t.Errorf("Bot.Typing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type testConnector struct {
	connections map[string]string
}

func (c *testConnector) Add(team, url string) error {
	c.connections[team] = url
	return nil
}

func (c *testConnector) Close() {}

func (c *testConnector) Typing(team, channel string) error {
	_, ok := c.connections[team]
	if !ok {
		return errors.New("Not Found")
	}

	return nil
}

func (c *testConnector) Messages() <-chan *connector.MessagePayload {
	return make(chan *connector.MessagePayload)
}

var _ Connector = &testConnector{}

func TestBot_Say(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/chat.postMessage" {
			t.Fatal("expected path to be /chat.postMessage")
		}

		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}

		m := &chat.Message{}

		if err := json.Unmarshal(d, m); err != nil {
			t.Fatal(err)
		}

		m.Ts = "12345"

		x := &struct {
			Message *chat.Message `json:"message"`
			OK      bool          `json:"ok"`
		}{
			m,
			true,
		}

		ans, err := json.Marshal(x)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprint(res, string(ans))
	}))

	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	type args struct {
		msg *chat.Message
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     *chat.Message
		wantErr  bool
		override bool
	}{
		{
			"",
			fields{},
			args{chat.TextMessage("test")},
			nil,
			true,
			true,
		},
		{
			"",
			fields{},
			args{&chat.Message{
				Text:    "test",
				Channel: "C12345",
			}},
			&chat.Message{
				Channel: "C12345",
				Text:    "test",
				Ts:      "12345",
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

			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			got, err := bot.Say(tt.args.msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("Bot.Say() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bot.Say() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_Reply(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/chat.postMessage" {
			t.Fatal("expected path to be /chat.postMessage")
		}

		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}

		m := &chat.Message{}

		if err := json.Unmarshal(d, m); err != nil {
			t.Fatal(err)
		}

		m.Ts = "12345"

		x := &struct {
			Message *chat.Message `json:"message"`
			OK      bool          `json:"ok"`
		}{
			m,
			true,
		}

		ans, err := json.Marshal(x)

		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprint(res, string(ans))
	}))

	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	type args struct {
		msg      *rtm.Message
		response *chat.Message
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     *chat.Message
		wantErr  bool
		override bool
	}{
		{
			"",
			fields{},
			args{&rtm.Message{Text: "orig"}, chat.TextMessage("test")},
			nil,
			true,
			true,
		},
		{
			"",
			fields{},
			args{&rtm.Message{Text: "orig", Channel: "C12345"}, chat.TextMessage("test")},
			&chat.Message{
				Channel: "C12345",
				Text:    "test",
				Ts:      "12345",
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

			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			got, err := bot.Reply(tt.args.msg, tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bot.Reply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bot.Reply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_SayEphemeral(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/chat.postEphemeral" {
			t.Fatal("expected path to be /chat.postEphemeral")
		}

		fmt.Fprint(res, "{\"message_ts\":\"boo\",\"ok\":true}")
	}))

	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	type args struct {
		msg *chat.EphemeralMessage
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     string
		wantErr  bool
		override bool
	}{
		{
			"",
			fields{},
			args{&chat.EphemeralMessage{
				Text: "test",
			}},
			"",
			true,
			true,
		},
		{
			"",
			fields{},
			args{&chat.EphemeralMessage{
				Text:    "test",
				Channel: "C12345",
			}},
			"boo",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.override {
				api.SLACK_API_ROOT = s.URL
			}

			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			got, err := bot.SayEphemeral(tt.args.msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("Bot.SayEphemeral() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Bot.SayEphemeral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_ReplyEphemeral(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/chat.postEphemeral" {
			t.Fatal("expected path to be /chat.postEphemeral")
		}

		fmt.Fprint(res, "{\"message_ts\":\"boo\",\"ok\":true}")
	}))

	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	type args struct {
		msg      *rtm.Message
		response *chat.EphemeralMessage
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     string
		wantErr  bool
		override bool
	}{
		{
			"",
			fields{},
			args{&rtm.Message{}, &chat.EphemeralMessage{}},
			"",
			true,
			true,
		},
		{
			"",
			fields{},
			args{&rtm.Message{Channel: "C12345"}, &chat.EphemeralMessage{}},
			"boo",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.override {
				api.SLACK_API_ROOT = s.URL
			}

			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			got, err := bot.ReplyEphemeral(tt.args.msg, tt.args.response)

			if (err != nil) != tt.wantErr {
				t.Errorf("Bot.ReplyEphemeral() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Bot.ReplyEphemeral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_ReplyInThread(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/chat.postMessage" {
			t.Fatal("expected path to be /chat.postMessage")
		}

		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}

		m := &chat.Message{}

		if err := json.Unmarshal(d, m); err != nil {
			t.Fatal(err)
		}

		m.Ts = "12345"

		x := &struct {
			Message *chat.Message `json:"message"`
			OK      bool          `json:"ok"`
		}{
			m,
			true,
		}

		ans, err := json.Marshal(x)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprint(res, string(ans))
	}))

	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	type args struct {
		msg      *rtm.Message
		response *chat.Message
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     *chat.Message
		wantErr  bool
		override bool
	}{
		{
			"",
			fields{},
			args{&rtm.Message{Text: "test"}, &chat.Message{Text: "new"}},
			nil,
			true,
			true,
		},
		{
			"",
			fields{},
			args{
				&rtm.Message{Channel: "C12345", Text: "test", Ts: "1"},
				&chat.Message{Text: "new"},
			},
			&chat.Message{
				Channel:  "C12345",
				Text:     "new",
				ThreadTs: "1",
				Ts:       "12345",
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

			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			got, err := bot.ReplyInThread(tt.args.msg, tt.args.response)

			if (err != nil) != tt.wantErr {
				t.Errorf("Bot.ReplyInThread() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bot.ReplyInThread() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_Update(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/chat.update" {
			t.Fatalf("expected path to be /chat.update, got %v", req.URL.Path)
		}

		x := &struct {
			Ok      bool   `json:"ok"`
			Text    string `json:"text"`
			Ts      string `json:"ts"`
			Channel string `json:"channel"`
		}{
			true,
			"test",
			"12345",
			"C12345",
		}

		ans, err := json.Marshal(x)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprint(res, string(ans))
	}))

	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	type args struct {
		ts  string
		msg *chat.Message
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     string
		wantErr  bool
		override bool
	}{
		{
			"",
			fields{},
			args{"12345", chat.TextMessage("test")},
			"12345",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.override {
				api.SLACK_API_ROOT = s.URL
			}

			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			got, err := bot.Update(tt.args.ts, tt.args.msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("Bot.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Bot.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_StartConversation(t *testing.T) {
	type fields struct {
		id     string
		teamID string
		token  string
		c      Connector
		convs  map[string]*Conversation
		cs     ConversationStore
	}

	type args struct {
		user    string
		channel string
		name    string
	}

	convs, cs := make(map[string]*Conversation), NewMemoryConversationStore()
	convs["x"] = NewConversation()
	convs["x"].On("start", func(m *chat.Message, c *Controls) {})

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"",
			fields{convs: convs, cs: cs},
			args{"U12345", "C12345", "x"},
			false,
		},
		{
			"",
			fields{convs: convs},
			args{"U12345", "C12345", "y"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &Bot{
				id:     tt.fields.id,
				teamID: tt.fields.teamID,
				token:  tt.fields.token,
				c:      tt.fields.c,
				convs:  tt.fields.convs,
				cs:     tt.fields.cs,
			}

			if err := bot.StartConversation(tt.args.user, tt.args.channel, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Bot.StartConversation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
