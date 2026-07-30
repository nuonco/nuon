package service

import (
	"strings"
	"testing"
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
