package sync

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/stretchr/testify/require"
)

type cfgOptions struct {
	emptySandbox    bool
	emptyComponents bool
}

type mockNotFoundError struct {
	error
}

func (e mockNotFoundError) IsCode(code int) bool {
	return code == 404
}
func (e mockNotFoundError) IsServerError() bool {
	return false
}
func (e mockNotFoundError) GetPayload() *models.StderrErrResponse {
	return nil
}

func getTestCfg(options cfgOptions) *config.AppConfig {
	baseConfg := &config.AppConfig{
		Version:     "v1",
		Description: "description",
		DisplayName: "displayName",
		Installer:   nil,
		Sandbox: &config.AppSandboxConfig{
			Source: "source",
			ConnectedRepo: &config.ConnectedRepoConfig{
				Repo:      "powertools/mono",
				Directory: "aws-ecs-byovpc",
				Branch:    "main",
			},
			PublicRepo: nil,
			VarMap: map[string]string{
				"vpc_id": "{{.nuon.install.inputs.vpc_id}}",
			},
			AWSDelegationIAMRoleARN: "arn:aws:iam::xxxxxxxxxxxx:role/nuon-aws-ecs-install-access",
		},
		Inputs: &config.AppInputConfig{
			Groups: []config.AppInputGroup{
				{
					Name:        "sandbox",
					Description: "Sandbox inputs",
					DisplayName: "Sandbox inputs",
				},
			},
			Inputs: []config.AppInput{
				{
					Name:        "vpc_id",
					Description: "vpc_id to install application into",
					DisplayName: "VPC ID",
					Required:    false,
					Default:     "",
					Group:       "sandbox",
				},
				{
					Name:        "api_key",
					Description: "API key",
					DisplayName: "API key",
					Required:    true,
					Default:     "",
					Group:       "sandbox",
				},
			},
		},
		Runner: &config.AppRunnerConfig{
			RunnerType: "aws-ecs",
			EnvVarMap: map[string]string{
				"runner-env-var": "runner-env-var",
			},
		},
		Components: []*config.Component{
			{
				Name:    "terraform1",
				Type:    config.TerraformModuleComponentType,
				VarName: "1",
				TerraformModule: &config.TerraformModuleComponentConfig{
					TerraformVersion: "0.12",
					EnvVarMap: map[string]string{
						"env-var": "env-var",
					},
					VarsMap: map[string]string{
						"var": "var",
					},
					ConnectedRepo: &config.ConnectedRepoConfig{
						Repo:      "powertools/mono",
						Directory: "aws-ecs-byovpc",
						Branch:    "main",
					},
				},
			},
			{
				Name:    "helm2",
				Type:    config.HelmChartComponentType,
				VarName: "2",
				HelmChart: &config.HelmChartComponentConfig{
					ValuesMap: map[string]string{
						"key": "value",
					},
					ConnectedRepo: &config.ConnectedRepoConfig{
						Repo:      "powertools/mono",
						Directory: "aws-ecs-byovpc",
						Branch:    "main",
					},
				},
			},
			{
				Name:    "docker3",
				Type:    config.DockerBuildComponentType,
				VarName: "3",
				DockerBuild: &config.DockerBuildComponentConfig{
					ConnectedRepo: &config.ConnectedRepoConfig{
						Repo:      "powertools/mono",
						Directory: "aws-ecs-byovpc",
						Branch:    "main",
					},
					EnvVarMap: map[string]string{
						"env-var": "env-var",
					},
				},
			},
		},
	}

	if options.emptySandbox {
		baseConfg.Sandbox = nil
	}

	if options.emptyComponents {
		baseConfg.Components = nil
	}

	return baseConfg
}

func getTestAppConfig() *models.AppAppConfig {
	return &models.AppAppConfig{
		ID:      "appID",
		OrgID:   "orgID",
		Status:  "active",
		State:   "",
		Version: 1,
	}
}

