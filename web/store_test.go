package web

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewMemoryControllerStore(t *testing.T) {
	tests := []struct {
		name string
		want *MemoryControllerStore
	}{
		{"", &MemoryControllerStore{make(map[BotID]ItemStore)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemoryControllerStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemoryControllerStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryControllerStore_Add(t *testing.T) {
	type args struct {
		id BotID
	}

	s := NewMemoryControllerStore()

	tests := []struct {
		name    string
		s       *MemoryControllerStore
		args    args
		wantErr bool
	}{
		{"", s, args{0}, false},
		{"", s, args{0}, true},
		{"", s, args{1}, false},
		{"", s, args{1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Add(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("MemoryControllerStore.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryControllerStore_Get(t *testing.T) {
	type args struct {
		id BotID
	}

	s := NewMemoryControllerStore()
	s.Add(0)

	tests := []struct {
		name    string
		s       *MemoryControllerStore
		args    args
		want    ItemStore
		wantErr bool
	}{
		{"", s, args{0}, NewMemoryItemStore(), false},
		{"", s, args{1}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryControllerStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryControllerStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryControllerStore_Remove(t *testing.T) {
	type args struct {
		id BotID
	}

	s := NewMemoryControllerStore()
	s.Add(0)

	tests := []struct {
		name    string
		s       *MemoryControllerStore
		args    args
		wantErr bool
	}{
		{"", s, args{0}, false},
		{"", s, args{1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Remove(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("MemoryControllerStore.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMemoryItemStore(t *testing.T) {
	tests := []struct {
		name string
		want *MemoryItemStore
	}{
		{"", &MemoryItemStore{messages: make(map[ItemID]*ItemSet)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemoryItemStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemoryItemStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryItemStore_Add(t *testing.T) {
	type args struct {
		item Item
	}

	s := NewMemoryItemStore()

	tests := []struct {
		name    string
		store   *MemoryItemStore
		args    args
		wantErr bool
	}{
		{"nil item", s, args{nil}, true},
		{"simple message", s, args{TextMessage("add")}, false},
		{"thread with id 0", s, args{newThread(TextMessage("thread add"))}, true},
		{"new thread from message", s, args{newThread(&Message{Text: "thread add", ID: 21, ThreadID: 42})}, false},
	}

	var wg sync.WaitGroup
	wg.Add(len(tests))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func(tt struct {
				name    string
				store   *MemoryItemStore
				args    args
				wantErr bool
			}, t *testing.T) {
				if err := tt.store.Add(tt.args.item); (err != nil) != tt.wantErr {
					t.Errorf("MemoryItemStore.Add() error = %v, wantErr %v", err, tt.wantErr)
				}

				wg.Done()
			}(tt, t)
		})
	}

	wg.Wait()

	if len(s.messages) != 2 {
		t.Errorf("MemoryItemStore.Add() expected to add %v threads, added %v", 2, len(s.messages))
	}

	if len(s.messages[0].items) != 1 {
		t.Errorf("MemoryItemStore.Add() expected to have 1 item in thread 0, got %v", len(s.messages[0].items))
	}

	if len(s.messages[42].items) != 1 {
		t.Errorf("MemoryItemStore.Add() expected to have 1 item in thread 42, got %v", len(s.messages[42].items))
	}
}

func TestMemoryItemStore_Update(t *testing.T) {
	type args struct {
		item Item
	}

	s := NewMemoryItemStore()
	m := TextMessage("a")
	m.ID = 42

	if err := s.Add(m); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		store   *MemoryItemStore
		args    args
		wantErr bool
	}{
		{"", s, args{nil}, true},
		{"", s, args{TextMessage("random")}, true},
		{"", s, args{m}, false},
		{"", s, args{&Message{Text: "update", ID: m.ID}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func(tt struct {
				name    string
				store   *MemoryItemStore
				args    args
				wantErr bool
			}) {
				if err := tt.store.Update(tt.args.item); (err != nil) != tt.wantErr {
					t.Errorf("MemoryItemStore.Update() error = %v, wantErr %v", err, tt.wantErr)
				}
			}(tt)
		})
	}
}

func TestMemoryItemStore_Get(t *testing.T) {
	type args struct {
		id     ItemID
		thread ItemID
	}

	s := NewMemoryItemStore()

	if err := s.Add(&Message{ID: 0, ThreadID: 0}); err != nil {
		t.Fatal(err)
	}

	if err := s.Add(&Thread{ID: 2, ThreadID: 1}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		store   *MemoryItemStore
		args    args
		want    Item
		wantErr bool
	}{
		{"", s, args{0, 0}, &Message{ID: 0, ThreadID: 0}, false},
		{"", s, args{0, 1}, nil, true},
		{"", s, args{2, 1}, &Thread{ID: 2, ThreadID: 1}, false},
		{"", s, args{2, 0}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.store.Get(tt.args.id, tt.args.thread)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryItemStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryItemStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryItemStore_All(t *testing.T) {
	type args struct {
		thread ItemID
	}

	s := NewMemoryItemStore()
	if err := s.Add(&Message{ID: 0, ThreadID: 0}); err != nil {
		t.Fatal(err)
	}

	if err := s.Add(&Thread{ID: 2, ThreadID: 1}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		store   *MemoryItemStore
		args    args
		want    []Item
		wantErr bool
	}{
		{"", s, args{0}, []Item{&Message{ID: 0, ThreadID: 0}}, false},
		{"", s, args{1}, []Item{&Thread{ID: 2, ThreadID: 1}}, false},
		{"", s, args{2}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.store.All(tt.args.thread)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryItemStore.All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryItemStore.All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewItemSet(t *testing.T) {
	tests := []struct {
		name string
		want *ItemSet
	}{
		{"", &ItemSet{nil, make(map[ItemID]Item)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewItemSet(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewItemSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItemSet_Add(t *testing.T) {
	type args struct {
		item Item
	}

	s := NewItemSet()

	tests := []struct {
		name string
		set  *ItemSet
		args args
	}{
		{"", s, args{&Message{ID: 0}}},
		{"", s, args{&Thread{ID: 1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.set.Add(tt.args.item)
			if !reflect.DeepEqual(tt.set.items[tt.args.item.ItemID()], tt.args.item) {
				t.Errorf("ItemSet.Add() = %v, want %v", tt.set.items[tt.args.item.ItemID()], tt.args.item)
			}
		})
	}
}

func TestItemSet_Get(t *testing.T) {
	type args struct {
		id ItemID
	}

	s, m1, t1, m2 := NewItemSet(), &Message{ID: 1}, &Thread{ID: 2}, &Message{ID: 3}
	s.Add(m1)
	s.Add(t1)
	s.Add(m2)

	tests := []struct {
		name    string
		set     *ItemSet
		args    args
		want    Item
		wantErr bool
	}{
		{"", s, args{1}, &Message{ID: 1, ThreadID: 0, Next: &Cursor{ThreadItemType, 2}}, false},
		{"", s, args{2}, &Thread{ID: 2, ThreadID: 0, Prev: &Cursor{MessageItemType, 1}, Next: &Cursor{MessageItemType, 3}}, false},
		{"", s, args{4}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.set.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ItemSet.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ItemSet.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItemSet_Set(t *testing.T) {
	type args struct {
		i Item
	}

	s := NewItemSet()
	s.Add(&Message{Text: "a", ID: 1})

	tests := []struct {
		name    string
		set     *ItemSet
		args    args
		want    Item
		wantErr bool
	}{
		{"", s, args{&Message{Text: "b", ID: 1}}, &Message{Text: "b", ID: 1}, false},
		{"", s, args{&Message{Text: "b", ID: 2}}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.set.Set(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("ItemSet.Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got, err := tt.set.Get(tt.args.i.ItemID()); (err != nil) != tt.wantErr || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ItemSet.Set() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItemSet_Len(t *testing.T) {
	s := NewItemSet()
	s.Add(TextMessage("a"))
	s.Add(&Thread{ID: 2})

	tests := []struct {
		name string
		set  *ItemSet
		want int
	}{
		{"", s, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.set.Len(); got != tt.want {
				t.Errorf("ItemSet.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMemoryConversationStore(t *testing.T) {
	tests := []struct {
		name string
		want *MemoryConversationStore
	}{
		{"", &MemoryConversationStore{make(map[BotID]*convdata), make(map[BotID]map[string]string)}},
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
		bot BotID
		id  string
	}

	s := NewMemoryConversationStore()

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		wantErr bool
	}{
		{"", s, args{1, "test"}, false},
		{"", s, args{1, "test"}, true},
		{"", s, args{1, "test 2"}, true},
		{"", s, args{2, "test 2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Start(tt.args.bot, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryConversationStore_IsActive(t *testing.T) {
	type args struct {
		bot BotID
	}

	s := NewMemoryConversationStore()

	if err := s.Start(1, "test"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		s    *MemoryConversationStore
		args args
		want bool
	}{
		{"", s, args{1}, true},
		{"", s, args{2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsActive(tt.args.bot); got != tt.want {
				t.Errorf("MemoryConversationStore.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryConversationStore_Active(t *testing.T) {
	type args struct {
		bot BotID
	}

	s := NewMemoryConversationStore()

	if err := s.Start(1, "test"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		s         *MemoryConversationStore
		args      args
		wantID    string
		wantState string
		wantErr   bool
	}{
		{"", s, args{1}, "test", "start", false},
		{"", s, args{2}, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotState, err := tt.s.Active(tt.args.bot)
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
		bot   BotID
		state string
	}

	s := NewMemoryConversationStore()

	if err := s.Start(1, "test"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		wantErr bool
	}{
		{"", s, args{1, "start"}, false},
		{"", s, args{1, "next"}, false},
		{"", s, args{2, "next"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.SetState(tt.args.bot, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.SetState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryConversationStore_SetData(t *testing.T) {
	type args struct {
		bot   BotID
		key   string
		value string
	}

	s := NewMemoryConversationStore()

	if err := s.Start(1, "test"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		wantErr bool
	}{
		{"", s, args{1, "a", "b"}, false},
		{"", s, args{1, "a", "c"}, false},
		{"", s, args{2, "a", "c"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.SetData(tt.args.bot, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.SetData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryConversationStore_GetData(t *testing.T) {
	type args struct {
		bot BotID
		key string
	}

	s := NewMemoryConversationStore()

	if err := s.Start(1, "test"); err != nil {
		t.Fatal(err)
	}

	if err := s.SetData(1, "a", "b"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		want    string
		wantErr bool
	}{
		{"", s, args{1, "a"}, "b", false},
		{"", s, args{2, "a"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetData(tt.args.bot, tt.args.key)
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
		bot BotID
	}

	s := NewMemoryConversationStore()

	if err := s.Start(1, "test"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		s       *MemoryConversationStore
		args    args
		wantErr bool
	}{
		{"", s, args{1}, false},
		{"", s, args{2}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.End(tt.args.bot); (err != nil) != tt.wantErr {
				t.Errorf("MemoryConversationStore.End() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
