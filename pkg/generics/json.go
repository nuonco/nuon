package generics

import "encoding/json"

func IsJSONStr(str string) bool {
	return IsJSON([]byte(str))
}

func IsJSON(s []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(s, &js) == nil
}
