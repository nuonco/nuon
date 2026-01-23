package service

import "errors"

var (
	errNoURL               = errors.New("no destination URL requested")
	errInvalidURL          = errors.New("requested destination URL appears to be invalid")
	errURLNotHTTP          = errors.New("requested destination URL must begin with http:// or https://")
	errDangerousQS         = errors.New("requested destination URL has a dangerous query string")
	errURLDomainNotAllowed = errors.New("requested destination URL domain not allowed")
	errTooManyRedirects    = errors.New("too many unsuccessful authorization attempts")
	errInvalidState        = errors.New("invalid state parameter")
	errSessionNotFound     = errors.New("session not found")
	errStateMismatch       = errors.New("state parameter mismatch")
	errNoAuthCode          = errors.New("no authorization code in request")
)
