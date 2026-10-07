package nuon

func optionalRequestID(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}