func TestSync(t *testing.T) {
	tests := []struct {
		name                          string
		appID                         string
		cfg                           *config.AppConfig
		err                           error
		existingComponentIDs          []string
		expectedLatestConfig          *models.AppAppConfig
		expectedLatestConfigErr       error
		expectSyncSandBox             bool
		expectSyncInputs              bool
		expectSyncRunner              bool
		expectSyncInstaller           bool
		expectGetComponentLatestBuild bool
		expectSyncComponents          bool
		expectCmpBuildScheduled       []ComponentState
		expectedFinish                bool
	}{
		{
			name:  "fails on nil cfg",
			appID: "appID",
			cfg:   nil,
			err: SyncInternalErr{
				Description: "nil config",
				Err:         fmt.Errorf("config is nil"),
			},
			expectedLatestConfig:    nil,
			expectedLatestConfigErr: nil,
			expectCmpBuildScheduled: []ComponentState{},
			expectedFinish:          false,
		},
		{
			name:  "fails on nil sandbox",
			appID: "appID",
			cfg:   getTestCfg(cfgOptions{emptySandbox: true, emptyComponents: true}),
			err: SyncAPIErr{
				Resource: "app-sandbox",
				Err:      fmt.Errorf("sandbox config is nil"),
			},
			expectedLatestConfig:    getTestAppConfig(),
			expectedLatestConfigErr: nil,
			expectSyncInputs:        true,
			expectCmpBuildScheduled: []ComponentState{},
			expectedFinish:          true,
		},
		{
			name:                          "syncs as expected",
			appID:                         "appID",
			cfg:                           getTestCfg(cfgOptions{emptySandbox: false, emptyComponents: false}),
			err:                           nil,
			expectedLatestConfig:          getTestAppConfig(),
			expectedLatestConfigErr:       nil,
			expectSyncInputs:              true,
			expectSyncSandBox:             true,
			expectSyncRunner:              true,
			expectSyncInstaller:           true,
			expectGetComponentLatestBuild: true,
			expectCmpBuildScheduled: []ComponentState{
				{ID: "idterraform1"},
				{ID: "idhelm2"},
				{ID: "iddocker3"},
			},
			expectedFinish: true,
		},
		{
			name:                          "syncs with updated component",
			appID:                         "appID",
			cfg:                           getTestCfg(cfgOptions{emptySandbox: false, emptyComponents: false}),
			err:                           nil,
			existingComponentIDs:          []string{"idterraform1"},
			expectedLatestConfig:          getTestAppConfig(),
			expectedLatestConfigErr:       nil,
			expectSyncInputs:              true,
			expectSyncSandBox:             true,
			expectSyncRunner:              true,
			expectSyncInstaller:           true,
			expectGetComponentLatestBuild: true,
			expectCmpBuildScheduled: []ComponentState{
				{ID: "idterraform1"},
				{ID: "idhelm2"},
				{ID: "iddocker3"},
			},
			expectedFinish: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)
			mockApiClient := nuon.NewMockClient(mockCtl)

			defer func() {
				syncer := New(mockApiClient, tt.appID, tt.cfg)
				err := syncer.Sync(ctx)
				require.Equal(t, tt.err, err)
				cmpBuildScheduled := syncer.GetComponentsScheduled()
				require.Equal(t, len(tt.expectCmpBuildScheduled), len(cmpBuildScheduled))
				for idx, cmp := range tt.expectCmpBuildScheduled {
					require.Contains(t, cmp.ID, cmpBuildScheduled[idx].ID)
				}
			}()

			if tt.cfg == nil {
				return
			}

			if tt.expectedLatestConfig != nil {
				mockApiClient.EXPECT().GetAppLatestConfig(ctx, tt.appID).Return(tt.expectedLatestConfig, tt.expectedLatestConfigErr)
			}

			if tt.expectedLatestConfig != nil {
				mockApiClient.EXPECT().CreateAppConfig(ctx, tt.appID, gomock.Any()).Return(tt.expectedLatestConfig, tt.expectedLatestConfigErr)
				mockApiClient.EXPECT().UpdateApp(ctx, tt.appID, gomock.Any()).Return(&models.AppApp{ID: tt.appID}, nil)
			}

			if tt.expectSyncSandBox {
				mockApiClient.EXPECT().CreateAppSandboxConfig(ctx, tt.appID, gomock.Any()).Return(&models.AppAppSandboxConfig{ID: "sandboxId"}, nil)
			}

			if tt.expectSyncInputs {
				mockApiClient.EXPECT().CreateAppInputConfig(ctx, tt.appID, gomock.Any()).Return(&models.AppAppInputConfig{ID: "inputId"}, nil)
			}

			if tt.expectSyncRunner {
				mockApiClient.EXPECT().CreateAppRunnerConfig(ctx, tt.appID, gomock.Any()).Return(&models.AppAppRunnerConfig{ID: "runnerId"}, nil)
			}

			// validate we are syncing dependencies in the correct order
			deps := make([]string, 0)

			for _, comp := range tt.cfg.Components {
				mockId := "id" + comp.Name
				if slices.Contains(tt.existingComponentIDs, mockId) {
					mockApiClient.EXPECT().GetAppComponent(ctx, tt.appID, comp.Name).Return(&models.AppComponent{ID: mockId, Type: models.AppComponentType(comp.Type)}, nil)
					mockApiClient.EXPECT().UpdateComponent(ctx, mockId, &models.ServiceUpdateComponentRequest{
						Dependencies: deps,
						Name:         &comp.Name,
						VarName:      comp.VarName,
					}).Return(nil, nil)
				} else {
					mockApiClient.EXPECT().GetAppComponent(ctx, tt.appID, comp.Name).Return(nil, mockNotFoundError{})
					mockApiClient.EXPECT().CreateComponent(ctx, tt.appID, &models.ServiceCreateComponentRequest{
						Dependencies: deps,
						Name:         &comp.Name,
						VarName:      comp.VarName,
					}).Return(&models.AppComponent{ID: mockId}, nil)
				}
				if tt.expectGetComponentLatestBuild {
					mockApiClient.EXPECT().GetComponentLatestBuild(ctx, mockId).Return(&models.AppComponentBuild{
						ID:     mockId,
						Status: "active",
					}, nil)
				}
				switch comp.Type {
				case config.TerraformModuleComponentType:
					mockApiClient.EXPECT().CreateTerraformModuleComponentConfig(ctx, gomock.Any(), gomock.Any()).Return(&models.AppTerraformModuleComponentConfig{ID: mockId}, nil)
				case config.HelmChartComponentType:
					mockApiClient.EXPECT().CreateHelmComponentConfig(ctx, gomock.Any(), gomock.Any()).Return(&models.AppHelmComponentConfig{ID: mockId}, nil)
				case config.DockerBuildComponentType:
					mockApiClient.EXPECT().CreateDockerBuildComponentConfig(ctx, gomock.Any(), gomock.Any()).Return(&models.AppDockerBuildComponentConfig{ID: mockId}, nil)
				case config.JobComponentType:
					mockApiClient.EXPECT().CreateJobComponentConfig(ctx, gomock.Any(), gomock.Any()).Return(&models.AppJobComponentConfig{ID: mockId}, nil)
				case config.ExternalImageComponentType:
					mockApiClient.EXPECT().CreateExternalImageComponentConfig(ctx, gomock.Any(), gomock.Any()).Return(&models.AppExternalImageComponentConfig{ID: mockId}, nil)
				}
				deps = []string{mockId}
			}

			if tt.expectedFinish {
				mockApiClient.EXPECT().UpdateAppConfig(ctx, tt.appID, tt.expectedLatestConfig.ID, gomock.Any()).Return(&models.AppAppConfig{ID: tt.appID}, nil)
			}
		})
	}
}
