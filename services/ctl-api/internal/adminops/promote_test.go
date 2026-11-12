package adminops

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/promotion"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type promoteTestDeps struct {
	fx.In

	DB  *gorm.DB `name:"psql"`
	Ops *Ops
}

type PromoteTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps promoteTestDeps
	ctx  context.Context

	dispatches  atomic.Int64
	dispatchErr atomic.Pointer[error]
}

func TestPromoteTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(PromoteTestSuite))
}

func (s *PromoteTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	mockTC := temporal.NewMockClient(gomock.NewController(s.T()))
	mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(&mockWorkflowRun{}, nil).AnyTimes()
	mockTC.EXPECT().SignalWithStartWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).DoAndReturn(func(context.Context, string, string, string, any, tclient.StartWorkflowOptions, any, any) (tclient.WorkflowRun, error) {
		if err := s.dispatchErr.Load(); err != nil {
			return nil, *err
		}
		s.dispatches.Add(1)
		return &mockWorkflowRun{}, nil
	}).AnyTimes()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			Mocks:           &tests.TestMocks{MockTC: mockTC},
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.deps),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.deps.DB)
}

func (s *PromoteTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	ctx, err := s.deps.Ops.WithActor(context.Background())
	s.Require().NoError(err)
	s.ctx = ctx
	s.reset()
}

func (s *PromoteTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *PromoteTestSuite) reset() {
	tests.ClearQueueSignals(s.T(), s.deps.DB)
	s.dispatches.Store(0)
	s.dispatchErr.Store(nil)
}

func (s *PromoteTestSuite) setStatus(queueSignalID string, status app.Status) {
	res := s.deps.DB.WithContext(s.ctx).
		Model(&app.QueueSignal{}).
		Where(app.QueueSignal{ID: queueSignalID}).
		Update("status", app.NewCompositeStatus(s.ctx, status))
	s.Require().NoError(res.Error)
	s.Require().EqualValues(1, res.RowsAffected)
}

func (s *PromoteTestSuite) signals() []app.QueueSignal {
	return tests.GetQueueSignals(s.T(), s.deps.DB)
}

func (s *PromoteTestSuite) TestPromoteOnceCreatesAndDispatches() {
	res, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
	s.Require().NoError(err)
	s.False(res.Deduplicated)
	s.Equal("v1", res.Tag)

	signals := s.signals()
	s.Require().Len(signals, 1)
	s.Equal(res.QueueSignalID, signals[0].ID)
	s.True(signals[0].Enqueued)
	s.NotEmpty(signals[0].CreatedByID)

	sig, ok := signals[0].Signal.Signal.(*promotion.Signal)
	s.Require().True(ok)
	s.Equal("v1", sig.Tag)
	s.EqualValues(1, s.dispatches.Load())
}

func (s *PromoteTestSuite) TestPromoteOnceByLatestStatus() {
	testCases := map[string]struct {
		status   app.Status
		wantDupe bool
	}{
		"queued is reused":      {status: app.StatusQueued, wantDupe: true},
		"pending is reused":     {status: app.StatusPending, wantDupe: true},
		"in progress is reused": {status: app.StatusInProgress, wantDupe: true},
		"success is reused":     {status: app.StatusSuccess, wantDupe: true},
		"error is retried":      {status: app.StatusError, wantDupe: false},
		"cancelled is retried":  {status: app.StatusCancelled, wantDupe: false},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.reset()

			first, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
			s.Require().NoError(err)
			s.setStatus(first.QueueSignalID, tc.status)

			second, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
			s.Require().NoError(err)
			s.Equal(tc.wantDupe, second.Deduplicated)

			if tc.wantDupe {
				s.Equal(first.QueueSignalID, second.QueueSignalID)
				s.Len(s.signals(), 1)
				s.EqualValues(1, s.dispatches.Load())
				return
			}

			s.NotEqual(first.QueueSignalID, second.QueueSignalID)
			s.Len(s.signals(), 2)
			s.EqualValues(2, s.dispatches.Load())
		})
	}
}

