package generics

import "fmt"

func MergeMap[K comparable, T any](src map[K]T, vals ...map[K]T) map[K]T {
	for _, val := range vals {
		for k, v := range val {
			src[k] = v
		}
	}

	return src
}

func ToStringMap(inp map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range inp {
		vStr, ok := v.(string)
		if ok {
			out[k] = vStr
			continue
		}

		out[k] = fmt.Sprintf("%v", v)
	}

	return out
}
