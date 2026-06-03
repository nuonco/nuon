package labels

import (
	"testing"
)

func TestSelector_Matches(t *testing.T) {
	tests := []struct {
		name string
		sel  *Selector
		set  Labels
		want bool
	}{
		{
			name: "nil selector matches everything",
			sel:  nil,
			set:  Labels{"env": "prod"},
			want: true,
		},
		{
			name: "exact equality hit",
			sel:  &Selector{MatchLabels: Labels{"env": "prod"}},
			set:  Labels{"env": "prod", "team": "platform"},
			want: true,
		},
		{
			name: "exact equality miss on value",
			sel:  &Selector{MatchLabels: Labels{"env": "prod"}},
			set:  Labels{"env": "stage"},
			want: false,
		},
		{
			name: "missing key returns false",
			sel:  &Selector{MatchLabels: Labels{"env": "prod"}},
			set:  Labels{"team": "platform"},
			want: false,
		},
		{
			name: "wildcard requires key to exist",
			sel:  &Selector{MatchLabels: Labels{"env": "*"}},
			set:  Labels{"env": "anything"},
			want: true,
		},
		{
			name: "wildcard miss when key absent",
			sel:  &Selector{MatchLabels: Labels{"env": "*"}},
			set:  Labels{"team": "platform"},
			want: false,
		},
		{
			name: "AND across multiple keys hit",
			sel:  &Selector{MatchLabels: Labels{"env": "prod", "team": "platform"}},
			set:  Labels{"env": "prod", "team": "platform", "region": "us-west-2"},
			want: true,
		},
		{
			name: "AND across multiple keys miss on one",
			sel:  &Selector{MatchLabels: Labels{"env": "prod", "team": "platform"}},
			set:  Labels{"env": "prod", "team": "ops"},
			want: false,
		},
		{
			name: "wildcard AND exact mixed",
			sel:  &Selector{MatchLabels: Labels{"env": "prod", "team": "*"}},
			set:  Labels{"env": "prod", "team": "anything"},
			want: true,
		},
		{
			name: "empty MatchLabels matches anything (defensive)",
			sel:  &Selector{MatchLabels: Labels{}},
			set:  Labels{"env": "prod"},
			want: true,
		},
		{
			name: "set is nil and selector requires key",
			sel:  &Selector{MatchLabels: Labels{"env": "prod"}},
			set:  nil,
			want: false,
		},
		{
			name: "not-match excludes set with matching key+value",
			sel:  &Selector{NotMatchLabels: Labels{"env": "stage"}},
			set:  Labels{"env": "stage"},
			want: false,
		},
		{
			name: "not-match allows set with different value for same key",
			sel:  &Selector{NotMatchLabels: Labels{"env": "stage"}},
			set:  Labels{"env": "prod"},
			want: true,
		},
		{
			name: "not-match allows set without the key",
			sel:  &Selector{NotMatchLabels: Labels{"env": "stage"}},
			set:  Labels{"team": "platform"},
			want: true,
		},
		{
			name: "not-match wildcard excludes set with key present",
			sel:  &Selector{NotMatchLabels: Labels{"env": "*"}},
			set:  Labels{"env": "anything"},
			want: false,
		},
		{
			name: "not-match wildcard allows set with key absent",
			sel:  &Selector{NotMatchLabels: Labels{"env": "*"}},
			set:  Labels{"team": "platform"},
			want: true,
		},
		{
			name: "include AND exclude combined: include hit + exclude not triggered",
			sel:  &Selector{MatchLabels: Labels{"team": "platform"}, NotMatchLabels: Labels{"env": "stage"}},
			set:  Labels{"team": "platform", "env": "prod"},
			want: true,
		},
		{
			name: "include AND exclude combined: include hit but exclude triggered",
			sel:  &Selector{MatchLabels: Labels{"team": "platform"}, NotMatchLabels: Labels{"env": "stage"}},
			set:  Labels{"team": "platform", "env": "stage"},
			want: false,
		},
		{
			name: "only NotMatchLabels: empty set is allowed (env=stage exclusion not triggered)",
			sel:  &Selector{NotMatchLabels: Labels{"env": "stage"}},
			set:  Labels{},
			want: true,
		},
		{
			name: "multiple NotMatchLabels: any match excludes",
			sel:  &Selector{NotMatchLabels: Labels{"env": "stage", "tier": "experimental"}},
			set:  Labels{"env": "prod", "tier": "experimental"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sel.Matches(tt.set)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelector_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sel     *Selector
		wantErr bool
	}{
		{
			name:    "nil selector is invalid",
			sel:     nil,
			wantErr: true,
		},
		{
			name:    "empty MatchLabels rejected",
			sel:     &Selector{MatchLabels: Labels{}},
			wantErr: true,
		},
		{
			name:    "nil MatchLabels rejected",
			sel:     &Selector{},
			wantErr: true,
		},
		{
			name:    "blank key rejected",
			sel:     &Selector{MatchLabels: Labels{"": "prod"}},
			wantErr: true,
		},
		{
			name:    "blank NotMatchLabels key rejected",
			sel:     &Selector{NotMatchLabels: Labels{"": "stage"}},
			wantErr: true,
		},
		{
			name:    "only NotMatchLabels is valid",
			sel:     &Selector{NotMatchLabels: Labels{"env": "stage"}},
			wantErr: false,
		},
		{
			name:    "MatchLabels + NotMatchLabels both populated is valid",
			sel:     &Selector{MatchLabels: Labels{"team": "platform"}, NotMatchLabels: Labels{"env": "stage"}},
			wantErr: false,
		},
		{
			name:    "whitespace-only key rejected",
			sel:     &Selector{MatchLabels: Labels{"   ": "prod"}},
			wantErr: true,
		},
		{
			name:    "valid single key",
			sel:     &Selector{MatchLabels: Labels{"env": "prod"}},
			wantErr: false,
		},
		{
			name:    "valid wildcard",
			sel:     &Selector{MatchLabels: Labels{"env": "*"}},
			wantErr: false,
		},
		{
			name:    "valid empty value (still a constraint)",
			sel:     &Selector{MatchLabels: Labels{"env": ""}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sel.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSelector_Canonical_StableAcrossMapOrder(t *testing.T) {
	a := &Selector{MatchLabels: Labels{"env": "prod", "team": "platform", "region": "us-west-2"}}
	b := &Selector{MatchLabels: Labels{"region": "us-west-2", "env": "prod", "team": "platform"}}

	if a.Canonical() != b.Canonical() {
		t.Errorf("Canonical not stable across map order:\n  a=%s\n  b=%s", a.Canonical(), b.Canonical())
	}

	if a.Canonical() == "" {
		t.Errorf("Canonical of populated selector should not be empty")
	}
}

func TestSelector_Canonical_NilEmpty(t *testing.T) {
	var nilSel *Selector
	if got := nilSel.Canonical(); got != "" {
		t.Errorf("nil Canonical = %q, want empty string", got)
	}
}

func TestSelector_Canonical_DistinguishesValues(t *testing.T) {
	a := &Selector{MatchLabels: Labels{"env": "prod"}}
	b := &Selector{MatchLabels: Labels{"env": "stage"}}
	if a.Canonical() == b.Canonical() {
		t.Errorf("Canonical should differ for different values")
	}
}

func TestSelector_Canonical_DistinguishesIncludeFromExclude(t *testing.T) {
	include := &Selector{MatchLabels: Labels{"env": "stage"}}
	exclude := &Selector{NotMatchLabels: Labels{"env": "stage"}}
	if include.Canonical() == exclude.Canonical() {
		t.Errorf("Canonical of include vs exclude must differ:\n  include=%s\n  exclude=%s",
			include.Canonical(), exclude.Canonical())
	}
}

func TestSelector_Canonical_NotMatchLabelsStable(t *testing.T) {
	a := &Selector{NotMatchLabels: Labels{"env": "stage", "tier": "experimental"}}
	b := &Selector{NotMatchLabels: Labels{"tier": "experimental", "env": "stage"}}
	if a.Canonical() != b.Canonical() {
		t.Errorf("Canonical not stable across NotMatchLabels map order:\n  a=%s\n  b=%s",
			a.Canonical(), b.Canonical())
	}
}
