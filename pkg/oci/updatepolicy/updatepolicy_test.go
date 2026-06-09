package updatepolicy

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name       string
		constraint string
		wantErr    bool
	}{
		{"tilde", "~1.25.0", false},
		{"caret", "^2", false},
		{"range", ">=1.0.0,<2.0.0", false},
		{"exact", "1.2.3", false},
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"garbage", "not a constraint", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.constraint)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q, got nil", tc.constraint)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error for %q, got %v", tc.constraint, err)
			}
		})
	}
}

func TestSelectHighestMatching(t *testing.T) {
	cases := []struct {
		name       string
		tags       []string
		constraint string
		want       string
		wantErr    error
	}{
		{
			name:       "tilde picks highest patch",
			tags:       []string{"1.25.0", "1.25.3", "1.25.7", "1.26.0", "latest"},
			constraint: "~1.25.0",
			want:       "1.25.7",
		},
		{
			name:       "caret picks highest minor.patch",
			tags:       []string{"2.0.0", "2.1.5", "2.10.0", "3.0.0"},
			constraint: "^2",
			want:       "2.10.0",
		},
		{
			name:       "non-semver tags skipped silently",
			tags:       []string{"latest", "stable", "main", "develop", "1.2.3"},
			constraint: "^1.0.0",
			want:       "1.2.3",
		},
		{
			name:       "preserves leading v in original tag",
			tags:       []string{"v1.0.0", "v1.1.0"},
			constraint: "^1.0.0",
			want:       "v1.1.0",
		},
		{
			name:       "no match returns ErrNoMatchingTag",
			tags:       []string{"1.0.0", "1.1.0"},
			constraint: "^2.0.0",
			wantErr:    ErrNoMatchingTag,
		},
		{
			name:       "empty input returns ErrNoMatchingTag",
			tags:       nil,
			constraint: "^1.0.0",
			wantErr:    ErrNoMatchingTag,
		},
		{
			name:       "all non-semver returns ErrNoMatchingTag",
			tags:       []string{"latest", "main"},
			constraint: "^1.0.0",
			wantErr:    ErrNoMatchingTag,
		},
		{
			name:       "exact constraint",
			tags:       []string{"1.2.3", "1.2.4"},
			constraint: "1.2.3",
			want:       "1.2.3",
		},
		{
			name:       "prerelease excluded by default",
			tags:       []string{"1.0.0", "1.1.0-rc.1"},
			constraint: "^1.0.0",
			want:       "1.0.0",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelectHighestMatching(tc.tags, tc.constraint)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v (result %q)", tc.wantErr, err, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestSelectHighestMatching_InvalidConstraint(t *testing.T) {
	_, err := SelectHighestMatching([]string{"1.0.0"}, "not a constraint")
	if err == nil {
		t.Fatalf("expected error for invalid constraint")
	}
}
