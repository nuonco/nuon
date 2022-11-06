package build

import (
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-waypoint/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func Test_logEventWriter_Write(t *testing.T) {
	log := &mockLogger{}
	log.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	event := job.WaypointJobEvent{
		JobID: uuid.NewString(),
	}
	writer := newLogEventWriter(log)
	err := writer.Write(event)
	assert.Nil(t, err)

	log.AssertNumberOfCalls(t, "Info", 1)
}
