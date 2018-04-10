package connector

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestNewConnector(t *testing.T) {
	tests := []struct {
		name string
		want *Connector
	}{
		{"", &Connector{make(map[string]*connection), nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConnector(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConnector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnector_Open(t *testing.T) {
	u := websocket.Upgrader{
		ReadBufferSize: 1024,
	}

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err := u.Upgrade(res, req, nil)
		if err != nil {
			t.Fatal(err)
		}
	}))

	c := NewConnector()

	type args struct {
		team string
		url  string
	}

	tests := []struct {
		name    string
		c       *Connector
		args    args
		wantErr bool
	}{
		{"", c, args{"T123456", ""}, true},
		{"", c, args{"T123456", strings.Replace(s.URL, "http", "ws", 1)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Open(tt.args.team, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("Connector.Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnector_Typing(t *testing.T) {
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
		team    string
		channel string
	}

	c := NewConnector()
	c.Open("T12345678", strings.Replace(s.URL, "http", "ws", 1))

	tests := []struct {
		name    string
		c       *Connector
		args    args
		wantErr bool
	}{
		{"", c, args{"T12345678", "C12345678"}, false},
		{"", c, args{"T87654321", "C12345678"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Typing(tt.args.team, tt.args.channel); (err != nil) != tt.wantErr {
				t.Errorf("Connector.Typing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: figure this out
//
// func TestConnector_readConn(t *testing.T) {
// 	type args struct {
// 		conn *websocket.Conn
// 		team string
// 	}

// 	tests := []struct {
// 		name string
// 		c    *Connector
// 		args args
// 	}{
// 		// TODO: Add test cases.
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.c.readConn(tt.args.conn, tt.args.team)
// 		})
// 	}
// }
