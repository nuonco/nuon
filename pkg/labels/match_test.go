package labels

import (
	"testing"
)

func TestSubscriptionMatch_Matches(t *testing.T) {
	tests := []struct {
		name string
		m    *SubscriptionMatch
		t    EventTargets
		want bool
	}{
		{
			name: "nil receiver matches everything",
			m:    nil,
			t:    EventTargets{InstallID: "ins_1"},
			want: true,
		},
		{
			name: "no kinds populated never matches",
			m:    &SubscriptionMatch{},
			t:    EventTargets{InstallID: "ins_1"},
			want: false,
		},
		{
			name: "empty install TargetMatch matches when install id present",
			m:    &SubscriptionMatch{Installs: &TargetMatch{}},
			t:    EventTargets{InstallID: "ins_1"},
			want: true,
		},
		{
			name: "empty install TargetMatch misses when install id absent",
			m:    &SubscriptionMatch{Installs: &TargetMatch{}},
			t:    EventTargets{ComponentID: "cmp_1"},
			want: false,
		},
		{
			name: "id-only hit",
			m:    &SubscriptionMatch{Installs: &TargetMatch{IDs: []string{"ins_1", "ins_2"}}},
			t:    EventTargets{InstallID: "ins_2"},
			want: true,
		},
		{
			name: "id-only miss",
			m:    &SubscriptionMatch{Installs: &TargetMatch{IDs: []string{"ins_1"}}},
			t:    EventTargets{InstallID: "ins_other"},
			want: false,
		},
		{
			name: "selector-only hit",
			m: &SubscriptionMatch{Installs: &TargetMatch{
				Selector: &Selector{MatchLabels: Labels{"env": "prod"}},
			}},
			t:    EventTargets{InstallID: "ins_1", InstallLabels: Labels{"env": "prod"}},
			want: true,
		},
		{
			name: "selector-only miss",
			m: &SubscriptionMatch{Installs: &TargetMatch{
				Selector: &Selector{MatchLabels: Labels{"env": "prod"}},
			}},
			t:    EventTargets{InstallID: "ins_1", InstallLabels: Labels{"env": "stage"}},
			want: false,
		},
		{
			name: "id hit beats selector miss within target (within-target OR)",
			m: &SubscriptionMatch{Installs: &TargetMatch{
				IDs:      []string{"ins_1"},
				Selector: &Selector{MatchLabels: Labels{"env": "prod"}},
			}},
			t:    EventTargets{InstallID: "ins_1", InstallLabels: Labels{"env": "stage"}},
			want: true,
		},
		{
			name: "selector hit beats id miss within target",
			m: &SubscriptionMatch{Installs: &TargetMatch{
				IDs:      []string{"ins_other"},
				Selector: &Selector{MatchLabels: Labels{"env": "prod"}},
			}},
			t:    EventTargets{InstallID: "ins_1", InstallLabels: Labels{"env": "prod"}},
			want: true,
		},
		{
			name: "OR across kinds: installs miss, components hit",
			m: &SubscriptionMatch{
				Installs:   &TargetMatch{IDs: []string{"ins_other"}},
				Components: &TargetMatch{IDs: []string{"cmp_1"}},
			},
			t:    EventTargets{InstallID: "ins_1", ComponentID: "cmp_1"},
			want: true,
		},
		{
			name: "OR across kinds: both miss",
			m: &SubscriptionMatch{
				Installs:   &TargetMatch{IDs: []string{"ins_other"}},
				Components: &TargetMatch{IDs: []string{"cmp_other"}},
			},
			t:    EventTargets{InstallID: "ins_1", ComponentID: "cmp_1"},
			want: false,
		},
		{
			name: "actions kind",
			m:    &SubscriptionMatch{Actions: &TargetMatch{IDs: []string{"act_1"}}},
			t:    EventTargets{ActionID: "act_1"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Matches(tt.t)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTargetMatch_matches_EmptyIDShortCircuits(t *testing.T) {
	// Even the permissive empty TargetMatch{} must reject id == "" — otherwise a
	// component-only event would falsely satisfy an installs filter.
	tm := &TargetMatch{}
	if tm.matches("", Labels{"env": "prod"}) {
		t.Errorf("empty TargetMatch should not match empty id")
	}

	tmIDs := &TargetMatch{IDs: []string{"x"}}
	if tmIDs.matches("", nil) {
		t.Errorf("TargetMatch with IDs should not match empty id")
	}

	tmSel := &TargetMatch{Selector: &Selector{MatchLabels: Labels{"env": "prod"}}}
	if tmSel.matches("", Labels{"env": "prod"}) {
		t.Errorf("TargetMatch with Selector should not match empty id")
	}
}

func TestSubscriptionMatch_Validate(t *testing.T) {
	tests := []struct {
		name    string
		m       *SubscriptionMatch
		wantErr bool
	}{
		{
			name:    "nil rejected",
			m:       nil,
			wantErr: true,
		},
		{
			name:    "all kinds nil rejected",
			m:       &SubscriptionMatch{},
			wantErr: true,
		},
		{
			name:    "empty TargetMatch is valid",
			m:       &SubscriptionMatch{Installs: &TargetMatch{}},
			wantErr: false,
		},
		{
			name: "valid IDs",
			m: &SubscriptionMatch{
				Installs: &TargetMatch{IDs: []string{"ins_1"}},
			},
			wantErr: false,
		},
		{
			name: "empty id in IDs rejected",
			m: &SubscriptionMatch{
				Installs: &TargetMatch{IDs: []string{"ins_1", ""}},
			},
			wantErr: true,
		},
		{
			name: "selector with empty match labels rejected",
			m: &SubscriptionMatch{
				Components: &TargetMatch{Selector: &Selector{}},
			},
			wantErr: true,
		},
		{
			name: "valid selector",
			m: &SubscriptionMatch{
				Components: &TargetMatch{Selector: &Selector{MatchLabels: Labels{"env": "prod"}}},
			},
			wantErr: false,
		},
		{
			name: "multiple kinds populated valid",
			m: &SubscriptionMatch{
				Installs: &TargetMatch{IDs: []string{"ins_1"}},
				Actions:  &TargetMatch{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubscriptionMatch_Canonical(t *testing.T) {
	t.Run("nil and zero produce empty", func(t *testing.T) {
		var m *SubscriptionMatch
		if got := m.Canonical(); got != "" {
			t.Errorf("nil Canonical = %q, want empty", got)
		}
		zero := &SubscriptionMatch{}
		if got := zero.Canonical(); got != "" {
			t.Errorf("zero Canonical = %q, want empty", got)
		}
	})

	t.Run("stable across IDs order", func(t *testing.T) {
		a := &SubscriptionMatch{Installs: &TargetMatch{IDs: []string{"a", "b", "c"}}}
		b := &SubscriptionMatch{Installs: &TargetMatch{IDs: []string{"c", "a", "b"}}}
		if a.Canonical() != b.Canonical() {
			t.Errorf("Canonical not stable across ids order:\n  a=%s\n  b=%s", a.Canonical(), b.Canonical())
		}
	})

	t.Run("stable across selector key order", func(t *testing.T) {
		a := &SubscriptionMatch{Components: &TargetMatch{
			Selector: &Selector{MatchLabels: Labels{"env": "prod", "team": "platform"}},
		}}
		b := &SubscriptionMatch{Components: &TargetMatch{
			Selector: &Selector{MatchLabels: Labels{"team": "platform", "env": "prod"}},
		}}
		if a.Canonical() != b.Canonical() {
			t.Errorf("Canonical not stable across selector key order:\n  a=%s\n  b=%s", a.Canonical(), b.Canonical())
		}
	})

	t.Run("differs for semantically different matches", func(t *testing.T) {
		a := &SubscriptionMatch{Installs: &TargetMatch{IDs: []string{"ins_1"}}}
		b := &SubscriptionMatch{Components: &TargetMatch{IDs: []string{"ins_1"}}}
		if a.Canonical() == b.Canonical() {
			t.Errorf("Canonical should differ when kind differs")
		}
	})
}

func TestSubscriptionMatch_ScanValueRoundTrip(t *testing.T) {
	t.Run("nil scan", func(t *testing.T) {
		var m SubscriptionMatch
		if err := m.Scan(nil); err != nil {
			t.Fatalf("Scan(nil) err = %v", err)
		}
		if !m.isZero() {
			t.Errorf("Scan(nil) should leave zero match")
		}
	})

	t.Run("empty bytes scan", func(t *testing.T) {
		var m SubscriptionMatch
		if err := m.Scan([]byte{}); err != nil {
			t.Fatalf("Scan(empty bytes) err = %v", err)
		}
		if !m.isZero() {
			t.Errorf("Scan(empty bytes) should leave zero match")
		}
	})

	t.Run("empty string scan", func(t *testing.T) {
		var m SubscriptionMatch
		if err := m.Scan(""); err != nil {
			t.Fatalf("Scan(empty string) err = %v", err)
		}
		if !m.isZero() {
			t.Errorf("Scan(empty string) should leave zero match")
		}
	})

	t.Run("zero match values to nil", func(t *testing.T) {
		var m SubscriptionMatch
		v, err := m.Value()
		if err != nil {
			t.Fatalf("Value() err = %v", err)
		}
		if v != nil {
			t.Errorf("zero Value() = %v, want nil", v)
		}
	})

	t.Run("populated round-trips", func(t *testing.T) {
		orig := SubscriptionMatch{
			Installs: &TargetMatch{IDs: []string{"ins_1"}},
			Components: &TargetMatch{
				Selector: &Selector{MatchLabels: Labels{"env": "prod"}},
			},
		}
		v, err := orig.Value()
		if err != nil {
			t.Fatalf("Value() err = %v", err)
		}
		bytes, ok := v.([]byte)
		if !ok {
			t.Fatalf("Value() = %T, want []byte", v)
		}
		var got SubscriptionMatch
		if err := got.Scan(bytes); err != nil {
			t.Fatalf("Scan() err = %v", err)
		}
		if got.Canonical() != orig.Canonical() {
			t.Errorf("round-trip mismatch:\n  orig=%s\n  got=%s", orig.Canonical(), got.Canonical())
		}
	})

	t.Run("scan unsupported type", func(t *testing.T) {
		var m SubscriptionMatch
		if err := m.Scan(123); err == nil {
			t.Errorf("Scan(int) expected error")
		}
	})
}

func TestSubscriptionMatch_GormDataType(t *testing.T) {
	if got := (SubscriptionMatch{}).GormDataType(); got != "jsonb" {
		t.Errorf("GormDataType() = %q, want %q", got, "jsonb")
	}
}
