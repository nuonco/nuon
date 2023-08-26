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
func SliceToGroups[T any](vals []T, grpSize int) [][]T {
	if grpSize < 1 {
		grpSize = 1
	}

	grps := make([][]T, 0)
	grp := make([]T, 0, grpSize)
	for _, val := range vals {
		grp = append(grp, val)

		if len(grp) >= grpSize {
			grps = append(grps, grp)
			grp = make([]T, 0, grpSize)
		}
	}

	return grps
}
