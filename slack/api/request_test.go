package api

import (
	"testing"
)

func TestRequest(t *testing.T) {
	type args struct {
		method string
		body   interface{}
		isJSON bool
		res    interface{}
		token  string
	}

	res := &struct {
		OK bool `json:"ok" url:"ok"`
	}{
		OK: false,
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{"api.test", nil, true, res, ""}, false},
		{"", args{"api.test", nil, false, res, ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Request(tt.args.method, tt.args.body, tt.args.isJSON, tt.args.res, tt.args.token); (err != nil) != tt.wantErr || !res.OK {
				t.Errorf("Request() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
