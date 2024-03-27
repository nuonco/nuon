package sandboxes

// NOTE: structpb does not support []string type, so we have to use interface{} here
func ToStringSlice(vals []interface{}) []string {
	strVals := make([]string, len(vals))
	for idx, val := range vals {
		v := val
		strVals[idx] = v.(string)
	}

	return strVals
}
