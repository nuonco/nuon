package statejwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const DefaultTTL = 10 * time.Minute

type Claims struct {
	AccountID string `json:"acc"`
	OrgID     string `json:"org"`
	Nonce     string `json:"nonce"`
	jwt.RegisteredClaims
}

type Encoder struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string) (*Encoder, error) {
	if secret == "" {
		return nil, errors.New("statejwt: secret must not be empty")
	}
	return &Encoder{
		secret: []byte(secret),
		ttl:    DefaultTTL,
	}, nil
}

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
