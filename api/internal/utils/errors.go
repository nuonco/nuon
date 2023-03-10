package utils

import (
	"errors"

	"github.com/jackc/pgconn"
)

func IsDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)
	if !ok {
		return false
	}

	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	return pgErr.Code == "23505"
}
