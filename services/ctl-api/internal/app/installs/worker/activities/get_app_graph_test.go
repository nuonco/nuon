package activities

import "testing"

func TestAppGraphConfigID(t *testing.T) {
	tests := []struct {
		name               string
		req                GetAppGraphRequest
		installAppConfigID string
		want               string
	}{
		{
			name:               "falls back to the install's pinned config",
			req:                GetAppGraphRequest{InstallID: "install-1"},
			installAppConfigID: "cfg-old",
			want:               "cfg-old",
		},
		{
			name:               "override orders against the config being rolled out",
			req:                GetAppGraphRequest{InstallID: "install-1", AppConfigID: "cfg-new"},
			installAppConfigID: "cfg-old",
			want:               "cfg-new",
		},
		{
			name: "override works on an install with no config yet",
			req:  GetAppGraphRequest{InstallID: "install-1", AppConfigID: "cfg-new"},
			want: "cfg-new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appGraphConfigID(tt.req, tt.installAppConfigID); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
