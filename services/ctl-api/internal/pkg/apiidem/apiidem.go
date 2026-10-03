package apiidem

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func ValidateRequestID(id string) error {
	if id == "" {
		return nil
	}
	if utf8.RuneCountInString(id) > 255 {
		return stderr.ErrUser{
			Err:         fmt.Errorf("request_id must be at most 255 characters"),
			Description: "invalid request_id",
		}
	}
	return nil
}

func Hash(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("hash request: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func DedupeKey(operation, requestID string) string {
	return "api-request:" + operation + ":" + requestID
}

func Check(idem *app.WorkflowIdempotency, requestHash, currentAppConfigID string) error {
	if idem == nil || idem.RequestHash != requestHash {
		return stderr.ErrConflict{
			Err:         fmt.Errorf("request_id was already used with a different request"),
			Description: "request_id was already used with a different request body",
		}
	}
	if idem.PinnedAppConfigID != currentAppConfigID {
		return stderr.ErrConflict{
			Err:         fmt.Errorf("install app config moved from %s to %s", idem.PinnedAppConfigID, currentAppConfigID),
			Description: fmt.Sprintf("install app config moved from %s to %s", idem.PinnedAppConfigID, currentAppConfigID),
		}
	}
	return nil
}

func IsDuplicateKey(err error) bool {
	for i := 0; err != nil && i < 20; i++ {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return true
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return true
		}
		if u := errors.Unwrap(err); u != nil {
			err = u
			continue
		}
		type causer interface{ Cause() error }
		if c, ok := err.(causer); ok && c.Cause() != nil {
			err = c.Cause()
			continue
		}
		return false
	}
	return false
}
