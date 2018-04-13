package web

import (
	"reflect"
	"testing"
)

func TestControls_Get(t *testing.T) {
	type args struct {
		key string
	}

	cs := NewMemoryConversationStore()
	cs.Start(42, "x")
	cs.SetData(42, "foo", "bar")

	tests := []struct {
		name    string
		c       *Controls
		args    args
		want    string
		wantErr bool
	}{
		{"", &Controls{&Bot{id: 42, convs: cs}}, args{"foo"}, "bar", false},
		{"", &Controls{&Bot{id: 42, convs: cs}}, args{"baz"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Get(tt.args.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("Controls.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Controls.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControls_Set(t *testing.T) {
	type args struct {
		key   string
		value string
	}

	cs := NewMemoryConversationStore()
	cs.Start(42, "x")

	tests := []struct {
		name    string
		c       *Controls
		args    args
		wantErr bool
	}{
		{"", &Controls{&Bot{id: 42, convs: cs}}, args{"foo", "bar"}, false},
		{"", &Controls{&Bot{id: 42, convs: cs}}, args{"foo", "baz"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Controls.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControls_To(t *testing.T) {
	type args struct {
		state string
	}

	cs := NewMemoryConversationStore()
	cs.Start(42, "x")

	tests := []struct {
		name    string
		c       *Controls
		args    args
		wantErr bool
	}{
		{"", &Controls{&Bot{id: 42, convs: cs}}, args{"foo"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.To(tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("Controls.To() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControls_End(t *testing.T) {
	cs := NewMemoryConversationStore()
	cs.Start(42, "x")

	tests := []struct {
		name    string
		c       *Controls
		wantErr bool
	}{
		{"", &Controls{&Bot{id: 42, convs: cs}}, false},
		{"", &Controls{&Bot{id: 42, convs: cs}}, true},
		{"", &Controls{&Bot{id: 21, convs: cs}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.End(); (err != nil) != tt.wantErr {
				t.Errorf("Controls.End() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControls_Bot(t *testing.T) {
	cs := NewMemoryConversationStore()
	cs.Start(42, "x")

	tests := []struct {
		name string
		c    *Controls
		want *Bot
	}{
		{"", &Controls{&Bot{id: 42, convs: cs}}, &Bot{id: 42, convs: cs}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Bot(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Controls.Bot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConversation(t *testing.T) {
	tests := []struct {
		name string
		want *Conversation
	}{
		{"", &Conversation{make(map[string]ConversationHandler)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConversation(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConversation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConversation_On(t *testing.T) {
	type args struct {
		state   string
		handler ConversationHandler
	}

	conv := NewConversation()
	conv.On("a", func(*Message, *Controls) {})

	tests := []struct {
		name    string
		s       *Conversation
		args    args
		wantErr bool
	}{
		{"", conv, args{"b", nil}, true},
		{"", conv, args{"start", func(*Message, *Controls) {}}, false},
		{"", conv, args{"a", func(*Message, *Controls) {}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.On(tt.args.state, tt.args.handler); (err != nil) != tt.wantErr {
				t.Errorf("Conversation.On() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
