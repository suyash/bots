package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestConnector_AddHandler(t *testing.T) {
	u := websocket.Upgrader{
		ReadBufferSize: 1024,
	}

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err := u.Upgrade(res, req, nil)
		if err != nil {
			t.Fatal(err)
		}
	}))

	c := NewConnector("")

	tests := []struct {
		name       string
		c          *Connector
		req        *http.Request
		wantStatus int
	}{
		{"", c, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusUnauthorized},
		{"", c, httptest.NewRequest(http.MethodPost, "/", nil), http.StatusInternalServerError},
		{"", c, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"team":"T12345678","url":"`+strings.Replace(s.URL, "http", "ws", 1)+`"}`)), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			tt.c.AddHandler()(res, tt.req)

			if res.Code != tt.wantStatus {
				t.Errorf("Connector.AddHandler() status = %v, want %v", res.Code, tt.wantStatus)
			}
		})
	}
}

func TestConnector_TypingHandler(t *testing.T) {
	u := websocket.Upgrader{
		ReadBufferSize: 1024,
	}

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err := u.Upgrade(res, req, nil)
		if err != nil {
			t.Fatal(err)
		}
	}))

	c := NewConnector("")

	tests := []struct {
		name       string
		c          *Connector
		req        *http.Request
		wantStatus int
	}{
		{"", c, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusUnauthorized},
		{"", c, httptest.NewRequest(http.MethodPost, "/", nil), http.StatusInternalServerError},
		{"", c, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"team":"T12345678","url":"`+strings.Replace(s.URL, "http", "ws", 1)+`"}`)), http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			tt.c.TypingHandler()(res, tt.req)

			if res.Code != tt.wantStatus {
				t.Errorf("Connector.TypingHandler() status = %v, want %v", res.Code, tt.wantStatus)
			}
		})
	}
}
