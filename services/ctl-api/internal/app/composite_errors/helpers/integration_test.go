package helpers_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/pkg/composite_error/catalog"
	"github.com/nuonco/nuon/pkg/composite_error/unknown"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cehelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// HelpersDeps lets fx populate the dependencies the suite needs.
type HelpersDeps struct {
	fx.In
	DB      *gorm.DB `name:"psql"`
	V       *validator.Validate
	L       *zap.Logger
	Seeder  *testseed.Seeder
	Helpers *cehelpers.Helpers
}

type CompositeErrorsHelpersTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps HelpersDeps

	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
}

func TestCompositeErrorsHelpersSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(CompositeErrorsHelpersTestSuite))
}

func (s *CompositeErrorsHelpersTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Populate(&s.deps),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *CompositeErrorsHelpersTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CompositeErrorsHelpersTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	ctx := context.Background()
	ctx, s.testAcc = s.deps.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.deps.Seeder.EnsureOrg(ctx, s.T())
	s.ctx = ctx
}

// stableOwner generates a unique owner id for each test so rows don't bleed
// across test cases via the polymorphic owner index.
func stableOwner(prefix string) (ownerType, ownerID string) {
	return prefix, domains.NewWorkflowStepID()
}

// TestRecordRoundTrip persists a typed unknown_error, then reads it back
// via Get() and verifies the typed instance hydrates correctly.
//
// This is the spec's Phase 1 round-trip test (encode → store → hydrate → render).
func (s *CompositeErrorsHelpersTestSuite) TestRecordRoundTrip() {
	exit := 137
	original := &unknown.Error{Message: "boom", ExitCode: &exit}
	ownerType, ownerID := stableOwner("install_workflow_steps")

	row, err := s.deps.Helpers.Record(s.ctx, cehelpers.RecordInput{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Error:     original,
		Source:    composite_error.Source{ParserName: "test", Snippet: "raw"},
	})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), row.ID)
	assert.Equal(s.T(), unknown.Type, row.Type)
	assert.Equal(s.T(), composite_error.DomainNuon, row.Domain)
	assert.Equal(s.T(), composite_error.SeverityError, row.Severity)
	assert.Equal(s.T(), s.testOrg.ID, row.OrgID, "OrgID populated from context via BeforeCreate")
	assert.Equal(s.T(), "boom", row.TitleCached)
	// unknown_error renders a bare title only (no apologetic summary copy)
	// so SummaryCached is intentionally empty for this type.
	assert.Empty(s.T(), row.SummaryCached)

	// Read back, hydrate, render.
	loaded, typed, err := s.deps.Helpers.Get(s.ctx, row.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), loaded)
	require.NotNil(s.T(), typed)

	hydrated, ok := typed.(*unknown.Error)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "boom", hydrated.Message)
	require.NotNil(s.T(), hydrated.ExitCode)
	assert.Equal(s.T(), 137, *hydrated.ExitCode)

	rendered := typed.Render(context.Background())
	assert.Equal(s.T(), "boom", rendered.Title)
}

// TestRecordPersistsCausesAsEdges verifies the cause graph is materialized
// as composite_error_causes rows during Record().
func (s *CompositeErrorsHelpersTestSuite) TestRecordPersistsCausesAsEdges() {
	ownerType, ownerID := stableOwner("install_workflow_steps")

	root, err := s.deps.Helpers.Record(s.ctx, cehelpers.RecordInput{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Error:     &unknown.Error{Message: "root"},
		Causes: []cehelpers.RecordCause{
			{Error: &unknown.Error{Message: "child1"}, IsPrimary: true},
			{Error: &unknown.Error{Message: "child2"}, IsPrimary: true /* coerced to false */},
			{Error: &unknown.Error{Message: "child3"}, Causes: []cehelpers.RecordCause{
				{Error: &unknown.Error{Message: "grandchild"}, IsPrimary: true},
			}},
		},
	})
	require.NoError(s.T(), err)

	var edges []app.CompositeErrorCause
	require.NoError(s.T(), s.deps.DB.Where("parent_id = ?", root.ID).Order("idx asc").Find(&edges).Error)
	require.Len(s.T(), edges, 3)
	assert.True(s.T(), edges[0].IsPrimary)
	assert.False(s.T(), edges[1].IsPrimary, "second IsPrimary=true should be coerced to false")
	assert.False(s.T(), edges[2].IsPrimary)

	// child3 should have its own grandchild edge.
	var grandEdges []app.CompositeErrorCause
	require.NoError(s.T(), s.deps.DB.Where("parent_id = ?", edges[2].ChildID).Find(&grandEdges).Error)
	require.Len(s.T(), grandEdges, 1)
	assert.True(s.T(), grandEdges[0].IsPrimary)
}

