package config

import "strings"

func (c *Config) envFromAPIURL(url string) string {
	if strings.Contains(url, "api.nuon.co") {
		return "production"
	}

	if strings.Contains(url, "stage.api.nuon.co") {
		return "stage"
	}

	return "dev"
}
