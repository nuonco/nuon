package service

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type explainRequest struct {
	SQL    string `json:"sql" binding:"required"`
	DBType string `json:"db_type" binding:"required"`
}

// placeholderRe matches PostgreSQL positional parameters ($1, $2, …).
var placeholderRe = regexp.MustCompile(`\$\d+`)

func (s *service) ExplainQuery(c *gin.Context) {
	var req explainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sql and db_type are required"})
		return
	}

	sql := strings.TrimSpace(req.SQL)
	if sql == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sql must not be empty"})
		return
	}

	// Only allow SELECT statements to be explained.
	upper := strings.ToUpper(sql)
	if !strings.HasPrefix(upper, "SELECT") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only SELECT statements can be explained"})
		return
	}

	db := s.db
	explainPrefix := "EXPLAIN "
	if req.DBType == "ch" {
		db = s.chDB
	} else {
		// Replace $N placeholders with NULL so EXPLAIN can plan without values.
		sql = placeholderRe.ReplaceAllString(sql, "NULL")
	}

	explainSQL := fmt.Sprintf("%s%s", explainPrefix, sql)

	var rows []map[string]interface{}
	if res := db.WithContext(c).Raw(explainSQL).Scan(&rows); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rows": rows})
}
