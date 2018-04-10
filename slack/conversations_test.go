package slack

import (
	"reflect"
	"testing"

	"suy.io/bots/slack/api/chat"
)

func TestControls_Get(t *testing.T) {
	type fields struct {
		b       *Bot
		user    string
		channel string
	}

	type args struct {
		key string
	}

	cs := NewMemoryConversationStore()

	if err := cs.Start("U12345678", "C12345678", "T12345678", "test"); err != nil {
		t.Fatal(err)
	}

	if err := cs.SetData("U12345678", "C12345678", "T12345678", "foo", "bar"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"", fields{&Bot{cs: cs, teamID: "T12345678"}, "U12345678", "C12345678"}, args{"foo"}, "bar", false},
		{"", fields{&Bot{cs: cs, teamID: "T12345678"}, "U12345678", "C12345678"}, args{"baz"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controls{
				b:       tt.fields.b,
				user:    tt.fields.user,
				channel: tt.fields.channel,
			}

			got, err := c.Get(tt.args.key)
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
	type fields struct {
		b       *Bot
		user    string
		channel string
	}

	type args struct {
		key   string
		value string
	}

	cs := NewMemoryConversationStore()

	if err := cs.Start("U12345678", "C12345678", "T12345678", "test"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"", fields{&Bot{cs: cs, teamID: "T12345678"}, "U12345678", "C12345678"}, args{"foo", "bar"}, false},
		{"", fields{&Bot{cs: cs, teamID: "T12345678"}, "U12345678", "C12345678"}, args{"foo", "baz"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controls{
				b:       tt.fields.b,
				user:    tt.fields.user,
				channel: tt.fields.channel,
			}

			if err := c.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Controls.Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			if cs.data[tt.fields.user+"_"+tt.fields.channel+"_"+tt.fields.b.teamID][tt.args.key] != tt.args.value {
				t.Errorf("Controls.Get() = %v, want %v", cs.data[tt.fields.user+"_"+tt.fields.channel+"_"+tt.fields.b.teamID][tt.args.key], tt.args.value)
			}
		})
	}
}

func TestControls_To(t *testing.T) {
	type fields struct {
		b       *Bot
		user    string
		channel string
	}

	type args struct {
		state string
	}

	cs := NewMemoryConversationStore()

	if err := cs.Start("U12345678", "C12345678", "T12345678", "test"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"", fields{&Bot{cs: cs, teamID: "T12345678"}, "U12345678", "C12345678"}, args{"next"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controls{
				b:       tt.fields.b,
				user:    tt.fields.user,
				channel: tt.fields.channel,
			}

			if err := c.To(tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("Controls.To() error = %v, wantErr %v", err, tt.wantErr)
			}

			if cs.active[tt.fields.user+"_"+tt.fields.channel+"_"+tt.fields.b.teamID].state != tt.args.state {
				t.Errorf("Controls.To() = %v, want %v", cs.active[tt.fields.user+"_"+tt.fields.channel+"_"+tt.fields.b.teamID].state, tt.args.state)
			}
		})
	}
}

func TestControls_End(t *testing.T) {
	type fields struct {
		b       *Bot
		user    string
		channel string
	}

	cs := NewMemoryConversationStore()

	if err := cs.Start("U12345678", "C12345678", "T12345678", "test"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"", fields{&Bot{cs: cs, teamID: "T12345678"}, "U12345678", "C12345678"}, false},
		{"", fields{&Bot{cs: cs, teamID: "T12345678"}, "U12345678", "C12345678"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controls{
				b:       tt.fields.b,
				user:    tt.fields.user,
				channel: tt.fields.channel,
			}

			if err := c.End(); (err != nil) != tt.wantErr {
				t.Errorf("Controls.End() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestControls_Bot(t *testing.T) {
	type fields struct {
		b       *Bot
		user    string
		channel string
	}

	tests := []struct {
		name   string
		fields fields
		want   *Bot
	}{
		{"", fields{&Bot{}, "", ""}, &Bot{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controls{
				b:       tt.fields.b,
				user:    tt.fields.user,
				channel: tt.fields.channel,
			}

			if got := c.Bot(); !reflect.DeepEqual(got, tt.want) {
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
	type fields struct {
		mp map[string]ConversationHandler
	}

	type args struct {
		state   string
		handler ConversationHandler
	}

	mp := make(map[string]ConversationHandler)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"", fields{mp}, args{"next", func(msg *chat.Message, controls *Controls) {}}, false},
		{"", fields{mp}, args{"next", func(msg *chat.Message, controls *Controls) {}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Conversation{
				mp: tt.fields.mp,
			}

			if err := s.On(tt.args.state, tt.args.handler); (err != nil) != tt.wantErr {
				t.Errorf("Conversation.On() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConversationRegistry(t *testing.T) {
	tests := []struct {
		name string
		want ConversationRegistry
	}{
		{"", make(ConversationRegistry)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConversationRegistry(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConversationRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConversationRegistry_Add(t *testing.T) {
	type args struct {
		name string
		conv *Conversation
	}

	c := NewConversationRegistry()

	conv := NewConversation()
	conv.On("start", func(*chat.Message, *Controls) {})

	tests := []struct {
		name    string
		c       ConversationRegistry
		args    args
		wantErr bool
	}{
		{"", c, args{"", NewConversation()}, true},
		{"", c, args{"", conv}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Add(tt.args.name, tt.args.conv); (err != nil) != tt.wantErr {
				t.Errorf("ConversationRegistry.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConversationRegistry_Get(t *testing.T) {
	type args struct {
		name string
	}

	conv := NewConversation()
	conv.On("start", func(*chat.Message, *Controls) {})

	c := NewConversationRegistry()
	if err := c.Add("foo", conv); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		c       ConversationRegistry
		args    args
		want    *Conversation
		wantErr bool
	}{
		{"", c, args{"foo"}, conv, false},
		{"", c, args{"bar"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Get(tt.args.name)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConversationRegistry.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConversationRegistry.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
