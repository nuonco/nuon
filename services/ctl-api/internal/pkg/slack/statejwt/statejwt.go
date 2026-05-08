// Package statejwt encodes and decodes the OAuth `state` parameter used in
// the Slack install / link flows.
//
// The state parameter binds the user's dashboard session (account_id, org_id)
// to the redirect that will land at Slack's OAuth callback, so that when the
// callback fires we know which Nuon account+org initiated the install. It is
// signed with HS256 using the SlackStateJWTSecret config value and carries a
// short TTL.
package statejwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// DefaultTTL bounds how long an issued state value is accepted at the
// callback. OAuth round-trips through Slack are interactive and complete in
// seconds; 10 minutes is a generous ceiling that tolerates redirects and slow
// approval flows.
const DefaultTTL = 10 * time.Minute

// Claims is the payload encoded into the OAuth state parameter.
type Claims struct {
	AccountID string `json:"acc"`
	OrgID     string `json:"org"`
	// Nonce is a per-issuance random string so two state values issued in
	// the same second still differ; the caller supplies it.
	Nonce string `json:"nonce"`
	jwt.RegisteredClaims
}

// Encoder issues and verifies signed state values. It holds the signing
// secret so callers don't have to thread it through every call site.
type Encoder struct {
	secret []byte
	ttl    time.Duration
}

// New constructs an Encoder. Returns an error if secret is empty.
func New(secret string) (*Encoder, error) {
	if secret == "" {
		return nil, errors.New("statejwt: secret must not be empty")
	}
	return &Encoder{
		secret: []byte(secret),
		ttl:    DefaultTTL,
	}, nil
}

// Issue produces a signed state value bound to the given account + org +
// nonce. The TTL is enforced at decode time by jwt's standard exp claim.
func (e *Encoder) Issue(accountID, orgID, nonce string) (string, error) {
	now := time.Now()
	claims := Claims{
		AccountID: accountID,
		OrgID:     orgID,
		Nonce:     nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(e.ttl)),
			Issuer:    "nuon-slack",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(e.secret)
	if err != nil {
		return "", fmt.Errorf("statejwt: sign: %w", err)
	}
	return signed, nil
}

// Decode verifies signature + expiry and returns the embedded claims.
func (e *Encoder) Decode(state string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(state, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("statejwt: unexpected signing method %v", t.Header["alg"])
		}
		return e.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("statejwt: decode: %w", err)
	}
	return claims, nil
}
