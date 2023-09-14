package generics

func ToIntSlice[T any](vals []T) []interface{} {
	intVals := make([]interface{}, len(vals))
	for idx, val := range vals {
		intVals[idx] = val
	}

	return intVals
}

// SliceToGroups turns a slice into a set of groups, with the group size. If the group size is less than 1, we set it to
// one.
func SliceToGroups[T any](slice []T, limit int) [][]T {
	if limit < 1 {
		limit = 1
	}

	matrix := [][]T{{}}
	row := 0
	for _, val := range slice {
		matrix[row] = append(matrix[row], val)
		if len(matrix[row]) == limit {
			row++
			matrix = append(matrix, []T{})
		}
	}

	return matrix
}
