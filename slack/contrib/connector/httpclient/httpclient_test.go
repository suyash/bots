package httpclient

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestNewConnector(t *testing.T) {
	type args struct {
		host string
	}

	tests := []struct {
		name    string
		args    args
		want    *Connector
		wantErr bool
	}{
		{"", args{"sadasd"}, nil, true},
		{"", args{"http://localhost:8080"}, &Connector{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConnector(tt.args.host)

			if err == nil {
				got.msgs = nil
				got.url = nil
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("NewConnector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConnector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sendExternalRequest(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("sendExternalRequest() Method = %v, wantMethod %v", req.Method, http.MethodPost)
		}

		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("sendExternalRequest() Content-Type = %v, want Content-Type %v", req.Header.Get("Content-Type"), "application/json")
		}
	}))

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		url  *url.URL
		data interface{}
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{u, "aaa"}, false},

		{
			"",
			args{
				u,
				&struct {
					Text string `json:"text"`
				}{
					"xxx",
				},
			},
			false,
		},

		{
			"",
			args{
				u,
				&struct {
					text string
				}{
					"xxx",
				},
			},
			false,
		},

		{
			"",
			args{
				u,
				struct {
					text string
				}{
					"xxx",
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendExternalRequest(tt.args.url, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("sendExternalRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnector_Add(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("Connector.Add() Method = %v, wantMethod %v", req.Method, http.MethodPost)
		}

		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Connector.Add() Content-Type = %v, want Content-Type %v", req.Header.Get("Content-Type"), "application/json")
		}

		if req.URL.Path != "/slack/add" {
			t.Errorf("Connector.Add() Path = %v, want Content-Type %v", req.URL.Path, "/slack/add")
		}
	}))

	c, err := NewConnector(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		team      string
		socketURL string
	}

	tests := []struct {
		name    string
		c       *Connector
		args    args
		wantErr bool
	}{
		{"", c, args{"T12345678", ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Add(tt.args.team, tt.args.socketURL); (err != nil) != tt.wantErr {
				t.Errorf("Connector.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnector_Typing(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("Connector.Typing() Method = %v, wantMethod %v", req.Method, http.MethodPost)
		}

		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Connector.Typing() Content-Type = %v, want Content-Type %v", req.Header.Get("Content-Type"), "application/json")
		}

		if req.URL.Path != "/slack/typing" {
			t.Errorf("Connector.Typing() Path = %v, want Content-Type %v", req.URL.Path, "/slack/typing")
		}
	}))

	c, err := NewConnector(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		team    string
		channel string
	}

	tests := []struct {
		name    string
		c       *Connector
		args    args
		wantErr bool
	}{
		{"", c, args{"T12345678", "C12345678"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Typing(tt.args.team, tt.args.channel); (err != nil) != tt.wantErr {
				t.Errorf("Connector.Typing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnector_ServeHTTP(t *testing.T) {
	type args struct {
		res *httptest.ResponseRecorder
		req *http.Request
	}

	c, err := NewConnector("http://a.a")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		c          *Connector
		args       args
		wantStatus int
	}{
		{
			"",
			c,
			args{httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil)},
			http.StatusInternalServerError,
		},
		{
			"",
			c,
			args{httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"team": "T123", "message":"bob-lob-law"}`))},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.ServeHTTP(tt.args.res, tt.args.req)

			if tt.args.res.Code != tt.wantStatus {
				t.Errorf("Connector.ServeHTTP() status = %v, wantStatus %v", tt.args.res.Code, tt.wantStatus)
			}
		})
	}
}
