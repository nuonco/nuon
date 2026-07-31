package kafka

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// writeKeyPair emits a self-signed cert/key pair and returns the cert's serial,
// which the tests use to tell one generation of the cert from the next.
func writeKeyPair(t *testing.T, certPath, keyPath string, serial int64) int64 {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{CommonName: "ctl-api"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	require.NoError(t, os.WriteFile(certPath, certPEM, 0o600))

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0o600))

	return serial
}

// setMTime pins mtimes explicitly so reload detection isn't at the mercy of
// filesystem timestamp resolution.
func setMTime(t *testing.T, path string, ts time.Time) {
	t.Helper()
	require.NoError(t, os.Chtimes(path, ts, ts))
}

func serialOf(t *testing.T, r *tlsReloader) int64 {
	t.Helper()

	cert, err := r.clientCertificate(nil)
	require.NoError(t, err)
	require.NotEmpty(t, cert.Certificate)

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)

	return parsed.SerialNumber.Int64()
}

func TestTLSReloaderReloadsOnChange(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "user.crt")
	keyPath := filepath.Join(dir, "user.key")

	writeKeyPair(t, certPath, keyPath, 1)
	base := time.Now().Add(-time.Hour)
	setMTime(t, certPath, base)
	setMTime(t, keyPath, base)

	r, err := newTLSReloader(Config{TLSCertPath: certPath, TLSKeyPath: keyPath}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, int64(1), serialOf(t, r))

	// unchanged mtime: must not re-read
	require.Equal(t, int64(1), serialOf(t, r))

	writeKeyPair(t, certPath, keyPath, 2)
	rotated := base.Add(time.Minute)
	setMTime(t, certPath, rotated)
	setMTime(t, keyPath, rotated)

	require.Equal(t, int64(2), serialOf(t, r), "rotated cert should be picked up without a restart")
}

func TestTLSReloaderKeepsLastGoodCertOnFailure(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "user.crt")
	keyPath := filepath.Join(dir, "user.key")

	writeKeyPair(t, certPath, keyPath, 7)
	base := time.Now().Add(-time.Hour)
	setMTime(t, certPath, base)
	setMTime(t, keyPath, base)

	r, err := newTLSReloader(Config{TLSCertPath: certPath, TLSKeyPath: keyPath}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, int64(7), serialOf(t, r))

	require.NoError(t, os.WriteFile(certPath, []byte("not a certificate"), 0o600))
	setMTime(t, certPath, base.Add(time.Minute))

	cert, err := r.clientCertificate(nil)
	require.NoError(t, err, "a corrupt cert on disk must not fail the handshake")
	require.Equal(t, int64(7), serialOf(t, r), "should still serve the last good cert")
	require.NotEmpty(t, cert.Certificate)
}

func TestTLSReloaderKeepsLastGoodCertOnMissingFile(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "user.crt")
	keyPath := filepath.Join(dir, "user.key")

	writeKeyPair(t, certPath, keyPath, 9)

	r, err := newTLSReloader(Config{TLSCertPath: certPath, TLSKeyPath: keyPath}, zap.NewNop())
	require.NoError(t, err)

	// mid-rotation the secret dir symlink can briefly not resolve
	require.NoError(t, os.Remove(certPath))

	require.Equal(t, int64(9), serialOf(t, r))
}

func TestTLSReloaderNoClientCertConfigured(t *testing.T) {
	r, err := newTLSReloader(Config{}, zap.NewNop())
	require.NoError(t, err)

	cert, err := r.clientCertificate(nil)
	require.NoError(t, err)
	require.Empty(t, cert.Certificate, "no cert configured should present none, not error")
}

func TestTLSReloaderRequiresCertAndKeyTogether(t *testing.T) {
	_, err := newTLSReloader(Config{TLSCertPath: "/tmp/only.crt"}, zap.NewNop())
	require.Error(t, err)

	_, err = newTLSReloader(Config{TLSKeyPath: "/tmp/only.key"}, zap.NewNop())
	require.Error(t, err)
}

