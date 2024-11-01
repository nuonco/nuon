package logv1

func (l *LogConfiguration) LogRecordAttrs() map[string]string {
	attrs := make(map[string]string, 0)
	for _, attr := range l.Attrs {
		attrs[attr.Key] = attr.Value
	}

	return attrs
}

func NewAttrs(vals map[string]string) []*LogRecordAttr {
	attrs := make([]*LogRecordAttr, 0)

	for k, v := range vals {
		attrs = append(attrs, &LogRecordAttr{
			Key:   k,
			Value: v,
		})
	}

	return attrs
}
