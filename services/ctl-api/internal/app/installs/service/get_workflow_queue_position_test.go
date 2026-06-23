package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const workflowSignalOwnerType = "install_workflows"

func (s *InstallsServiceTestSuite) createQueue(installID string) *app.Queue {
	q := &app.Queue{
		ID:          domains.NewQueueID(),
		CreatedByID: s.testAcc.ID,
		OrgID:       &s.testOrg.ID,
		OwnerID:     installID,
		OwnerType:   "installs",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(q).Error)
	return q
}

func (s *InstallsServiceTestSuite) createQueueSignal(queueID, ownerID string, status app.Status, createdAt time.Time) *app.QueueSignal {
	sig := &app.QueueSignal{
		CreatedByID: s.testAcc.ID,
		OrgID:       &s.testOrg.ID,
		QueueID:     queueID,
		OwnerID:     ownerID,
		OwnerType:   workflowSignalOwnerType,
		Status:      app.NewCompositeStatus(s.ctx, status),
		CreatedAt:   createdAt,
		Workflow:    signaldb.WorkflowRef{IDTemplate: "wf-%s"},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(sig).Error)
	return sig
}

func (s *InstallsServiceTestSuite) getQueuePosition(workflowID string) WorkflowQueuePositionResponse {
	path := fmt.Sprintf("/v1/workflows/%s/queue-position", workflowID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var result WorkflowQueuePositionResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &result))
	return result
}

func (s *InstallsServiceTestSuite) TestGetWorkflowQueuePositionWithSignalsAhead() {
	install := s.createTestInstall()
	queue := s.createQueue(install.ID)
	base := time.Now().UTC().Add(-time.Hour)

	wfA := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	wfB := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	wfDone := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	wfTarget := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	s.createQueueSignal(queue.ID, wfA.ID, app.StatusQueued, base)
	// A parked (pending) signal still occupies the queue and must be counted.
	s.createQueueSignal(queue.ID, wfB.ID, app.StatusPending, base.Add(time.Minute))
	// A completed signal is no longer in the queue and must be excluded.
	s.createQueueSignal(queue.ID, wfDone.ID, app.StatusSuccess, base.Add(2*time.Minute))
	s.createQueueSignal(queue.ID, wfTarget.ID, app.StatusQueued, base.Add(3*time.Minute))

	result := s.getQueuePosition(wfTarget.ID)

	assert.Equal(s.T(), 3, result.Position, "two non-terminal signals ahead + self")
	assert.Equal(s.T(), 3, result.QueueDepth, "three non-terminal signals in the queue")
	require.Len(s.T(), result.SignalsAhead, 2, "completed signal excluded")
	assert.Equal(s.T(), wfA.ID, result.SignalsAhead[0].WorkflowID, "ordered front to back")
	assert.Equal(s.T(), wfB.ID, result.SignalsAhead[1].WorkflowID)
	assert.Equal(s.T(), app.WorkflowTypeReprovision, result.SignalsAhead[0].WorkflowType)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowQueuePositionAtFront() {
	install := s.createTestInstall()
	queue := s.createQueue(install.ID)
	base := time.Now().UTC().Add(-time.Hour)

	wfDone := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	wfTarget := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	s.createQueueSignal(queue.ID, wfDone.ID, app.StatusSuccess, base)
	s.createQueueSignal(queue.ID, wfTarget.ID, app.StatusQueued, base.Add(time.Minute))

	result := s.getQueuePosition(wfTarget.ID)

	assert.Equal(s.T(), 1, result.Position)
	assert.Equal(s.T(), 1, result.QueueDepth)
	assert.Empty(s.T(), result.SignalsAhead)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowQueuePositionNoSignal() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	result := s.getQueuePosition(workflow.ID)

	assert.Equal(s.T(), 0, result.Position)
	assert.Equal(s.T(), 0, result.QueueDepth)
	assert.Empty(s.T(), result.SignalsAhead)
}
