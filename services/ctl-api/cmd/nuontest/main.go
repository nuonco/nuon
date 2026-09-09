// nuontest sets up disposable test databases and runs the ctl-api integration tests.
//
// With no arguments it only creates and migrates the databases named by DB_NAME and
// CLICKHOUSE_DB_NAME. With arguments it additionally runs `go test`, forwarding every
// argument after the nuontest flags.
//
// Use -shards=N to split the packages across N parallel lanes, each against its own
// Postgres and ClickHouse database (DB_NAME_i / CLICKHOUSE_DB_NAME_i) and its own
// blob-cache directory. Lane 0 runs the flow testworker package; the migrations
// package runs last, alone, on the final lane's database. Each lane runs its packages
// serially, so pass -p=1 in the forwarded go test flags.
//
//	nuontest -shards=2 -v -timeout=40m -count=1 -p=1 -skip 'TestSuite/TestPin' <packages...>
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/tests"
)

const (
	flowPackageSuffix       = "internal/pkg/flow/testworker"
	migrationsPackageSuffix = "internal/pkg/db/psql/migrations"
	maxShards               = 8
)

// go test flags that take a separate value argument, needed to tell
// `-skip foo` apart from a package path. Flags passed as `-flag=value`
// are handled regardless of this table.
var valueFlags = map[string]bool{
	"-skip": true, "-run": true, "-timeout": true, "-count": true,
	"-p": true, "-parallel": true, "-cpu": true, "-bench": true, "-benchtime": true,
}

// Shared-output flags that make no sense across concurrent lanes.
var rejectedFlags = map[string]bool{"-coverprofile": true, "-outputdir": true, "-o": true}

func main() {
	shards, args, err := parseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("nuontest: %v", err)
	}

	log.Println("setting up test databases...")

	dbCfg, err := tests.LoadDBConfig()
	if err != nil {
		log.Fatalf("failed to load db config: %v", err)
	}
	chCfg, err := tests.LoadCHConfig()
	if err != nil {
		log.Fatalf("failed to load clickhouse config: %v", err)
	}

	plan, err := buildPlan(shards, args, dbCfg, chCfg)
	if err != nil {
		log.Fatalf("nuontest: %v", err)
	}

	if err := plan.setupDatabases(); err != nil {
		log.Fatalf("failed to setup test databases: %v", err)
	}
	log.Println("test database setup complete")

	// No go test args: setup-only mode.
	if len(args) == 0 {
		return
	}

	os.Exit(plan.run())
}

// shard is one parallel lane: a package list bound to its own databases.
type shard struct {
	index    int
	dbCfg    tests.DBConfig
	chCfg    tests.CHConfig
	packages []string
	flags    []string

	blobCacheDir string
	phaseLogPath func(phase string) string
}

type plan struct {
	shards     []shard
	migrations []string // migrations package flags+path, run last on the final shard
}

// parseArgs extracts -shards=N from the argument list and passes the rest through
// for go test flag/package classification.
func parseArgs(args []string) (int, []string, error) {
	shards := 1
	rest := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-shards" || arg == "--shards":
			if i+1 >= len(args) {
				return 0, nil, errors.New("-shards requires a value")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 || n > maxShards {
				return 0, nil, fmt.Errorf("invalid -shards %q (want 1..%d)", args[i], maxShards)
			}
			shards = n
		case strings.HasPrefix(arg, "-shards=") || strings.HasPrefix(arg, "--shards="):
			v := arg[strings.IndexByte(arg, '=')+1:]
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 || n > maxShards {
				return 0, nil, fmt.Errorf("invalid -shards %q (want 1..%d)", v, maxShards)
			}
			shards = n
		default:
			rest = append(rest, arg)
		}
	}
	return shards, rest, nil
}

// splitFlagsAndPackages classifies forwarded go test arguments.
func splitFlagsAndPackages(args []string) (flags, packages []string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			packages = append(packages, arg)
			continue
		}
		name, hasValue := arg, strings.Contains(arg, "=")
		if idx := strings.IndexByte(arg, '='); idx >= 0 {
			name = arg[:idx]
		}
		if rejectedFlags[name] {
			return nil, nil, fmt.Errorf("%s cannot be used with -shards (shared output path)", name)
		}
		flags = append(flags, arg)
		if !hasValue && valueFlags[name] {
			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("flag %s requires a value", name)
			}
			i++
			flags = append(flags, args[i])
		}
	}
	return flags, packages, nil
}

