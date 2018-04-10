package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"suy.io/bots/slack/api"
	"suy.io/bots/slack/api/chat"
	"suy.io/bots/slack/api/dialog"
)

func TestInteraction_OpenDialog(t *testing.T) {
	type args struct {
		d *dialog.Dialog
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
			t.Fatalf("expected the request to happen at /dialog.open, got %v", req.URL.Path)
		}

		fmt.Fprint(res, "{\"ok\": true}")
	}))

	iact := &Interaction{}

	tests := []struct {
		name     string
		iact     *Interaction
		args     args
		wantErr  bool
		override bool
	}{
		{"", iact, args{d}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.override {
				api.SLACK_API_ROOT = s.URL
			}

			if err := tt.iact.OpenDialog(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("Interaction.OpenDialog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInteraction_RespondImmediately(t *testing.T) {
	type args struct {
		msg *chat.Message
	}

	iact := &Interaction{
		immediateResponse: make(chan []byte),
	}

	msg := chat.TextMessage("test")

	tests := []struct {
		name    string
		iact    *Interaction
		args    args
		wantErr bool
	}{
		{"", iact, args{msg}, false},
		{"", iact, args{nil}, true},
	}

	var wg sync.WaitGroup
	wg.Add(len(tests))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func(tt struct {
				name    string
				iact    *Interaction
				args    args
				wantErr bool
			}) {
				if err := tt.iact.RespondImmediately(tt.args.msg); (err != nil) != tt.wantErr {
					t.Errorf("Interaction.RespondImmediately() error = %v, wantErr %v", err, tt.wantErr)
				}

				wg.Done()
			}(tt)

			if !tt.wantErr {
				data := <-iact.immediateResponse
				t.Log(string(data))

				msg := &chat.Message{}
				if err := json.Unmarshal(data, msg); err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tt.args.msg, msg) {
					t.Errorf("Interaction.RespondImmediately() = %v, want %v", msg, tt.args.msg)
				}
			}
		})
	}

	wg.Wait()
}

func TestInteraction_Respond(t *testing.T) {
	type args struct {
		msg *chat.Message
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

	iact := &Interaction{
		ResponseURL: s.URL,
	}

	tests := []struct {
		name    string
		iact    *Interaction
		args    args
		wantErr bool
	}{
		{"", iact, args{chat.TextMessage("text")}, false},
		{"", iact, args{chat.TextMessage("text")}, false},
		{"", iact, args{chat.TextMessage("text")}, false},
		{"", iact, args{chat.TextMessage("text")}, false},
		{"", iact, args{chat.TextMessage("text")}, false},
		{"", iact, args{chat.TextMessage("text")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.iact.Respond(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Interaction.Respond() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			d := <-reqdata
			t.Log(string(d))

			m := &chat.Message{}

			if err := json.Unmarshal(d, m); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tt.args.msg, m) {
				t.Errorf("Interaction.RespondImmediately() = %v, want %v", m, tt.args.msg)
			}
		})
	}
}

func TestInteraction_RespondWithEmptyBody(t *testing.T) {
	iact := &Interaction{
		immediateResponse: make(chan []byte),
	}

	tests := []struct {
		name    string
		iact    *Interaction
		wantErr bool
	}{
		{"", iact, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.iact.RespondWithEmptyBody(); (err != nil) != tt.wantErr {
				t.Errorf("Interaction.RespondWithEmptyBody() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, ok := <-iact.immediateResponse; ok {
				t.Errorf("expected channel to be closed")
			}
		})
	}
}

func TestInteraction_RespondWithErrors(t *testing.T) {
	iact := &Interaction{
		immediateResponse: make(chan []byte),
	}

	type args struct {
		errs []*dialog.Error
	}

	tests := []struct {
		name    string
		iact    *Interaction
		args    args
		wantErr bool
	}{
		{"", iact, args{[]*dialog.Error{
			&dialog.Error{Name: "name1", Error: "error1"},
		}}, false},

		{"", iact, args{[]*dialog.Error{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func(tt struct {
				name    string
				iact    *Interaction
				args    args
				wantErr bool
			}) {
				if err := tt.iact.RespondWithErrors(tt.args.errs); (err != nil) != tt.wantErr {
					t.Errorf("Interaction.RespondWithErrors() error = %v, wantErr %v", err, tt.wantErr)
				}
			}(tt)

			if tt.wantErr {
				return
			}

			s := &struct {
				Errors json.RawMessage `json:"errors"`
			}{json.RawMessage("")}

			v := <-iact.immediateResponse

			if err := json.Unmarshal(v, s); err != nil {
				t.Fatal(err)
			}

			if len(s.Errors) == 0 {
				t.Errorf("Expected errors key to be set in the sent payload")
			}
		})
	}
}

func TestInteractionOptions_Respond(t *testing.T) {
	type args struct {
		msg *InteractionOptionsResponse
	}

	iactopt := &InteractionOptions{
		immediateResponse: make(chan []byte),
	}

	tests := []struct {
		name    string
		iactopt *InteractionOptions
		args    args
		wantErr bool
	}{
		{"", iactopt, args{msg: &InteractionOptionsResponse{
			Options: []*chat.Option{
				{"foo", "bar", "baz"},
			},
		}}, false},
	}

	var wg sync.WaitGroup
	wg.Add(len(tests))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				if err := tt.iactopt.Respond(tt.args.msg); (err != nil) != tt.wantErr {
					t.Errorf("InteractionOptions.Respond() error = %v, wantErr %v", err, tt.wantErr)
				}

				wg.Done()
			}()

			if tt.wantErr {
				return
			}

			d := <-iactopt.immediateResponse
			s := &struct {
				Options json.RawMessage `json:"options"`
			}{json.RawMessage("")}

			if err := json.Unmarshal(d, s); err != nil {
				t.Fatal(err)
			}

			if len(s.Options) == 0 {
				t.Errorf("InteractionOptions.Respond() expected options key to be set in sent response")
			}
		})
	}

	wg.Wait()
}
