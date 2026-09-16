package githubevent

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestMatchesRunConfig(t *testing.T) {
	tests := []struct {
		name   string
		config app.AppBranchRunConfig
		event  fanOutRequest
		want   bool
	}{
		{name: "default push", event: fanOutRequest{EventType: "push"}, want: true},
		{name: "default PR", event: fanOutRequest{EventType: "pull_request"}, want: true},
		{name: "default tag", event: fanOutRequest{EventType: "tag", HeadRef: "foobar/v1"}},
		{name: "matching tag", config: app.AppBranchRunConfig{Mode: app.AppBranchRunModeTagPrefix, TagPrefix: "foobar/"}, event: fanOutRequest{EventType: "tag", HeadRef: "foobar/v1"}, want: true},
		{name: "other tag", config: app.AppBranchRunConfig{Mode: app.AppBranchRunModeTagPrefix, TagPrefix: "foobar/"}, event: fanOutRequest{EventType: "tag", HeadRef: "other/v1"}},
		{name: "tag mode push", config: app.AppBranchRunConfig{Mode: app.AppBranchRunModeTagPrefix, TagPrefix: "foobar/"}, event: fanOutRequest{EventType: "push"}},
		{name: "label mode push candidate", config: app.AppBranchRunConfig{Mode: app.AppBranchRunModeGithubLabel, GithubLabel: "daily"}, event: fanOutRequest{EventType: "push"}, want: true},
		{name: "label mode PR", config: app.AppBranchRunConfig{Mode: app.AppBranchRunModeGithubLabel, GithubLabel: "daily"}, event: fanOutRequest{EventType: "pull_request"}},
		{name: "manual mode push", config: app.AppBranchRunConfig{Mode: app.AppBranchRunModeManualOnly}, event: fanOutRequest{EventType: "push"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := matchesRunConfig(test.config, test.event); got != test.want {
				t.Fatalf("matchesRunConfig() = %v, want %v", got, test.want)
			}
		})
	}
}
