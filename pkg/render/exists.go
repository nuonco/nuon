package render

import (
	"github.com/Jeffail/gabs"
)

func Exists(path string, data map[string]interface{}) bool {
	json, _ := gabs.Consume(data)
	return json.ExistsP(path)
}