// TestListByOwnerOrdersByCreatedAtAsc verifies the read-side ordering: rows
// come back in insertion order. Severity-aware re-ranking is intentionally
// a UI concern (see ListByOwner doc).
func (s *CompositeErrorsHelpersTestSuite) TestListByOwnerOrdersByCreatedAtAsc() {
	ownerType, ownerID := stableOwner("install_workflow_steps")

	// Insert in mixed severities; the order should match insertion order.
	wantSev := []composite_error.Severity{
		composite_error.SeverityWarning,
		composite_error.SeverityError,
		composite_error.SeverityWarning,
	}
	for _, sev := range wantSev {
		_, err := s.deps.Helpers.Record(s.ctx, cehelpers.RecordInput{
			OwnerType: ownerType,
			OwnerID:   ownerID,
			Error:     &severityForcer{Sev: sev, Msg: string(sev)},
		})
		require.NoError(s.T(), err)
		time.Sleep(2 * time.Millisecond) // make created_at deterministic
	}

	rows, err := s.deps.Helpers.ListByOwner(s.ctx, ownerType, ownerID)
	require.NoError(s.T(), err)
	require.Len(s.T(), rows, 3)
	for i, sev := range wantSev {
		assert.Equal(s.T(), sev, rows[i].Severity, "row %d preserves insertion order", i)
	}
	assert.True(s.T(), rows[0].CreatedAt.Before(rows[1].CreatedAt))
	assert.True(s.T(), rows[1].CreatedAt.Before(rows[2].CreatedAt))
}

// TestPrimaryReturnsOldestError verifies that Primary returns the oldest
// row attached to the owner regardless of severity.
func (s *CompositeErrorsHelpersTestSuite) TestPrimaryReturnsOldestError() {
	ownerType, ownerID := stableOwner("install_workflow_steps")

	firstRow, err := s.deps.Helpers.Record(s.ctx, cehelpers.RecordInput{
		OwnerType: ownerType, OwnerID: ownerID,
		Error: &severityForcer{Sev: composite_error.SeverityWarning, Msg: "w"},
	})
	require.NoError(s.T(), err)
	time.Sleep(2 * time.Millisecond)

	_, err = s.deps.Helpers.Record(s.ctx, cehelpers.RecordInput{
		OwnerType: ownerType, OwnerID: ownerID,
		Error: &severityForcer{Sev: composite_error.SeverityError, Msg: "e"},
	})
	require.NoError(s.T(), err)

	primary, err := s.deps.Helpers.Primary(s.ctx, ownerType, ownerID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), primary)
	assert.Equal(s.T(), firstRow.ID, primary.ID, "oldest error wins regardless of severity")
}

// TestTreeReturnsCausesGraph verifies depth-bounded traversal.
func (s *CompositeErrorsHelpersTestSuite) TestTreeReturnsCausesGraph() {
	ownerType, ownerID := stableOwner("install_workflow_steps")

	root, err := s.deps.Helpers.Record(s.ctx, cehelpers.RecordInput{
		OwnerType: ownerType, OwnerID: ownerID,
		Error: &unknown.Error{Message: "root"},
		Causes: []cehelpers.RecordCause{
			{Error: &unknown.Error{Message: "child"}, IsPrimary: true, Causes: []cehelpers.RecordCause{
				{Error: &unknown.Error{Message: "grand"}, IsPrimary: true},
			}},
		},
	})
	require.NoError(s.T(), err)

	tree, err := s.deps.Helpers.Tree(s.ctx, root.ID, 5)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), tree)
	require.Equal(s.T(), root.ID, tree.Root.ID)

	require.Len(s.T(), tree.Children, 1)
	child := tree.Children[0]
	assert.True(s.T(), child.IsPrimary)
	assert.Equal(s.T(), "child", child.Error.TitleCached)
	require.Len(s.T(), child.Children, 1)
	assert.Equal(s.T(), "grand", child.Children[0].Error.TitleCached)

	shallow, err := s.deps.Helpers.Tree(s.ctx, root.ID, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), shallow.Children, 1)
	assert.Empty(s.T(), shallow.Children[0].Children)
}

// TestRecordFromError exercises the convenience entrypoint via the unknown
// fallback path.
func (s *CompositeErrorsHelpersTestSuite) TestRecordFromError() {
	ownerType, ownerID := stableOwner("install_workflow_steps")

	row, typed, err := s.deps.Helpers.RecordFromError(s.ctx, ownerType, ownerID, "unhandled/context",
		composite_error.ParseInput{
			Raw: []byte("first line\nrest of output"),
		},
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), row)
	require.NotNil(s.T(), typed)

	assert.Equal(s.T(), unknown.Type, row.Type)
	assert.Equal(s.T(), "first line", row.TitleCached)
	assert.Equal(s.T(), "unknown_error.builder", row.Source.ParserName)
}

// severityForcer is a minimal CompositeError used to insert rows with
// arbitrary severities for sort/primary tests. Registered into the catalog
// in init() so the helper's Hydrate path round-trips it correctly.
type severityForcer struct {
	Sev composite_error.Severity `json:"sev"`
	Msg string                   `json:"msg"`
}

func (s *severityForcer) Type() composite_error.Type {
	return composite_error.Type("test_severity_forcer")
}
func (s *severityForcer) Domain() composite_error.Domain     { return composite_error.DomainNuon }
func (s *severityForcer) Severity() composite_error.Severity { return s.Sev }
func (s *severityForcer) Render(_ context.Context) composite_error.Render {
	return composite_error.Render{Title: s.Msg}
}

func init() {
	catalog.RegisterType(composite_error.Type("test_severity_forcer"),
		func() composite_error.CompositeError { return &severityForcer{} })
}
