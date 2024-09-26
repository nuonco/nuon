package analytics

import (
	segment "github.com/segmentio/analytics-go/v3"
)

type SegmentClient struct {
	client segment.Client
}

func NewSegmentClient(writeKey string) (*SegmentClient, error) {
	client := SegmentClient{
		client: segment.New(writeKey),
	}
	return &client, nil
}

func (sc *SegmentClient) Identify(userId, name, email string) {
	sc.client.Enqueue(segment.Identify{
		UserId: userId,
		Traits: segment.NewTraits().
			SetName(name).
			SetEmail(email),
	})

}

func (sc *SegmentClient) Group(userId, groupId, name string) {
	sc.client.Enqueue(segment.Group{
		UserId:  userId,
		GroupId: groupId,
		Traits: map[string]interface{}{
			"name": name,
		},
	})

}

func (sc *SegmentClient) Track(userId string, event Event, properties map[string]interface{}) {
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
