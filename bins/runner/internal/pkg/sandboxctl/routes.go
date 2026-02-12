package sandboxctl

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed ui
var uiFS embed.FS

func (s *Server) registerRoutes(router *gin.Engine) {
	router.GET("/", s.handleUI)

	api := router.Group("/api/v1")
	api.GET("/health", s.handleHealth)
	api.GET("/state", s.handleGetState)
	api.PUT("/state", s.handleUpdateState)
	api.PUT("/state/variants/:category", s.handleSetVariant)
	api.PUT("/state/failure-mode", s.handleSetFailureMode)
	api.GET("/fixtures", s.handleListFixtures)
	api.GET("/fixtures/:category/:variant", s.handlePreviewFixture)
}

func (s *Server) handleUI(c *gin.Context) {
	data, err := uiFS.ReadFile("ui/index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to load UI")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"sandbox_mode": true,
	})
}

func (s *Server) handleGetState(c *gin.Context) {
	c.JSON(http.StatusOK, s.state.Snapshot())
}

func (s *Server) handleUpdateState(c *gin.Context) {
	var snap StateSnapshot
	if err := c.ShouldBindJSON(&snap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.state.Apply(snap)
	c.JSON(http.StatusOK, s.state.Snapshot())
}

func (s *Server) handleSetVariant(c *gin.Context) {
	cat := JobCategory(c.Param("category"))

	var body struct {
		Variant ResponseVariant `json:"variant"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate category
	valid := false
	for _, validCat := range AllCategories() {
		if cat == validCat {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}

	s.state.SetVariant(cat, body.Variant)
	c.JSON(http.StatusOK, s.state.Snapshot())
}

func (s *Server) handleSetFailureMode(c *gin.Context) {
	var body struct {
		FailureMode FailureMode `json:"failure_mode"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.state.SetFailureMode(body.FailureMode)
	c.JSON(http.StatusOK, s.state.Snapshot())
}

func (s *Server) handleListFixtures(c *gin.Context) {
	c.JSON(http.StatusOK, s.fixtures.AvailableVariants())
}

func (s *Server) handlePreviewFixture(c *gin.Context) {
	cat := JobCategory(c.Param("category"))
	variant := ResponseVariant(c.Param("variant"))

	f := s.fixtures.Get(cat, variant)
	if f == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fixture not found"})
		return
	}

	c.JSON(http.StatusOK, f)
}
