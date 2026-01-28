package activities

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/oci/metadata"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PrepExternalImagePolicyRequest struct {
	BuildID       string                  `validate:"required"`
	ImageMetadata *metadata.ImageMetadata `validate:"required"`
}

type ExternalImagePolicyViolation struct {
	PolicyID string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	Message  string `json:"message" temporaljson:"message,omitempty"`
	Severity string `json:"severity" temporaljson:"severity,omitempty"`
}

type ExternalImagePolicyToEvaluate struct {
	PolicyID  string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	Contents  string `json:"contents" temporaljson:"contents,omitempty"`
	InputJSON []byte `json:"input_json" temporaljson:"input_json,omitempty"`
}

type PrepExternalImagePolicyResult struct {
	Policies    []ExternalImagePolicyToEvaluate `json:"policies" temporaljson:"policies,omitempty"`
	HasPolicies bool                            `json:"has_policies" temporaljson:"has_policies,omitempty"`
}

// @temporal-gen activity
// @max-retries 1
// @schedule-to-close-timeout 5m
// @start-to-close-timeout 4m
func (a *Activities) PrepExternalImagePolicy(ctx context.Context, req *PrepExternalImagePolicyRequest) (*PrepExternalImagePolicyResult, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("build_id", req.BuildID))

	l.Info("preparing external image policy evaluation")

	build, err := a.getBuildWithAppConfig(ctx, req.BuildID)
	if err != nil {
		l.Error("unable to get build with app config", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get build with app config")
	}

	appConfigs := build.ComponentConfigConnection.Component.App.AppConfigs
	if len(appConfigs) == 0 {
		l.Info("no app config found, skipping policy evaluation")
		return &PrepExternalImagePolicyResult{
			Policies:    []ExternalImagePolicyToEvaluate{},
			HasPolicies: false,
		}, nil
	}
	appConfigID := appConfigs[0].ID

	l = l.With(zap.String("app_config_id", appConfigID))

	policiesConfig, err := a.appsHelpers.GetPoliciesConfigByAppConfigID(ctx, appConfigID)
	if err != nil {
		l.Info("no policies config found, skipping policy evaluation")
		return &PrepExternalImagePolicyResult{
			Policies:    []ExternalImagePolicyToEvaluate{},
			HasPolicies: false,
		}, nil
	}

	componentName := build.ComponentConfigConnection.Component.Name
	applicablePolicies := a.filterContainerImagePolicies(policiesConfig.Policies, componentName)

	l.Info("filtered applicable container image policies", zap.Int("count", len(applicablePolicies)))

	if len(applicablePolicies) == 0 {
		l.Info("no applicable container image policies found")
		return &PrepExternalImagePolicyResult{
			Policies:    []ExternalImagePolicyToEvaluate{},
			HasPolicies: false,
		}, nil
	}

	policyInput := &metadata.ExternalImagePolicyInput{
		Image:    req.ImageMetadata.Image,
		Tag:      req.ImageMetadata.Tag,
		Digest:   req.ImageMetadata.Digest,
		Metadata: req.ImageMetadata,
	}

	inputJSON, err := json.Marshal(policyInput)
	if err != nil {
		l.Error("unable to marshal policy input", zap.Error(err))
		return nil, errors.Wrap(err, "unable to marshal policy input")
	}

	policies := make([]ExternalImagePolicyToEvaluate, 0, len(applicablePolicies))
	for _, policy := range applicablePolicies {
		policies = append(policies, ExternalImagePolicyToEvaluate{
			PolicyID:  policy.ID,
			Contents:  policy.Contents,
			InputJSON: inputJSON,
		})
	}

	l.Info("policy evaluation preparation complete",
		zap.Int("policies_count", len(policies)),
	)

	return &PrepExternalImagePolicyResult{
		Policies:    policies,
		HasPolicies: true,
	}, nil
}

func (a *Activities) getBuildWithAppConfig(ctx context.Context, buildID string) (*app.ComponentBuild, error) {
	var bld app.ComponentBuild

	res := a.db.WithContext(ctx).
		Preload("ComponentConfigConnection.Component.App.AppConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		First(&bld, "id = ?", buildID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get component build")
	}

	return &bld, nil
}

func (a *Activities) filterContainerImagePolicies(
	policies []app.AppPolicyConfig,
	componentName string,
) []app.AppPolicyConfig {
	var applicable []app.AppPolicyConfig

	for _, policy := range policies {
		if policy.Engine != config.AppPolicyEngineOPA {
			continue
		}

		if policy.Type != config.AppPolicyTypeContainerImage {
			continue
		}

		if len(policy.Components) == 0 {
			continue
		}

		for _, comp := range policy.Components {
			if comp == "*" || comp == componentName {
				applicable = append(applicable, policy)
				break
			}
		}
	}

	return applicable
}
