package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

const tlsDialTimeout = 10 * time.Second

// tlsReloader re-reads the CA bundle and client keypair from disk when their
// mtimes change, so a rotated certificate is picked up without a restart.
//
// Certificates are renewed well before they expire and the previous one stays
// valid throughout the renewal window, so continuing to serve the last good
// certificate after a failed reload is safe — and strictly better than failing
// every handshake.
type tlsReloader struct {
	caPath   string
	certPath string
	keyPath  string
	l        *zap.Logger

	mu      sync.Mutex
	pool    *x509.CertPool
	caMod   time.Time
	cert    *tls.Certificate
	certMod time.Time
	keyMod  time.Time
}

// newTLSReloader loads the configured material once so a bad path or unreadable
// secret fails at startup rather than on the first produce.
func newTLSReloader(c Config, l *zap.Logger) (*tlsReloader, error) {
	r := &tlsReloader{
		caPath:   c.TLSCAPath,
		certPath: c.TLSCertPath,
		keyPath:  c.TLSKeyPath,
		l:        l.Named("tls"),
	}

	if (r.certPath == "") != (r.keyPath == "") {
		return nil, fmt.Errorf("kafka tls: cert and key paths must be set together")
	}

	if r.caPath != "" {
		if _, err := r.reloadCA(); err != nil {
			return nil, err
		}
	}
	if r.certPath != "" {
		if _, err := r.reloadCert(); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *tlsReloader) reloadCA() (*x509.CertPool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// kubelet swaps the whole secret directory symlink, so a stat can fail
	// transiently mid-rotation; keep the last good bundle rather than failing.
	fi, err := os.Stat(r.caPath)
	if err != nil {
		return r.pool, fmt.Errorf("kafka tls: stat ca %q: %w", r.caPath, err)
	}

	if r.pool != nil && fi.ModTime().Equal(r.caMod) {
		return r.pool, nil
	}

	pem, err := os.ReadFile(r.caPath)
	if err != nil {
		return r.pool, fmt.Errorf("kafka tls: read ca %q: %w", r.caPath, err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return r.pool, fmt.Errorf("kafka tls: no certificates found in %q", r.caPath)
	}

	if r.pool != nil {
		r.l.Info("reloaded kafka ca bundle", zap.String("path", r.caPath))
	}
	r.pool, r.caMod = pool, fi.ModTime()

	return r.pool, nil
}

func (r *tlsReloader) reloadCert() (*tls.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	certFI, err := os.Stat(r.certPath)
	if err != nil {
		return r.cert, fmt.Errorf("kafka tls: stat cert %q: %w", r.certPath, err)
	}
	keyFI, err := os.Stat(r.keyPath)
	if err != nil {
		return r.cert, fmt.Errorf("kafka tls: stat key %q: %w", r.keyPath, err)
	}

	if r.cert != nil && certFI.ModTime().Equal(r.certMod) && keyFI.ModTime().Equal(r.keyMod) {
		return r.cert, nil
	}

	cert, err := tls.LoadX509KeyPair(r.certPath, r.keyPath)
	if err != nil {
		return r.cert, fmt.Errorf("kafka tls: load keypair: %w", err)
	}

	if r.cert != nil {
		r.l.Info("reloaded kafka client certificate",
			zap.String("path", r.certPath),
			zap.Time("not_after", notAfter(&cert)),
		)
	}
	r.cert, r.certMod, r.keyMod = &cert, certFI.ModTime(), keyFI.ModTime()

	return r.cert, nil
}

// clientCertificate is invoked per TLS handshake. An empty certificate tells the
// server we have none to present, which is correct when it doesn't ask for one.
func (r *tlsReloader) clientCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	if r.certPath == "" {
		return &tls.Certificate{}, nil
	}

	cert, err := r.reloadCert()
	if err != nil {
		if cert == nil {
			return nil, err
		}
		r.l.Warn("kafka client certificate reload failed; using previously loaded certificate",
			zap.Error(err),
			zap.Time("not_after", notAfter(cert)),
		)
	}

	return cert, nil
}

func (r *tlsReloader) tlsConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion:           tls.VersionTLS12,
		GetClientCertificate: r.clientCertificate,
	}

	if r.caPath == "" {
		return cfg, nil
	}

	pool, err := r.reloadCA()
	if err != nil {
		if pool == nil {
			return nil, err
		}
		r.l.Warn("kafka ca reload failed; using previously loaded bundle", zap.Error(err))
	}
	cfg.RootCAs = pool

	return cfg, nil
}

// dial replaces franz-go's DialTLSConfig, which snapshots a single tls.Config at
// client construction and would pin RootCAs for the process lifetime.
func (r *tlsReloader) dial(ctx context.Context, network, host string) (net.Conn, error) {
	cfg, err := r.tlsConfig()
	if err != nil {
		return nil, err
	}

	server, _, err := net.SplitHostPort(host)
	if err != nil {
		return nil, fmt.Errorf("kafka tls: split %q: %w", host, err)
	}
	cfg.ServerName = server

	d := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: tlsDialTimeout},
		Config:    cfg,
	}

	return d.DialContext(ctx, network, host)
}

func notAfter(cert *tls.Certificate) time.Time {
	if cert == nil || len(cert.Certificate) == 0 {
		return time.Time{}
	}
	leaf := cert.Leaf
	if leaf == nil {
		parsed, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return time.Time{}
		}
		leaf = parsed
	}

	return leaf.NotAfter
}
