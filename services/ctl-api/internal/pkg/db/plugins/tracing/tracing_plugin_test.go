package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type testAuthor struct {
	ID    uint
	Name  string
	Books []testBook `gorm:"foreignKey:AuthorID"`
}

type testBook struct {
	ID       uint
	AuthorID uint
	Title    string
}

func setupTest(t *testing.T) (*gorm.DB, *tracetest.SpanRecorder) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&testAuthor{}, &testBook{}))
	require.NoError(t, db.Use(NewTracingPlugin("psql")))

	return db, recorder
}

func attrMap(span sdktrace.ReadOnlySpan) map[attribute.Key]attribute.Value {
	m := map[attribute.Key]attribute.Value{}
	for _, kv := range span.Attributes() {
		m[kv.Key] = kv.Value
	}
	return m
}

func findSpan(t *testing.T, spans []sdktrace.ReadOnlySpan, name, table string) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, s := range spans {
		if s.Name() == name && attrMap(s)["db.sql.table"].AsString() == table {
			return s
		}
	}
	t.Fatalf("span %q for table %q not found in %d spans", name, table, len(spans))
	return nil
}

func TestQuerySpanAttributes(t *testing.T) {
	db, recorder := setupTest(t)

	require.NoError(t, db.Create(&testAuthor{Name: "amy"}).Error)

	var authors []testAuthor
	require.NoError(t, db.Where(testAuthor{Name: "amy"}).Find(&authors).Error)

	span := findSpan(t, recorder.Ended(), "gorm.query", "test_authors")
	attrs := attrMap(span)

	require.Equal(t, "postgresql.query", attrs["operation.name"].AsString())
	require.Equal(t, "query test_authors", attrs["resource.name"].AsString())
	require.Equal(t, "postgresql", attrs["db.system"].AsString())
	require.Equal(t, "query", attrs["db.operation"].AsString())
	require.Equal(t, "test_authors", attrs["db.sql.table"].AsString())
	require.Contains(t, attrs["db.statement"].AsString(), "SELECT")
	require.Contains(t, attrs["db.statement"].AsString(), "?")
	require.NotContains(t, attrs["db.statement"].AsString(), "amy")
	require.Equal(t, int64(1), attrs["db.rows_affected"].AsInt64())
	require.Equal(t, int64(1), attrs["nuon.db.response_size"].AsInt64())
	require.Equal(t, codes.Unset, span.Status().Code)
}

func TestRequestContextAttributesAndParenting(t *testing.T) {
	db, recorder := setupTest(t)

	metricCtx := &cctx.MetricContext{
		Endpoint:  "/v1/apps/:app_id",
		Method:    "GET",
		OrgID:     "org123",
		Context:   "public",
		Namespace: "ns1",
	}
	ctx := context.WithValue(context.Background(), keys.MetricsKey, metricCtx)

	ctx, parent := otel.Tracer("test").Start(ctx, "http.request")

	var authors []testAuthor
	require.NoError(t, db.WithContext(ctx).Find(&authors).Error)
	parent.End()

	span := findSpan(t, recorder.Ended(), "gorm.query", "test_authors")
	attrs := attrMap(span)

	require.Equal(t, "/v1/apps/:app_id", attrs["http.route"].AsString())
	require.Equal(t, "GET", attrs["http.method"].AsString())
	require.Equal(t, "org123", attrs["nuon.org_id"].AsString())
	require.Equal(t, "public", attrs["nuon.context"].AsString())
	require.Equal(t, "ns1", attrs["nuon.namespace"].AsString())

	require.Equal(t, parent.SpanContext().SpanID(), span.Parent().SpanID())
	require.Equal(t, parent.SpanContext().TraceID(), span.SpanContext().TraceID())
}

func TestPreloadQueriesBecomeChildSpans(t *testing.T) {
	db, recorder := setupTest(t)

	require.NoError(t, db.Create(&testAuthor{
		Name:  "amy",
		Books: []testBook{{Title: "one"}, {Title: "two"}},
	}).Error)

	var authors []testAuthor
	require.NoError(t, db.Preload("Books").Find(&authors).Error)

	var root, preload sdktrace.ReadOnlySpan
	for _, s := range recorder.Ended() {
		if s.Name() != "gorm.query" {
			continue
		}
		switch attrMap(s)["db.sql.table"].AsString() {
		case "test_authors":
			if !s.Parent().IsValid() {
				root = s
			}
		case "test_books":
			preload = s
		}
	}
	require.NotNil(t, root, "root query span not found")
	require.NotNil(t, preload, "preload query span not found")

	require.Equal(t, root.SpanContext().SpanID(), preload.Parent().SpanID())
	require.Equal(t, root.SpanContext().TraceID(), preload.SpanContext().TraceID())
	require.Equal(t, int64(1), attrMap(root)["nuon.db.preload_count"].AsInt64())
}

