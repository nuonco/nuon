package generics

func SliceContains[T comparable](val T, vals []T) bool {
	for _, v := range vals {
		if val == v {
			return true
		}
	}

	return false
}

func ToIntSlice[T any](vals []T) []interface{} {
	intVals := make([]interface{}, len(vals))
	for idx, val := range vals {
		intVals[idx] = val
	}

	return intVals
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