func (s *PromoteTestSuite) TestPromoteOnceRedispatchesAfterDispatchFailure() {
	dispatchErr := errors.New("temporal unavailable")
	s.dispatchErr.Store(&dispatchErr)

	_, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
	s.Require().ErrorIs(err, dispatchErr)

	signals := s.signals()
	s.Require().Len(signals, 1)
	s.False(signals[0].Enqueued)

	s.dispatchErr.Store(nil)

	res, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
	s.Require().NoError(err)
	s.True(res.Deduplicated)
	s.Equal(signals[0].ID, res.QueueSignalID)

	signals = s.signals()
	s.Require().Len(signals, 1)
	s.True(signals[0].Enqueued)
	s.EqualValues(1, s.dispatches.Load())
}

func (s *PromoteTestSuite) TestPromoteOncePromotesAgainAfterRollback() {
	first, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
	s.Require().NoError(err)
	s.setStatus(first.QueueSignalID, app.StatusSuccess)

	second, err := s.deps.Ops.PromoteOnce(s.ctx, "v2")
	s.Require().NoError(err)
	s.False(second.Deduplicated)
	s.setStatus(second.QueueSignalID, app.StatusSuccess)

	rollback, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
	s.Require().NoError(err)
	s.False(rollback.Deduplicated)
	s.NotEqual(first.QueueSignalID, rollback.QueueSignalID)

	s.Len(s.signals(), 3)
	s.EqualValues(3, s.dispatches.Load())
}

func (s *PromoteTestSuite) TestPromoteOnceConcurrentRunsCreateOneSignal() {
	testCases := map[string]struct {
		failedPredecessor bool
		wantSignals       int
	}{
		"no previous promotion":    {failedPredecessor: false, wantSignals: 1},
		"after a failed promotion": {failedPredecessor: true, wantSignals: 2},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.reset()

			_, err := s.deps.Ops.generalHelpers.EnsureGeneralQueue(s.ctx)
			s.Require().NoError(err)

			if tc.failedPredecessor {
				failed, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
				s.Require().NoError(err)
				s.setStatus(failed.QueueSignalID, app.StatusError)
			}

			const runs = 8
			results := make([]*PromoteResult, runs)
			errs := make([]error, runs)

			var wg sync.WaitGroup
			for i := range runs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					results[i], errs[i] = s.deps.Ops.PromoteOnce(s.ctx, "v1")
				}()
			}
			wg.Wait()

			for i := range runs {
				s.Require().NoError(errs[i])
				s.Equal(results[0].QueueSignalID, results[i].QueueSignalID)
			}
			s.Len(s.signals(), tc.wantSignals)
		})
	}
}

func (s *PromoteTestSuite) TestPromoteAlwaysCreatesASignal() {
	first, err := s.deps.Ops.Promote(s.ctx, "v1")
	s.Require().NoError(err)
	s.setStatus(first.QueueSignalID, app.StatusSuccess)

	second, err := s.deps.Ops.Promote(s.ctx, "v1")
	s.Require().NoError(err)
	s.False(second.Deduplicated)
	s.NotEqual(first.QueueSignalID, second.QueueSignalID)
	s.setStatus(second.QueueSignalID, app.StatusSuccess)

	s.Len(s.signals(), 2)
	s.EqualValues(2, s.dispatches.Load())

	once, err := s.deps.Ops.PromoteOnce(s.ctx, "v1")
	s.Require().NoError(err)
	s.True(once.Deduplicated)
	s.Equal(second.QueueSignalID, once.QueueSignalID)
}

type mockWorkflowRun struct{}

func (m *mockWorkflowRun) GetID() string    { return "mock-workflow-id" }
func (m *mockWorkflowRun) GetRunID() string { return "mock-run-id" }
func (m *mockWorkflowRun) Get(context.Context, any) error {
	return nil
}

func (m *mockWorkflowRun) GetWithOptions(context.Context, any, tclient.WorkflowRunGetOptions) error {
	return nil
}