func TestStatementSpanExcludesPreloads(t *testing.T) {
	db, recorder := setupTest(t)

	require.NoError(t, db.Create(&testAuthor{
		Name:  "amy",
		Books: []testBook{{Title: "one"}, {Title: "two"}},
	}).Error)

	var operation, statement, preload sdktrace.ReadOnlySpan
	var authors []testAuthor
	require.NoError(t, db.Preload("Books").Find(&authors).Error)

	for _, s := range recorder.Ended() {
		table := attrMap(s)["db.sql.table"].AsString()
		switch {
		case s.Name() == "gorm.query" && table == "test_authors" && !s.Parent().IsValid():
			operation = s
		case s.Name() == "gorm.query.statement" && table == "test_authors":
			statement = s
		case s.Name() == "gorm.query" && table == "test_books":
			preload = s
		}
	}
	require.NotNil(t, operation, "operation span not found")
	require.NotNil(t, statement, "statement span not found")
	require.NotNil(t, preload, "preload span not found")

	require.Equal(t, operation.SpanContext().SpanID(), statement.Parent().SpanID())
	require.Equal(t, operation.SpanContext().TraceID(), statement.SpanContext().TraceID())

	// The statement span must close before the preload runs, and the operation span after it:
	// that difference is what separates statement_latency from gorm_operation_latency.
	require.False(t, statement.EndTime().After(preload.StartTime()))
	require.False(t, operation.EndTime().Before(preload.EndTime()))

	attrs := attrMap(statement)
	require.Equal(t, "postgresql.statement", attrs["operation.name"].AsString())
	require.Equal(t, "query test_authors", attrs["resource.name"].AsString())
	require.Contains(t, attrs["db.statement"].AsString(), "SELECT")
	require.NotContains(t, attrs["db.statement"].AsString(), "amy")
}

func TestStatementSpanEmittedForEachOperation(t *testing.T) {
	db, recorder := setupTest(t)

	author := testAuthor{Name: "amy"}
	require.NoError(t, db.Create(&author).Error)
	require.NoError(t, db.Model(&author).Update("name", "amy2").Error)
	require.NoError(t, db.Delete(&author).Error)

	got := map[string]bool{}
	for _, s := range recorder.Ended() {
		if attrMap(s)["operation.name"].AsString() == "postgresql.statement" {
			got[attrMap(s)["db.operation"].AsString()] = true
		}
	}

	for _, op := range []string{"create", "update", "delete"} {
		require.True(t, got[op], "no statement span for %s", op)
	}
}

func TestErrorsAreRecorded(t *testing.T) {
	db, recorder := setupTest(t)

	require.Error(t, db.Raw("SELECT * FROM missing_table").Scan(&map[string]any{}).Error)

	var errSpan sdktrace.ReadOnlySpan
	for _, s := range recorder.Ended() {
		if s.Status().Code == codes.Error {
			errSpan = s
		}
	}
	require.NotNil(t, errSpan, "expected a span with error status")
	require.NotEmpty(t, errSpan.Events(), "expected recorded error event")
}

func TestRecordNotFoundIsNotAnError(t *testing.T) {
	db, recorder := setupTest(t)

	var author testAuthor
	err := db.Where(testAuthor{Name: "ghost"}).First(&author).Error
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	span := findSpan(t, recorder.Ended(), "gorm.query", "test_authors")
	require.Equal(t, codes.Unset, span.Status().Code)
}

func TestNoopProviderAddsNoOverheadPath(t *testing.T) {
	otel.SetTracerProvider(trace.NewNoopTracerProvider())

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&testAuthor{}))
	require.NoError(t, db.Use(NewTracingPlugin("psql")))

	var authors []testAuthor
	require.NoError(t, db.Find(&authors).Error)
}

func TestGinContextQueryParentsToRootSpan(t *testing.T) {
	db, recorder := setupTest(t)
	require.NoError(t, db.Create(&testAuthor{Name: "gin"}).Error)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = httptest.NewRequest(http.MethodGet, "/v1/things", nil)
	ginCtx.Set(keys.MetricsKey, &cctx.MetricContext{
		Endpoint: "/v1/things",
		Method:   "GET",
		OrgID:    "org123",
	})

	rootCtx, rootSpan := otel.Tracer("test").Start(context.Background(), "GET /v1/things")
	ginCtx.Request = ginCtx.Request.WithContext(rootCtx)

	var authors []testAuthor
	require.NoError(t, db.WithContext(ginCtx).Where(testAuthor{Name: "gin"}).Find(&authors).Error)
	rootSpan.End()

	span := findSpan(t, recorder.Ended(), "gorm.query", "test_authors")
	require.Equal(t, rootSpan.SpanContext().TraceID(), span.SpanContext().TraceID())
	require.Equal(t, rootSpan.SpanContext().SpanID(), span.Parent().SpanID())

	attrs := attrMap(span)
	require.Equal(t, "/v1/things", attrs["http.route"].AsString())
	require.Equal(t, "org123", attrs["nuon.org_id"].AsString())
}

func TestWrappedGinContextStillParentsToRootSpan(t *testing.T) {
	db, recorder := setupTest(t)
	require.NoError(t, db.Create(&testAuthor{Name: "wrapped"}).Error)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = httptest.NewRequest(http.MethodGet, "/v1/things", nil)

	rootCtx, rootSpan := otel.Tracer("test").Start(context.Background(), "GET /v1/things")
	ginCtx.Request = ginCtx.Request.WithContext(rootCtx)

	type wrapperKey string
	wrapped := context.WithValue(ginCtx, wrapperKey("routing_decision"), "replica")

	var authors []testAuthor
	require.NoError(t, db.WithContext(wrapped).Where(testAuthor{Name: "wrapped"}).Find(&authors).Error)
	rootSpan.End()

	span := findSpan(t, recorder.Ended(), "gorm.query", "test_authors")
	require.Equal(t, rootSpan.SpanContext().TraceID(), span.SpanContext().TraceID())
	require.Equal(t, rootSpan.SpanContext().SpanID(), span.Parent().SpanID())
}
