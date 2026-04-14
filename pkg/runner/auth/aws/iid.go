package aws

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	mozpkcs7 "go.mozilla.org/pkcs7"
)

var accountIDPattern = regexp.MustCompile(`^\d{12}$`)

// InstanceIdentityDocument represents the JSON document returned by
// the EC2 IMDS at /latest/dynamic/instance-identity/document.
type InstanceIdentityDocument struct {
	AccountID        string    `json:"accountId"`
	Architecture     string    `json:"architecture"`
	AvailabilityZone string    `json:"availabilityZone"`
	ImageID          string    `json:"imageId"`
	InstanceID       string    `json:"instanceId"`
	InstanceType     string    `json:"instanceType"`
	PendingTime      time.Time `json:"pendingTime"`
	PrivateIP        string    `json:"privateIp"`
	Region           string    `json:"region"`
	Version          string    `json:"version"`
}

// ParseAndValidateIID parses a raw IID JSON document and validates
// that required fields are present and well-formed.
func ParseAndValidateIID(document string) (*InstanceIdentityDocument, error) {
	var iid InstanceIdentityDocument
	if err := json.Unmarshal([]byte(document), &iid); err != nil {
		return nil, fmt.Errorf("failed to parse IID document: %w", err)
	}

	if iid.Region == "" || iid.AccountID == "" || iid.InstanceID == "" {
		return nil, fmt.Errorf("identity document missing required fields")
	}

	if !accountIDPattern.MatchString(iid.AccountID) {
		return nil, fmt.Errorf("invalid AWS account ID format: %s", iid.AccountID)
	}

	return &iid, nil
}

// VerifyIIDSignature verifies the PKCS7 signature of an instance
// identity document using the AWS public certificate for the given
// region. The signature from IMDS /instance-identity/rsa2048 is a
// PKCS7/SMIME signed message.
func VerifyIIDSignature(certStore *IIDCertStore, region string, document []byte, signatureB64 string) error {
	cert, err := certStore.GetCert(region)
	if err != nil {
		return fmt.Errorf("no certificate for region %s: %w", region, err)
	}

	sigDER, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	p7, err := mozpkcs7.Parse(sigDER)
	if err != nil {
		return fmt.Errorf("failed to parse PKCS7 signature: %w", err)
	}

	p7.Certificates = []*x509.Certificate{cert}

	if err := p7.Verify(); err != nil {
		return fmt.Errorf("PKCS7 signature verification failed: %w", err)
	}

	return nil
}
