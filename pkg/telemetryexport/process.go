package telemetryexport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
)

// CollectorOptions contains local process configuration. Only explicitly supplied
// environment values and allowlisted TLS/proxy settings reach the Collector.
type CollectorOptions struct {
	Binary      string
	Config      []byte
	HealthURL   string
	Environment []string
	Args        []string
	Logger      *zap.Logger
}

// Collector owns a child process and its temporary config. The supervisor owns
// restart policy; Done reports exit, and Stop joins shutdown and removes the config.
type Collector struct {
	cmd           *exec.Cmd
	tempDir       string
	done          chan struct{}
	startedAt     time.Time
	outputWriters []io.Closer
}

// StartCollector replaces the previous child only after preparing the new config.
// Always retain the returned handle: a preparation error leaves the previous child
// running, while a failed launch returns nil after cleanup. Cancellation terminates
// the child with SIGTERM, escalating to SIGKILL after five seconds. Callers must
// eventually Stop the handle to join exit and remove its temporary configuration.
func StartCollector(ctx context.Context, previous *Collector, options CollectorOptions) (*Collector, error) {
	if err := ctx.Err(); err != nil {
		return previous, err
	}
	if _, err := os.Stat(options.Binary); err != nil {
		return previous, err
	}
	tempDir, err := os.MkdirTemp("", "nuon-telemetry-export-")
	if err != nil {
		return previous, err
	}
	path := filepath.Join(tempDir, "collector.yaml")
	if err := os.WriteFile(path, options.Config, 0o600); err != nil {
		_ = os.RemoveAll(tempDir)
		return previous, err
	}
	args := append([]string{"--config", path}, options.Args...)
	cmd := exec.CommandContext(ctx, options.Binary, args...)
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 5 * time.Second
	cmd.Env = childEnvironment(options.Environment)
	stdout := &zapio.Writer{Log: options.Logger.With(zap.String("stream", "stdout")), Level: zapcore.WarnLevel}
	stderr := &zapio.Writer{Log: options.Logger.With(zap.String("stream", "stderr")), Level: zapcore.WarnLevel}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	previous.Stop()
	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		_ = os.RemoveAll(tempDir)
		return nil, err
	}
	child := &Collector{cmd: cmd, tempDir: tempDir, done: make(chan struct{}), startedAt: time.Now(), outputWriters: []io.Closer{stdout, stderr}}
	go func() {
		_ = cmd.Wait()
		for _, writer := range child.outputWriters {
			_ = writer.Close()
		}
		close(child.done)
	}()
	if err := waitForCollector(ctx, child, options.HealthURL); err != nil {
		child.Stop()
		return nil, err
	}
	return child, nil
}

func (c *Collector) Done() <-chan struct{} { return c.done }

func (c *Collector) StartedAt() time.Time { return c.startedAt }

// Stop is safe for a nil child, and waits for process exit before removing its config.
func (c *Collector) Stop() {
	if c == nil {
		return
	}
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-c.done:
		case <-time.After(5 * time.Second):
			_ = c.cmd.Process.Kill()
			<-c.done
		}
	}
	_ = os.RemoveAll(c.tempDir)
}

func waitForCollector(ctx context.Context, child *Collector, healthURL string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client := &http.Client{Timeout: 50 * time.Millisecond}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
				return nil
			}
		}
		select {
		case <-child.done:
			return fmt.Errorf("telemetry export collector exited before becoming healthy")
		case <-ctx.Done():
			return fmt.Errorf("telemetry export collector did not become healthy: %w", ctx.Err())
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// Do not pass the parent's bootstrap or cloud credentials into the Collector.
func childEnvironment(extra []string) []string {
	allowed := map[string]struct{}{
		"SSL_CERT_FILE": {}, "SSL_CERT_DIR": {},
		"HTTP_PROXY": {}, "HTTPS_PROXY": {}, "NO_PROXY": {},
		"http_proxy": {}, "https_proxy": {}, "no_proxy": {},
	}
	environment := make([]string, 0, len(extra)+len(allowed))
	for _, value := range os.Environ() {
		name, _, ok := strings.Cut(value, "=")
		if !ok {
			continue
		}
		if _, ok := allowed[name]; ok {
			environment = append(environment, value)
		}
	}
	return append(environment, extra...)
}
