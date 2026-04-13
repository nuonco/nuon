package aws

import (
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// iid_certs/ contains one PEM file per AWS region, named <region>.pem.
// Source: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/regions-certs.html
//
//go:embed iid_certs/*.pem
var iidCertsFS embed.FS

// IIDCertStore provides parsed x509 certificates for IID verification.
type IIDCertStore struct {
	certs map[string]*x509.Certificate
}

// NewIIDCertStore loads IID verification certificates. If certsDir is
// non-empty and the directory exists, PEM files from it override the
// embedded defaults.
func NewIIDCertStore(l *zap.Logger, certsDir string) (*IIDCertStore, error) {
	certs := make(map[string]*x509.Certificate)

	// Load embedded certs as defaults.
	embeddedFS, _ := fs.Sub(iidCertsFS, "iid_certs")
	embedded := loadCertsFromFS(l, embeddedFS, certs)
	l.Info("loaded AWS IID certificates from embedded",
		zap.Int("count", embedded))

	// Override with certs from config dir (if configured).
	if certsDir != "" {
		loaded := loadCertsFromFS(l, os.DirFS(certsDir), certs)
		if loaded > 0 {
			l.Info("loaded AWS IID certificate overrides from config dir",
				zap.String("dir", certsDir),
				zap.Int("count", loaded))
		}
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no valid AWS IID certificates found")
	}

	l.Info("total AWS IID certificates loaded",
		zap.Int("count", len(certs)))

	return &IIDCertStore{certs: certs}, nil
}

// loadCertsFromFS reads PEM files from an fs.FS into certs.
// Returns the number of certs successfully loaded.
func loadCertsFromFS(l *zap.Logger, fsys fs.FS, certs map[string]*x509.Certificate) int {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		l.Warn("unable to read IID certs dir",
			zap.Error(err))
		return 0
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pem" {
			continue
		}
		region := strings.TrimSuffix(entry.Name(), ".pem")

		data, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			l.Warn("failed to read cert",
				zap.String("region", region),
				zap.Error(err))
			continue
		}

		cert, err := parsePEM(data)
		if err != nil {
			l.Warn("failed to parse cert",
				zap.String("region", region),
				zap.Error(err))
			continue
		}

		certs[region] = cert
		loaded++
	}
	return loaded
}

func parsePEM(data []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	return x509.ParseCertificate(block.Bytes)
}

// GetCert returns the certificate for the given AWS region.
func (s *IIDCertStore) GetCert(region string) (*x509.Certificate, error) {
	cert, ok := s.certs[region]
	if !ok {
		return nil, fmt.Errorf("no certificate for region %s", region)
	}
	return cert, nil
}
