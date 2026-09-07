package v2

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/nuonco/nuon/pkg/generics"
	dataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

const (
	oldAppConfigID = "cfg-old"
	newAppConfigID = "cfg-new"
	existingCompID = "comp-existing"
	addedCompID    = "comp-added"
)

// filterComponentsByDiff can only ever return a subset of the order it is given,
// so a component the update adds survives only when that order came from the new
// app config.
func TestFilterComponentsByDiff(t *testing.T) {
	newAppCfg := &app.AppConfig{
		ComponentIDs: pq.StringArray{existingCompID, addedCompID},
	}
	diff := &app.InstallConfigDiff{
		Added:     []app.ComponentDiffEntry{{ComponentID: addedCompID}},
		Changed:   []app.ComponentDiffEntry{{ComponentID: existingCompID}},
		Unchanged: []app.ComponentDiffEntry{{ComponentID: "comp-untouched"}},
	}

	tests := []struct {
		name         string
		componentIDs []string
		diff         *app.InstallConfigDiff
		want         []string
	}{
		{
			name:         "keeps a component only the new config has",
			componentIDs: []string{existingCompID, addedCompID},
			diff:         diff,
			want:         []string{existingCompID, addedCompID},
		},
		{
			name:         "preserves the order it was given",
			componentIDs: []string{addedCompID, existingCompID},
			diff:         diff,
			want:         []string{addedCompID, existingCompID},
		},
		{
			name:         "an order taken from the old config drops the added component",
			componentIDs: []string{existingCompID},
			diff:         diff,
			want:         []string{existingCompID},
		},
		{
			name:         "untouched components are not redeployed",
			componentIDs: []string{existingCompID, "comp-untouched", addedCompID},
			diff:         diff,
			want:         []string{existingCompID, addedCompID},
		},
		{
			name:         "components dropped from the new config are excluded",
			componentIDs: []string{existingCompID, "comp-removed"},
			diff: &app.InstallConfigDiff{
				Changed: []app.ComponentDiffEntry{{ComponentID: existingCompID}, {ComponentID: "comp-removed"}},
			},
			want: []string{existingCompID},
		},
		{
			name:         "no diff deploys every component in the new config",
			componentIDs: []string{existingCompID, addedCompID},
			want:         []string{existingCompID, addedCompID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, filterComponentsByDiff(tt.componentIDs, newAppCfg, tt.diff))
		})
	}
}

type appBranchConfigUpdateSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment

	// graphRequests records what the workflow asked the app graph to be ordered
	// against.
	graphRequests []activities.GetAppGraphRequest
}

func TestAppBranchConfigUpdateSuite(t *testing.T) {
	suite.Run(t, new(appBranchConfigUpdateSuite))
}

func (s *appBranchConfigUpdateSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()

	// app.* models carry their Temporal payload shape on `temporaljson` tags, so
	// the SDK's default converter would drop fields the step generator reads —
	// notably ComponentConfigConnection.Component, which is `json:"-"`.
	s.env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		dataconverter.NewJSONConverter(),
	))
	s.env.SetWorkerOptions(worker.Options{
		DeadlockDetectionTimeout: time.Minute,
	})
	s.graphRequests = nil

	// Registering by name is what lets the test env deserialize each call's
	// arguments before handing them to the mock; without it the mock sees zero
	// values.
	a := &activities.Activities{}
	for name, fn := range map[string]any{
		"Get":                                  a.Get,
		"HasFeature":                           a.HasFeature,
		"GetAppConfig":                         a.GetAppConfig,
		"GetComponentBuildForConfigConnection": a.GetComponentBuildForConfigConnection,
		"GetAppGraph":                          a.GetAppGraph,
		"GetInstallAppConfigVersionDiff":       a.GetInstallAppConfigVersionDiff,
		"GetActionWorkflows":                   a.GetActionWorkflows,
		"GetInstallComponentsBatch":            a.GetInstallComponentsBatch,
	} {
		s.env.RegisterActivityWithOptions(fn, activity.RegisterOptions{Name: name})
	}
}

// newAppConfig is the config being rolled out: it carries a component the
// install's current config does not have.
func (s *appBranchConfigUpdateSuite) newAppConfig() *app.AppConfig {
	return &app.AppConfig{
		ID:           newAppConfigID,
		ComponentIDs: pq.StringArray{existingCompID, addedCompID},
		ComponentConfigConnections: []app.ComponentConfigConnection{
			{
				ID:          "ccc-existing",
				AppConfigID: newAppConfigID,
				ComponentID: existingCompID,
				Component:   app.Component{ID: existingCompID, Name: "existing", Type: app.ComponentTypeTerraformModule},
			},
			{
				ID:                     "ccc-added",
				AppConfigID:            newAppConfigID,
				ComponentID:            addedCompID,
				Component:              app.Component{ID: addedCompID, Name: "added", Type: app.ComponentTypeTerraformModule},
				ComponentDependencyIDs: pq.StringArray{existingCompID},
			},
		},
	}
}