func TestTLSReloaderFailsFastOnUnreadableMaterial(t *testing.T) {
	dir := t.TempDir()

	_, err := newTLSReloader(Config{TLSCAPath: filepath.Join(dir, "missing.crt")}, zap.NewNop())
	require.Error(t, err, "a bad path should fail at startup, not on first produce")

	_, err = newTLSReloader(Config{
		TLSCertPath: filepath.Join(dir, "missing.crt"),
		TLSKeyPath:  filepath.Join(dir, "missing.key"),
	}, zap.NewNop())
	require.Error(t, err)
}

func TestTLSReloaderCARejectsNonPEM(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	require.NoError(t, os.WriteFile(caPath, []byte("garbage"), 0o600))

	_, err := newTLSReloader(Config{TLSCAPath: caPath}, zap.NewNop())
	require.Error(t, err)
}

func TestTLSReloaderCAReloadsOnChange(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	certPath := filepath.Join(dir, "tmp.crt")
	keyPath := filepath.Join(dir, "tmp.key")

	writeKeyPair(t, certPath, keyPath, 11)
	first, err := os.ReadFile(certPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(caPath, first, 0o600))

	base := time.Now().Add(-time.Hour)
	setMTime(t, caPath, base)

	r, err := newTLSReloader(Config{TLSCAPath: caPath}, zap.NewNop())
	require.NoError(t, err)

	cfg, err := r.tlsConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg.RootCAs)
	before := cfg.RootCAs

	writeKeyPair(t, certPath, keyPath, 12)
	second, err := os.ReadFile(certPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(caPath, second, 0o600))
	setMTime(t, caPath, base.Add(time.Minute))

	cfg, err = r.tlsConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg.RootCAs)
	require.False(t, before.Equal(cfg.RootCAs), "rotated CA bundle should be picked up")
}

func TestTLSReloaderConcurrentRotation(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "user.crt")
	keyPath := filepath.Join(dir, "user.key")
	caPath := filepath.Join(dir, "ca.crt")

	writeKeyPair(t, certPath, keyPath, 1)
	pemBytes, err := os.ReadFile(certPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(caPath, pemBytes, 0o600))

	r, err := newTLSReloader(Config{
		TLSCAPath:   caPath,
		TLSCertPath: certPath,
		TLSKeyPath:  keyPath,
	}, zap.NewNop())
	require.NoError(t, err)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := r.clientCertificate(nil); err != nil {
					t.Error(err)
					return
				}
				if _, err := r.tlsConfig(); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}

	base := time.Now().Add(-time.Hour)
	for i := 2; i < 25; i++ {
		writeKeyPair(t, certPath, keyPath, int64(i))
		ts := base.Add(time.Duration(i) * time.Minute)
		setMTime(t, certPath, ts)
		setMTime(t, keyPath, ts)
		setMTime(t, caPath, ts)
	}

	close(stop)
	wg.Wait()
}

func TestBaseOptsMTLS(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "user.crt")
	keyPath := filepath.Join(dir, "user.key")
	caPath := filepath.Join(dir, "ca.crt")

	writeKeyPair(t, certPath, keyPath, 21)
	pemBytes, err := os.ReadFile(certPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(caPath, pemBytes, 0o600))

	cfg := Config{
		Brokers:          []string{"broker:9093"},
		SecurityProtocol: securitySSL,
		TLSCAPath:        caPath,
		TLSCertPath:      certPath,
		TLSKeyPath:       keyPath,
	}

	opts, err := cfg.baseOpts(zap.NewNop())
	require.NoError(t, err)
	require.NotEmpty(t, opts)
}

func TestBaseOptsMTLSMissingFilesFailsFast(t *testing.T) {
	cfg := Config{
		Brokers:          []string{"broker:9093"},
		SecurityProtocol: securitySSL,
		TLSCertPath:      "/nonexistent/user.crt",
		TLSKeyPath:       "/nonexistent/user.key",
	}

	_, err := cfg.baseOpts(zap.NewNop())
	require.Error(t, err)
}
