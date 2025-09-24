package generics

func AnyTrue(vals ...bool) bool {
	for _, v := range vals {
		if v {
			return true
		}
	}

	return false
}
