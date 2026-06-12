package useragent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCLI(t *testing.T) {
	cliAgents := []string{
		"nuon-cli/1.2.3",
		"Nuon/0.9.0 (darwin; arm64)",
		"Go-http-client/2.0",
		"curl/8.4.0",
		"Wget/1.21",
		"PostmanRuntime/7.36.0",
	}
	for _, ua := range cliAgents {
		assert.True(t, IsCLI(ua), ua)
	}

	browserAgents := []string{
		"",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0",
	}
	for _, ua := range browserAgents {
		assert.False(t, IsCLI(ua), ua)
	}
}
