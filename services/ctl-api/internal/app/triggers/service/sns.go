package service

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

const maxSNSCertificateSize = 1 << 20

var (
	snsRegionPattern = regexp.MustCompile(`^[a-z]{2,4}(?:-[a-z0-9]+)+-\d+$`)
	snsCertPattern   = regexp.MustCompile(`^SimpleNotificationService-[[:xdigit:]]+\.pem$`)
)

type snsMessage struct {
	Type             string `json:"Type"`
	MessageID        string `json:"MessageId"`
	TopicARN         string `json:"TopicArn"`
	Subject          string `json:"Subject,omitempty"`
	Message          string `json:"Message"`
	Timestamp        string `json:"Timestamp"`
	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`
	Token            string `json:"Token,omitempty"`
	SubscribeURL     string `json:"SubscribeURL,omitempty"`
}

type snsVerifier struct {
	client *http.Client
	mu     sync.RWMutex
	certs  map[string]*x509.Certificate
}

func newSNSVerifier(client *http.Client) *snsVerifier {
	if client == nil {
		client = http.DefaultClient
	}
	return &snsVerifier{client: client, certs: make(map[string]*x509.Certificate)}
}

func parseSNSMessage(body []byte) (*snsMessage, error) {
	var msg snsMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&msg); err != nil {
		return nil, fmt.Errorf("decode SNS message: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("invalid trailing SNS message data")
	}
	if msg.Type == "" || msg.MessageID == "" || msg.TopicARN == "" || msg.Message == "" || msg.Timestamp == "" || msg.SignatureVersion == "" || msg.Signature == "" || msg.SigningCertURL == "" {
		return nil, errors.New("SNS message is missing required fields")
	}
	switch msg.Type {
	case "Notification":
	case "SubscriptionConfirmation", "UnsubscribeConfirmation":
		if msg.Token == "" || msg.SubscribeURL == "" {
			return nil, errors.New("SNS confirmation is missing required fields")
		}
		if err := validateSNSSubscribeURL(msg.SubscribeURL, msg.TopicARN, msg.Token); err != nil {
			return nil, fmt.Errorf("invalid SNS SubscribeURL: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported SNS message type %q", msg.Type)
	}
	if msg.SignatureVersion != "1" && msg.SignatureVersion != "2" {
		return nil, fmt.Errorf("unsupported SNS signature version %q", msg.SignatureVersion)
	}
	if err := validateSNSURL(msg.SigningCertURL, true); err != nil {
		return nil, fmt.Errorf("invalid SNS SigningCertURL: %w", err)
	}
	return &msg, nil
}

func (m *snsMessage) canonicalString() (string, error) {
	var fields []string
	switch m.Type {
	case "Notification":
		fields = []string{"Message", m.Message, "MessageId", m.MessageID}
		if m.Subject != "" {
			fields = append(fields, "Subject", m.Subject)
		}
		fields = append(fields, "Timestamp", m.Timestamp, "TopicArn", m.TopicARN, "Type", m.Type)
	case "SubscriptionConfirmation", "UnsubscribeConfirmation":
		fields = []string{"Message", m.Message, "MessageId", m.MessageID, "SubscribeURL", m.SubscribeURL, "Timestamp", m.Timestamp, "Token", m.Token, "TopicArn", m.TopicARN, "Type", m.Type}
	default:
		return "", fmt.Errorf("unsupported SNS message type %q", m.Type)
	}
	return strings.Join(fields, "\n") + "\n", nil
}

func (v *snsVerifier) verify(ctx context.Context, msg *snsMessage) error {
	if msg == nil {
		return errors.New("nil SNS message")
	}
	expectedHost, err := snsHostForTopicARN(msg.TopicARN)
	if err != nil {
		return err
	}
	if err := validateSNSHost(msg.SigningCertURL, expectedHost); err != nil {
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

func (v *snsVerifier) certificate(ctx context.Context, rawURL, expectedHost string) (*x509.Certificate, error) {
	if err := validateSNSURL(rawURL, true); err != nil {
		return nil, fmt.Errorf("invalid SNS SigningCertURL: %w", err)
	}
	if err := validateSNSHost(rawURL, expectedHost); err != nil {
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
		if err := validateSNSURL(req.URL.String(), true); err != nil {
			return err
		}
		if err := validateSNSHost(req.URL.String(), expectedHost); err != nil {
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
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSNSCertificateSize+1))
	if err != nil {
		return nil, fmt.Errorf("read SNS signing certificate: %w", err)
	}
	if len(data) > maxSNSCertificateSize {
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

func validateSNSURL(rawURL string, certificate bool) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Port() != "" {
		return errors.New("SNS URL must use unadorned HTTPS on the default port")
	}
	host := strings.ToLower(u.Hostname())
	suffix := ".amazonaws.com"
	if strings.HasSuffix(host, ".amazonaws.com.cn") {
		suffix = ".amazonaws.com.cn"
	}
	if !strings.HasPrefix(host, "sns.") || !strings.HasSuffix(host, suffix) {
		return errors.New("SNS URL has an untrusted host")
	}
	region := strings.TrimSuffix(strings.TrimPrefix(host, "sns."), suffix)
	if !snsRegionPattern.MatchString(region) {
		return errors.New("SNS URL has an invalid region")
	}
	if certificate && (!snsCertPattern.MatchString(path.Base(u.EscapedPath())) || strings.Contains(strings.ToLower(u.EscapedPath()), "%2f")) {
		return errors.New("SNS certificate URL has an invalid path")
	}
	return nil
}

func validateSNSHost(rawURL, expectedHost string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if !strings.EqualFold(u.Hostname(), expectedHost) {
		return errors.New("SNS URL host does not match the topic region")
	}
	return nil
}

func snsHostForTopicARN(topicARN string) (string, error) {
	parts := strings.Split(topicARN, ":")
	if len(parts) < 6 || parts[0] != "arn" || parts[2] != "sns" || !snsRegionPattern.MatchString(parts[3]) {
		return "", errors.New("invalid SNS topic ARN")
	}
	suffix := "amazonaws.com"
	switch parts[1] {
	case "aws", "aws-us-gov":
	case "aws-cn":
		suffix = "amazonaws.com.cn"
	default:
		return "", errors.New("unsupported SNS topic ARN partition")
	}
	return "sns." + parts[3] + "." + suffix, nil
}

func validateSNSSubscribeURL(rawURL, topicARN, token string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	query := u.Query()
	if len(query) != 3 || len(query["Action"]) != 1 || len(query["TopicArn"]) != 1 || len(query["Token"]) != 1 ||
		query.Get("Action") != "ConfirmSubscription" || query.Get("TopicArn") != topicARN || query.Get("Token") != token {
		return errors.New("SNS SubscribeURL query does not match the signed message")
	}
	cleanURL := *u
	cleanURL.RawQuery = ""
	if err := validateSNSURL(cleanURL.String(), false); err != nil {
		return err
	}
	if u.EscapedPath() != "/" {
		return errors.New("SNS SubscribeURL has an invalid path")
	}
	return nil
}
