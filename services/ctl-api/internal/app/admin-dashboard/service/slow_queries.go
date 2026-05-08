package service

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AggregatedQuery groups all executions of the same SQL statement.
type AggregatedQuery struct {
	SQL         string  `json:"sql"`
	Table       string  `json:"table"`
	Operation   string  `json:"operation"`
	DBType      string  `json:"db_type"`
	Source      string  `json:"source"`
	Endpoint    string  `json:"endpoint"`
	Count       int     `json:"count"`
	TotalMS     float64 `json:"total_ms"`
	AvgMS       float64 `json:"avg_ms"`
	MinMS       float64 `json:"min_ms"`
	MaxMS       float64 `json:"max_ms"`
	TotalRows   int64   `json:"total_rows"`
	MaxRows     int64   `json:"max_rows"`
	MaxRespSize int     `json:"max_response_size"`
	LastError   string  `json:"last_error,omitempty"`
	Caller      string  `json:"caller"`
	CallerURL   string  `json:"caller_url,omitempty"`
	LastSeenAt  string  `json:"last_seen_at"`
}

func (s *service) Queries(c *gin.Context) {
	if s.queryCollector == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"queries": []any{},
			"tables":  []string{},
			"total":   0,
		})
		return
	}

	records := s.queryCollector.Records()
	search := strings.ToLower(c.Query("search"))
	table := c.Query("table")
	dbType := c.Query("db_type")
	source := c.Query("source")
	sortBy := c.DefaultQuery("sort", "max_ms")
	minDurationMS := c.Query("min_duration_ms")

	// Aggregate by SQL string.
	groups := make(map[string]*AggregatedQuery)
	var order []string

	for _, r := range records {
		if table != "" && r.Table != table {
			continue
		}
		if dbType != "" && r.DBType != dbType {
			continue
		}
		if source != "" && r.Source != source {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(r.SQL), search) &&
			!strings.Contains(strings.ToLower(r.Table), search) &&
			!strings.Contains(strings.ToLower(r.Caller), search) &&
			!strings.Contains(strings.ToLower(r.Endpoint), search) {
			continue
		}

		agg, ok := groups[r.SQL]
		if !ok {
			agg = &AggregatedQuery{
				SQL:       r.SQL,
				Table:     r.Table,
				Operation: r.Operation,
				DBType:    r.DBType,
				Source:    r.Source,
				Endpoint:  r.Endpoint,
				MinMS:     r.DurationMS,
				Caller:    r.Caller,
			}
			groups[r.SQL] = agg
			order = append(order, r.SQL)
		}

		agg.Count++
		agg.TotalMS += r.DurationMS
		if r.DurationMS < agg.MinMS {
			agg.MinMS = r.DurationMS
		}
		if r.DurationMS > agg.MaxMS {
			agg.MaxMS = r.DurationMS
		}
		agg.TotalRows += r.RowsAffected
		if r.RowsAffected > agg.MaxRows {
			agg.MaxRows = r.RowsAffected
		}
		if r.ResponseSize > agg.MaxRespSize {
			agg.MaxRespSize = r.ResponseSize
		}
		if r.Error != "" {
			agg.LastError = r.Error
		}
		if r.Caller != "" {
			agg.Caller = r.Caller
		}
		if r.CallerURL != "" {
			agg.CallerURL = r.CallerURL
		}
		if r.Endpoint != "" {
			agg.Endpoint = r.Endpoint
		}
		ts := r.Timestamp.Format("2006-01-02T15:04:05.000Z07:00")
		if ts > agg.LastSeenAt {
			agg.LastSeenAt = ts
		}
	}

	// Compute averages.
	result := make([]AggregatedQuery, 0, len(groups))
	for _, sql := range order {
		agg := groups[sql]
		agg.AvgMS = agg.TotalMS / float64(agg.Count)
		result = append(result, *agg)
	}

	// Filter by min duration (applied to max).
	if minDurationMS != "" {
		if minMS, err := strconv.ParseFloat(minDurationMS, 64); err == nil {
			filtered := result[:0]
			for _, q := range result {
				if q.MaxMS >= minMS {
					filtered = append(filtered, q)
				}
			}
			result = filtered
		}
	}

	// Sort.
	switch sortBy {
	case "max_ms":
		sort.Slice(result, func(i, j int) bool { return result[i].MaxMS > result[j].MaxMS })
	case "avg_ms":
		sort.Slice(result, func(i, j int) bool { return result[i].AvgMS > result[j].AvgMS })
	case "count":
		sort.Slice(result, func(i, j int) bool { return result[i].Count > result[j].Count })
	case "total_ms":
		sort.Slice(result, func(i, j int) bool { return result[i].TotalMS > result[j].TotalMS })
	case "last_seen":
		sort.Slice(result, func(i, j int) bool { return result[i].LastSeenAt > result[j].LastSeenAt })
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"queries": result,
		"tables":  s.queryCollector.Tables(),
		"total":   s.queryCollector.Total(),
	})
}

func (s *service) ClearQueries(c *gin.Context) {
	if s.queryCollector != nil {
		s.queryCollector.Clear()
	}
	c.JSON(http.StatusOK, gin.H{"cleared": true})
}
