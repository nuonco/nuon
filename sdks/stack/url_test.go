package stack

import "testing"

func TestParseURL(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		wantBase    string
		wantPhoneID string
		wantErr     bool
	}{
		{
			name:        "https",
			in:          "https://api.nuon.co/v1/stack-runs/awsabc",
			wantBase:    "https://api.nuon.co",
			wantPhoneID: "awsabc",
		},
		{
			name:        "trailing slash tolerated",
			in:          "https://api.nuon.co/v1/stack-runs/awsabc/",
			wantBase:    "https://api.nuon.co",
			wantPhoneID: "awsabc",
		},
		{
			name:        "localhost with port",
			in:          "http://localhost:8081/v1/stack-runs/awsabc",
			wantBase:    "http://localhost:8081",
			wantPhoneID: "awsabc",
		},
		{
			name:        "query string stripped",
			in:          "https://api.nuon.co/v1/stack-runs/awsabc?x=1",
			wantBase:    "https://api.nuon.co",
			wantPhoneID: "awsabc",
		},
		{name: "empty", in: "", wantErr: true},
		{name: "no scheme", in: "api.nuon.co/v1/stack-runs/abc", wantErr: true},
		{name: "wrong path", in: "https://api.nuon.co/v1/installs/abc", wantErr: true},
		{
			name:        "api prefix tolerated",
			in:          "https://ja.tail2117d3.ts.net/api/v1/stack-runs/awsabc",
			wantBase:    "https://ja.tail2117d3.ts.net/api",
			wantPhoneID: "awsabc",
		},
		{name: "missing phone_home_id", in: "https://api.nuon.co/v1/stack-runs/", wantErr: true},
		{name: "kind suffix rejected", in: "https://api.nuon.co/v1/stack-runs/abc/kind/provision", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			base, phoneID, err := parseURL(c.in)
			if c.wantErr {
				if err == nil {
					t.Fatalf("want error, got base=%q phone=%q", base, phoneID)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if base != c.wantBase {
				t.Errorf("base: got %q want %q", base, c.wantBase)
			}
			if phoneID != c.wantPhoneID {
				t.Errorf("phone: got %q want %q", phoneID, c.wantPhoneID)
			}
		})
	}
}
