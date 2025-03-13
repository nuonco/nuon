package generics

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func FirstNonEmptyString(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}

	return ""
}

func DisplayName(val string) string {
	str := strings.ReplaceAll(string(val), "_", " ")

	caser := cases.Title(language.English)
	str = caser.String(str)
	return str
}

func HasAnyPrefix(val string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(val, prefix) {
			return true
		}
	}

	return false
}

func SystemName(val string) string {
	str := strings.ReplaceAll(string(val), " ", "-")
	str = strings.ReplaceAll(string(val), "_", "-")

	caser := cases.Lower(language.English)
	str = caser.String(str)
	return str
}

func StringOneOf(val string, vals ...string) bool {
	for _, v := range vals {
		if v == val {
			return true
		}
	}

	return false
}
