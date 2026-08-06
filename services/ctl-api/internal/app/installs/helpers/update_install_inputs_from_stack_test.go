package helpers_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	temporal "github.com/nuonco/nuon/pkg/temporal/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type stackInputsDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *installhelpers.Helpers
}

type StackOutputInputsTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps stackInputsDeps
}

func TestStackOutputInputsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(StackOutputInputsTestSuite))
}

type stackInputsWorkflowRun struct{}

func (r *stackInputsWorkflowRun) GetID() string    { return "test-workflow-id" }
func (r *stackInputsWorkflowRun) GetRunID() string { return "test-run-id" }
func (r *stackInputsWorkflowRun) Get(context.Context, any) error {
	return nil
}
func (r *stackInputsWorkflowRun) GetWithOptions(context.Context, any, tclient.WorkflowRunGetOptions) error {
	return nil
}

func (s *StackOutputInputsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	ctrl := gomock.NewController(s.T())
	mockTC := temporal.NewMockClient(ctrl)
	mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(&stackInputsWorkflowRun{}, nil).AnyTimes()

	options := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T:     s.T(),
		Mocks: &tests.TestMocks{MockTC: mockTC},
	}), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *StackOutputInputsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

// Customer stack outputs must land on the row readers actually resolve. Keying the
// write on the input config the caller was handed writes to a row a newer config pin
// already supersedes, and the outputs never take effect.
func (s *StackOutputInputsTestSuite) TestOutputsMergeOntoNewestRow() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var oldInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: oldCfg.ID}).First(&oldInputCfg).Error)

	var oldGroup app.AppInputGroup
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputGroup{AppInputConfigID: oldInputCfg.ID}).First(&oldGroup).Error)

	// The stack can only write inputs it declares as customer-sourced.
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(&app.AppInput{
		AppInputConfigID: oldInputCfg.ID,
		AppInputGroupID:  oldGroup.ID,
		Name:             "account_id",
		DisplayName:      "Account ID",
		Description:      "customer account id, reported by the stack",
		Type:             app.AppInputTypeString,
		Source:           app.AppInputSourceCustomer,
	}).Error)

	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, oldInputCfg.ID,
		map[string]*string{"region": ptrTo("us-west-2")})

	// The install has since rolled onto a newer config, so the newest row — the one
	// every reader resolves — is pinned somewhere else.
	newCfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	newInputCfg := s.deps.Seed.CreateAppInputConfig(ctx, s.T(), testApp.ID, newCfg.ID)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, newInputCfg.ID,
		map[string]*string{"region": ptrTo("eu-central-1")})

	stackVersionID := "istv" + install.ID
	stackAccount := testseed.BuildAccount()
	stackAccount.Subject = stackVersionID
	stackAccount.AccountType = app.AccountTypeService
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(stackAccount).Error)

	_, err := s.deps.Helpers.UpdateInstallInputsFromStackOutputs(
		ctx,
		stackVersionID,
		install.ID,
		oldInputCfg.ID,
		map[string]string{"account_id": "123456789012"},
		true,
	)
	s.Require().NoError(err)

	var newest app.InstallInputs
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.InstallInputs{InstallID: install.ID}).
		Order("created_at DESC").
		First(&newest).Error)

	s.Require().NotNil(newest.Values["account_id"], "stack output did not reach the newest row")
	s.Equal("123456789012", *newest.Values["account_id"])
	s.Require().NotNil(newest.Values["region"])
	s.Equal("eu-central-1", *newest.Values["region"])
}
