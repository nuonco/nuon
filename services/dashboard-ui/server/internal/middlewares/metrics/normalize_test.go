package metrics

import "testing"

func TestNormalizeAPIPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"single id", "/v1/apps/app98e2wpzdxwoey393edtqj45", "/v1/apps/{app_id}"},
		{"id with trailing segment", "/v1/apps/app98e2wpzdxwoey393edtqj45/installs", "/v1/apps/{app_id}/installs"},
		{"nested ids", "/v1/apps/app98e2wpzdxwoey393edtqj45/installs/inl98e2wpzdxwoey393edtqj45", "/v1/apps/{app_id}/installs/{inl_id}"},
		{"org id", "/v1/orgs/orgrok933tcyzji01s7us3aeo3", "/v1/orgs/{org_id}"},
		{"uuid", "/v1/things/123e4567-e89b-12d3-a456-426614174000", "/v1/things/{uuid}"},
		{"numeric", "/v1/things/12345", "/v1/things/{id}"},
		{"no ids", "/v1/apps", "/v1/apps"},
		{"trailing slash", "/v1/apps/app98e2wpzdxwoey393edtqj45/", "/v1/apps/{app_id}"},
		{"root", "/", "/"},
		{"empty", "", "/"},
		{"short word not an id", "/v1/apps/latest", "/v1/apps/latest"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeAPIPath(tc.in); got != tc.want {
				t.Errorf("normalizeAPIPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