func buildPlan(shardCount int, args []string, dbCfg tests.DBConfig, chCfg tests.CHConfig) (*plan, error) {
	base := filepath.Base(dbCfg.DBName)
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(base) {
		return nil, fmt.Errorf("refusing to drop database with unsafe name %q", dbCfg.DBName)
	}

	flags, packages, err := splitFlagsAndPackages(args)
	if err != nil {
		return nil, err
	}

	// Lane assignment matches packages by import path, so patterns like ./...
	// must be expanded first. Without this, every lane would run the full
	// pattern set against one database.
	if shardCount > 1 && len(packages) > 0 {
		packages, err = expandPackages(packages)
		if err != nil {
			return nil, err
		}
		// go test's 10m default kills the flow testworker package (~12m)
		// before it finishes; keep a generous default unless one is passed.
		if !hasTimeoutFlag(flags) {
			flags = append(flags, "-timeout", "45m")
		}
	}

	p := &plan{shards: make([]shard, shardCount)}
	blobRoot := os.Getenv("TEMPORAL_BLOB_CACHE_DIR")
	if blobRoot == "" {
		blobRoot = "/tmp/temporal-blobs"
	}

	// Sort packages for deterministic lane assignment.
	sorted := append([]string(nil), packages...)
	sort.Strings(sorted)

	var flow, rest []string
	for _, pkg := range sorted {
		switch {
		case strings.HasSuffix(pkg, flowPackageSuffix):
			flow = append(flow, pkg)
		default:
			rest = append(rest, pkg)
		}
	}

	for i := 0; i < shardCount; i++ {
		s := &p.shards[i]
		s.index = i
		s.flags = flags
		s.blobCacheDir = filepath.Join(blobRoot, fmt.Sprintf("shard-%d", i))

		if shardCount == 1 {
			s.dbCfg, s.chCfg = dbCfg, chCfg
			s.packages = sorted
		} else {
			suffix := strconv.Itoa(i)
			s.dbCfg = withDBName(dbCfg, base+"_"+suffix)
			s.chCfg = withCHName(chCfg, base+"_"+suffix)
			if i == 0 {
				s.packages = flow
			} else {
				s.packages = assignEvenly(rest, shardCount-1, i-1)
			}
		}

		s.phaseLogPath = func(phase string) string {
			return filepath.Join(os.TempDir(), fmt.Sprintf("nuontest-shard%d-%s-%d.log", i, phase, os.Getpid()))
		}
	}

	// The migrations package mutates its shard's schema, so it runs last and
	// alone on the final lane's database — sequenced by the launcher, not by
	// go test argument order.
	if shardCount > 1 {
		last := &p.shards[shardCount-1]
		for _, pkg := range last.packages {
			if strings.HasSuffix(pkg, migrationsPackageSuffix) {
				p.migrations = append(last.flags, pkg)
				last.packages = removePackage(last.packages, pkg)
				break
			}
		}
	}

	return p, nil
}

func withDBName(cfg tests.DBConfig, name string) tests.DBConfig {
	cfg.DBName = name
	return cfg
}

// expandPackages resolves go test patterns (./..., ./internal/...) into import
// paths so lanes can be assigned by suffix. Only packages with test files are
// kept: linking a test binary for every no-test package under -p=1 adds
// minutes of pure overhead.
func expandPackages(patterns []string) ([]string, error) {
	out, err := exec.Command("go", append([]string{"list", "-f",
		`{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}`}, patterns...)...).Output()
	if err != nil {
		return nil, fmt.Errorf("expanding package patterns: %w", err)
	}
	var pkgs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			pkgs = append(pkgs, line)
		}
	}
	if len(pkgs) == 0 {
		return nil, errors.New("package patterns matched no packages")
	}
	return pkgs, nil
}

func hasTimeoutFlag(flags []string) bool {
	for _, f := range flags {
		if idx := strings.IndexByte(f, '='); idx >= 0 {
			f = f[:idx]
		}
		if f == "-timeout" || f == "--timeout" {
			return true
		}
	}
	return false
}

func withCHName(cfg tests.CHConfig, name string) tests.CHConfig {
	cfg.Name = name
	return cfg
}

// assignEvenly distributes len(packages) items across lanes as evenly as possible
// and returns lane i's share. Earlier lanes get the larger remainder.
func assignEvenly(packages []string, lanes, i int) []string {
	if lanes <= 0 {
		return nil
	}
	n := len(packages)
	base, extra := n/lanes, n%lanes
	start := 0
	for lane := 0; lane < i; lane++ {
		size := base
		if lane < extra {
			size++
		}
		start += size
	}
	size := base
	if i < extra {
		size++
	}
	return packages[start : start+size]
}

func removePackage(packages []string, pkg string) []string {
	out := packages[:0]
	for _, p := range packages {
		if p != pkg {
			out = append(out, p)
		}
	}
	return out
}

func (p *plan) setupDatabases() error {
	schemaDir := tests.SchemaSnapshotDir()
	for i := range p.shards {
		s := &p.shards[i]
		if snapshotPath, ok := tests.SchemaSnapshotPath(s.dbCfg); ok {
			log.Printf("shard %d: restoring postgres schema from %s", s.index, snapshotPath)
			if err := tests.RestoreDatabase(s.dbCfg, snapshotPath); err != nil {
				return fmt.Errorf("shard %d postgres restore: %w", s.index, err)
			}
		} else {
			log.Printf("shard %d: creating postgresql database %s...", s.index, s.dbCfg.DBName)
			if err := tests.CreateAndMigrateDatabase(s.dbCfg); err != nil {
				return fmt.Errorf("shard %d postgres: %w", s.index, err)
			}
			if schemaDir != "" {
				if err := tests.DumpSchema(s.dbCfg, schemaDir); err != nil {
					return fmt.Errorf("shard %d postgres schema dump: %w", s.index, err)
				}
			}
		}
		log.Printf("shard %d: creating clickhouse database %s...", s.index, s.chCfg.Name)
		if err := tests.CreateAndMigrateCHDatabase(s.chCfg, s.dbCfg); err != nil {
			return fmt.Errorf("shard %d clickhouse: %w", s.index, err)
		}
		if err := os.MkdirAll(s.blobCacheDir, 0o755); err != nil {
			return fmt.Errorf("shard %d blob cache dir: %w", s.index, err)
		}
	}
	return nil
}

