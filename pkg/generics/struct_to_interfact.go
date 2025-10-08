package generics


import (
	"encoding/json"
	"fmt"
	"strings"
)

func StructToMap(obj any) (map[string]string, error) {
	// ONLY use this IFF you know all of the values are strings or can be converted to strings
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
			// Convert array to comma-separated string
			var strSlice []string
			for _, item := range v {
				strSlice = append(strSlice, fmt.Sprintf("%v", item))
			}
			result[key] = strings.Join(strSlice, ", ")
		default:
			// Convert other types to string
			result[key] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}
