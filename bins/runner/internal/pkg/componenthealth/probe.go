package componenthealth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	providerProbe = "probe"

	probeKindHTTP = "http"
	probeKindTCP  = "tcp"
	probeKindExec = "exec"

	resourceKindHTTPProbe = "HTTPProbe"
	resourceKindTCPProbe  = "TCPProbe"
	resourceKindExecProbe = "ExecProbe"

	// probeTimeout is the hard per-probe budget: a probe can never hold up a
	// report cycle longer than this, whatever the target does.
	probeTimeout = 5 * time.Second

	// probeBodyDrainLimit bounds the response body read: it is irrelevant to the
	// verdict, drained only to free the connection.
	probeBodyDrainLimit = 4 * 1024

	// probeOutputLimit bounds how much of an exec probe's combined output travels
	// as the failure message.
	probeOutputLimit = 2 * 1024

	probeUserAgent = "nuon-runner/component-health"
)

// probeSpec is one synthetic probe declared on a component. Only specs built by
// newProbeSpec or newExecProbeSpec run, so a probe can never reach an unnamed target.
type probeSpec struct {
	kind   string
	target string
	// dialAddr is the host:port a tcp probe connects to, derived from target.
	dialAddr string
	// command is the argv an exec probe runs, never a shell string.
	command []string
	// name is the vendor's display label; falls back to the target or argv[0].
	name string
	// unresolved explains why a declared probe cannot run. Such a probe still
	// reports as unknown carrying this reason, rather than silently vanishing.
	unresolved string
}

// newProbeSpec validates and normalizes a declared http or tcp probe. Anything
// it rejects is never executed.
func newProbeSpec(kind, target string) (probeSpec, bool) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	target = strings.TrimSpace(target)
	if target == "" {
		return probeSpec{}, false
	}

	switch kind {
	case probeKindHTTP, "https":
		u, err := url.Parse(target)
		if err != nil || u.Host == "" {
			return probeSpec{}, false
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return probeSpec{}, false
		}
		return probeSpec{kind: probeKindHTTP, target: u.String()}, true
	case probeKindTCP:
		addr, ok := tcpDialAddr(target)
		if !ok {
			return probeSpec{}, false
		}
		return probeSpec{kind: probeKindTCP, target: addr, dialAddr: addr}, true
	default:
		return probeSpec{}, false
	}
}

// tcpDialAddr resolves a tcp probe target — either a bare host:port or a URL
// whose scheme implies the port — to a dialable address.
func tcpDialAddr(target string) (string, bool) {
	if u, err := url.Parse(target); err == nil && u.Host != "" && u.Scheme != "" {
		host := u.Hostname()
		port := u.Port()
		if port == "" {
			switch u.Scheme {
			case "https":
				port = "443"
			case "http":
				port = "80"
			default:
				return "", false
			}
		}
		return net.JoinHostPort(host, port), true
	}

	host, port, err := net.SplitHostPort(target)
	if err != nil || host == "" || port == "" {
		return "", false
	}
	return net.JoinHostPort(host, port), true
}

// newExecProbeSpec validates and normalizes a declared exec probe. The argv is
// taken verbatim — it is never joined into a shell string.
func newExecProbeSpec(command []string) (probeSpec, bool) {
	if len(command) == 0 || strings.TrimSpace(command[0]) == "" {
		return probeSpec{}, false
	}

	argv := slices.Clone(command)
	argv[0] = strings.TrimSpace(argv[0])

	return probeSpec{kind: probeKindExec, command: argv}, true
}

func (s probeSpec) resourceKind() string {
	switch s.kind {
	case probeKindTCP:
		return resourceKindTCPProbe
	case probeKindExec:
		return resourceKindExecProbe
	default:
		return resourceKindHTTPProbe
	}
}

// displayName is what the health row is named by.
func (s probeSpec) displayName() string {
	if s.name != "" {
		return s.name
	}
	if s.kind == probeKindExec && len(s.command) > 0 {
		return s.command[0]
	}
	return s.target
}

type probeResult struct {
	health     string
	message    string
	statusCode int
	exitCode   *int
	latency    time.Duration
}

// newProbeHTTPClient never follows redirects, so a probe cannot be bounced to an
// unnamed host, and owns its transport because the runner mutates the default one.
func newProbeHTTPClient() *http.Client {
	return &http.Client{
		Timeout: probeTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: probeTimeout}).DialContext,
			TLSHandshakeTimeout:   probeTimeout,
			ResponseHeaderTimeout: probeTimeout,
			DisableKeepAlives:     true,
		},
	}
}

func runProbe(ctx context.Context, client *http.Client, spec probeSpec) probeResult {
	if spec.unresolved != "" {
		return probeResult{health: healthUnknown, message: spec.unresolved}
	}

	switch spec.kind {
	case probeKindTCP:
		return runTCPProbe(ctx, spec.dialAddr)
	case probeKindExec:
		return runExecProbe(ctx, spec.command)
	default:
		return runHTTPProbe(ctx, client, spec.target)
	}
}

func runHTTPProbe(ctx context.Context, client *http.Client, target string) probeResult {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return probeResult{health: healthUnhealthy, message: err.Error()}
	}
	req.Header.Set("User-Agent", probeUserAgent)

	started := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return probeResult{health: healthUnhealthy, message: err.Error(), latency: time.Since(started)}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, probeBodyDrainLimit))

	res := probeResult{statusCode: resp.StatusCode, latency: time.Since(started)}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
		res.health = healthHealthy
		return res
	}
	res.health = healthUnhealthy
	res.message = fmt.Sprintf("HTTP %d from %s", resp.StatusCode, target)
	return res
}