func (s *shard) env() []string {
	return append(os.Environ(),
		"INTEGRATION=true",
		fmt.Sprintf("DB_NAME=%s", s.dbCfg.DBName),
		fmt.Sprintf("CLICKHOUSE_DB_NAME=%s", s.chCfg.Name),
		fmt.Sprintf("TEMPORAL_BLOB_CACHE_DIR=%s", s.blobCacheDir),
	)
}

func (p *plan) run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	results := make([]shardResult, len(p.shards))
	var wg sync.WaitGroup
	for i := range p.shards {
		wg.Add(1)
		go func(s *shard) {
			defer wg.Done()
			results[s.index] = s.runTests(ctx, "tests", s.packages)
		}(&p.shards[i])
	}
	wg.Wait()

	// Migrations run only after all lanes have finished, exclusively on the
	// final lane's database.
	if len(p.migrations) > 0 {
		last := &p.shards[len(p.shards)-1]
		res := last.runTests(ctx, "migrations", p.migrations)
		results[last.index].migrations = &res
	}

	return summarize(p, results)
}

type shardResult struct {
	elapsed    time.Duration
	status     int
	logPath    string
	migrations *shardResult
}

func (s *shard) runTests(ctx context.Context, phase string, packages []string) shardResult {
	start := time.Now()
	res := shardResult{status: 1}
	res.logPath = s.phaseLogPath(phase)

	if len(packages) == 0 {
		res.status = 0 // nothing assigned to this lane
		res.elapsed = time.Since(start)
		return res
	}

	args := append([]string{"test"}, s.flags...)
	args = append(args, packages...)
	cmd := exec.Command("go", args...)
	cmd.Env = s.env()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logFile, err := os.Create(res.logPath)
	if err != nil {
		log.Printf("shard %d %s: failed to create log %s: %v", s.index, phase, res.logPath, err)
		return res
	}
	defer logFile.Close()

	prefix := fmt.Sprintf("shard%d| ", s.index)
	out := &prefixWriter{prefix: prefix, w: os.Stdout, atLineStart: true}
	tee := io.MultiWriter(out, logFile)
	cmd.Stdout = tee
	cmd.Stderr = tee

	if err := cmd.Start(); err != nil {
		log.Printf("shard %d %s: failed to start go test: %v", s.index, phase, err)
		res.elapsed = time.Since(start)
		return res
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err == nil {
			res.status = 0
		}
	case <-ctx.Done():
		// Kill the whole child process group; go test spawns test binaries.
		syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		<-done
		log.Printf("shard %d %s: cancelled", s.index, phase)
	}

	res.elapsed = time.Since(start)
	return res
}

func summarize(p *plan, results []shardResult) int {
	failed := false
	fmt.Println("\nnuontest summary:")
	for i, r := range results {
		status := "ok"
		if r.status != 0 {
			status = "FAIL"
			failed = true
		}
		if len(p.shards[i].packages) == 0 && r.migrations == nil {
			status = "empty (no packages assigned)"
		}
		line := fmt.Sprintf("  shard %d: %s in %s (log: %s)", i, status, r.elapsed.Round(time.Second), r.logPath)
		if r.migrations != nil {
			mStatus := "ok"
			if r.migrations.status != 0 {
				mStatus = "FAIL"
				failed = true
			}
			line += fmt.Sprintf("\n  shard %d migrations: %s in %s (log: %s)", i, mStatus, r.migrations.elapsed.Round(time.Second), r.migrations.logPath)
		}
		fmt.Println(line)
	}
	if failed {
		return 1
	}
	return 0
}

// prefixWriter writes each line prefixed, safe for concurrent cmd.Stdout/cmd.Stderr.
type prefixWriter struct {
	mu     sync.Mutex
	prefix string
	w      io.Writer

	atLineStart bool
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	written := 0
	for len(b) > 0 {
		if p.atLineStart {
			if _, err := io.WriteString(p.w, p.prefix); err != nil {
				return written, err
			}
			p.atLineStart = false
		}
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			if _, err := p.w.Write(b); err != nil {
				return written, err
			}
			written += len(b)
			break
		}
		if _, err := p.w.Write(b[:i+1]); err != nil {
			return written, err
		}
		written += i + 1
		b = b[i+1:]
		p.atLineStart = true
	}
	return written, nil
}
