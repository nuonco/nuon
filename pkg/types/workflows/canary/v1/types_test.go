package canaryv1

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	interval "google.golang.org/genproto/googleapis/type/interval"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_Step(t *testing.T) {
	t.Run("latency returns the duration between start and finish", func(t *testing.T) {
		ts := time.Now()
		obj := &Step{
			Status: Status_STATUS_OK,
			Name:   uuid.NewString(),
			Error:  nil,
			Interval: &interval.Interval{
				StartTime: timestamppb.New(ts),
				EndTime:   timestamppb.New(ts.Add(time.Second)),
			},
		}
		err := obj.Validate()
		assert.NoError(t, err)
		assert.Equal(t, time.Second, obj.Duration())
	})
}