func runTCPProbe(ctx context.Context, addr string) probeResult {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	started := time.Now()
	conn, err := (&net.Dialer{Timeout: probeTimeout}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return probeResult{health: healthUnhealthy, message: err.Error(), latency: time.Since(started)}
	}
	_ = conn.Close()

	return probeResult{health: healthHealthy, latency: time.Since(started)}
}

// execProbeEnv is deliberately minimal — the runner's own env carries its
// control-plane token and cloud credentials, which a probe must not inherit.
func execProbeEnv() []string {
	path := os.Getenv("PATH")
	if path == "" {
		path = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	}
	return []string{"PATH=" + path, "HOME=" + os.TempDir(), "TMPDIR=" + os.TempDir()}
}

// runExecProbe runs the declared argv directly — no shell, so nothing in the
// config is interpreted as a pipeline, redirect or expansion.
func runExecProbe(ctx context.Context, command []string) probeResult {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Env = execProbeEnv()

	started := time.Now()
	out, err := cmd.CombinedOutput()
	res := probeResult{latency: time.Since(started)}
	tail := outputTail(out)

	if ctx.Err() != nil {
		res.health = healthUnhealthy
		res.message = joinProbeMessage(
			fmt.Sprintf("timed out after %s", res.latency.Round(time.Millisecond)), tail)
		return res
	}

	if err == nil {
		res.health = healthHealthy
		return res
	}

	res.health = healthUnhealthy

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()
		res.exitCode = &code
		res.message = joinProbeMessage(fmt.Sprintf("exit code %d", code), tail)
		return res
	}

	res.message = joinProbeMessage(err.Error(), tail)
	return res
}

// outputTail returns the last probeOutputLimit bytes of an exec probe's output:
// failures report at the end.
func outputTail(out []byte) string {
	if len(out) > probeOutputLimit {
		out = out[len(out)-probeOutputLimit:]
	}
	return strings.TrimSpace(string(out))
}

func joinProbeMessage(reason, tail string) string {
	if tail == "" {
		return reason
	}
	return reason + ": " + tail
}

func probeResourceRow(spec probeSpec, res probeResult) *models.ServiceComponentHealthResource {
	detail := map[string]any{
		"type": spec.kind,
	}
	if spec.target != "" {
		detail["target"] = spec.target
	}
	if len(spec.command) > 0 {
		detail["command"] = spec.command
	}
	if res.statusCode > 0 {
		detail["status_code"] = res.statusCode
	}
	if res.exitCode != nil {
		detail["exit_code"] = *res.exitCode
	}
	if res.latency > 0 {
		detail["latency_ms"] = res.latency.Milliseconds()
	}

	details := map[string]any{"probe": detail}
	// A failing probe is the reason a component went bad, so it also travels as
	// a diagnosis — that is the copy the evaluator puts on a transition.
	if res.health != healthHealthy {
		details["diagnosis"] = map[string]any{"probe": detail}
	}

	blob := ""
	if b, err := json.Marshal(details); err == nil && len(b) <= maxDetailsBytes {
		blob = string(b)
	}

	return &models.ServiceComponentHealthResource{
		Provider: providerProbe,
		Kind:     spec.resourceKind(),
		Name:     spec.displayName(),
		Health:   res.health,
		Message:  res.message,
		Details:  blob,
	}
}

// probeSpecsFor maps a component's declared probes onto executable specs. One
// that cannot be built becomes an unresolved spec reporting unknown with the
// reason, rather than being silently skipped.
func probeSpecsFor(c *models.ServiceRunnerInstallComponent) []probeSpec {
	if c == nil || len(c.Probes) == 0 {
		return nil
	}

	specs := make([]probeSpec, 0, len(c.Probes))
	for _, p := range c.Probes {
		if p == nil {
			continue
		}

		kind := strings.ToLower(strings.TrimSpace(p.Type))

		// Checked before validity, or an exec probe whose argv still holds a
		// template would run with a literal "{{...}}" as input.
		var (
			spec probeSpec
			ok   bool
		)
		if !probeTargetsResolved(p) {
			ok = false
		} else if kind == probeKindExec {
			spec, ok = newExecProbeSpec(p.Command)
		} else {
			spec, ok = newProbeSpec(p.Type, p.URL)
		}
		if !ok {
			spec = probeSpec{
				kind:       kind,
				target:     strings.TrimSpace(p.URL),
				command:    p.Command,
				unresolved: unresolvedProbeReason(p),
			}
		}

		spec.name = strings.TrimSpace(p.Name)
		specs = append(specs, spec)
	}

	if len(specs) == 0 {
		return nil
	}
	return specs
}

// unresolvedProbeReason explains in the vendor's terms why a probe cannot run.
func unresolvedProbeReason(p *models.ServiceRunnerComponentProbe) string {
	target := strings.TrimSpace(p.URL)
	if strings.Contains(target, "{{") || anyTemplated(p.Command) {
		return "probe target could not be resolved from install state yet — it references a value the install has not produced"
	}
	switch strings.ToLower(strings.TrimSpace(p.Type)) {
	case probeKindExec:
		return "probe declares no command to run"
	case probeKindHTTP, "https", probeKindTCP:
		if target == "" {
			return "probe declares no url to check"
		}
		return "probe url is not valid: " + target
	default:
		return "probe type is not supported: " + strings.TrimSpace(p.Type)
	}
}

// probeTargetsResolved reports whether every templated value in the probe was
// substituted before it reached the runner.
func probeTargetsResolved(p *models.ServiceRunnerComponentProbe) bool {
	return !strings.Contains(p.URL, "{{") && !anyTemplated(p.Command)
}

func anyTemplated(args []string) bool {
	for _, a := range args {
		if strings.Contains(a, "{{") {
			return true
		}
	}
	return false
}
