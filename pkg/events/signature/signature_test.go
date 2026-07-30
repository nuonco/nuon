package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"
)

func TestVerifyHMACAuthenticatesRawBody(t *testing.T) {
	payload := []byte(`{"ref":"main"}`)
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write(payload)
	signature := mac.Sum(nil)
	if !VerifyHMAC("secret", payload, signature, "sha256") {
		t.Fatal("raw body signature rejected")
	}
	if VerifyHMAC("wrong", payload, signature, "sha256") {
		t.Fatal("wrong secret accepted")
	}
}

func TestHMACVerifierMatchesSecretIndex(t *testing.T) {
	payload := []byte(`{"ref":"main"}`)
	mac := hmac.New(sha256.New, []byte("second"))
	mac.Write(payload)
	headers := http.Header{"X-Nuon-Signature": {"v1=" + hex.EncodeToString(mac.Sum(nil))}}
	verifier := HMAC{Header: "X-Nuon-Signature", Prefix: "v1=", Algorithm: "sha256", Encoding: "hex"}
	idx, err := verifier.Verify(headers, payload, []string{"first", "second"}, time.Now())
	if err != nil || idx != 1 {
		t.Fatalf("Verify = %d, %v; want 1, nil", idx, err)
	}
	if _, err := verifier.Verify(headers, payload, []string{"first"}, time.Now()); err == nil {
		t.Fatal("unmatched signature accepted")
	}
}

func TestValidSlackRequestTimestamp(t *testing.T) {
	now := time.Unix(1785254400, 0)
	if !ValidSlackRequestTimestamp("1785254400", now) || !ValidSlackRequestTimestamp("1785254100", now) || ValidSlackRequestTimestamp("1785254099", now) || ValidSlackRequestTimestamp("not-a-time", now) {
		t.Fatal("Slack timestamp window was not enforced")
	}
}
