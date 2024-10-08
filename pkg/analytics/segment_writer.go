package analytics

import (
	segment "github.com/segmentio/analytics-go/v3"
)

type SegmentWriter struct {
	client segment.Client
}

func NewSegmentWriter(writeKey string) *SegmentWriter {
	return &SegmentWriter{
		client: segment.New(writeKey),
	}
}

func (sc *SegmentWriter) Identify(userId, email string) {
	sc.client.Enqueue(segment.Identify{
		UserId: userId,
		Traits: segment.NewTraits().
			SetEmail(email),
	})

}

func (sc *SegmentWriter) Group(userId, groupId, name string) {
	sc.client.Enqueue(segment.Group{
		UserId:  userId,
		GroupId: groupId,
		Traits: map[string]interface{}{
			"name": name,
		},
	})

}

func (sc *SegmentWriter) Track(userId string, event Event, properties map[string]interface{}) {
	segmentProperties := segment.NewProperties()
	for k, v := range properties {
		segmentProperties.Set(k, v)
	}

	sc.client.Enqueue(segment.Track{
		UserId:     userId,
		Event:      string(event),
		Properties: segmentProperties,
	})

}
