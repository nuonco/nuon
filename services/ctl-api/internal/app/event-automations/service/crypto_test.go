package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestCredentialGenerationAndHash(t *testing.T) {
	first, err := generateCredential()
	if err != nil {
		t.Fatal(err)
	}
	second, err := generateCredential()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 43 || first == second || strings.Contains(first, "=") {
		t.Fatalf("unexpected credentials %q and %q", first, second)
	}
	if hashIngressKey(first) == first || len(hashIngressKey(first)) != 64 {
		t.Fatal("ingress key was not SHA-256 hex hashed")
	}
}

func TestSignatureParsingAndVerification(t *testing.T) {
	body, timestamp, secret := []byte("{\"ok\":true}"), "1720000000", "secret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)
	signature, err := parseSignature("v1=" + hex.EncodeToString(mac.Sum(nil)))
	if err != nil || !verifySignature(secret, timestamp, body, signature) {
		t.Fatal("valid signature rejected")
	}
	for _, malformed := range []string{"", "sha256=" + strings.Repeat("0", 64), "v1=XYZ", "v1=" + strings.Repeat("A", 64)} {
		if _, err := parseSignature(malformed); err == nil {
			t.Fatalf("accepted malformed signature %q", malformed)
		}
	}
}

func TestTimestampSkew(t *testing.T) {
	now := time.Unix(1720000000, 0)
	if _, err := parseTimestamp("1720000000", now); err != nil {
		t.Fatal(err)
	}
	if _, err := parseTimestamp("1719999699", now); err == nil {
		t.Fatal("accepted stale timestamp")
	}
}

func TestGenericHMACPayloadAuthenticatesEventIdentity(t *testing.T) {
	payload, err := hmacPayload(app.EventEnvelopeTypeNone, "delivery-1", "push", []byte(`{"ref":"main"}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != "delivery-1\npush\n{\"ref\":\"main\"}" {
		t.Fatalf("unexpected signed payload %q", payload)
	}
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte("1720000000."))
	mac.Write(payload)
	signature := mac.Sum(nil)
	alteredPayload, err := hmacPayload(app.EventEnvelopeTypeNone, "delivery-1", "delete", []byte(`{"ref":"main"}`))
	if err != nil {
		t.Fatal(err)
	}
	if verifySignature("secret", "1720000000", alteredPayload, signature) {
		t.Fatal("accepted a signature after event type was altered")
	}
	if _, err := hmacPayload(app.EventEnvelopeTypeNone, "", "push", []byte(`{}`)); err == nil {
		t.Fatal("accepted HMAC generic event without an ID")
	}
}
