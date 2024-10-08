package analytics

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_writer.go -source=writer.go -package=analytics
type Writer interface {
	Identify(string, string)
	Group(string, string, string)
	Track(string, Event, map[string]interface{})
}

func NewWriter(writeKey string) Writer {
	return NewSegmentWriter(writeKey)
}
