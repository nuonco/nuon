package cmd

import "testing"

func TestFetchTokenDenied(t *testing.T) {
	cases := []struct {
		name string
		pid  int
		cmd  string
		want bool
	}{
		{name: "action step inside the runner container", pid: 20, cmd: "/bin/runner\x00install\x00", want: true},
		{name: "fetch-token is pid 1", pid: 1, cmd: "/bin/runner\x00mng\x00fetch-token\x00", want: false},
		{name: "host systemd", pid: 400, cmd: "/sbin/init\x00", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := fetchTokenDenied(tc.pid, tc.cmd); got != tc.want {
				t.Fatalf("fetchTokenDenied(%d, %q) = %v, want %v", tc.pid, tc.cmd, got, tc.want)
			}
		})
	}
}
