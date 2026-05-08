package labels

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantSet    map[string]string
		wantRemove []string
		wantErr    bool
	}{
		{
			name:    "empty",
			args:    nil,
			wantSet: map[string]string{},
		},
		{
			name:    "single set",
			args:    []string{"env=prod"},
			wantSet: map[string]string{"env": "prod"},
		},
		{
			name:    "multiple set",
			args:    []string{"env=prod", "team=platform"},
			wantSet: map[string]string{"env": "prod", "team": "platform"},
		},
		{
			name:       "single remove",
			args:       []string{"env-"},
			wantSet:    map[string]string{},
			wantRemove: []string{"env"},
		},
		{
			name:       "mixed",
			args:       []string{"env=prod", "owner-", "team=platform"},
			wantSet:    map[string]string{"env": "prod", "team": "platform"},
			wantRemove: []string{"owner"},
		},
		{
			name:    "value with equals",
			args:    []string{"url=https://x.com/y=z"},
			wantSet: map[string]string{"url": "https://x.com/y=z"},
		},
		{
			name:    "empty value allowed",
			args:    []string{"foo="},
			wantSet: map[string]string{"foo": ""},
		},
		{
			name:    "no separator",
			args:    []string{"foo"},
			wantErr: true,
		},
		{
			name:    "empty key",
			args:    []string{"=value"},
			wantErr: true,
		},
		{
			name:    "bare dash",
			args:    []string{"-"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set, remove, err := ParseArgs(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if !reflect.DeepEqual(set, tc.wantSet) {
				t.Errorf("set = %v, want %v", set, tc.wantSet)
			}
			if !reflect.DeepEqual(remove, tc.wantRemove) {
				t.Errorf("remove = %v, want %v", remove, tc.wantRemove)
			}
		})
	}
}
