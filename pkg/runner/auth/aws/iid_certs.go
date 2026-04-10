package aws

import (
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
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

// NewIIDCertStore parses the embedded certs on startup.
func NewIIDCertStore(l *zap.Logger) (*IIDCertStore, error) {
	entries, err := iidCertsFS.ReadDir("iid_certs")
	if err != nil {
		return nil, fmt.Errorf("reading embedded iid_certs dir: %w", err)
	}

	certs := make(map[string]*x509.Certificate, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".pem" {
			continue
		}
		region := strings.TrimSuffix(name, ".pem")

		data, err := iidCertsFS.ReadFile("iid_certs/" + name)
		if err != nil {
			l.Warn("failed to read embedded cert",
				zap.String("region", region),
				zap.Error(err))
			continue
		}

		block, _ := pem.Decode(data)
		if block == nil {
			l.Warn("failed to decode embedded cert PEM",
				zap.String("region", region))
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			l.Warn("failed to parse embedded cert",
				zap.String("region", region),
				zap.Error(err))
			continue
		}
		certs[region] = cert
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no valid AWS IID certificates found")
	}

	l.Info("loaded AWS IID certificates",
		zap.Int("count", len(certs)))

	return &IIDCertStore{certs: certs}, nil
}

// GetCert returns the certificate for the given AWS region.
func (s *IIDCertStore) GetCert(region string) (*x509.Certificate, error) {
	cert, ok := s.certs[region]
	if !ok {
		return nil, fmt.Errorf("no certificate for region %s", region)
	}
	return cert, nil
}
