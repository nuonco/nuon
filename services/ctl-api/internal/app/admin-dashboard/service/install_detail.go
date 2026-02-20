package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) InstallDetail(c *gin.Context) {
	install, err := s.getInstall(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "install not found"})
			return
		}
		s.l.Error("failed to fetch install", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch install"})
		return
	}

	activeDeployments, err := s.getActiveDeployments(c, install.ID)
	if err != nil {
		s.l.Error("failed to fetch active deployments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch active deployments"})
		return
	}

	component := views.InstallDetail(install, activeDeployments)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// InstallActiveDeploymentsTable handles the polling endpoint for active deployments
func (s *service) InstallActiveDeploymentsTable(c *gin.Context) {
	installID := c.Param("id")
	if installID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "install ID required"})
		return
	}

	deployments, err := s.getActiveDeployments(c, installID)
	if err != nil {
		s.l.Error("failed to fetch active deployments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch deployments"})
		return
	}

	component := views.InstallActiveDeploymentsTable(installID, deployments)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// getInstall fetches an install by ID with necessary preloads
func (s *service) getInstall(c *gin.Context) (*app.Install, error) {
	installID := c.Param("id")
	if installID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var install app.Install
	err := s.db.
		Preload("Org").
		Preload("App").
		Preload("AppConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("AppRunnerConfig").
		Where("id = ?", installID).
		First(&install).Error

	if err != nil {
		return nil, err
	}

	return &install, nil
}

// getActiveDeployments fetches active deployments for an install
func (s *service) getActiveDeployments(c *gin.Context, installID string) ([]app.InstallDeploy, error) {
	activeStatuses := []app.InstallDeployStatus{
		app.InstallDeployStatusPlanning,
		app.InstallDeployStatusSyncing,
		app.InstallDeployStatusExecuting,
		app.InstallDeployStatusQueued,
		app.InstallDeployStatusPending,
		app.InstallDeployStatusPendingApproval,
	}

	var deployments []app.InstallDeploy
	err := s.db.
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where("install_components.install_id = ?", installID).
		Where("install_deploys.status IN ?", activeStatuses).
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Order("install_deploys.created_at DESC").
		Find(&deployments).Error

	if err != nil {
		return nil, err
	}

	for i := range deployments {
		if deployments[i].InstallComponent.Component.Name != "" {
			deployments[i].ComponentName = deployments[i].InstallComponent.Component.Name
		}
	}

	return deployments, nil
}
