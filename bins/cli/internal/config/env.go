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
		return "YaAuCiuua7rbkJOuDf6wRoTuDuM4BjFY"
	case "stage":
		return "2zX2oZQCkdHaPkoQW79fOtW20pzH7DQ2"
	case "dev":
		return "GHW8n0dIkdycQoRTzl2hQCvsRV8p8bGt"
	}

	return ""
}
