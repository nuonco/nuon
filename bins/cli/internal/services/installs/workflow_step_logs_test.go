package installs

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type stepLogPage struct {
	records []*models.AppOtelLogRecord
	next    string
}

type stepLogsAPI struct {
	nuon.Client
	pages        []stepLogPage
	offsets      []string
	open         bool
	streamCalls  int
	cancel       context.CancelFunc
	cancelOnRead int
}

func (a *stepLogsAPI) LogStreamReadLogsWithNextOffset(_ context.Context, _, offset, _ string, _ *nuon.LogStreamLogFilters) ([]*models.AppOtelLogRecord, string, error) {
	a.offsets = append(a.offsets, offset)
	if a.cancel != nil && a.cancelOnRead > 0 && len(a.offsets) == a.cancelOnRead {
		a.cancel()
	}
	i := len(a.offsets) - 1
	if i >= len(a.pages) {
		return []*models.AppOtelLogRecord{}, "", nil
	}
	return a.pages[i].records, a.pages[i].next, nil
}

func (a *stepLogsAPI) GetLogStream(context.Context, string) (*models.AppLogStream, error) {
	a.streamCalls++
	return &models.AppLogStream{Open: a.open}, nil
}

func TestStreamStepLogsDrainsUntilNextIsEmpty(t *testing.T) {
	api := &stepLogsAPI{pages: twoLogPages()}
	svc := &Service{api: api}

	out := captureStdout(t, func() {
		err := svc.streamStepLogs(context.Background(), "ls_1", WorkflowStepLogsOptions{})
		require.NoError(t, err)
	})

	require.Equal(t, []string{"", "111"}, api.offsets)
	require.Zero(t, api.streamCalls)
	require.Contains(t, out, "first page")
	require.Contains(t, out, "second page")
}

func TestStreamStepLogsTailClosedStreamDrainsThenStops(t *testing.T) {
	api := &stepLogsAPI{pages: twoLogPages(), open: false}
	svc := &Service{api: api}

	out := captureStdout(t, func() {
		err := svc.streamStepLogs(context.Background(), "ls_1", WorkflowStepLogsOptions{Follow: true})
		require.NoError(t, err)
	})

	require.Equal(t, []string{"", "111"}, api.offsets)
	require.Equal(t, 1, api.streamCalls)
	require.Contains(t, out, "first page")
	require.Contains(t, out, "second page")
}

func TestStreamStepLogsTailOpenStreamFollowsOnce(t *testing.T) {
	orig := stepLogFollowWait
	stepLogFollowWait = 0
	t.Cleanup(func() { stepLogFollowWait = orig })

	ctx, cancel := context.WithCancel(context.Background())
	api := &stepLogsAPI{
		pages:        twoLogPages(),
		open:         true,
		cancel:       cancel,
		cancelOnRead: 3,
	}
	svc := &Service{api: api}

	out := captureStdout(t, func() {
		err := svc.streamStepLogs(ctx, "ls_1", WorkflowStepLogsOptions{Follow: true})
		require.ErrorIs(t, err, context.Canceled)
	})

	last, err := time.Parse(time.RFC3339Nano, "2026-10-21T21:00:00.123456789Z")
	require.NoError(t, err)
	require.Equal(t, []string{"", "111", unixNanoCursor(last.Format(time.RFC3339Nano))}, api.offsets)
	require.Contains(t, out, "first page")
	require.Contains(t, out, "second page")
	require.NotContains(t, out, "follow page")
}

func TestStreamStepLogsLimitStopsAcrossPages(t *testing.T) {
	api := &stepLogsAPI{pages: []stepLogPage{
		{records: []*models.AppOtelLogRecord{logRec("one"), logRec("two")}, next: "222"},
		{records: []*models.AppOtelLogRecord{logRec("three"), logRec("four")}, next: ""},
	}}
	svc := &Service{api: api}

	out := captureStdout(t, func() {
		err := svc.streamStepLogs(context.Background(), "ls_1", WorkflowStepLogsOptions{Limit: 3})
		require.NoError(t, err)
	})

	require.Equal(t, []string{"", "222"}, api.offsets)
	require.Contains(t, out, "one")
	require.Contains(t, out, "two")
	require.Contains(t, out, "three")
	require.NotContains(t, out, "four")
}

func TestReadAllStepLogsDrainsPages(t *testing.T) {
	api := &stepLogsAPI{pages: twoLogPages()}
	svc := &Service{api: api}

	records, err := svc.readAllStepLogs(context.Background(), "ls_1")
	require.NoError(t, err)
	require.Equal(t, []string{"", "111"}, api.offsets)
	require.Len(t, records, 2)
	require.Equal(t, "first page", records[0].Body)
	require.Equal(t, "second page", records[1].Body)
}

func twoLogPages() []stepLogPage {
	return []stepLogPage{
		{records: []*models.AppOtelLogRecord{logRec("first page")}, next: "111"},
		{records: []*models.AppOtelLogRecord{{
			Timestamp:    "2026-10-21T21:00:00.123456789Z",
			SeverityText: "Info",
			ServiceName:  "runner",
			Body:         "second page",
		}}, next: ""},
	}
}

func logRec(body string) *models.AppOtelLogRecord {
	return &models.AppOtelLogRecord{
		Timestamp:    "2026-10-21T21:00:00Z",
		SeverityText: "Info",
		ServiceName:  "runner",
		Body:         body,
	}
}

func TestUnixNanoCursorParsesTimestamp(t *testing.T) {
	got := unixNanoCursor("2026-10-21T21:00:00.123456789Z")
	parsed, err := time.Parse(time.RFC3339Nano, "2026-10-21T21:00:00.123456789Z")
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(got), got)
	require.NotEmpty(t, got)
	require.Equal(t, unixNanoCursor(parsed.Format(time.RFC3339Nano)), got)
}
