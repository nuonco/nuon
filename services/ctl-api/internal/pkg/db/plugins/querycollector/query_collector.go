package querycollector

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type contextKey string

const ctxKey contextKey = "query_collector_start"

// QueryRecord holds the captured data for a single query execution.
type QueryRecord struct {
	Table        string        `json:"table"`
	Operation    string        `json:"operation"`
	SQL          string        `json:"sql"`
	Duration     time.Duration `json:"duration_ns"`
	DurationMS   float64       `json:"duration_ms"`
	RowsAffected int64         `json:"rows_affected"`
	ResponseSize int           `json:"response_size"`
	PreloadCount int           `json:"preload_count"`
	Timestamp    time.Time     `json:"timestamp"`
	Error        string        `json:"error,omitempty"`
	Caller       string        `json:"caller"`
	CallerURL    string        `json:"caller_url,omitempty"`
	DBType       string        `json:"db_type"`
	Source       string        `json:"source"`
	Endpoint     string        `json:"endpoint"`
}

// Collector accumulates query records in a fixed-size ring buffer.
type Collector struct {
	mu      sync.RWMutex
	records []QueryRecord
	maxSize int
	pos     int
	total   int
	writer  *Writer
}

// NewCollector creates a new collector with the given max buffer size.
func NewCollector(maxSize int) *Collector {
	return &Collector{
		records: make([]QueryRecord, 0, maxSize),
		maxSize: maxSize,
	}
}

// SetWriter attaches a Writer that persists records to ClickHouse.
func (c *Collector) SetWriter(w *Writer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writer = w
}

// callerURL builds a GitHub permalink for the given caller string (file:line).
func (c *Collector) callerURL(caller string) string {
	if caller == "" {
		return ""
	}
	idx := strings.LastIndex(caller, ":")
	if idx < 0 {
		return ""
	}
	filePath := caller[:idx]
	line := caller[idx+1:]
	return fmt.Sprintf("https://github.com/nuonco/nuon/tree/main/%s#L%s", filePath, line)
}

// Add inserts a record into the ring buffer and forwards to the writer if set.
func (c *Collector) Add(r QueryRecord) {
	c.mu.Lock()
	w := c.writer
	c.total++
	if len(c.records) < c.maxSize {
		c.records = append(c.records, r)
	} else {
		c.records[c.pos] = r
	}
	c.pos = (c.pos + 1) % c.maxSize
	c.mu.Unlock()

	if w != nil {
		w.Write(r)
	}
}

// Records returns a snapshot of all collected records (newest first).
func (c *Collector) Records() []QueryRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	n := len(c.records)
	out := make([]QueryRecord, n)
	for i := 0; i < n; i++ {
		idx := (c.pos - 1 - i + n) % n
		out[i] = c.records[idx]
	}
	return out
}

// Total returns the total number of queries captured (including evicted).
func (c *Collector) Total() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.total
}

// Clear resets the collector.
func (c *Collector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = c.records[:0]
	c.pos = 0
	c.total = 0
}

// Tables returns the distinct table names seen in the buffer.
func (c *Collector) Tables() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	seen := make(map[string]struct{})
	for _, r := range c.records {
		if r.Table != "" {
			seen[r.Table] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	return out
}

// Plugin is a GORM plugin that records every query into a Collector.
type Plugin struct {
	collector *Collector
	dbType    string
}

var _ gorm.Plugin = (*Plugin)(nil)

// NewPlugin creates a plugin that tags records with the given dbType ("psql" or "ch").
func NewPlugin(collector *Collector, dbType string) *Plugin {
	return &Plugin{collector: collector, dbType: dbType}
}

func (p *Plugin) Name() string { return "query-collector-" + p.dbType }

func (p *Plugin) Initialize(db *gorm.DB) error {
	prefix := "qc_" + p.dbType + "_"

	before := func(tx *gorm.DB) {
		ctx := context.WithValue(tx.Statement.Context, ctxKey, time.Now())
		tx.Statement.Context = ctx
	}

	db.Callback().Create().Before("*").Register(prefix+"before_create", before)
	db.Callback().Create().After("*").Register(prefix+"after_create", func(tx *gorm.DB) { p.afterAll(tx, "Create") })
	db.Callback().Query().Before("*").Register(prefix+"before_query", before)
	db.Callback().Query().After("*").Register(prefix+"after_query", func(tx *gorm.DB) { p.afterAll(tx, "Query") })
	db.Callback().Update().Before("*").Register(prefix+"before_update", before)
	db.Callback().Update().After("*").Register(prefix+"after_update", func(tx *gorm.DB) { p.afterAll(tx, "Update") })
	db.Callback().Delete().Before("*").Register(prefix+"before_delete", before)
	db.Callback().Delete().After("*").Register(prefix+"after_delete", func(tx *gorm.DB) { p.afterAll(tx, "Delete") })
	db.Callback().Raw().Before("*").Register(prefix+"before_raw", before)
	db.Callback().Raw().After("*").Register(prefix+"after_raw", func(tx *gorm.DB) { p.afterAll(tx, "Raw") })

	return nil
}

func (p *Plugin) afterAll(tx *gorm.DB, operation string) {
	val := tx.Statement.Context.Value(ctxKey)
	if val == nil {
		return
	}
	startTS := val.(time.Time)
	dur := time.Since(startTS)

	table := ""
	if tx.Statement.Schema != nil {
		table = tx.Statement.Schema.Table
	}

	respSize := 0
	if tx.Statement.ReflectValue.IsValid() {
		if tx.Statement.ReflectValue.Kind() == reflect.Slice {
			respSize = tx.Statement.ReflectValue.Len()
		} else if !tx.Statement.ReflectValue.IsZero() {
			respSize = 1
		}
	}

	errStr := ""
	if tx.Error != nil {
		errStr = tx.Error.Error()
	}

	source := "worker"
	endpoint := ""
	if mc, err := cctx.MetricsContextFromGinContext(tx.Statement.Context); err == nil {
		endpoint = mc.Endpoint
		if mc.Context != "" {
			source = mc.Context
		}
	}

	caller := findCaller()
	p.collector.Add(QueryRecord{
		Table:        table,
		Operation:    operation,
		SQL:          tx.Statement.SQL.String(),
		Duration:     dur,
		DurationMS:   float64(dur.Nanoseconds()) / 1e6,
		RowsAffected: tx.RowsAffected,
		ResponseSize: respSize,
		PreloadCount: len(tx.Statement.Preloads),
		Timestamp:    startTS,
		Error:        errStr,
		Caller:       caller,
		CallerURL:    p.collector.callerURL(caller),
		DBType:       p.dbType,
		Source:       source,
		Endpoint:     endpoint,
	})
}

// skipPrefixes are package paths that belong to GORM/plugin internals.
var skipPrefixes = []string{
	"gorm.io/",
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/",
	"moul.io/zapgorm2",
}

// findCaller walks the stack to find the first frame outside GORM and plugin internals.
func findCaller() string {
	for i := 4; i < 20; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		name := fn.Name()
		skip := false
		for _, prefix := range skipPrefixes {
			if strings.Contains(name, prefix) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if idx := strings.Index(file, "services/"); idx >= 0 {
			file = file[idx:]
		} else if idx := strings.Index(file, "pkg/"); idx >= 0 {
			file = file[idx:]
		}
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}
