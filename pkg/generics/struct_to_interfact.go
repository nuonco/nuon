package generics

import (
	"encoding/json"
	"fmt"
	"strings"
)

func StructToMap(obj any) (map[string]string, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var intermediate map[string]any
	err = json.Unmarshal(jsonBytes, &intermediate)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for key, value := range intermediate {
		switch v := value.(type) {
		case string:
			result[key] = v
		case []any:
			var strSlice []string
			for _, item := range v {
				strSlice = append(strSlice, fmt.Sprintf("%v", item))
			}
			result[key] = strings.Join(strSlice, ", ")
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}
