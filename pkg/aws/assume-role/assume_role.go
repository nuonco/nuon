package iam

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	// why: by default, the maximum number of time that you can assume a role, via a chained assume (like we do in all of
	// our processes), is 3600 seconds. However, due to rounding issues, when this was originally set to time.Hour,
	// this failed because it would come out as slightly larger than 3600 seconds and aws would reject the role
	// assume step.
	defaultRoleSessionDuration time.Duration = time.Second * 3600

	maxSessionDuration time.Duration = time.Second * 3600
)

// TwoStepConfig is a configuration used when we need to jump through another IAM role, before assuming a customer role.
//
// The IAMRoleARN provided will be assumed, by either using default creds, the SrcAccessKeyID or the
// SrcStaticCredentials
type TwoStepConfig struct {
	IAMRoleARN string `cty:"iam_role_arn" hcl:"iam_role_arn" mapstructure:"iam_role_arn,omitempty" json:"iam_role_arn"`

	// Root Credentials
	SrcStaticCredentials StaticCredentials `cty:"src_static_credentials" hcl:"src_static_credentials" mapstructure:"src_static_credentials,omitempty" json:"src_static_credentials"`
	SrcIAMRoleARN        string            `cty:"src_iam_role_arn" hcl:"src_iam_role_arn" mapstructure:"src_iam_role_arn,omitempty" json:"src_iam_role_arn"`
}

type StaticCredentials struct {
	AccessKeyID     string `cty:"access_key_id" hcl:"access_key_id" mapstructure:"access_key_id,omitempty" json:"access_key_id"`
	SecretAccessKey string `cty:"secret_access_key" hcl:"secret_access_key" mapstructure:"secret_access_key,omitempty" json:"secret_access_key"`
	SessionToken    string `cty:"session_token" hcl:"session_token" mapstructure:"session_token,omitempty" json:"session_token"`
}

type Settings struct {
	RoleARN             string `validate:"required"`
	RoleSessionName     string `validate:"required"`
	RoleSessionDuration time.Duration
	ExternalID          string

	TwoStepConfig *TwoStepConfig

	UseGithubOIDC bool
	UseGCPOIDC    bool

	Region string
}

func (s Settings) Validate(v *validator.Validate) error {
	return v.Struct(s)
}

type assumer struct {
	RoleARN             string `validate:"required"`
	RoleSessionName     string `validate:"required"`
	RoleSessionDuration time.Duration
	ExternalID          string

	Region string

	TwoStepConfig *TwoStepConfig
	UseGithubOIDC bool
	UseGCPOIDC    bool

	v *validator.Validate
}

type assumerOptions func(*assumer) error

func New(v *validator.Validate, opts ...assumerOptions) (*assumer, error) {
	a := &assumer{
		RoleSessionDuration: defaultRoleSessionDuration,
	}

	if v == nil {
		return nil, fmt.Errorf("error instantiating assumer: validator is nil")
	}
	a.v = v

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}
	if err := a.v.Struct(a); err != nil {
		return nil, err
	}

	if a.RoleSessionDuration > maxSessionDuration {
		return nil, fmt.Errorf("role session duration must be less than %d", maxSessionDuration)
	}
	if a.ExternalID != "" && (a.UseGithubOIDC || a.UseGCPOIDC) {
		return nil, fmt.Errorf("external ID cannot be used with OIDC")
	}

	return a, nil
}

func WithSettings(s Settings) assumerOptions {
	return func(a *assumer) error {
		if err := s.Validate(a.v); err != nil {
			return fmt.Errorf("settings are invalid: %w", err)
		}

		a.RoleARN = s.RoleARN
		a.RoleSessionName = s.RoleSessionName
		a.ExternalID = s.ExternalID
		a.TwoStepConfig = s.TwoStepConfig
		a.UseGithubOIDC = s.UseGithubOIDC
		a.UseGCPOIDC = s.UseGCPOIDC
		a.Region = s.Region

		if s.RoleSessionDuration > 0 {
			a.RoleSessionDuration = s.RoleSessionDuration
		}

		return nil
	}
}

func WithRoleARN(s string) assumerOptions {
	return func(a *assumer) error {
		a.RoleARN = s
		return nil
	}
}

func WithRoleSessionName(s string) assumerOptions {
	return func(a *assumer) error {
		a.RoleSessionName = s
		return nil
	}
}

func WithRoleSessionDuration(s time.Duration) assumerOptions {
	return func(a *assumer) error {
		a.RoleSessionDuration = s
		return nil
	}
}

func WithTwoStepConfig(s *TwoStepConfig) assumerOptions {
	return func(a *assumer) error {
		a.TwoStepConfig = s
		return nil
	}
}

func WithRegion(s string) assumerOptions {
	return func(a *assumer) error {
		a.Region = s
		return nil
	}
}
