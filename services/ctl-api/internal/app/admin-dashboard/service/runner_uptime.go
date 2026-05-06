package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// InstallUptimeEntry groups all runner data for a single install.
type InstallUptimeEntry struct {
	InstallID   string `json:"install_id"`
	InstallName string `json:"install_name"`
	OrgID       string `json:"org_id"`
	OrgName     string `json:"org_name"`

	RunnerCreatedAt string `json:"runner_created_at,omitempty"`

	InstallProcesses []ProcessUptime `json:"install_processes"`
	MngProcesses     []ProcessUptime `json:"mng_processes"`

	InstallMetrics UptimeMetrics `json:"install_metrics"`
	MngMetrics     UptimeMetrics `json:"mng_metrics"`

	Jobs JobSummary `json:"jobs"`
}

type ProcessUptime struct {
	ProcessID       string  `json:"process_id"`
	RunnerID        string  `json:"runner_id"`
	Type            string  `json:"type"`
	Status          string  `json:"status"`
	Version         string  `json:"version"`
	StartedAt       string  `json:"started_at,omitempty"`
	LastHeartbeat   string  `json:"last_heartbeat,omitempty"`
	UptimeMS        float64 `json:"uptime_ms"`
	UptimeStr       string  `json:"uptime_str"`
	Heartbeats      int64   `json:"heartbeats"`
	HealthChecks    int64   `json:"health_checks"`
	HealthyChecks   int64   `json:"healthy_checks"`
	UnhealthyChecks int64   `json:"unhealthy_checks"`
}

type JobSummary struct {
	Total     int64 `json:"total"`
	Finished  int64 `json:"finished"`
	Failed    int64 `json:"failed"`
	TimedOut  int64 `json:"timed_out"`
	Cancelled int64 `json:"cancelled"`
	Other     int64 `json:"other"`
}

// UptimeMetrics holds aggregate numbers for the pie charts.
type UptimeMetrics struct {
	// Effective window (adjusted if runner created after window start).
	EffectiveWindowMS float64 `json:"effective_window_ms"`
	TotalUptimeMS     float64 `json:"total_uptime_ms"`
	TotalProcs        int     `json:"total_procs"`
	Restarts          int     `json:"restarts"`

	TotalHeartbeats      int64 `json:"total_heartbeats"`
	ExpectedHeartbeats   int64 `json:"expected_heartbeats"`
	TotalHealthChecks    int64 `json:"total_health_checks"`
	HealthyChecks        int64 `json:"healthy_checks"`
	UnhealthyChecks      int64 `json:"unhealthy_checks"`
	ExpectedHealthChecks int64 `json:"expected_health_checks"`
}

