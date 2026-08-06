package day2run

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
)

type Scheduler struct {
	dispatcher *dispatcher
	actions    []airgap.ActionTemplate
	writeLocal func(string, []byte) error
	mu         sync.Mutex
	cursors    map[string]day2.ScheduleCursor
}

func (s *Scheduler) Run(ctx context.Context) {
	for _, action := range s.actions {
		if action.CronSchedule == "" {
			continue
		}
		go s.runAction(ctx, action)
	}
	<-ctx.Done()
}

func (s *Scheduler) runAction(ctx context.Context, action airgap.ActionTemplate) {
	for {
		next, err := day2.NextAfter(action.CronSchedule, s.dispatcher.now())
		if err != nil {
			s.dispatcher.logger.Error("compute day-2 schedule", zap.String("action_id", action.ID), zap.Error(err))
			return
		}
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			s.tick(ctx, action, next)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, action airgap.ActionTemplate, tick time.Time) {
	now := s.dispatcher.now().UTC()
	s.mu.Lock()
	if s.cursors == nil {
		s.cursors = map[string]day2.ScheduleCursor{}
	}
	cursor := s.cursors[action.ID]
	cursor.ScheduleID, cursor.UpdatedAt = action.ID, now
	if s.dispatcher.executor.Busy() {
		cursor.Skipped++
		cursor.LastSkippedAt = &tick
		s.cursors[action.ID] = cursor
		s.mu.Unlock()
		s.writeCursor(action.ID, cursor)
		return
	}
	s.mu.Unlock()
	id := day2.OccurrenceID(s.dispatcher.deploymentID, s.dispatcher.digest, action.ID, tick)
	req := day2.Request{SchemaVersion: day2.SchemaVersion, DeploymentID: s.dispatcher.deploymentID, BundleDigest: s.dispatcher.digest, RefID: action.ID, DispatchID: id, Source: day2.SourceCron, ScheduledAt: &tick, CreatedAt: now}
	if err := s.dispatcher.mailbox.PutRequest(ctx, req); err != nil {
		s.dispatcher.logger.Warn("write scheduled day-2 request", zap.Error(err))
		return
	}
	if err := s.dispatcher.handle(ctx, id, req); err != nil {
		s.dispatcher.logger.Warn("execute scheduled day-2 request", zap.Error(err))
	}
	cursor.LastFiredAt = &tick
	s.mu.Lock()
	s.cursors[action.ID] = cursor
	s.mu.Unlock()
	s.writeCursor(action.ID, cursor)
}

func (s *Scheduler) writeCursor(id string, cursor day2.ScheduleCursor) {
	if s.writeLocal == nil {
		return
	}
	b, err := json.MarshalIndent(cursor, "", "  ")
	if err == nil {
		err = s.writeLocal(day2.ScheduleCursorKey(id), append(b, '\n'))
	}
	if err != nil {
		s.dispatcher.logger.Warn("write schedule cursor", zap.String("action_id", id), zap.Error(err))
	}
}
