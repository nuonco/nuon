package render

import (
	"strings"

	"github.com/Jeffail/gabs"
)

func Exists(path string, data map[string]interface{}) bool {
	path = strings.TrimPrefix(path, ".")

	json, _ := gabs.Consume(data)
	return json.ExistsP(path)
}
