package slack

import (
	"reflect"
	"testing"

	"suy.io/bots/slack/api/oauth"
)

func TestNewMemoryBotStore(t *testing.T) {
	tests := []struct {
		name string
		want *MemoryBotStore
	}{
		{"", &MemoryBotStore{make(map[string]*oauth.AccessResponse)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemoryBotStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemoryBotStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryBotStore_AddBot(t *testing.T) {
	type args struct {
		p *oauth.AccessResponse
	}

	bs := NewMemoryBotStore()

	tests := []struct {
		name    string
		bs      *MemoryBotStore
		args    args
		wantErr bool
	}{
		{"", bs, args{&oauth.AccessResponse{TeamID: "T1234567"}}, false},
		{"", bs, args{&oauth.AccessResponse{TeamID: "T1234567"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bs.AddBot(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("MemoryBotStore.AddBot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryBotStore_GetBot(t *testing.T) {
	type args struct {
		team string
	}

	bs := NewMemoryBotStore()
	bs.AddBot(&oauth.AccessResponse{
		TeamID: "T7654321",
	})

	tests := []struct {
		name    string
		bs      *MemoryBotStore
		args    args
		want    *oauth.AccessResponse
		wantErr bool
	}{
		{"", bs, args{"T1234567"}, nil, true},
		{"", bs, args{"T7654321"}, &oauth.AccessResponse{TeamID: "T7654321"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bs.GetBot(tt.args.team)

			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryBotStore.GetBot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryBotStore.GetBot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryBotStore_RemoveBot(t *testing.T) {
	type args struct {
		team string
	}

	bs := NewMemoryBotStore()
	bs.AddBot(&oauth.AccessResponse{
		TeamID: "T7654321",
	})

	tests := []struct {
		name    string
		bs      *MemoryBotStore
		args    args
		wantErr bool
	}{
		{"", bs, args{"T1234567"}, true},
		{"", bs, args{"T7654321"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bs.RemoveBot(tt.args.team); (err != nil) != tt.wantErr {
				t.Errorf("MemoryBotStore.RemoveBot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryBotStore_AllBots(t *testing.T) {
	bs := NewMemoryBotStore()
	bs.AddBot(&oauth.AccessResponse{
		TeamID: "T7654321",
	})

	tests := []struct {
		name    string
		bs      *MemoryBotStore
		want    []*oauth.AccessResponse
		wantErr bool
	}{
		{"", bs, []*oauth.AccessResponse{
			{TeamID: "T7654321"},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bs.AllBots()

			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryBotStore.AllBots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryBotStore.AllBots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMemoryConversationStore(t *testing.T) {
	tests := []struct {
		name string
		want *MemoryConversationStore
	}{
		{"", NewMemoryConversationStore()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemoryConversationStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemoryConversationStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryConversationStore_Start(t *testing.T) {
	type args struct {
		user    string
		channel string
		team    string
		id      string
	}

	s := NewMemoryConversationStore()

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		wantErr bool
	}{
		{"", s, args{"U1234567", "C1234567", "T1234567", "test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Start(tt.args.user, tt.args.channel, tt.args.team, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryConversationStore_IsActive(t *testing.T) {
	type args struct {
		user    string
		channel string
		team    string
	}

	s := NewMemoryConversationStore()
	s.Start("U1234567", "C1234567", "T1234567", "test")

	tests := []struct {
		name string
		s    *MemoryConversationStore
		args args
		want bool
	}{
		{"", s, args{"U1234567", "C1234567", "T1234567"}, true},
		{"", s, args{"U1234567", "C1234567", "T1234568"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsActive(tt.args.user, tt.args.channel, tt.args.team); got != tt.want {
				t.Errorf("MemoryConversationStore.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryConversationStore_Active(t *testing.T) {
	type args struct {
		user    string
		channel string
		team    string
	}

	s := NewMemoryConversationStore()
	s.Start("U1234567", "C1234567", "T1234567", "test")

	tests := []struct {
		name      string
		s         *MemoryConversationStore
		args      args
		wantID    string
		wantState string
		wantErr   bool
	}{
		{"", s, args{"U1234567", "C1234567", "T1234567"}, "test", "start", false},
		{"", s, args{"U1234567", "C1234567", "T1234568"}, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotState, err := tt.s.Active(tt.args.user, tt.args.channel, tt.args.team)

			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.Active() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotID != tt.wantID {
				t.Errorf("MemoryConversationStore.Active() gotId = %v, want %v", gotID, tt.wantID)
			}

			if gotState != tt.wantState {
				t.Errorf("MemoryConversationStore.Active() gotState = %v, want %v", gotState, tt.wantState)
			}
		})
	}
}

func TestMemoryConversationStore_SetState(t *testing.T) {
	type args struct {
		user    string
		channel string
		team    string
		state   string
	}

	s := NewMemoryConversationStore()
	s.Start("U1234567", "C1234567", "T1234567", "test")

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		wantErr bool
	}{
		{"", s, args{"U1234567", "C1234567", "T1234567", "next"}, false},
		{"", s, args{"U1234567", "C1234567", "T1234568", "next"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.SetState(tt.args.user, tt.args.channel, tt.args.team, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.SetState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryConversationStore_SetData(t *testing.T) {
	type args struct {
		user    string
		channel string
		team    string
		key     string
		value   string
	}

	s := NewMemoryConversationStore()
	s.Start("U1234567", "C1234567", "T1234567", "test")

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		wantErr bool
	}{
		{"", s, args{"U1234567", "C1234567", "T1234567", "foo", "bar"}, false},
		{"", s, args{"U1234567", "C1234567", "T1234568", "foo", "bar"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.SetData(tt.args.user, tt.args.channel, tt.args.team, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.SetData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryConversationStore_GetData(t *testing.T) {
	type args struct {
		user    string
		channel string
		team    string
		key     string
	}

	s := NewMemoryConversationStore()
	s.Start("U1234567", "C1234567", "T1234567", "test")
	s.SetData("U1234567", "C1234567", "T1234567", "foo", "bar")

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		want    string
		wantErr bool
	}{
		{"", s, args{"U1234567", "C1234567", "T1234567", "foo"}, "bar", false},
		{"", s, args{"U1234567", "C1234567", "T1234567", "baz"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetData(tt.args.user, tt.args.channel, tt.args.team, tt.args.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.GetData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("MemoryConversationStore.GetData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryConversationStore_End(t *testing.T) {
	type args struct {
		user    string
		channel string
		team    string
	}

	s := NewMemoryConversationStore()
	s.Start("U1234567", "C1234567", "T1234567", "test")

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		wantErr bool
	}{
		{"", s, args{"U1234567", "C1234567", "T1234567"}, false},
		{"", s, args{"U1234567", "C1234567", "T1234568"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.End(tt.args.user, tt.args.channel, tt.args.team); (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.End() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
