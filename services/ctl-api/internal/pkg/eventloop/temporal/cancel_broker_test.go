package temporal

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/metrics"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap/zaptest"
)

func TestCancelBroker(t *testing.T) {
	clientOpts := new(client.Options)
	var err error
	clientOpts.HostPort, err = getFreeHostPort()
	if err != nil {
		t.Fatalf("failed to get free host port: %v", err)
	}

	srv, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		ClientOptions: clientOpts,
	})
	if err != nil {
		t.Fatalf("failed to set up dev server: %v", err)
	}

	t.Cleanup(func() {
		srv.Client().Close()
		srv.Stop()
	})

	if err = registerNS(*clientOpts, "other"); err != nil {
		t.Fatalf("failed to register namespace: %v", err)
	}

	ctx := context.Background()
	mockCtl, _ := gomock.WithContext(ctx, t)
	mw := metrics.NewMockWriter(mockCtl)
	mw.EXPECT().Incr(gomock.Any(), gomock.Any()).AnyTimes()

	wf := &CancelFlows{
		t: t,
		evClient: New(Params{
			L:  zaptest.NewLogger(t),
			MW: mw,
			V:  validator.New(),
		}),
	}

	wrk := worker.New(srv.Client(), "default", worker.Options{})
	wrk.RegisterWorkflow(wf.Root)
	wrk.RegisterWorkflow(wf.ReceiveWorkflow)
	err = wrk.Start()
	if err != nil {
		t.Fatalf("failed to start worker: %v", err)
	}

	run, err := srv.Client().ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "cancel-broker-test",
		TaskQueue: "default",
	}, wf.Root, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = run.Get(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
}

type CancelFlows struct {
	t        *testing.T
	evClient Client
	cb       *CancelBroker[string]
}

func (w *CancelFlows) Root(ctx workflow.Context, _ string) error {
	// 1. test basic cancellation
	// 2. test pre-cancellation
	// 3. test multi-in-flight cancellation
	// 3a. test with different ids
	// 3b. test with same ids

	cctx, cancel := workflow.WithCancel(ctx)
	rfut := workflow.ExecuteChildWorkflow(workflow.WithChildOptions(cctx, workflow.ChildWorkflowOptions{
		WorkflowID: "receive-wf",
		Namespace:  "default",
	}), w.ReceiveWorkflow, "")
	workflow.Sleep(ctx, 10*time.Second)
	cancel()

	return rfut.Get(ctx, nil)

	// // NOTE(sdboyer) none of this works if we're in a replay. check on that if this test ever flakes
	// cctx, ccancel := workflow.WithCancel(ctx)
	// boop(ctx, "first")

	// // boot up one child workflow we'll use throughout the test
	// wg := workflow.NewWaitGroup(ctx)
	// wg.Add(1)
	// workflow.Go(ctx, func(ctx workflow.Context) {
	// 	fut, err := w.evClient.SendAsync(cctx, "receive.workflow", newTestSignal("doesnamatta", "default"))
	// 	if err != nil {
	// 		w.t.Fatal("failed to send async")
	// 		return
	// 	}
	// 	wg.Done()

	// })
	// wg.Wait()
	// ccancel()
	return nil
}

func boop(ctx workflow.Context, msg string) error {
	return workflow.SignalExternalWorkflow(
		ctx,
		"receive-wf",
		"",
		CancelChannelName,
		msg,
	).Get(ctx, nil)
}

func (w *CancelFlows) ReceiveWorkflow(ctx workflow.Context, _ string) error {
	w.cb = NewCancelBroker[string](ctx)
	workflow.Sleep(ctx, time.Hour)
	return nil
}
