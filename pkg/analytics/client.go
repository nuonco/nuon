package analytics

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client.go -source=client.go -package=analytics
type Client interface {
	Identify(string, string, string)
	Group(string, string, string)
	Track(string, Event, map[string]interface{})
}

func New(writeKey string) (Client, error) {
	return NewSegmentClient(writeKey)
}
