package service

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func (s *service) SignalCatalog(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("panic: %v", r)})
		}
	}()

	search := strings.TrimSpace(c.Query("search"))

	infos := catalog.InspectAll()
	sort.Slice(infos, func(i, j int) bool {
		return string(infos[i].Type) < string(infos[j].Type)
	})

	// Filter by search query.
	if search != "" {
		lower := strings.ToLower(search)
		var filtered []catalog.SignalTypeInfo
		for _, info := range infos {
			if strings.Contains(strings.ToLower(string(info.Type)), lower) ||
				strings.Contains(strings.ToLower(info.Operation), lower) ||
				strings.Contains(strings.ToLower(info.Namespace), lower) {
				filtered = append(filtered, info)
			}
		}
		infos = filtered
	}

	// Group by namespace.
	grouped := make(map[string][]catalog.SignalTypeInfo)
	var namespaces []string
	for _, info := range infos {
		if _, ok := grouped[info.Namespace]; !ok {
			namespaces = append(namespaces, info.Namespace)
		}
		grouped[info.Namespace] = append(grouped[info.Namespace], info)
	}
	sort.Strings(namespaces)

	c.JSON(http.StatusOK, gin.H{
		"grouped":    grouped,
		"namespaces": namespaces,
	})
}

func (s *service) SignalCatalogDetail(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("panic: %v", r)})
		}
	}()

	signalType := c.Param("signal_type")
	// URL-decode since signal types contain dashes but might be encoded.
	decoded, err := url.PathUnescape(signalType)
	if err == nil {
		signalType = decoded
	}

	info, err := catalog.InspectType(signal.SignalType(signalType))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Signal type not found: " + signalType})
		return
	}

	// Fetch recent signals of this type from the database.
	var recentSignals []app.QueueSignal
	res := s.db.WithContext(c.Request.Context()).
		Model(&app.QueueSignal{}).
		Where("type = ?", signalType).
		Order("created_at desc").
		Limit(50).
		Find(&recentSignals)
	if res.Error != nil {
		s.l.Error("failed to fetch recent signals", zap.String("type", signalType), zap.Error(res.Error))
	}

	c.JSON(http.StatusOK, gin.H{
		"info":           info,
		"recent_signals": recentSignals,
	})
}

func signalAttributesForType(signalType signal.SignalType) *catalog.SignalTypeInfo {
	info, err := catalog.InspectType(signalType)
	if err != nil {
		return nil
	}
	return &info
}
