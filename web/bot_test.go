package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
)

func Test_newBot(t *testing.T) {
	type args struct {
		id               BotID
		conn             *websocket.Conn
		sanitizer        *bluemonday.Policy
		incomingMessages chan *MessagePair
		is               ItemStore
		conversations    map[string]*Conversation
		convs            ConversationStore
		errhandler       ErrorHandler
	}

	tests := []struct {
		name string
		args args
		want *Bot
	}{
		{"", args{
			conversations: make(map[string]*Conversation),
		}, &Bot{
			conversations: make(map[string]*Conversation),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newBot(tt.args.id, tt.args.conn, tt.args.sanitizer, tt.args.incomingMessages, tt.args.is, tt.args.conversations, tt.args.convs, tt.args.errhandler)
			got.incomingMessages = nil
			got.outgoingMessages = nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newBot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_ID(t *testing.T) {
	tests := []struct {
		name string
		bot  *Bot
		want BotID
	}{
		{"", &Bot{id: BotID(2)}, BotID(2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bot.ID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bot.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestBot_read(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		bot  *Bot
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.bot.read()
// 		})
// 	}
// }

func TestBot_handleMessage(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		msg *Message
	}

	tests := []struct {
		name string
		bot  *Bot
		args args
	}{
		{"", b, args{TextMessage("test")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.bot.handleMessage(tt.args.msg)

			if got := <-tt.bot.incomingMessages; !reflect.DeepEqual(got.Message, tt.args.msg) {
				t.Errorf("Bot.handleMessage() = %v, want %v", got, tt.args.msg)
			}
		})
	}
}

// TODO: figure this out
//
// func TestBot_write(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		bot  *Bot
// 	}{
// 		// TODO: Add test cases.
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.bot.write()
// 		})
// 	}
// }

func TestBot_send(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		i Item
	}

	tests := []struct {
		name string
		bot  *Bot
		args args
	}{
		{"", b, args{TextMessage("test")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.bot.send(tt.args.i)

			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			if string(msg) != `{"attachments":null,"source":"","type":"","text":"test","threadId":0,"replyId":0,"id":0,"prev":null,"next":null}` {
				t.Errorf("Bot.handleMessage() = %v, want %v", string(msg), "{}")
			}
		})
	}
}

func TestBot_Send(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		item Item
	}

	tests := []struct {
		name    string
		bot     *Bot
		args    args
		wantErr bool
	}{
		{"", b, args{TextMessage("test")}, true},
		{"", b, args{&Message{Text: "test", ID: 2}}, true},
		{"", b, args{&Message{Text: "test", ID: 2, Source: BotItemSource}}, true},
		{"", b, args{&Message{Text: "test", ID: 2, Source: BotItemSource, Type: MessageItemType}}, false},
		{"", b, args{&Thread{ID: 2}}, true},
		{"", b, args{&Thread{ID: 2, Type: ThreadItemType}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bot.Send(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("Bot.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBot_say(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		i Item
	}

	tests := []struct {
		name string
		bot  *Bot
		args args
	}{
		{"", b, args{TextMessage("test")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.bot.say(tt.args.i)

			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			if string(msg) != `{"attachments":null,"source":"","type":"","text":"test","threadId":0,"replyId":0,"id":0,"prev":null,"next":null}` {
				t.Errorf("Bot.handleMessage() = %v, want %v", string(msg), "{}")
			}
		})
	}
}

func TestBot_Say(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		msg *Message
	}

	tests := []struct {
		name string
		bot  *Bot
		args args
	}{
		{"", b, args{TextMessage("test")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.bot.Say(tt.args.msg)

			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			m := &Message{}
			if err := json.Unmarshal(msg, m); err != nil {
				t.Fatal(err)
			}

			if m.ID == 0 {
				t.Errorf("Bot.handleMessage() wanted ID to not be zero")
			}

			if m.Source != BotItemSource {
				t.Errorf("Bot.handleMessage() source = %v, want %v", string(m.Source), BotItemSource)
			}

			if m.Type != MessageItemType {
				t.Errorf("Bot.handleMessage() type = %v, want %v", string(m.Type), MessageItemType)
			}

			if m.Text != tt.args.msg.Text {
				t.Errorf("Bot.handleMessage() type = %v, want %v", m.Text, tt.args.msg.Text)
			}
		})
	}
}

func TestBot_Reply(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		reply *Message
	}

	tests := []struct {
		name string
		bot  *Bot
		args args
	}{
		{"", b, args{TextMessage("reply")}},
		{"", b, args{&Message{Text: "a", ThreadID: 222}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"text\": \"hey\" }")); err != nil {
				t.Fatal(err)
			}

			// skip id echo
			if _, _, err := conn.ReadMessage(); err != nil {
				t.Fatal(err)
			}

			mp := <-b.incomingMessages

			mp.Reply(tt.args.reply)

			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			m := &Message{}
			if err := json.Unmarshal(msg, m); err != nil {
				t.Fatal(err)
			}

			if m.Text != tt.args.reply.Text {
				t.Fatalf("Bot.Reply() %v, want %v", m.Text, "reply")
			}

			if m.ReplyID != mp.Message.ID {
				t.Fatalf("Bot.Reply() %v, want %v", m.ReplyID, mp.Message.ID)
			}
		})
	}
}

func TestBot_ReplyInThread(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		reply *Message
	}

	tests := []struct {
		name string
		bot  *Bot
		args args
	}{
		{"", b, args{TextMessage("text")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"text\": \"hey\" }")); err != nil {
				t.Fatal(err)
			}

			// skip id echo
			if _, _, err := conn.ReadMessage(); err != nil {
				t.Fatal(err)
			}

			mp := <-b.incomingMessages

			mp.ReplyInThread(tt.args.reply)

			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			th := &Thread{}
			if err := json.Unmarshal(msg, th); err != nil {
				t.Fatal(err)
			}

			if th.ID != mp.Message.ID {
				t.Errorf("Bot.ReplyInThread() %v, want %v", th.ID, mp.Message.ID)
			}

			_, msg, err = conn.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			m := &Message{}
			if err := json.Unmarshal(msg, m); err != nil {
				t.Fatal(err)
			}

			if m.Text != tt.args.reply.Text {
				t.Fatalf("Bot.ReplyInThread() %v, want %v", m.Text, "reply")
			}

			if m.ReplyID != mp.Message.ID {
				t.Fatalf("Bot.ReplyInThread() %v, want %v", m.ReplyID, mp.Message.ID)
			}
		})
	}
}

func TestBot_Update(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		msg *Message
	}

	tests := []struct {
		name string
		bot  *Bot
		args args
	}{
		{"", b, args{TextMessage("test")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"text\": \"hey\" }")); err != nil {
				t.Fatal(err)
			}

			// skip id echo
			if _, _, err := conn.ReadMessage(); err != nil {
				t.Fatal(err)
			}

			mp := <-b.incomingMessages

			mp.Message.Text = "random"

			mp.Bot.Update(mp.Message)

			// skip id echo
			if _, _, err := conn.ReadMessage(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBot_StartConversation(t *testing.T) {
	s, cb := newTestEchoServer(t)

	url := strings.Replace(s.URL, "http://", "ws://", -1)
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := <-cb

	type args struct {
		name string
	}

	tests := []struct {
		name    string
		bot     *Bot
		args    args
		wantErr bool
	}{
		{"", b, args{"foo"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bot.StartConversation(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Bot.StartConversation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newTestEchoServer(t *testing.T) (*httptest.Server, chan *Bot) {
	c := make(chan *Bot)

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			t.Fatalf("cannot start server, error: %v", err)
		}

		b := newBot(
			BotID(0),
			conn,
			bluemonday.UGCPolicy(),
			make(chan *MessagePair),
			NewMemoryItemStore(),
			nil,
			NewMemoryConversationStore(),
			func(err error) {
				t.Logf("internal error: %v", err)
				// t.Errorf("internal error: %v", err)
			},
		)

		b.remove = func() {}

		go b.read()
		go b.write()

		c <- b
	}))

	return s, c
}