func (s *service) RunnerUptime(c *gin.Context) {
	ctx := c.Request.Context()

	orgID := c.Query("org_id")
	installName := c.Query("install_name")
	label := c.Query("label")
	window := c.DefaultQuery("window", "today")

	since := windowStart(window)
	now := time.Now()
	windowMS := float64(now.Sub(since).Milliseconds())

	// Fetch installs.
	type installRow struct {
		InstallID   string
		InstallName string
		OrgID       string
		OrgName     string
	}

	installQuery := s.db.WithContext(ctx).
		Table("installs i").
		Select("i.id as install_id, i.name as install_name, i.org_id, o.name as org_name").
		Joins("JOIN orgs o ON o.id = i.org_id AND o.deleted_at = 0").
		Where("i.deleted_at = 0")

	if orgID != "" {
		installQuery = installQuery.Where("i.org_id = ?", orgID)
	}
	if installName != "" {
		installQuery = installQuery.Where("i.name ILIKE ?", "%"+installName+"%")
	}
	if label != "" {
		if parts := strings.SplitN(label, ":", 2); len(parts) == 2 {
			installQuery = installQuery.Where("i.labels->>? = ?", parts[0], parts[1])
		} else {
			installQuery = installQuery.Where("i.labels::text ILIKE ?", "%"+label+"%")
		}
	}

	var installs []installRow
	if err := installQuery.Order("o.name, i.name").Find(&installs).Error; err != nil {
		s.l.Error("runner uptime: failed to fetch installs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch installs"})
		return
	}

	if len(installs) == 0 {
		c.JSON(http.StatusOK, gin.H{"installs": []any{}, "window": window, "since": since.Format(time.RFC3339), "window_ms": windowMS})
		return
	}

	installIDs := make([]string, len(installs))
	for i, inst := range installs {
		installIDs[i] = inst.InstallID
	}

	// Fetch runners for these installs.
	type runnerMapping struct {
		RunnerID  string
		CreatedAt time.Time
		OwnerID   string
	}
	var runners []runnerMapping
	s.db.WithContext(ctx).
		Table("runners r").
		Select("r.id as runner_id, r.created_at, rg.owner_id").
		Joins("JOIN runner_groups rg ON rg.id = r.runner_group_id AND rg.deleted_at = 0").
		Where("r.deleted_at = 0").
		Where("rg.owner_type = 'installs'").
		Where("rg.owner_id IN ?", installIDs).
		Find(&runners)

	installRunnerMap := make(map[string][]string)
	runnerCreatedAt := make(map[string]time.Time)
	for _, rm := range runners {
		installRunnerMap[rm.OwnerID] = append(installRunnerMap[rm.OwnerID], rm.RunnerID)
		runnerCreatedAt[rm.RunnerID] = rm.CreatedAt
	}

	allRunnerIDs := make([]string, 0)
	for _, ids := range installRunnerMap {
		allRunnerIDs = append(allRunnerIDs, ids...)
	}

	if len(allRunnerIDs) == 0 {
		entries := make([]InstallUptimeEntry, len(installs))
		for i, inst := range installs {
			entries[i] = InstallUptimeEntry{
				InstallID: inst.InstallID, InstallName: inst.InstallName,
				OrgID: inst.OrgID, OrgName: inst.OrgName,
				InstallProcesses: []ProcessUptime{}, MngProcesses: []ProcessUptime{},
			}
		}
		c.JSON(http.StatusOK, gin.H{"installs": entries, "window": window, "since": since.Format(time.RFC3339), "window_ms": windowMS})
		return
	}

	// Fetch processes.
	var processes []app.RunnerProcess
	s.db.WithContext(ctx).
		Where("runner_id IN ?", allRunnerIDs).
		Where("deleted_at = 0").
		Where("started_at IS NOT NULL").
		Order("runner_id, type, started_at DESC").
		Find(&processes)

	// Fetch per-process heartbeat counts and last heartbeat time from ClickHouse.
	type hbRow struct {
		RunnerID      string    `gorm:"column:runner_id"`
		ProcessID     string    `gorm:"column:process_id"`
		Count         int64     `gorm:"column:cnt"`
		LastHeartbeat time.Time `gorm:"column:last_hb"`
	}
	hbMap := make(map[string]hbRow)
	if time.Since(since) <= 48*time.Hour {
		var hbRows []hbRow
		s.chDB.WithContext(ctx).
			Table("runner_heart_beats").
			Select("runner_id, process_id, count(*) as cnt, max(created_at) as last_hb").
			Where("runner_id IN ?", allRunnerIDs).
			Where("created_at >= ?", since).
			Group("runner_id, process_id").
			Find(&hbRows)
		for _, hb := range hbRows {
			hbMap[hb.RunnerID+"|"+hb.ProcessID] = hb
		}
	}

	// Fetch per-process health check counts with healthy/unhealthy breakdown.
	type hcRow struct {
		RunnerID  string `gorm:"column:runner_id"`
		ProcessID string `gorm:"column:process_id"`
		Status    string `gorm:"column:runner_status"`
		Count     int64  `gorm:"column:cnt"`
	}
	type hcAgg struct {
		Total     int64
		Healthy   int64
		Unhealthy int64
	}
	hcMap := make(map[string]*hcAgg)
	if time.Since(since) <= 24*time.Hour {
		var hcRows []hcRow
		s.chDB.WithContext(ctx).
			Table("runner_health_checks").
			Select("runner_id, process_id, runner_status, count(*) as cnt").
			Where("runner_id IN ?", allRunnerIDs).
			Where("created_at >= ?", since).
			Group("runner_id, process_id, runner_status").
			Find(&hcRows)
		for _, hc := range hcRows {
			key := hc.RunnerID + "|" + hc.ProcessID
			agg, ok := hcMap[key]
			if !ok {
				agg = &hcAgg{}
				hcMap[key] = agg
			}
			agg.Total += hc.Count
			if app.RunnerStatus(hc.Status).IsHealthy() {
				agg.Healthy += hc.Count
			} else {
				agg.Unhealthy += hc.Count
			}
		}
	}

	// Fetch runner job status breakdown.
	type jobRow struct {
		RunnerID string `gorm:"column:runner_id"`
		Status   string `gorm:"column:status"`
		Count    int64  `gorm:"column:cnt"`
	}
	var jobRows []jobRow
	s.db.WithContext(ctx).
		Table("runner_jobs").
		Select("runner_id, status, count(*) as cnt").
		Where("runner_id IN ?", allRunnerIDs).
		Where("created_at >= ?", since).
		Where("deleted_at = 0").
		Group("runner_id, status").
		Find(&jobRows)

	jobMap := make(map[string]*JobSummary)
	for _, jr := range jobRows {
		js, ok := jobMap[jr.RunnerID]
		if !ok {
			js = &JobSummary{}
			jobMap[jr.RunnerID] = js
		}
		js.Total += jr.Count
		switch app.RunnerJobStatus(jr.Status) {
		case app.RunnerJobStatusFinished:
			js.Finished += jr.Count
		case app.RunnerJobStatusFailed:
			js.Failed += jr.Count
		case app.RunnerJobStatusTimedOut:
			js.TimedOut += jr.Count
		case app.RunnerJobStatusCancelled:
			js.Cancelled += jr.Count
		default:
			js.Other += jr.Count
		}
	}

	// Build ProcessUptime per runner.
	procByRunner := make(map[string][]ProcessUptime)
	for _, p := range processes {
		key := p.RunnerID + "|" + p.ID
		hb := hbMap[key]
		hc := hcMap[key]

		// Uptime = process created_at to last heartbeat (or updated_at if no HB data).
		var uptimeMS float64
		var uptimeStr, lastHBStr string
		if p.StartedAt != nil {
			end := p.UpdatedAt // fallback
			if !hb.LastHeartbeat.IsZero() {
				end = hb.LastHeartbeat
				lastHBStr = hb.LastHeartbeat.Format(time.RFC3339)
			}
			dur := end.Sub(*p.StartedAt)
			if dur < 0 {
				dur = 0
			}
			uptimeMS = float64(dur.Milliseconds())
			uptimeStr = fmtDuration(dur)
		}

		startedStr := ""
		if p.StartedAt != nil {
			startedStr = p.StartedAt.Format(time.RFC3339)
		}

		pu := ProcessUptime{
			ProcessID:     p.ID,
			RunnerID:      p.RunnerID,
			Type:          string(p.Type),
			Status:        string(p.CompositeStatus.Status),
			Version:       p.Version,
			StartedAt:     startedStr,
			LastHeartbeat: lastHBStr,
			UptimeMS:      uptimeMS,
			UptimeStr:     uptimeStr,
			Heartbeats:    hb.Count,
		}
		if hc != nil {
			pu.HealthChecks = hc.Total
			pu.HealthyChecks = hc.Healthy
			pu.UnhealthyChecks = hc.Unhealthy
		}

		procByRunner[p.RunnerID] = append(procByRunner[p.RunnerID], pu)
	}

	// Collect processes by type and compute metrics.
	// effectiveSince is max(since, runner.created_at) for each install.
	collectByType := func(runnerIDs []string, processType string, effectiveSince time.Time) ([]ProcessUptime, UptimeMetrics) {
		effectiveWindowMS := float64(now.Sub(effectiveSince).Milliseconds())
		if effectiveWindowMS < 0 {
			effectiveWindowMS = 0
		}

		var procs []ProcessUptime
		var totalUptimeMS float64
		var totalHB, totalHC, healthyHC, unhealthyHC int64

		for _, rid := range runnerIDs {
			for _, pu := range procByRunner[rid] {
				if pu.Type != processType {
					continue
				}
				procs = append(procs, pu)
				totalUptimeMS += pu.UptimeMS
				totalHB += pu.Heartbeats
				totalHC += pu.HealthChecks
				healthyHC += pu.HealthyChecks
				unhealthyHC += pu.UnhealthyChecks
			}
		}

		if procs == nil {
			procs = []ProcessUptime{}
		}

		totalProcs := len(procs)
		// Restarts = total processes minus 1 (the current one). If 0 procs, 0 restarts.
		restarts := 0
		if totalProcs > 1 {
			restarts = totalProcs - 1
		}

		// Expected heartbeats: 1 per 5s per effective window.
		expectedHB := int64(effectiveWindowMS / 5000)
		// Expected health checks: 1 per 60s per effective window.
		expectedHC := int64(effectiveWindowMS / 60000)

		return procs, UptimeMetrics{
			EffectiveWindowMS:    effectiveWindowMS,
			TotalUptimeMS:        totalUptimeMS,
			TotalProcs:           totalProcs,
			Restarts:             restarts,
			TotalHeartbeats:      totalHB,
			ExpectedHeartbeats:   expectedHB,
			TotalHealthChecks:    totalHC,
			HealthyChecks:        healthyHC,
			UnhealthyChecks:      unhealthyHC,
			ExpectedHealthChecks: expectedHC,
		}
	}

	mergeJobs := func(runnerIDs []string) JobSummary {
		var js JobSummary
		for _, rid := range runnerIDs {
			if j, ok := jobMap[rid]; ok {
				js.Total += j.Total
				js.Finished += j.Finished
				js.Failed += j.Failed
				js.TimedOut += j.TimedOut
				js.Cancelled += j.Cancelled
				js.Other += j.Other
			}
		}
		return js
	}

	// Assemble per-install entries.
	entries := make([]InstallUptimeEntry, 0, len(installs))
	for _, inst := range installs {
		runnerIDs := installRunnerMap[inst.InstallID]

		// Effective since = max(since, earliest runner created_at).
		effectiveSince := since
		var earliestCreated string
		for _, rid := range runnerIDs {
			if ca, ok := runnerCreatedAt[rid]; ok {
				if ca.After(effectiveSince) {
					effectiveSince = ca
				}
				ts := ca.Format(time.RFC3339)
				if earliestCreated == "" || ts < earliestCreated {
					earliestCreated = ts
				}
			}
		}

		installProcs, installMetrics := collectByType(runnerIDs, "install", effectiveSince)
		mngProcs, mngMetrics := collectByType(runnerIDs, "build", effectiveSince)

		entries = append(entries, InstallUptimeEntry{
			InstallID:        inst.InstallID,
			InstallName:      inst.InstallName,
			OrgID:            inst.OrgID,
			OrgName:          inst.OrgName,
			RunnerCreatedAt:  earliestCreated,
			InstallProcesses: installProcs,
			MngProcesses:     mngProcs,
			InstallMetrics:   installMetrics,
			MngMetrics:       mngMetrics,
			Jobs:             mergeJobs(runnerIDs),
		})
	}

	// Fetch distinct orgs for filter.
	var orgs []struct {
		ID   string
		Name string
	}
	s.db.WithContext(ctx).
		Table("orgs").
		Select("id, name").
		Where("deleted_at = 0").
		Order("name").
		Find(&orgs)

	// Fetch label options.
	type labelRow struct {
		Labels *string
	}
	var labelRows []labelRow
	s.db.WithContext(ctx).
		Table("installs").
		Select("labels::text as labels").
		Where("deleted_at = 0").
		Where("labels IS NOT NULL").
		Where("labels::text != '{}'").
		Where("labels::text != 'null'").
		Find(&labelRows)

	labelMap := make(map[string]map[string]bool)
	for _, lr := range labelRows {
		if lr.Labels == nil {
			continue
		}
		var parsed map[string]string
		if err := json.Unmarshal([]byte(*lr.Labels), &parsed); err != nil {
			continue
		}
		for k, v := range parsed {
			if _, ok := labelMap[k]; !ok {
				labelMap[k] = make(map[string]bool)
			}
			labelMap[k][v] = true
		}
	}

	type labelOption struct {
		Key    string   `json:"key"`
		Values []string `json:"values"`
	}
	var labelOptions []labelOption
	for k, vs := range labelMap {
		vals := make([]string, 0, len(vs))
		for v := range vs {
			vals = append(vals, v)
		}
		sort.Strings(vals)
		labelOptions = append(labelOptions, labelOption{Key: k, Values: vals})
	}
	sort.Slice(labelOptions, func(i, j int) bool { return labelOptions[i].Key < labelOptions[j].Key })

	c.JSON(http.StatusOK, gin.H{
		"installs":      entries,
		"window":        window,
		"since":         since.Format(time.RFC3339),
		"window_ms":     windowMS,
		"orgs":          orgs,
		"label_options": labelOptions,
	})
}

func windowStart(window string) time.Time {
	now := time.Now()
	switch window {
	case "today":
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	case "quarter":
		return now.AddDate(0, -3, 0)
	default:
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	}
}

func fmtDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}
