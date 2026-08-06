package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const source = "nuonctl export-airgap-plan"

type envelope struct {
	Version               string          `json:"version"`
	OrgID                 string          `json:"org_id"`
	AppID                 string          `json:"app_id"`
	InstallID             string          `json:"install_id"`
	CreatedAt             time.Time       `json:"created_at"`
	Source                string          `json:"source"`
	AppConfig             json.RawMessage `json:"app_config"`
	Inputs                []input         `json:"inputs"`
	ForceDefaultCloudAuth bool            `json:"force_default_cloud_auth,omitempty"`
	Steps                 []step          `json:"steps"`
}

type input struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Secret      bool   `json:"secret"`
	Required    bool   `json:"required"`
}

type step struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	JobType       string          `json:"job_type"`
	JobOperation  string          `json:"job_operation"`
	JobGroup      string          `json:"job_group"`
	DependsOn     []string        `json:"depends_on"`
	PlanFromStep  string          `json:"plan_from_step,omitempty"`
	CompositePlan json.RawMessage `json:"composite_plan"`
}

type install struct {
	ID          string
	OrgID       string
	AppID       string
	AppConfigID string
}

type job struct {
	ID            string
	JobType       string
	Operation     string
	Group         string
	CompositePlan []byte
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("export-airgap-plan", flag.ContinueOnError)
	installID := flags.String("install-id", "", "install ID to export")
	out := flags.String("out", "./airgap-envelope.json", "output path")
	jobIDs := flags.String("job-ids", "", "ordered comma-separated runner job IDs to export (default: all plan-bearing jobs)")
	forceDefaultCloudAuth := flags.Bool("force-default-cloud-auth", false, "rewrite cloud-auth blocks to use the offline runner's ambient credentials")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *installID == "" {
		return errors.New("--install-id is required")
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}

	dsn := os.Getenv("CTL_API_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://ctl_api@localhost:5432/ctl_api?sslmode=disable"
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open local ctl-api database: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to local ctl-api database: %w", err)
	}

	inst, err := getInstall(ctx, db, *installID)
	if err != nil {
		return err
	}
	appConfig, err := getAppConfig(ctx, db, inst.AppConfigID)
	if err != nil {
		return err
	}
	inputs, err := getInputs(ctx, db, inst.AppConfigID)
	if err != nil {
		return err
	}
	steps, skipped, err := getSteps(ctx, db, inst.ID)
	if err != nil {
		return err
	}
	if *jobIDs != "" {
		steps, err = selectSteps(steps, strings.Split(*jobIDs, ","))
		if err != nil {
			return err
		}
	}
	if len(skipped) > 0 {
		log.Printf("warning: skipped runner jobs with empty plans: %s", strings.Join(skipped, ", "))
	}
	if len(steps) == 0 {
		return fmt.Errorf("install %s has no runner jobs with plans in sandbox, deploy, or sync groups", inst.ID)
	}

	contents, err := json.MarshalIndent(envelope{
		Version:               "v0",
		OrgID:                 inst.OrgID,
		AppID:                 inst.AppID,
		InstallID:             inst.ID,
		CreatedAt:             time.Now().UTC(),
		Source:                source,
		AppConfig:             appConfig,
		Inputs:                inputs,
		ForceDefaultCloudAuth: *forceDefaultCloudAuth,
		Steps:                 steps,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	contents = append(contents, '\n')
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(*out, contents, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", *out, err)
	}
	log.Printf("wrote %d steps to %s", len(steps), *out)
	return nil
}

// selectSteps filters to the given ordered job IDs, rebuilds the linear
// dependency chain, and points each apply-plan step at the create-*-plan step
// immediately before it so the offline runner chains the freshly rendered
// plan instead of replaying a stale one.
func selectSteps(steps []step, ids []string) ([]step, error) {
	byID := make(map[string]step, len(steps))
	for _, s := range steps {
		byID[s.ID] = s
	}
	selected := make([]step, 0, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		s, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("job %s not found among the install's plan-bearing runner jobs", id)
		}
		s.DependsOn = nil
		s.PlanFromStep = ""
		if len(selected) > 0 {
			prev := selected[len(selected)-1]
			s.DependsOn = []string{prev.ID}
			if s.JobOperation == "apply-plan" && s.JobType == prev.JobType && strings.HasPrefix(prev.JobOperation, "create-") {
				s.PlanFromStep = prev.ID
			}
		}
		selected = append(selected, s)
	}
	if len(selected) == 0 {
		return nil, errors.New("--job-ids selected no jobs")
	}
	return selected, nil
}

func getInstall(ctx context.Context, db *sql.DB, installID string) (install, error) {
	var inst install
	err := db.QueryRowContext(ctx, `
SELECT id, org_id, app_id, app_config_id
FROM installs
WHERE id = $1 AND deleted_at = 0`, installID).Scan(&inst.ID, &inst.OrgID, &inst.AppID, &inst.AppConfigID)
	if errors.Is(err, sql.ErrNoRows) {
		return install{}, fmt.Errorf("install %s not found", installID)
	}
	if err != nil {
		return install{}, fmt.Errorf("query install %s: %w", installID, err)
	}
	if inst.AppConfigID == "" {
		return install{}, fmt.Errorf("install %s has no app config", installID)
	}
	return inst, nil
}

func getAppConfig(ctx context.Context, db *sql.DB, appConfigID string) (json.RawMessage, error) {
	var raw []byte
	err := db.QueryRowContext(ctx, `
SELECT row_to_json(app_config)::jsonb || jsonb_build_object(
	'sandbox',
	(
		SELECT row_to_json(sandbox)::jsonb
		FROM app_sandbox_configs AS sandbox
		WHERE sandbox.app_config_id = app_config.id AND sandbox.deleted_at = 0
		ORDER BY sandbox.created_at DESC
		LIMIT 1
	)
)
FROM app_configs AS app_config
WHERE id = $1 AND deleted_at = 0`, appConfigID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("app config %s not found", appConfigID)
	}
	if err != nil {
		return nil, fmt.Errorf("query app config %s: %w", appConfigID, err)
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("app config %s did not produce valid JSON", appConfigID)
	}
	var probe struct {
		Sandbox json.RawMessage `json:"sandbox"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("parse app config %s: %w", appConfigID, err)
	}
	if len(probe.Sandbox) == 0 || string(probe.Sandbox) == "null" {
		return nil, fmt.Errorf("app config %s has no sandbox config; airgap replay of sandbox jobs requires one", appConfigID)
	}
	return raw, nil
}

func getInputs(ctx context.Context, db *sql.DB, appConfigID string) ([]input, error) {
	rows, err := db.QueryContext(ctx, `
SELECT i.name, i.type, i.description, i.sensitive, i.required
FROM app_input_configs AS c
JOIN app_inputs AS i ON i.app_input_config_id = c.id AND i.deleted_at = 0
WHERE c.app_config_id = $1 AND c.deleted_at = 0
ORDER BY i.index, i.created_at, i.id`, appConfigID)
	if err != nil {
		return nil, fmt.Errorf("query inputs for app config %s: %w", appConfigID, err)
	}
	defer rows.Close()

	inputs := make([]input, 0)
	for rows.Next() {
		var value input
		if err := rows.Scan(&value.Name, &value.Type, &value.Description, &value.Secret, &value.Required); err != nil {
			return nil, fmt.Errorf("scan app input: %w", err)
		}
		inputs = append(inputs, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate app inputs: %w", err)
	}
	return inputs, nil
}

func getSteps(ctx context.Context, db *sql.DB, installID string) ([]step, []string, error) {
	rows, err := db.QueryContext(ctx, `
SELECT j.id, j.type, j.operation, j."group", COALESCE(p.composite_plan, 'null'::jsonb)
FROM runner_groups AS rg
JOIN runners AS r ON r.runner_group_id = rg.id AND r.deleted_at = 0
JOIN runner_jobs AS j ON j.runner_id = r.id AND j.deleted_at = 0
LEFT JOIN runner_job_plans AS p ON p.runner_job_id = j.id AND p.deleted_at = 0
WHERE rg.owner_id = $1
  AND rg.owner_type = 'installs'
  AND rg.deleted_at = 0
  AND j."group" IN ('sandbox', 'deploy', 'sync')
ORDER BY CASE j."group"
  WHEN 'sandbox' THEN 0
  WHEN 'deploy' THEN 1
  WHEN 'sync' THEN 2
  ELSE 3
END, j.created_at, j.id`, installID)
	if err != nil {
		return nil, nil, fmt.Errorf("query runner jobs for install %s: %w", installID, err)
	}
	defer rows.Close()

	jobs := make([]job, 0)
	skipped := make([]string, 0)
	for rows.Next() {
		var value job
		if err := rows.Scan(&value.ID, &value.JobType, &value.Operation, &value.Group, &value.CompositePlan); err != nil {
			return nil, nil, fmt.Errorf("scan runner job: %w", err)
		}
		trimmed := strings.TrimSpace(string(value.CompositePlan))
		if trimmed == "" || trimmed == "null" || trimmed == "{}" {
			skipped = append(skipped, value.ID)
			continue
		}
		if !json.Valid(value.CompositePlan) {
			return nil, nil, fmt.Errorf("runner job %s has invalid composite_plan JSON", value.ID)
		}
		jobs = append(jobs, value)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate runner jobs: %w", err)
	}

	steps := make([]step, 0, len(jobs))
	for index, value := range jobs {
		dependsOn := make([]string, 0, 1)
		if index > 0 {
			dependsOn = append(dependsOn, jobs[index-1].ID)
		}
		steps = append(steps, step{
			ID:            value.ID,
			Name:          strings.TrimSpace(value.JobType + " " + value.Operation),
			JobType:       value.JobType,
			JobOperation:  value.Operation,
			JobGroup:      value.Group,
			DependsOn:     dependsOn,
			CompositePlan: value.CompositePlan,
		})
	}
	return steps, skipped, nil
}
