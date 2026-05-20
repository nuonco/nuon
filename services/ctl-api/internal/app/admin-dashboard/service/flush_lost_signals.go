package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *service) FlushLostSignals(c *gin.Context) {
	cutoff := time.Now().Add(-1 * time.Hour)
	now := time.Now().Unix()

	// Mark unenqueued signals older than 1 hour as error and soft-delete them
	// so they stop appearing in sweep queries.
	res := s.db.WithContext(c.Request.Context()).Exec(`
		UPDATE queue_signals
		SET
			status = jsonb_set(
				jsonb_set(status, '{status}', '"error"'),
				'{status_human_description}',
				'"marked as lost by admin: unenqueued for over 1 hour"'
			),
			deleted_at = ?
		WHERE deleted_at = 0
			AND enqueued = false
			AND created_at < ?
	`, now, cutoff)

	if res.Error != nil {
		s.l.Error("failed to flush lost signals", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to flush lost signals"})
		return
	}

	s.l.Info("flushed lost signals",
		zap.Int64("count", res.RowsAffected),
		zap.Time("cutoff", cutoff))

	c.JSON(http.StatusOK, gin.H{
		"status":  "flushed",
		"flushed": res.RowsAffected,
	})
}
