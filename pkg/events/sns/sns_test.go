package sns

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCanonicalStrings(t *testing.T) {
	tests := []struct {
		msg  Message
		want string
	}{
		{Message{Type: "Notification", Message: "hello", MessageID: "id", Subject: "subject", Timestamp: "time", TopicARN: "arn"}, "Message\nhello\nMessageId\nid\nSubject\nsubject\nTimestamp\ntime\nTopicArn\narn\nType\nNotification\n"},
		{Message{Type: "SubscriptionConfirmation", Message: "confirm", MessageID: "id", SubscribeURL: "url", Timestamp: "time", Token: "token", TopicARN: "arn"}, "Message\nconfirm\nMessageId\nid\nSubscribeURL\nurl\nTimestamp\ntime\nToken\ntoken\nTopicArn\narn\nType\nSubscriptionConfirmation\n"},
	}
	for _, tt := range tests {
		got, err := tt.msg.canonicalString()
		if err != nil || got != tt.want {
			t.Fatalf("canonicalString() = %q, %v; want %q", got, err, tt.want)
		}
	}
}

func TestValidateURL(t *testing.T) {
	accepted := []string{
		"https://sns.us-east-1.amazonaws.com/SimpleNotificationService-deadBEEF.pem",
		"https://sns.cn-north-1.amazonaws.com.cn/SimpleNotificationService-123.pem",
		"https://sns.us-gov-west-1.amazonaws.com/SimpleNotificationService-ab.pem",
	}
	for _, raw := range accepted {
		if err := validateURL(raw, true); err != nil {
			t.Errorf("validateURL(%q): %v", raw, err)
		}
	}
	rejected := []string{
		"http://sns.us-east-1.amazonaws.com/SimpleNotificationService-ab.pem",
		"https://evil.example/SimpleNotificationService-ab.pem",
		"https://sns.us-east-1.amazonaws.com:443/SimpleNotificationService-ab.pem",
		"https://user@sns.us-east-1.amazonaws.com/SimpleNotificationService-ab.pem",
		"https://sns.us-east-1.amazonaws.com/cert.pem",
		"https://sns.not-a-region.amazonaws.com/SimpleNotificationService-ab.pem",
		"https://sns.us-east-1.amazonaws.com/SimpleNotificationService-ab.pem?q=1",
		"https://sns.us-east-1.amazonaws.com/SimpleNotificationService-ab.pem#x",
	}
	for _, raw := range rejected {
		if err := validateURL(raw, true); err == nil {
			t.Errorf("validateURL(%q) unexpectedly succeeded", raw)
		}
	}
}

func TestParseMessageValidatesConfirmationURL(t *testing.T) {
	body := `{"Type":"SubscriptionConfirmation","MessageId":"id","TopicArn":"arn","Message":"message","Timestamp":"time","SignatureVersion":"1","Signature":"AA==","SigningCertURL":"https://sns.us-east-1.amazonaws.com/SimpleNotificationService-ab.pem","Token":"token","SubscribeURL":"https://evil.example/confirm"}`
	if _, err := ParseMessage([]byte(body)); err == nil {
		t.Fatal("ParseMessage unexpectedly accepted untrusted SubscribeURL")
	}
	body = strings.Replace(body, "https://evil.example/confirm", "https://sns.us-east-1.amazonaws.com/?Action=ConfirmSubscription&Token=token&TopicArn=arn", 1)
	if _, err := ParseMessage([]byte(body)); err != nil {
		t.Fatalf("ParseMessage rejected valid message: %v", err)
	}
	body = strings.Replace(body, "TopicArn=arn", "TopicArn=other", 1)
	if _, err := ParseMessage([]byte(body)); err == nil {
		t.Fatal("ParseMessage accepted a SubscribeURL for another topic")
	}
}

func TestVerifierVersionsAndCache(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "sns.us-east-1.amazonaws.com"}, DNSNames: []string{"sns.us-east-1.amazonaws.com"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write(certPEM)
	}))
	defer server.Close()

	client := server.Client()
	client.Transport = rewriteHostTransport{base: client.Transport, target: server.URL}
	verifier := NewVerifier(client)
	expiredTemplate := *template
	expiredTemplate.SerialNumber = big.NewInt(2)
	expiredTemplate.NotBefore = time.Now().Add(-2 * time.Hour)
	expiredTemplate.NotAfter = time.Now().Add(-time.Hour)
	expiredDER, err := x509.CreateCertificate(rand.Reader, &expiredTemplate, &expiredTemplate, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	expiredCert, err := x509.ParseCertificate(expiredDER)
	if err != nil {
		t.Fatal(err)
	}
	certURL := "https://sns.us-east-1.amazonaws.com/SimpleNotificationService-deadbeef.pem"
	verifier.certs[certURL] = expiredCert
	for _, version := range []string{"1", "2"} {
		msg := &Message{Type: "Notification", Message: "hello", MessageID: "id", Timestamp: "time", TopicARN: "arn:aws:sns:us-east-1:123456789012:events", SignatureVersion: version, SigningCertURL: certURL}
		canonical, _ := msg.canonicalString()
		var hash crypto.Hash
		var digest []byte
		if version == "1" {
			sum := sha1.Sum([]byte(canonical))
			hash, digest = crypto.SHA1, sum[:]
		} else {
			sum := sha256.Sum256([]byte(canonical))
			hash, digest = crypto.SHA256, sum[:]
		}
		signature, err := rsa.SignPKCS1v15(rand.Reader, key, hash, digest)
		if err != nil {
			t.Fatal(err)
		}
		msg.Signature = base64.StdEncoding.EncodeToString(signature)
		if err := verifier.Verify(context.Background(), msg); err != nil {
			t.Fatalf("verify version %s: %v", version, err)
		}
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("certificate requests = %d; want 1", got)
	}
}

func TestVerifierRejectsCertificateHostnameAndTopicRegionMismatch(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	makeCert := func(host string) *x509.Certificate {
		t.Helper()
		template := &x509.Certificate{SerialNumber: big.NewInt(3), Subject: pkix.Name{CommonName: host}, DNSNames: []string{host}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature}
		der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
		if err != nil {
			t.Fatal(err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		return cert
	}

	certURL := "https://sns.us-east-1.amazonaws.com/SimpleNotificationService-deadbeef.pem"
	msg := &Message{Type: "Notification", Message: "hello", MessageID: "id", Timestamp: "time", TopicARN: "arn:aws:sns:us-east-1:123456789012:events", SignatureVersion: "2", SigningCertURL: certURL, Signature: "AA=="}
	verifier := NewVerifier(nil)
	verifier.certs[certURL] = makeCert("sns.us-west-2.amazonaws.com")
	if err := verifier.Verify(context.Background(), msg); err == nil || !strings.Contains(err.Error(), "hostname verification failed") {
		t.Fatalf("wrong certificate hostname error = %v", err)
	}

	verifier.certs[certURL] = makeCert("sns.us-east-1.amazonaws.com")
	msg.TopicARN = "arn:aws:sns:us-west-2:123456789012:events"
	if err := verifier.Verify(context.Background(), msg); err == nil || !strings.Contains(err.Error(), "does not match the topic region") {
		t.Fatalf("topic region mismatch error = %v", err)
	}
}

type rewriteHostTransport struct {
	base   http.RoundTripper
	target string
}

func (t rewriteHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	copy := req.Clone(req.Context())
	target := strings.TrimPrefix(t.target, "https://")
	copy.URL.Host = target
	copy.Host = target
	return t.base.RoundTrip(copy)
}
