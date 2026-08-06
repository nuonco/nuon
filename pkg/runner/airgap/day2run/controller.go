package day2run

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
)

type ControllerConfig struct {
	Mailbox      *Mailbox
	Envelope     *airgap.Envelope
	Digest       string
	DeploymentID string
	Owner        string
	Executor     RefExecutor
	Logger       *zap.Logger
	FlushRun     func(context.Context, string) error
	WriteLocal   func(string, []byte) error
	PollInterval time.Duration
}

type Controller struct {
	cfg        ControllerConfig
	dispatcher *dispatcher
}

func NewController(cfg ControllerConfig) (*Controller, error) {
	if cfg.Mailbox == nil || cfg.Envelope == nil || cfg.Executor == nil {
		return nil, fmt.Errorf("mailbox, envelope, and executor are required")
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaultPollInterval
	}
	d := &dispatcher{mailbox: cfg.Mailbox, envelope: cfg.Envelope, digest: cfg.Digest, deploymentID: cfg.DeploymentID, owner: cfg.Owner, executor: cfg.Executor, flushRun: cfg.FlushRun, logger: cfg.Logger, now: time.Now}
	return &Controller{cfg: cfg, dispatcher: d}, nil
}

func (c *Controller) Run(ctx context.Context) error {
	if err := c.publishCatalog(ctx); err != nil {
		return err
	}
	poller := &Poller{dispatcher: c.dispatcher, interval: c.cfg.PollInterval}
	scheduler := &Scheduler{dispatcher: c.dispatcher, actions: c.cfg.Envelope.Actions, writeLocal: c.cfg.WriteLocal}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); poller.Run(ctx) }()
	go func() { defer wg.Done(); scheduler.Run(ctx) }()
	wg.Wait()
	return ctx.Err()
}

func (c *Controller) publishCatalog(ctx context.Context) error {
	catalog := day2.Catalog{SchemaVersion: day2.SchemaVersion, DeploymentID: c.cfg.DeploymentID, BundleDigest: c.cfg.Digest, GeneratedAt: time.Now().UTC()}
	for _, action := range c.cfg.Envelope.Actions {
		catalog.Refs = append(catalog.Refs, day2.CatalogRef{ID: action.ID, Kind: day2.RefKindAction, Name: action.Name, CronSchedule: action.CronSchedule})
	}
	for _, drift := range c.cfg.Envelope.Drift {
		catalog.Refs = append(catalog.Refs, day2.CatalogRef{ID: drift.ID, Kind: day2.RefKindDrift, Name: drift.ComponentName, Component: drift.ComponentName})
	}
	for _, book := range c.cfg.Envelope.Runbooks {
		catalog.Refs = append(catalog.Refs, day2.CatalogRef{ID: book.ID, Kind: day2.RefKindRunbook, Name: book.Name, Steps: len(book.Steps)})
	}
	if c.cfg.WriteLocal != nil {
		b, err := json.MarshalIndent(catalog, "", "  ")
		if err != nil {
			return err
		}
		if err := c.cfg.WriteLocal(day2.CatalogKey, append(b, '\n')); err != nil {
			return fmt.Errorf("write local day-2 catalog: %w", err)
		}
	}
	if err := c.cfg.Mailbox.PutCatalog(ctx, catalog); err != nil {
		return fmt.Errorf("publish day-2 catalog: %w", err)
	}
	return nil
}
