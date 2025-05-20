package temporal

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/metrics"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	// "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap/zaptest"
)

type TestSignal struct {
	name string
	eventloop.BaseSignal
}

func (s *TestSignal) Name() string {
	return s.name
}

func (s *TestSignal) SignalType() eventloop.SignalType {
	return "test-signal-type"
}

func (s *TestSignal) Namespace() string {
	return "default"
}

func (s *TestSignal) Stop() bool {
	return false
}

func (s *TestSignal) Restart() bool {
	return false
}

func (s *TestSignal) Start() bool {
	return false
}

func (s *TestSignal) Validate(v *validator.Validate) error {
	return nil
}

type TestELSignal struct {
	*TestSignal
	eventloop.EventLoopRequest
}

func newTestSignal(id string) *TestSignal {
	return &TestSignal{
		name:       id,
		BaseSignal: eventloop.BaseSignal{},
	}
}

func newTestELSignal(id string) *TestELSignal {
	return &TestELSignal{
		TestSignal:       newTestSignal(id),
		EventLoopRequest: eventloop.EventLoopRequest{},
	}
}

type SendTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

type testEnv struct {
	env *testsuite.TestWorkflowEnvironment
	sar *SendAndReceive
}

// TODO(sdboyer) errors.Is() is not currently working across de/serialization, so we have a weird string for Contains-checking
var testErr = errors.New("sentinel test error BEEP BOOP BEEP")

// func registerNS(opts client.Options, ns string) error {
// 	namespaceClient, err := client.NewNamespaceClient(opts)
// 	if err != nil {
// 		return errors.Wrap(err, "Failed to create namespace client")
// 	}

// 	err = namespaceClient.Register(context.Background(), &workflowservice.RegisterNamespaceRequest{
// 		Namespace: ns,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	namespaceClient.Close()
// 	return nil
// }

// Test sending over a matrix of possibilities:
// - same vs. cross-ns
// - SendAsync vs SendAndWait
// - notify with (nil, nil), (val, nil), and (nil, err)
// - request signals pattern vs. base pattern
// func (s *SendTestSuite) Test_Send() {
func TestSend(t *testing.T) {
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

	// if err = registerNS(*clientOpts, "other"); err != nil {
	// 	t.Fatalf("failed to register namespace: %v", err)
	// }

	worker := worker.New(srv.Client(), "default", worker.Options{})

	ctx := context.Background()
	mockCtl, _ := gomock.WithContext(ctx, t)
	mw := metrics.NewMockWriter(mockCtl)
	mw.EXPECT().Incr(gomock.Any(), gomock.Any()).AnyTimes()

	sar := &SendAndReceive{
		t: t,
		evClient: New(Params{
			L:  zaptest.NewLogger(t),
			MW: mw,
			V:  validator.New(),
		}),
	}

	worker.RegisterWorkflow(sar.Root)
	worker.RegisterWorkflow(sar.SendAsyncWorkflow)
	worker.RegisterWorkflow(sar.SendAndWaitWorkflow)
	worker.RegisterWorkflow(sar.ReceiveWorkflow)
	err = worker.Start()
	if err != nil {
		t.Fatalf("failed to start worker: %v", err)
	}
	t.Cleanup(worker.Stop)

	// Assemble groups of disjoint tests
	type caseopts struct {
		name   string
		optsfn func(SRTestOptions) SRTestOptions
	}
	type testgroup struct {
		groupname string
		cases     []caseopts
	}
	testGroups := []testgroup{
		{
			groupname: "method",
			cases: []caseopts{
				{
					name: "async",
					optsfn: func(opts SRTestOptions) SRTestOptions {
						opts.Await = false
						return opts
					},
				},
				{
					name: "await",
					optsfn: func(opts SRTestOptions) SRTestOptions {
						opts.Await = true
						return opts
					},
				},
			},
		},
		// {
		// 	groupname: "namespace",
		// 	cases: []caseopts{
		// 		{
		// 			name: "same",
		// 			optsfn: func(opts SRTestOptions) SRTestOptions {
		// 				opts.SenderNS = "default"
		// 				opts.ReceiverNS = "default"
		// 				return opts
		// 			},
		// 		},
		// 		{
		// 			name: "cross",
		// 			optsfn: func(opts SRTestOptions) SRTestOptions {
		// 				opts.SenderNS = "default"
		// 				opts.ReceiverNS = "other"
		// 				return opts
		// 			},
		// 		},
		// 	},
		// },
		{
			groupname: "notify",
			cases: []caseopts{
				{
					name: "nil-nil",
					optsfn: func(opts SRTestOptions) SRTestOptions {
						opts.Response = eventloop.SignalDoneMessage{}
						return opts
					},
				},
				{
					name: "val-nil",
					optsfn: func(opts SRTestOptions) SRTestOptions {
						opts.Response = eventloop.SignalDoneMessage{
							Result: "success",
						}
						return opts
					},
				},
				{
					name: "nil-err",
					optsfn: func(opts SRTestOptions) SRTestOptions {
						opts.Response = eventloop.SignalDoneMessage{
							Error: testErr,
						}
						return opts
					},
				},
			},
		},
		{
			groupname: "method",
			cases: []caseopts{
				{
					name: "request", optsfn: func(opts SRTestOptions) SRTestOptions {
						opts.UseRequestSignal = true
						return opts
					},
				},
				{
					name: "base",
					optsfn: func(opts SRTestOptions) SRTestOptions {
						opts.UseRequestSignal = false
						return opts
					},
				},
			},
		},
	}

	// recursive func, walks the group list to assemble it into a matrix
	var runMatrix func(*testing.T, int, SRTestOptions)
	runMatrix = func(t *testing.T, groupidx int, opts SRTestOptions) {
		if groupidx >= len(testGroups) {
			// When we reach the end of the groups, we've hydrated a full matrix of options and can run the test
			t.Logf("%+v\n", opts)
			opts.ID = t.Name()

			run, err := srv.Client().ExecuteWorkflow(context.Background(),
				client.StartWorkflowOptions{
					ID:                       "root",
					TaskQueue:                "default",
					WorkflowExecutionTimeout: 3 * time.Second,
				},
				sar.Root,
				opts)
			if err != nil {
				t.Fatalf("failed to start workflow: %v", err)
			}

			var v any
			err = run.Get(ctx, v)
			if opts.Response.Error != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				} else if !strings.Contains(err.Error(), testErr.Error()) {
					t.Fatal("errors not same")
				}
			} else {
				assert.NoError(t, run.Get(context.Background(), nil))
			}
			return
		}

		for _, tcase := range testGroups[groupidx].cases {
			t.Run(fmt.Sprintf("%s-%s", testGroups[groupidx].groupname, tcase.name), func(t *testing.T) {
				runMatrix(t, groupidx+1, tcase.optsfn(opts))
			})
		}
	}

	// All other options besides ID are set by the opts fns
	runMatrix(t, 0, SRTestOptions{
		ID: "default-id",
	})
}

