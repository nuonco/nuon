package vcsv1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Directory(t *testing.T) {
	tests := map[string]struct {
		cfgFn    func() *Config
		expected string
	}{
		"public-git-config": {
			cfgFn: func() *Config {
				return &Config{
					Cfg: &Config_PublicGitConfig{
						PublicGitConfig: &PublicGitConfig{
							Directory: "public-git-dir",
						},
					},
				}
			},
			expected: "public-git-dir",
		},
		"connected-github-config": {
			cfgFn: func() *Config {
				return &Config{
					Cfg: &Config_ConnectedGithubConfig{
						ConnectedGithubConfig: &ConnectedGithubConfig{
							Directory: "connected-github-dir",
						},
					},
				}
			},
			expected: "connected-github-dir",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := test.cfgFn()
			dir := cfg.Directory()
			assert.Equal(t, test.expected, dir)
		})
	}
}

func TestConfig_Repo(t *testing.T) {
	tests := map[string]struct {
		cfgFn    func() *Config
		expected string
	}{
		"public-git-config": {
			cfgFn: func() *Config {
				return &Config{
					Cfg: &Config_PublicGitConfig{
						PublicGitConfig: &PublicGitConfig{
							Repo: "git@github.com:jonmorehouse/empty.git",
						},
					},
				}
			},
			expected: "git@github.com:jonmorehouse/empty.git",
		},
		"connected-github-config": {
			cfgFn: func() *Config {
				return &Config{
					Cfg: &Config_ConnectedGithubConfig{
						ConnectedGithubConfig: &ConnectedGithubConfig{
							Repo: "jonmorehouse/empty",
						},
					},
				}
			},
			expected: "jonmorehouse/empty",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := test.cfgFn()
			dir := cfg.Repo()
			assert.Equal(t, test.expected, dir)
		})
	}
}

func TestConfig_GitRef(t *testing.T) {
	tests := map[string]struct {
		cfgFn    func() *Config
		expected string
	}{
		"public-git-config": {
			cfgFn: func() *Config {
				return &Config{
					Cfg: &Config_PublicGitConfig{
						PublicGitConfig: &PublicGitConfig{
							GitRef: "HEAD",
						},
					},
				}
			},
			expected: "HEAD",
		},
		"connected-github-config": {
			cfgFn: func() *Config {
				return &Config{
					Cfg: &Config_ConnectedGithubConfig{
						ConnectedGithubConfig: &ConnectedGithubConfig{
							GitRef: "HEAD",
						},
					},
				}
			},
			expected: "HEAD",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := test.cfgFn()
			dir := cfg.GitRef()
			assert.Equal(t, test.expected, dir)
		})
	}
}
