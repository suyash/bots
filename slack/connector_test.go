package slack

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"suy.io/bots/slack/connector"
)

func Test_newInternalConnector(t *testing.T) {
	tests := []struct {
		name string
		want *internalConnector
	}{
		{"", &internalConnector{connector.NewConnector(), nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newInternalConnector()
			got.conn = connector.NewConnector()
			got.msgs = nil

			if !reflect.DeepEqual(got.conn, tt.want.conn) {
				t.Errorf("newInternalConnector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_internalConnector_Add(t *testing.T) {
	u := websocket.Upgrader{
		ReadBufferSize: 1024,
	}

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err := u.Upgrade(res, req, nil)
		if err != nil {
			t.Fatal(err)
		}
	}))

	type args struct {
		team string
		url  string
	}

	c := newInternalConnector()

	tests := []struct {
		name    string
		c       *internalConnector
		args    args
		wantErr bool
	}{
		{"", c, args{"T1234567", strings.Replace(s.URL, "http", "ws", 1)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Add(tt.args.team, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("internalConnector.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_internalConnector_handleMessage(t *testing.T) {
	conn, msgs := connector.NewConnector(), make(chan *connector.MessagePayload)

	type fields struct {
		conn *connector.Connector
		msgs chan *connector.MessagePayload
	}

	type args struct {
		msg  []byte
		team string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"", fields{conn, msgs}, args{[]byte(`{"text":"ok"}`), "T12345678"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &internalConnector{
				conn: tt.fields.conn,
				msgs: tt.fields.msgs,
			}

			c.handleMessage(tt.args.msg, tt.args.team)

			m := <-c.msgs

			if m.Team != tt.args.team || string(m.Message) != string(tt.args.msg) {
				t.Errorf("internalConnector.handleMessage() created invalid payload to send to connector")
			}
		})
	}
}

func Test_internalConnector_Close(t *testing.T) {
	c := newInternalConnector()

	tests := []struct {
		name string
		c    *internalConnector
	}{
		{"", c},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Close()

			if c.msgs != nil {
				t.Error("internalConnector.Close(): expected msgs to be nil")
			}
		})
	}
}

// TODO: figure this out
//
// func Test_internalConnector_Typing(t *testing.T) {
// 	type fields struct {
// 		conn *connector.Connector
// 		msgs chan *connector.MessagePayload
// 	}
// 	type args struct {
// 		team    string
// 		channel string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := &internalConnector{
// 				conn: tt.fields.conn,
// 				msgs: tt.fields.msgs,
// 			}
// 			if err := c.Typing(tt.args.team, tt.args.channel); (err != nil) != tt.wantErr {
// 				t.Errorf("internalConnector.Typing() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