type SendAndReceive struct {
	evClient Client
	t        *testing.T
}

type SRTestOptions struct {
	ID                   string
	SenderNS, ReceiverNS string
	UseRequestSignal     bool
	Await                bool
	Response             eventloop.SignalDoneMessage
}

func (o SRTestOptions) getSignal() eventloop.Signal {
	if o.UseRequestSignal {
		return newTestELSignal(o.ID)
	}
	return newTestSignal(o.ID)
}

// Need a root workflow that starts the others b/c otherwise the test harness can't handle
// signal passing between them
func (w *SendAndReceive) Root(ctx workflow.Context, opts SRTestOptions) error {
	var sfut, rfut workflow.Future
	rfut = workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: "event-loop-receive.workflow",
		Namespace:  opts.ReceiverNS,
	}), w.ReceiveWorkflow, opts)

	if opts.Await {
		sfut = workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: "send.workflow",
			Namespace:  opts.SenderNS,
		}), w.SendAndWaitWorkflow, opts)
	} else {
		sfut = workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: "send.workflow",
			Namespace:  opts.SenderNS,
		}), w.SendAsyncWorkflow, opts)
	}

	if err := sfut.Get(ctx, nil); err != nil {
		return errors.Wrap(err, "error from sending workflow")
	}
	if err := rfut.Get(ctx, nil); err != nil {
		return errors.Wrap(err, "error from receiving workflow")
	}
	return nil
}

func (w *SendAndReceive) SendAsyncWorkflow(ctx workflow.Context, opts SRTestOptions) error {
	defer func() { w.t.Log("SEND WF: returning from send") }()
	w.t.Log("SEND WF: sending")
	fut, err := w.evClient.SendAsync(ctx, "receive.workflow", opts.getSignal())
	w.t.Log("SEND WF: got future")
	if err != nil {
		return err
	}

	if fut == nil {
		return errors.New("future was nil")
	}
	w.t.Log("SEND WF: blocking on future")
	err = fut.Get(ctx, nil)
	return err
}

func (w *SendAndReceive) SendAndWaitWorkflow(ctx workflow.Context, opts SRTestOptions) error {
	defer func() { w.t.Log("SEND WF: returning from send") }()
	err := w.evClient.SendAndWait(ctx, "receive.workflow", opts.getSignal())
	w.t.Log("SEND WF: wait finished")
	return err
}

// This impl imitates the receiver-side logic that's in the eventloop/loop
// generic implementation. There is/should be another test in that package
// which directly exercises the logic. This separate test impl exists so that
// the SUT is solely the SendAsync implementation.
func (w *SendAndReceive) ReceiveWorkflow(ctx workflow.Context, opts SRTestOptions) error {
	defer func() { w.t.Log("RECEIVE WF: returning from receive") }()
	schan := workflow.GetSignalChannel(ctx, "receive.workflow")

	// In the real event loop, this is handled with a generic type
	var signal eventloop.Signal
	if opts.UseRequestSignal {
		signal = new(TestELSignal)
	} else {
		signal = new(TestSignal)
	}

	more := schan.Receive(ctx, signal)
	if !more {
		return errors.New("signal channel was closed")
	}
	w.t.Log("RECEIVE WF: got signal")
	if signal == nil {
		return errors.New("signal was nil")
	}

	// res := eventloop.SignalDoneMessage{
	// 	Result: opts.Response.Result,
	// }
	// if opts.Response.Error != "" {
	// 	res.Error = errors.New(opts.Response.Error)
	// }

	var listenErrs []error
	for _, listener := range signal.Listeners() {
		lctx := workflow.WithWorkflowNamespace(ctx, listener.Namespace)
		w.t.Logf("RECEIVE WF: sending response to %+v\n", listener)
		lerr := workflow.SignalExternalWorkflow(
			lctx,
			listener.WorkflowID,
			"",
			listener.SignalName,
			opts.Response,
		).Get(lctx, nil)
		if lerr != nil {
			listenErrs = append(listenErrs, lerr)
		}
	}

	if len(listenErrs) > 0 {
		return errors.Wrap(errors.Join(listenErrs...), "error notifying signal listeners: %v")
	}
	return nil
}
