package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"suy.io/bots/slack/api"
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/dialog"
)

func Test_newCommand(t *testing.T) {
	type args struct {
		vals url.Values
	}

	v := make(url.Values)
	v.Add("team_id", "a")
	v.Add("team_domain", "b")
	v.Add("enterprise_id", "c")
	v.Add("enterprise_name", "d")
	v.Add("channel_id", "e")
	v.Add("channel_name", "f")
	v.Add("user_id", "g")
	v.Add("command", "h")
	v.Add("text", "i")
	v.Add("response_url", "j")
	v.Add("trigger_id", "k")

	tests := []struct {
		name string
		args args
		want *Command
	}{
		{"", args{v}, &Command{
			TeamID:         "a",
			TeamDomain:     "b",
			EnterpriseID:   "c",
			EnterpriseName: "d",
			ChannelID:      "e",
			ChannelName:    "f",
			UserID:         "g",
			Command:        "h",
			Text:           "i",
			ResponseURL:    "j",
			TriggerID:      "k",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newCommand(tt.args.vals)
			got.immediateResponse = nil
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommand_RespondImmediately(t *testing.T) {
	type args struct {
		msg       *chat.Message
		inChannel bool
	}

	c := &Command{
		immediateResponse: make(chan []byte),
		TeamID:            "a",
		TeamDomain:        "b",
		EnterpriseID:      "c",
		EnterpriseName:    "d",
		ChannelID:         "e",
		ChannelName:       "f",
		UserID:            "g",
		Command:           "h",
		Text:              "i",
		ResponseURL:       "j",
		TriggerID:         "k",
	}

	tests := []struct {
		name    string
		command *Command
		args    args
		wantErr bool
	}{
		{"", c, args{chat.TextMessage("test"), true}, false},
		{"", c, args{chat.TextMessage("test"), false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func(tt struct {
				name    string
				command *Command
				args    args
				wantErr bool
			}) {
				if err := tt.command.RespondImmediately(tt.args.msg, tt.args.inChannel); (err != nil) != tt.wantErr {
					t.Errorf("Command.RespondImmediately() error = %v, wantErr %v", err, tt.wantErr)
				}
			}(tt)

			if !tt.wantErr {
				data := <-c.immediateResponse
				t.Log(string(data))

				cr := &struct {
					Text         string `json:"text"`
					ResponseType string `json:"response_type"`
				}{"", ""}

				if err := json.Unmarshal(data, cr); err != nil {
					t.Fatal(err)
				}

				if tt.args.inChannel && cr.ResponseType != "in_channel" {
					t.Errorf("expected in_channel response")
				} else if !tt.args.inChannel && cr.ResponseType != "ephemeral" {
					t.Errorf("expected ephemeral response")
				}

				if cr.Text != "test" {
					t.Errorf("message not properly encoded")
				}
			}
		})
	}
}

func TestCommand_Respond(t *testing.T) {
	type args struct {
		msg       *chat.Message
		inChannel bool
	}

	reqdata := make(chan []byte)

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		d, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
		go func() { reqdata <- d }()
	}))

	c := &Command{
		TeamID:         "a",
		TeamDomain:     "b",
		EnterpriseID:   "c",
		EnterpriseName: "d",
		ChannelID:      "e",
		ChannelName:    "f",
		UserID:         "g",
		Command:        "h",
		Text:           "i",
		ResponseURL:    s.URL,
		TriggerID:      "k",
	}

	tests := []struct {
		name    string
		command *Command
		args    args
		wantErr bool
	}{
		{"", c, args{chat.TextMessage("test"), true}, false},
		{"", c, args{chat.TextMessage("test"), false}, false},
		{"", c, args{chat.TextMessage("test"), true}, false},
		{"", c, args{chat.TextMessage("test"), false}, false},
		{"", c, args{chat.TextMessage("test"), true}, false},
		{"Responding to the same command 6th time", c, args{chat.TextMessage("test"), false}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.command.Respond(tt.args.msg, tt.args.inChannel); (err != nil) != tt.wantErr {
				t.Errorf("Command.Respond() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			d := <-reqdata
			t.Log(string(d))

			cr := &struct {
				Text         string `json:"text"`
				ResponseType string `json:"response_type"`
			}{"", ""}

			if err := json.Unmarshal(d, cr); err != nil {
				t.Fatal(err)
			}

			if tt.args.inChannel && cr.ResponseType != "in_channel" {
				t.Error("expected in_channel response")
			} else if !tt.args.inChannel && cr.ResponseType != "ephemeral" {
				t.Error("expected ephemeral response")
			}

			if cr.Text != tt.args.msg.Text {
				t.Errorf("expected text to be %v, got %v", cr.Text, tt.args.msg.Text)
			}
		})
	}
}

func TestCommand_OpenDialog(t *testing.T) {
	c := &Command{
		TeamID:         "a",
		TeamDomain:     "b",
		EnterpriseID:   "c",
		EnterpriseName: "d",
		ChannelID:      "e",
		ChannelName:    "f",
		UserID:         "g",
		Command:        "h",
		Text:           "i",
		ResponseURL:    "j",
		TriggerID:      "k",
	}

	d := &dialog.Dialog{
		CallbackID: "test",
		Title:      "test",
		Elements: []*dialog.Element{
			{
				Name:  "text",
				Type:  dialog.TextElementType,
				Label: "text",
			},
		},
	}

	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/dialog.open" {
			t.Errorf("expected the request to happen at /dialog.open, got %v", req.URL.Path)
		}

		fmt.Fprint(res, "{\"ok\": true}")
	}))

	type args struct {
		d     *dialog.Dialog
		token string
	}

	tests := []struct {
		name     string
		command  *Command
		args     args
		wantErr  bool
		override bool
	}{
		{"", c, args{d, "x"}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.override {
				api.SLACK_API_ROOT = s.URL
			}

			if err := tt.command.OpenDialog(tt.args.d, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("Command.OpenDialog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	s.Close()
}

func TestCommand_RespondWithEmptyBody(t *testing.T) {
	c := &Command{
		immediateResponse: make(chan []byte),
		TeamID:            "a",
		TeamDomain:        "b",
		EnterpriseID:      "c",
		EnterpriseName:    "d",
		ChannelID:         "e",
		ChannelName:       "f",
		UserID:            "g",
		Command:           "h",
		Text:              "i",
		ResponseURL:       "j",
		TriggerID:         "k",
	}

	tests := []struct {
		name    string
		command *Command
		wantErr bool
	}{
		{"", c, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.command.RespondWithEmptyBody(); (err != nil) != tt.wantErr {
				t.Errorf("Command.RespondWithEmptyBody() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, ok := <-c.immediateResponse; ok {
				t.Errorf("expected channel to be closed")
			}
		})
	}
}

func TestCommand_RespondWithErrors(t *testing.T) {
	c := &Command{
		immediateResponse: make(chan []byte),
		TeamID:            "a",
		TeamDomain:        "b",
		EnterpriseID:      "c",
		EnterpriseName:    "d",
		ChannelID:         "e",
		ChannelName:       "f",
		UserID:            "g",
		Command:           "h",
		Text:              "i",
		ResponseURL:       "j",
		TriggerID:         "k",
	}

	type args struct {
		errs []*dialog.Error
	}

	tests := []struct {
		name    string
		command *Command
		args    args
		wantErr bool
	}{
		{"", c, args{[]*dialog.Error{
			&dialog.Error{Name: "name1", Error: "error1"},
		}}, false},

		{"", c, args{[]*dialog.Error{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func(tt struct {
				name    string
				command *Command
				args    args
				wantErr bool
			}) {
				if err := tt.command.RespondWithErrors(tt.args.errs); (err != nil) != tt.wantErr {
					t.Errorf("Command.RespondWithErrors() error = %v, wantErr %v", err, tt.wantErr)
				}
			}(tt)

			if tt.wantErr {
				return
			}

			s := &struct {
				Errors json.RawMessage `json:"errors"`
			}{json.RawMessage("")}

			v := <-c.immediateResponse

			if err := json.Unmarshal(v, s); err != nil {
				t.Fatal(err)
			}

			if len(s.Errors) == 0 {
				t.Errorf("Expected errors key to be set in the sent payload")
			}
		})
	}
}
