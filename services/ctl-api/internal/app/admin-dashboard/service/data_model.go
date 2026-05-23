package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Caps to keep payloads bounded — the data-model page is a structural overview,
// not a complete log.
const (
	dataModelInstallsLimit       = 100
	dataModelRunnersLimit        = 100
	dataModelComponentsLimit     = 200
	dataModelWorkflowsPerInstall = 1  // latest non-terminal (or latest period) per install
	dataModelSignalsPerQueue     = 20 // most recent N signals per queue
	dataModelEmittersPerQueue    = 20
)

// DataModel returns a single aggregate payload describing the product +
// queue/signal data model for one org, used by the admin dashboard's
// interactive data-model diagram.
//
// GET /api/data-model?org_id=<id>
func (s *service) DataModel(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Query("org_id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "org_id query param is required"})
		return
	}

	var (
		org        *app.Org
		apps       []app.App
		components []app.Component
		installs   []app.Install
		runners    []app.Runner
		workflows  []app.Workflow
		steps      []app.WorkflowStep
		queues     []app.Queue
		emitters   []app.QueueEmitter
		signals    []app.QueueSignal
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var o app.Org
		if err := s.readDB().WithContext(gCtx).Where("id = ?", orgID).First(&o).Error; err != nil {
			return fmt.Errorf("fetch org: %w", err)
		}
		org = &o
		return nil
	})

	g.Go(func() error {
		return s.readDB().WithContext(gCtx).
			Where("org_id = ?", orgID).
			Order("created_at ASC").
			Find(&apps).Error
	})

	g.Go(func() error {
		return s.readDB().WithContext(gCtx).
			Where("org_id = ?", orgID).
			Order("created_at ASC").
			Limit(dataModelComponentsLimit).
			Find(&components).Error
	})

	g.Go(func() error {
		return s.readDB().WithContext(gCtx).
			Where("org_id = ?", orgID).
			Order("created_at DESC").
			Limit(dataModelInstallsLimit).
			Find(&installs).Error
	})

	g.Go(func() error {
		return s.readDB().WithContext(gCtx).
			Where("org_id = ?", orgID).
			Order("created_at ASC").
			Limit(dataModelRunnersLimit).
			Find(&runners).Error
	})

	g.Go(func() error {
		return s.readDB().WithContext(gCtx).
			Where("org_id = ?", orgID).
			Order("created_at DESC").
			Find(&queues).Error
	})

	if err := g.Wait(); err != nil {
		s.l.Error("data model: failed to fetch base entities",
			zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch data model"})
		return
	}

	// Second-stage fetches depend on installs / queues for IN clauses.
	installIDs := make([]string, 0, len(installs))
	for _, i := range installs {
		installIDs = append(installIDs, i.ID)
	}
	queueIDs := make([]string, 0, len(queues))
	for _, q := range queues {
		queueIDs = append(queueIDs, q.ID)
	}

	g2, g2Ctx := errgroup.WithContext(ctx)

	if len(installIDs) > 0 {
		g2.Go(func() error {
			return s.latestWorkflowsForInstalls(g2Ctx, installIDs, &workflows)
		})
	}

	if len(queueIDs) > 0 {
		g2.Go(func() error {
			return s.readDB().WithContext(g2Ctx).
				Where("queue_id IN ?", queueIDs).
				Order("created_at DESC").
				Limit(dataModelEmittersPerQueue * len(queueIDs)).
				Find(&emitters).Error
		})

		g2.Go(func() error {
			return s.recentSignalsForQueues(g2Ctx, queueIDs, &signals)
		})
	}

	if err := g2.Wait(); err != nil {
		s.l.Error("data model: failed to fetch dependent entities",
			zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch data model"})
		return
	}

	// Fetch steps for the workflows we ended up keeping.
	if len(workflows) > 0 {
		workflowIDs := make([]string, 0, len(workflows))
		for _, w := range workflows {
			workflowIDs = append(workflowIDs, w.ID)
		}
		if err := s.readDB().WithContext(ctx).
			Where("owner_id IN ? AND owner_type = ?", workflowIDs, (&app.Workflow{}).TableName()).
			Order("idx ASC").
			Find(&steps).Error; err != nil {
			s.l.Warn("data model: failed to fetch steps", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"org":        org,
		"apps":       apps,
		"components": components,
		"installs":   installs,
		"runners":    runners,
		"workflows":  workflows,
		"steps":      steps,
		"queues":     queues,
		"emitters":   emitters,
		"signals":    signals,
	})
}

// latestWorkflowsForInstalls picks the most recent N workflows per install
// (currently N=1) using a DISTINCT ON query.
func (s *service) latestWorkflowsForInstalls(ctx context.Context, installIDs []string, out *[]app.Workflow) error {
	// Postgres DISTINCT ON keeps the first row per (owner_id) per the ORDER BY.
	return s.readDB().WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (owner_id) *
			FROM install_workflows
			WHERE owner_type = ? AND owner_id IN ? AND deleted_at = 0
			ORDER BY owner_id, created_at DESC
		`, "installs", installIDs).
		Scan(out).Error
}

// recentSignalsForQueues fetches the most recent N signals per queue using a
// window function so one busy queue can't drown the others.
func (s *service) recentSignalsForQueues(ctx context.Context, queueIDs []string, out *[]app.QueueSignal) error {
	return s.readDB().WithContext(ctx).
		Raw(`
			SELECT * FROM (
				SELECT *, ROW_NUMBER() OVER (PARTITION BY queue_id ORDER BY created_at DESC) AS rn
				FROM queue_signals
				WHERE queue_id IN ? AND deleted_at = 0
			) ranked
			WHERE rn <= ?
		`, queueIDs, dataModelSignalsPerQueue).
		Scan(out).Error
}
