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
