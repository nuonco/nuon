// Package httpdebug provides an http.RoundTripper that traces every request
// and prints per-phase timings (DNS, connect, TLS, server wait, transfer) to
// stderr. Headers and bodies are never printed.
package httpdebug

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"sync"
	"time"
)

type Transport struct {
	base http.RoundTripper
}

func NewTransport(base http.RoundTripper) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &Transport{base: base}
}

// Printing happens off the request path: lines are handed to a printer
// goroutine so a slow stderr never delays a response. Flush drains pending
// lines before the process exits.
var (
	printCh     = make(chan string, 512)
	pending     sync.WaitGroup
	printerOnce sync.Once

	statsMu  sync.Mutex
	apiCount int
	apiTotal time.Duration
)

func enqueue(line string) {
	printerOnce.Do(func() {
		go func() {
			for l := range printCh {
				fmt.Fprintln(os.Stderr, l)
				pending.Done()
			}
		}()
	})

	pending.Add(1)
	select {
	case printCh <- line:
	default:
		pending.Done()
		fmt.Fprintln(os.Stderr, line)
	}
}

func Flush() {
	pending.Wait()
}

// PrintSummary flushes pending request lines, then prints cumulative API call
// stats alongside the run's wall-clock time.
func PrintSummary(w io.Writer, wall time.Duration) {
	Flush()
	statsMu.Lock()
	defer statsMu.Unlock()
	fmt.Fprintf(w, "[nuon debug] summary: api calls=%d api_total=%s wall=%s\n",
		apiCount, fmtDur(apiTotal), fmtDur(wall))
}

type timings struct {
	mu sync.Mutex

	start        time.Time
	dnsStart     time.Time
	dns          time.Duration
	connectStart time.Time
	connect      time.Duration
	tlsStart     time.Time
	tls          time.Duration
	wroteRequest time.Time
	firstByte    time.Time
	reused       bool
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	tm := &timings{start: time.Now()}

	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) {
			tm.mu.Lock()
			tm.dnsStart = time.Now()
			tm.mu.Unlock()
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			tm.mu.Lock()
			if !tm.dnsStart.IsZero() {
				tm.dns = time.Since(tm.dnsStart)
			}
			tm.mu.Unlock()
		},
		ConnectStart: func(_, _ string) {
			tm.mu.Lock()
			if tm.connectStart.IsZero() {
				tm.connectStart = time.Now()
			}
			tm.mu.Unlock()
		},
		ConnectDone: func(_, _ string, err error) {
			tm.mu.Lock()
			if err == nil && !tm.connectStart.IsZero() && tm.connect == 0 {
				tm.connect = time.Since(tm.connectStart)
			}
			tm.mu.Unlock()
		},
		TLSHandshakeStart: func() {
			tm.mu.Lock()
			tm.tlsStart = time.Now()
			tm.mu.Unlock()
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			tm.mu.Lock()
			if !tm.tlsStart.IsZero() {
				tm.tls = time.Since(tm.tlsStart)
			}
			tm.mu.Unlock()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			tm.mu.Lock()
			tm.reused = info.Reused
			tm.mu.Unlock()
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			tm.mu.Lock()
			tm.wroteRequest = time.Now()
			tm.mu.Unlock()
		},
		GotFirstResponseByte: func() {
			tm.mu.Lock()
			tm.firstByte = time.Now()
			tm.mu.Unlock()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		t.print(req, 0, tm, err)
		return resp, err
	}

	resp.Body = &tracedBody{
		body: resp.Body,
		done: func() { t.print(req, resp.StatusCode, tm, nil) },
	}
	return resp, nil
}

func (t *Transport) print(req *http.Request, status int, tm *timings, reqErr error) {
	total := time.Since(tm.start)
	statsMu.Lock()
	apiCount++
	apiTotal += total
	statsMu.Unlock()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	parts := []string{fmt.Sprintf("%s %s", req.Method, req.URL.Path)}
	if reqErr != nil {
		parts = append(parts, fmt.Sprintf("error=%q", reqErr.Error()))
	} else {
		parts = append(parts, fmt.Sprintf("status=%d", status))
	}

	if tm.reused {
		parts = append(parts, "conn=reused")
	} else {
		if tm.dns > 0 {
			parts = append(parts, "dns="+fmtDur(tm.dns))
		}
		if tm.connect > 0 {
			parts = append(parts, "connect="+fmtDur(tm.connect))
		}
		if tm.tls > 0 {
			parts = append(parts, "tls="+fmtDur(tm.tls))
		}
	}

	if !tm.wroteRequest.IsZero() && !tm.firstByte.IsZero() {
		parts = append(parts, "server="+fmtDur(tm.firstByte.Sub(tm.wroteRequest)))
	}
	if !tm.firstByte.IsZero() {
		parts = append(parts, "transfer="+fmtDur(total-tm.firstByte.Sub(tm.start)))
	}
	parts = append(parts, "total="+fmtDur(total))

	enqueue("[nuon debug] " + strings.Join(parts, " "))
}

func fmtDur(d time.Duration) string {
	return d.Round(100 * time.Microsecond).String()
}

// tracedBody defers timing output until the response body is fully consumed
// or closed, so transfer time covers the actual body read.
type tracedBody struct {
	body io.ReadCloser
	once sync.Once
	done func()
}

func (b *tracedBody) Read(p []byte) (int, error) {
	n, err := b.body.Read(p)
	if err == io.EOF {
		b.once.Do(b.done)
	}
	return n, err
}

func (b *tracedBody) Close() error {
	err := b.body.Close()
	b.once.Do(b.done)
	return err
}
