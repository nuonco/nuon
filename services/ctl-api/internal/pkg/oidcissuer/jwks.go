package oidcissuer

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
)

// JWK is a public RSA JSON web key as published in the JWKS endpoint.
type JWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

// JWKS is a JSON web key set of public keys.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

type privateJWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
	D         string `json:"d"`
	P         string `json:"p"`
	Q         string `json:"q"`
	DP        string `json:"dp"`
	DQ        string `json:"dq"`
	QI        string `json:"qi"`
}

type privateJWKSet struct {
	Keys []privateJWK `json:"keys"`
}

func ValidateIssuer(issuer string) error {
	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("issuer must be an absolute HTTP or HTTPS URL without userinfo, query, or fragment")
	}
	return nil
}

// ParseJWKS validates a JWKS document and returns its private signing key, the
// signing key ID, and the public keys safe to publish.
func ParseJWKS(value string) (*rsa.PrivateKey, string, JWKS, error) {
	var input privateJWKSet
	if err := json.Unmarshal([]byte(value), &input); err != nil {
		return nil, "", JWKS{}, fmt.Errorf("decode JWKS: %w", err)
	}
	if len(input.Keys) == 0 {
		return nil, "", JWKS{}, fmt.Errorf("JWKS contains no keys")
	}

	publicKeys := JWKS{Keys: make([]JWK, 0, len(input.Keys))}
	seenKeyIDs := make(map[string]struct{}, len(input.Keys))
	var signingKey *rsa.PrivateKey
	var signingKeyID string
	for _, key := range input.Keys {
		if key.KeyID == "" {
			return nil, "", JWKS{}, fmt.Errorf("JWK key ID is required")
		}
		if _, exists := seenKeyIDs[key.KeyID]; exists {
			return nil, "", JWKS{}, fmt.Errorf("JWKS contains duplicate key IDs")
		}
		seenKeyIDs[key.KeyID] = struct{}{}

		if key.KeyType != "RSA" || (key.Use != "" && key.Use != "sig") || (key.Algorithm != "" && key.Algorithm != jwt.SigningMethodRS256.Alg()) {
			return nil, "", JWKS{}, fmt.Errorf("JWKS supports only RSA signing keys using RS256")
		}

		publicKey, err := parseRSAPublicKey(key)
		if err != nil {
			return nil, "", JWKS{}, fmt.Errorf("parse JWK %q: %w", key.KeyID, err)
		}
		publicKeys.Keys = append(publicKeys.Keys, JWK{
			KeyType:   "RSA",
			KeyID:     key.KeyID,
			Use:       "sig",
			Algorithm: jwt.SigningMethodRS256.Alg(),
			Modulus:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
			Exponent:  encodeJWKInteger(int64(publicKey.E)),
		})

		if key.D == "" && key.P == "" && key.Q == "" && key.DP == "" && key.DQ == "" && key.QI == "" {
			continue
		}
		if signingKey != nil {
			return nil, "", JWKS{}, fmt.Errorf("JWKS must contain exactly one private signing key")
		}

		signingKey, err = parseRSAPrivateKey(key, publicKey)
		if err != nil {
			return nil, "", JWKS{}, fmt.Errorf("parse private JWK %q: %w", key.KeyID, err)
		}
		signingKeyID = key.KeyID
	}

	if signingKey == nil {
		return nil, "", JWKS{}, fmt.Errorf("JWKS does not contain a private signing key")
	}
	return signingKey, signingKeyID, publicKeys, nil
}

func parseRSAPublicKey(key privateJWK) (*rsa.PublicKey, error) {
	modulus, err := decodeJWKInteger(key.Modulus)
	if err != nil || modulus.Sign() <= 0 {
		return nil, fmt.Errorf("invalid RSA modulus")
	}
	if modulus.BitLen() < 2048 {
		return nil, fmt.Errorf("RSA modulus must be at least 2048 bits")
	}
	exponent, err := decodeJWKInteger(key.Exponent)
	if err != nil || !exponent.IsInt64() || exponent.Int64() < 3 || exponent.Int64() > int64(^uint(0)>>1) || exponent.Int64()%2 == 0 {
		return nil, fmt.Errorf("invalid RSA exponent")
	}
	return &rsa.PublicKey{N: modulus, E: int(exponent.Int64())}, nil
}

func parseRSAPrivateKey(key privateJWK, publicKey *rsa.PublicKey) (*rsa.PrivateKey, error) {
	if key.D == "" || key.P == "" || key.Q == "" {
		return nil, fmt.Errorf("RSA private key requires d, p, and q")
	}
	d, err := decodeJWKInteger(key.D)
	if err != nil {
		return nil, fmt.Errorf("invalid RSA private exponent")
	}
	p, err := decodeJWKInteger(key.P)
	if err != nil {
		return nil, fmt.Errorf("invalid first RSA prime")
	}
	q, err := decodeJWKInteger(key.Q)
	if err != nil {
		return nil, fmt.Errorf("invalid second RSA prime")
	}

	privateKey := &rsa.PrivateKey{
		PublicKey: *publicKey,
		D:         d,
		Primes:    []*big.Int{p, q},
	}
	if err := privateKey.Validate(); err != nil {
		return nil, fmt.Errorf("validate RSA private key: %w", err)
	}
	privateKey.Precompute()
	return privateKey, nil
}

func decodeJWKInteger(value string) (*big.Int, error) {
	if value == "" {
		return nil, errors.New("JWK integer is empty")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) == 0 {
		return nil, errors.New("JWK integer is not valid base64url")
	}
	return new(big.Int).SetBytes(decoded), nil
}

func encodeJWKInteger(value int64) string {
	return base64.RawURLEncoding.EncodeToString(big.NewInt(value).Bytes())
}
