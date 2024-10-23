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

func (c *Config) segmentWriteKey(env string) string {
	switch env {
	case "production":
		return "PzWHI08ttSjMYoJH21Z5GNqlURUqZED7"
	case "stage":
		return "L7i4QsG9TetBtV9queHK6ZICH6s0OCmA"
	case "dev":
		return "GHW8n0dIkdycQoRTzl2hQCvsRV8p8bGt"
	}

	return ""
}
