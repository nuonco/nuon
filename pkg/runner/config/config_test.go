package runnerconfig

import "testing"

func TestNewConfigTrimsTrailingSlashFromRunnerAPIURL(t *testing.T) {
	t.Setenv("RUNNER_ID", "runner-test")
	t.Setenv("HOST_IP", "127.0.0.1")

	tests := map[string]struct {
		url  string
		want string
	}{
		"trailing slash":  {url: "https://runner.example.com/", want: "https://runner.example.com"},
		"several slashes": {url: "https://runner.example.com///", want: "https://runner.example.com"},
		"no slash":        {url: "https://runner.example.com", want: "https://runner.example.com"},
		"path prefix":     {url: "https://runner.example.com/api/", want: "https://runner.example.com/api"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("RUNNER_API_URL", test.url)

			cfg, err := NewConfig()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.RunnerAPIURL != test.want {
				t.Errorf("RunnerAPIURL = %q, want %q", cfg.RunnerAPIURL, test.want)
			}
		})
	}
}