func (s *appBranchConfigUpdateSuite) mockActivities(diff *app.InstallConfigDiff, graphOrder []string) {
	s.env.OnActivity("Get", mock.Anything, mock.Anything).Return(
		&app.Install{ID: "install-1", AppConfigID: oldAppConfigID}, nil)
	s.env.OnActivity("HasFeature", mock.Anything, mock.Anything).Return(true, nil)
	s.env.OnActivity("GetAppConfig", mock.Anything, mock.Anything).Return(s.newAppConfig(), nil)
	s.env.OnActivity("GetComponentBuildForConfigConnection", mock.Anything, mock.Anything).Return(
		&app.ComponentBuild{ID: "bld-1"}, nil).Maybe()
	s.env.OnActivity("GetInstallAppConfigVersionDiff", mock.Anything, mock.Anything).Return(diff, nil)
	s.env.OnActivity("GetActionWorkflows", mock.Anything, mock.Anything).Return(
		[]*app.InstallActionWorkflow{}, nil)
	s.env.OnActivity("GetInstallComponentsBatch", mock.Anything, mock.Anything).Return(
		map[string]*app.InstallComponent{}, nil)

	s.env.OnActivity("GetAppGraph", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			s.graphRequests = append(s.graphRequests, args.Get(1).(activities.GetAppGraphRequest))
		}).
		Return(graphOrder, nil)
}

func (s *appBranchConfigUpdateSuite) workflow() *app.Workflow {
	return &app.Workflow{
		ID:       "flw-1",
		PlanOnly: true,
		Metadata: pgtype.Hstore{
			"install_id":               generics.ToPtr("install-1"),
			"new_app_config_id":        generics.ToPtr(newAppConfigID),
			"install_config_update_id": generics.ToPtr("iacv-1"),
		},
	}
}

func (s *appBranchConfigUpdateSuite) run(diff *app.InstallConfigDiff, graphOrder []string) *app.GenerateStepsResult {
	s.mockActivities(diff, graphOrder)
	s.env.ExecuteWorkflow(AppBranchConfigUpdate, s.workflow())

	s.Require().True(s.env.IsWorkflowCompleted())
	s.Require().NoError(s.env.GetWorkflowError())

	var result app.GenerateStepsResult
	s.Require().NoError(s.env.GetWorkflowResult(&result))
	return &result
}

// The install is still pinned to its old config while these steps are generated,
// so ordering the graph against the install leaves a newly added component with
// no vertex — and therefore no deploy step.
func (s *appBranchConfigUpdateSuite) TestOrdersGraphAgainstNewAppConfig() {
	s.run(
		&app.InstallConfigDiff{
			Added:   []app.ComponentDiffEntry{{ComponentID: addedCompID}},
			Changed: []app.ComponentDiffEntry{{ComponentID: existingCompID}},
		},
		[]string{existingCompID, addedCompID},
	)

	s.Require().Len(s.graphRequests, 1)
	s.Equal(newAppConfigID, s.graphRequests[0].AppConfigID)
	s.Equal("install-1", s.graphRequests[0].InstallID)
}

func (s *appBranchConfigUpdateSuite) TestDeploysComponentOnlyInNewConfig() {
	result := s.run(
		&app.InstallConfigDiff{
			Added:   []app.ComponentDiffEntry{{ComponentID: addedCompID}},
			Changed: []app.ComponentDiffEntry{{ComponentID: existingCompID}},
		},
		[]string{existingCompID, addedCompID},
	)

	s.Equal([]string{"existing", "added"}, deployedComponents(result.Steps))
}

// A component the update leaves alone stays off the deploy list even though the
// new config's graph includes it.
func (s *appBranchConfigUpdateSuite) TestSkipsUnchangedComponents() {
	result := s.run(
		&app.InstallConfigDiff{
			Added:     []app.ComponentDiffEntry{{ComponentID: addedCompID}},
			Unchanged: []app.ComponentDiffEntry{{ComponentID: existingCompID}},
		},
		[]string{existingCompID, addedCompID},
	)

	s.Equal([]string{"added"}, deployedComponents(result.Steps))
}

func deployedComponents(steps []*app.WorkflowStep) []string {
	const prefix = "sync and plan "

	names := make([]string, 0, len(steps))
	for _, step := range steps {
		if strings.HasPrefix(step.Name, prefix) {
			names = append(names, strings.TrimPrefix(step.Name, prefix))
		}
	}
	return names
}
