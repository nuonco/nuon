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

func (c *Config) sentryDSN(env string) string {
	switch env {
	case "production":
		return "https://54381d62bdf1c538b8f59f41feedc759@o4507623795523584.ingest.us.sentry.io/4508185736380416"
	case "stage":
		return "https://f7b3f6437277178bc9ee2520a1b91903@o4507623795523584.ingest.us.sentry.io/4508185738412032"
	case "dev":
		return "https://a1b830bf8c55b306e9f27e9937feca73@o4507623795523584.ingest.us.sentry.io/4508201572040704"
	}

	return ""
}
