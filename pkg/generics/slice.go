package generics

import "fmt"

func SliceContains[T comparable](val T, vals []T) bool {
	for _, v := range vals {
		if val == v {
			return true
		}
	}

	return false
}

func ToStringSlice(vals []interface{}) []string {
	strVals := make([]string, len(vals))
	for idx, val := range vals {
		strVal, ok := val.(string)
		if !ok {
			strVal = fmt.Sprintf("%s", val)
		}

		strVals[idx] = strVal
	}

	return strVals
}

func ToIntSlice[T any](vals []T) []interface{} {
	intVals := make([]interface{}, len(vals))
	for idx, val := range vals {
		intVals[idx] = val
	}

	return intVals
}

func MergeSlice[T comparable](a, b []T) []T {
	vals := make([]T, 0, len(a)+len(b))
	vals = append(vals, a...)
	vals = append(vals, b...)

	return vals
}

// SliceToGroups turns a slice into a set of groups, with the group size. If the group size is less than 1, we set it to
// one.
func SliceToGroups[T any](vals []T, grpSize int) [][]T {
	if grpSize < 1 {
		grpSize = len(vals)
	}

	var grps [][]T
	for i := 0; i < len(vals); i += grpSize {
		end := i + grpSize

		if end > len(vals) {
			end = len(vals)
		}

		grps = append(grps, vals[i:end])
	}
	return grps
}
