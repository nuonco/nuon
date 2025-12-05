package loop

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/tctest"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

var devenv *tctest.DevTestEnv

func TestLoopConcurrency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ns := strings.Replace(uuid.New().String(), "-", "", -1)
	r, err := devenv.NewRunInNamespace(t, ctx, ns)
	if err != nil {
		t.Fatal(err)
	}

	tstate := &testState{
		log:      &mlog{log: make([]string, 0)},
		mockCtrl: gomock.NewController(t),
		done:     make(chan any),
	}
	states[t.Name()] = tstate

	r.Worker.RegisterWorkflow(EventLoopTest)
	r.Worker.RegisterWorkflow(UtiliHandler)

	req := eventloop.EventLoopRequest{
		ID: t.Name(),
	}
	_, err = r.Client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       t.Name(),
		TaskQueue:                "default",
		WorkflowExecutionTimeout: 3 * time.Second,
	}, EventLoopTest, req, nil)
	if err != nil {
		t.Fatal("error starting workflow:", err)
	}

	sigs := []*UtiliSignal{
		{
			SigName: "one",
		},
		{
			SigName: "two",
			WaitFor: 100 * time.Millisecond,
		},
		{
			SigName:   "conc",
			ConcGroup: "other",
		},
		{
			SigName: "done",
			Done:    true,
		},
	}

	for _, sig := range sigs {
		sig.TestName = t.Name()
		sig.NS = ns
		err := r.Client.SignalWorkflow(ctx, t.Name(), "", t.Name(), sig)
		if err != nil {
			t.Fatalf("error signaling workflow: %v", err)
		}
	}

	expect := (&semiorderedLog{}).
		Add(sigs[0].done()).
		Add(sigs[2].done(), sigs[1].done()).
		Add("workflow completed successfully")

	<-tstate.done
	expect.CompareLog(t, tstate.log.log)
}

func TestContinueAsNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ns := strings.Replace(uuid.New().String(), "-", "", -1)
	r, err := devenv.NewRunInNamespace(t, ctx, ns)
	if err != nil {
		t.Fatal(err)
	}

	tstate := &testState{
		log:      &mlog{log: make([]string, 0)},
		mockCtrl: gomock.NewController(t),
		done:     make(chan any),
	}
	states[t.Name()] = tstate

	r.Worker.RegisterWorkflow(EventLoopTest)
	r.Worker.RegisterWorkflow(UtiliHandler)

	req := eventloop.EventLoopRequest{
		ID: t.Name(),
	}
	_, err = r.Client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                       t.Name(),
		TaskQueue:                "default",
		WorkflowExecutionTimeout: 3 * time.Second,
	}, EventLoopTest, req, nil)
	if err != nil {
		t.Fatal("error starting workflow:", err)
	}

	expect := (&semiorderedLog{})

	var i int
	for i = range maxSignals + 4 {
		sig := &UtiliSignal{
			SigName:  fmt.Sprintf("sig-%d", i),
			TestName: t.Name(),
		}

		// group the max signal and its predecessor together with
		// the continue as new event, as certain execution orderings
		// can cause the restart to happen before final signals are processed
		if i != 0 && (i+1)%maxSignals == 0 {
			expect.Add(sig.done(), "workflow continued as new")
		} else {
			expect.Add(sig.done())
		}

		err := r.Client.SignalWorkflow(ctx, t.Name(), "", t.Name(), sig)
		if err != nil {
			t.Fatalf("error signaling workflow: %v", err)
		}
	}

	time.Sleep(time.Second)
	// time.Sleep(500 * time.Millisecond)

	// Do 2+2*maxSignals more, each in their own queue. This will force a ContinueAsNew, but only one,
	// because pending signals are not counted against the maxSignals limit.
	i++
	var logs []string
	for init := i; i < init+(maxSignals*2+2); i++ {
		sig := &UtiliSignal{
			SigName:   fmt.Sprintf("sig-%d", i),
			ConcGroup: fmt.Sprintf("sig-%d", i),
			TestName:  t.Name(),
		}
		logs = append(logs, sig.done())
		err := r.Client.SignalWorkflow(ctx, t.Name(), "", t.Name(), sig)
		if err != nil {
			t.Fatalf("error signaling workflow: %v", err)
		}
	}
	expect.Add(append(logs, "workflow continued as new")...)

	err = r.Client.SignalWorkflow(ctx, t.Name(), "", t.Name(), &UtiliSignal{
		SigName:  "done",
		TestName: t.Name(),
		Done:     true,
	})
	if err != nil {
		t.Fatalf("error signaling workflow: %v", err)
	}
	expect.Add("workflow completed successfully")

	<-tstate.done
	expect.CompareLog(t, tstate.log.log)
}

const OperationUtilisignal eventloop.SignalType = "utilisignal"

type UtiliSignal struct {
	SigName, NS string
	TestName    string
	ConcGroup   string
	Done        bool

	WaitFor time.Duration
	eventloop.BaseSignal
}

func (s *UtiliSignal) Name() string {
	return s.SigName
}

func (s *UtiliSignal) SignalType() eventloop.SignalType {
	return OperationUtilisignal
}

func (s *UtiliSignal) Namespace() string {
	return s.NS
}

func (s *UtiliSignal) ConcurrencyGroup() string {
	return s.ConcGroup
}

func (s *UtiliSignal) Stop() bool {
	return s.Done
}

func (s *UtiliSignal) Restart() bool {
	return false
}

func (s *UtiliSignal) Start() bool {
	return false
}

func (s *UtiliSignal) Validate(v *validator.Validate) error {
	return nil
}

func (s *UtiliSignal) GetWorkflowContext(ctx workflow.Context) workflow.Context {
	return ctx
}

