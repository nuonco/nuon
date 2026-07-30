package sns

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const maxCertificateSize = 1 << 20

// Verifier verifies SNS message signatures, fetching and caching the AWS
// signing certificates.
type Verifier struct {
	client *http.Client
	mu     sync.RWMutex
	certs  map[string]*x509.Certificate
}

// NewVerifier returns a Verifier that fetches signing certificates with
// client, or http.DefaultClient when client is nil.
func NewVerifier(client *http.Client) *Verifier {
	if client == nil {
		client = http.DefaultClient
	}
	return &Verifier{client: client, certs: make(map[string]*x509.Certificate)}
}

// Verify checks msg's RSA signature against the AWS signing certificate for
// the topic's region.
func (v *Verifier) Verify(ctx context.Context, msg *Message) error {
	if msg == nil {
		return errors.New("nil SNS message")
	}
	expectedHost, err := hostForTopicARN(msg.TopicARN)
	if err != nil {
		return err
	}
	if err := validateHost(msg.SigningCertURL, expectedHost); err != nil {
		return fmt.Errorf("invalid SNS SigningCertURL: %w", err)
	}
	canonical, err := msg.canonicalString()
	if err != nil {
		return err
	}
	signature, err := base64.StdEncoding.DecodeString(msg.Signature)
	if err != nil {
		return fmt.Errorf("decode SNS signature: %w", err)
	}
	cert, err := v.certificate(ctx, msg.SigningCertURL, expectedHost)
	if err != nil {
		return err
	}
	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("SNS signing certificate does not contain an RSA public key")
	}
	var hash crypto.Hash
	var digest []byte
	switch msg.SignatureVersion {
	case "1":
		sum := sha1.Sum([]byte(canonical))
		hash, digest = crypto.SHA1, sum[:]
	case "2":
		sum := sha256.Sum256([]byte(canonical))
		hash, digest = crypto.SHA256, sum[:]
	default:
		return fmt.Errorf("unsupported SNS signature version %q", msg.SignatureVersion)
	}
	if err := rsa.VerifyPKCS1v15(publicKey, hash, digest, signature); err != nil {
		return fmt.Errorf("verify SNS signature: %w", err)
	}
	return nil
}

func (v *Verifier) certificate(ctx context.Context, rawURL, expectedHost string) (*x509.Certificate, error) {
	if err := validateURL(rawURL, true); err != nil {
		return nil, fmt.Errorf("invalid SNS SigningCertURL: %w", err)
	}
	if err := validateHost(rawURL, expectedHost); err != nil {
		return nil, fmt.Errorf("invalid SNS SigningCertURL: %w", err)
	}
	v.mu.RLock()
	cert := v.certs[rawURL]
	v.mu.RUnlock()
	if cert != nil && time.Now().Before(cert.NotAfter) {
		if err := cert.VerifyHostname(expectedHost); err != nil {
			return nil, fmt.Errorf("SNS signing certificate hostname verification failed: %w", err)
		}
		return cert, nil
	}

	client := *v.client
	previousCheck := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := validateURL(req.URL.String(), true); err != nil {
			return err
		}
		if err := validateHost(req.URL.String(), expectedHost); err != nil {
			return err
		}
		if previousCheck != nil {
			return previousCheck(req, via)
		}
		if len(via) >= 10 {
			return errors.New("too many SNS certificate redirects")
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch SNS signing certificate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch SNS signing certificate: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxCertificateSize+1))
	if err != nil {
		return nil, fmt.Errorf("read SNS signing certificate: %w", err)
	}
	if len(data) > maxCertificateSize {
		return nil, errors.New("SNS signing certificate is too large")
	}
	block, rest := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" || len(bytes.TrimSpace(rest)) != 0 {
		return nil, errors.New("invalid SNS signing certificate PEM")
	}
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse SNS signing certificate: %w", err)
	}
	if _, ok := cert.PublicKey.(*rsa.PublicKey); !ok {
		return nil, errors.New("SNS signing certificate does not contain an RSA public key")
	}
	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return nil, errors.New("SNS signing certificate is not currently valid")
	}
	if err := cert.VerifyHostname(expectedHost); err != nil {
		return nil, fmt.Errorf("SNS signing certificate hostname verification failed: %w", err)
	}
	v.mu.Lock()
	if cached := v.certs[rawURL]; cached != nil && time.Now().Before(cached.NotAfter) {
		cert = cached
	} else {
		v.certs[rawURL] = cert
	}
	v.mu.Unlock()
	if err := cert.VerifyHostname(expectedHost); err != nil {
		return nil, fmt.Errorf("SNS signing certificate hostname verification failed: %w", err)
	}
	return cert, nil
}
