package writer

import (
	"github.com/powertoolsdev/workers-executors/internal/executors/waypoint/event"
	"github.com/stretchr/testify/mock"
)

type mockEventWriter struct {
	mock.Mock
}

func (t *mockEventWriter) Write(ev event.WaypointJobEvent) error {
	args := t.Called(ev)
	return args.Error(0)
}

var _ EventWriter = (*mockEventWriter)(nil)
