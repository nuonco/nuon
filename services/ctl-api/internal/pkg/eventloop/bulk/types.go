package bulk

import signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"

type EventLoop struct {
	ID        string
	Namespace string

	WorkflowRef *signaldb.WorkflowRef
}