func (s *UtiliSignal) done() string {
	return fmt.Sprintf("sig %s finished", s.SigName)
}

// 1 - test that default behavior is serial
// 2 - test that separate groups run concurrently
// 3 - test internal state of conc groups obj (?)
// 4 - test pending sigs are retained in order over restart

func UtiliHandler(ctx workflow.Context, sig *UtiliSignal) error {
	if sig.Done || workflow.IsReplaying(ctx) {
		return nil
	}
	tstate := states[sig.TestName]
	if sig.WaitFor > 0 {
		workflow.Sleep(ctx, sig.WaitFor)
	}
	tstate.log.Append(sig.done())
	return nil
}

type semiorderedLog struct {
	solo [][]string
	idx  []int
}

func (s *semiorderedLog) Add(elems ...string) *semiorderedLog {
	if len(elems) == 0 {
		panic("provide at least one")
	}

	var prior int
	if len(s.idx) != 0 {
		prior = s.idx[len(s.idx)-1]
	}

	sort.StringSlice(elems).Sort()
	s.solo = append(s.solo, elems)
	s.idx = append(s.idx, len(elems)+prior)
	return s
}

func (s *semiorderedLog) covers(ord, unord []string) ([]string, []string) {
	var extra []string
	unord = slices.Clone(unord)
	for _, elem := range ord {
		i, has := slices.BinarySearch(unord, elem)
		if !has {
			extra = append(extra, elem)
		} else {
			unord = slices.Delete(unord, i, i+1)
		}
	}
	return extra, unord
}

func (s *semiorderedLog) CompareLog(t *testing.T, log []string) {
	t.Helper()

	// Build up a comparison list in case we need it for either failure mode
	buf := new(bytes.Buffer)
	tw := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tw, "IDX\tWANT\tIDX\tGOT")

	flat := make([]string, 0, s.idx[len(s.idx)-1])
	for _, elems := range s.solo {
		flat = append(flat, elems...)
	}

	for i, elem := range flat {
		var widx string
		if _, has := slices.BinarySearch(s.idx, i); has || i == 0 {
			widx = strconv.Itoa(i)
		}

		if i < len(log) {
			fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", widx, elem, i, log[i])
		} else {
			fmt.Fprintf(tw, "%s\t%s\t\t\n", widx, elem)
		}
	}
	if len(log) > len(flat) {
		for i := len(flat); i < len(log); i++ {
			fmt.Fprintf(tw, "\t\t%d\t%s\n", i, log[i])
		}
	}
	tw.Flush()

	t.Logf("Expected vs. actual event log:\n%s", buf.String())

	if len(log) != s.idx[len(s.idx)-1] {
		t.Fatalf("expected log had %d elements, but got %d", s.idx[len(s.idx)-1], len(log))
	}

	for i, idx := range s.idx {
		var bottom int
		if i != 0 {
			bottom = s.idx[i-1]
		}
		extra, unmatched := s.covers(log[bottom:idx], s.solo[i])
		if len(unmatched) > 0 {
			t.Fatalf("mismatch on indices %d:%d: expected values %q were unmatched, saw %q instead", bottom, idx, unmatched, extra)
		}
	}
}

type testState struct {
	log      *mlog
	mockCtrl *gomock.Controller
	done     chan any
}

var states = make(map[string]*testState) // key is test name

type mlog struct {
	mut sync.Mutex // 😬
	log []string
}

func (l *mlog) Append(elem string) {
	l.mut.Lock()
	l.log = append(l.log, elem)
	l.mut.Unlock()
}

func TestMain(m *testing.M) {
	var err error
	devenv, err = tctest.NewEnv(context.Background())
	if err != nil {
		fmt.Printf("Failed to create test environment: %v\n", err)
		os.Exit(1)
	}

	m.Run()

	devenv.Server.Client().Close()
	devenv.Server.Stop()
}

func EventLoopTest(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*UtiliSignal) error {
	tstate := states[req.ID]
	mw := metrics.NewMockWriter(tstate.mockCtrl)
	mw.EXPECT().Incr(gomock.Any(), gomock.Any()).AnyTimes()
	mw.EXPECT().Gauge(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mw.EXPECT().Timing(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	validator := validator.New()
	tmw, err := tmetrics.New(validator, tmetrics.WithMetricsWriter(mw))
	if err != nil {
		return err
	}

	el := Loop[*UtiliSignal, *UtiliSignal]{
		Cfg: &internal.Config{
			Version: "test",
		},
		MW: tmw,
		V:  validator,
		Handlers: map[eventloop.SignalType]func(workflow.Context, *UtiliSignal) error{
			OperationUtilisignal: UtiliHandler,
		},
		NewRequestSignal: func(er eventloop.EventLoopRequest, sig *UtiliSignal) *UtiliSignal {
			newsig := new(UtiliSignal)
			*newsig = *sig
			return newsig
		},
	}

	var replaying bool
	// Temporal is usually, but not always, triggering replays after CaN for
	// some unknown reason. It's a no-op within the loop execution, but adds an
	// extra entry to the log. To avoid test flakines, if this is a replay, don't
	// write to the log.
	if !workflow.IsReplaying(ctx) {
		replaying = true
	}
	err = el.RunWithConcurrency(ctx, req, pendingSignals)
	if err == nil {
		tstate.log.Append("workflow completed successfully")
		close(states[req.ID].done)
	} else if workflow.IsContinueAsNewError(err) {
		if !replaying {
			tstate.log.Append("workflow continued as new")
		}
	} else {
		tstate.log.Append("workflow errored out")
		close(states[req.ID].done)
	}
	return err
}
