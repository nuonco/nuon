// Package secretsmanager manages the phone-home secret in the management account's
// AWS Secrets Manager.
//
// The secret always lives in AWS regardless of which cloud the control plane runs
// on, because the reader is the customer's phone-home Lambda and Secrets Manager is
// the only store it can reach. Which credentials get us there is decided once, by
// internal.Config.ManagementSecretsCreds.
package secretsmanager

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssm "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// ErrUnsupportedCloud is returned when this control plane has no path to the
// management account's Secrets Manager. Callers treat it as "skip", not "fail".
var ErrUnsupportedCloud = errors.New("control plane cannot reach management secrets manager")

// IsPermanentInputError reports whether AWS rejected the request content, so a
// retry can only fail the same way.
func IsPermanentInputError(err error) bool {
	var malformed *types.MalformedPolicyDocumentException
	var invalidParam *types.InvalidParameterException
	var invalidReq *types.InvalidRequestException

	return errors.As(err, &malformed) ||
		errors.As(err, &invalidParam) ||
		errors.As(err, &invalidReq)
}

// Service manages secrets in the management account.
type Service interface {
	// EnsureSecret creates the secret or updates its value, and returns the full
	// ARN. The ARN is not derivable from the name — AWS appends a random 6-char
	// suffix and cross-account reads reject a bare name — so callers must persist
	// what this returns.
	EnsureSecret(ctx context.Context, input EnsureSecretInput) (*EnsureSecretOutput, error)

	// PutResourcePolicy replaces the secret's resource policy. Full replacement,
	// not a merge.
	PutResourcePolicy(ctx context.Context, secretID, policy string) error

	// DeleteSecret removes the secret without a recovery window. The default 7-30
	// day window would make re-provisioning the same install ID fail with
	// InvalidRequestException.
	DeleteSecret(ctx context.Context, secretID string) error
}

// api is the subset of the Secrets Manager SDK this package uses, so tests can
// substitute a fake.
type api interface {
	DescribeSecret(context.Context, *awssm.DescribeSecretInput, ...func(*awssm.Options)) (*awssm.DescribeSecretOutput, error)
	GetSecretValue(context.Context, *awssm.GetSecretValueInput, ...func(*awssm.Options)) (*awssm.GetSecretValueOutput, error)
	CreateSecret(context.Context, *awssm.CreateSecretInput, ...func(*awssm.Options)) (*awssm.CreateSecretOutput, error)
	UpdateSecret(context.Context, *awssm.UpdateSecretInput, ...func(*awssm.Options)) (*awssm.UpdateSecretOutput, error)
	PutSecretValue(context.Context, *awssm.PutSecretValueInput, ...func(*awssm.Options)) (*awssm.PutSecretValueOutput, error)
	PutResourcePolicy(context.Context, *awssm.PutResourcePolicyInput, ...func(*awssm.Options)) (*awssm.PutResourcePolicyOutput, error)
	DeleteSecret(context.Context, *awssm.DeleteSecretInput, ...func(*awssm.Options)) (*awssm.DeleteSecretOutput, error)
	RestoreSecret(context.Context, *awssm.RestoreSecretInput, ...func(*awssm.Options)) (*awssm.RestoreSecretOutput, error)
	TagResource(context.Context, *awssm.TagResourceInput, ...func(*awssm.Options)) (*awssm.TagResourceOutput, error)
}

type EnsureSecretInput struct {
	Name        string
	Value       string
	Description string
	// KMSKeyARN encrypts the secret. When empty the AWS-managed key is used, which
	// cannot be read cross-account — acceptable only before the shared CMK exists.
	KMSKeyARN string
	// Tags are applied on create and reconciled on every later call, because
	// CreateSecret is the only call that accepts them — secrets provisioned before a
	// tag was added would otherwise never get it. Only added and updated, never
	// removed: something outside this reconciler may have tagged the secret for cost
	// allocation or policy and deleting those would be a surprise.
	Tags map[string]string
}

type EnsureSecretOutput struct {
	ARN    string
	Region string
	// Wrote reports whether a new secret version was actually written. False means
	// the stored value already matched, which is the common case across repeated
	// stack generations.
	Wrote bool
}

type service struct {
	cfg *internal.Config
	l   *zap.Logger

	// newAPI is overridden in tests.
	newAPI func(ctx context.Context) (api, error)
}

func NewService(cfg *internal.Config, l *zap.Logger) Service {
	s := &service{cfg: cfg, l: l}
	s.newAPI = s.newAWSAPI

	return s
}

func (s *service) newAWSAPI(ctx context.Context) (api, error) {
	credsCfg := s.cfg.ManagementSecretsCreds()
	if credsCfg == nil {
		return nil, ErrUnsupportedCloud
	}

	// Assumed-role credentials expire, so they are fetched per call rather than
	// held for process lifetime; CacheID makes repeated calls in one reconcile
	// reuse the same session.
	credsCfg.CacheID = "phone-home-secrets"

	awsCfg, err := credentials.Fetch(ctx, credsCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch credentials: %w", err)
	}

	return awssm.NewFromConfig(awsCfg), nil
}

func (s *service) EnsureSecret(ctx context.Context, input EnsureSecretInput) (*EnsureSecretOutput, error) {
	client, err := s.newAPI(ctx)
	if err != nil {
		return nil, err
	}

	out := &EnsureSecretOutput{Region: s.cfg.ManagementRegion}

	described, err := client.DescribeSecret(ctx, &awssm.DescribeSecretInput{
		SecretId: aws.String(input.Name),
	})
	switch {
	case err != nil && isNotFound(err):
		created, cerr := client.CreateSecret(ctx, &awssm.CreateSecretInput{
			Name:         aws.String(input.Name),
			SecretString: aws.String(input.Value),
			Description:  stringOrNil(input.Description),
			KmsKeyId:     stringOrNil(input.KMSKeyARN),
			Tags:         awsTags(input.Tags),
		})
		// Lost a race with a concurrent provision; fall through to the update path.
		if cerr != nil && isAlreadyExists(cerr) {
			described, err = client.DescribeSecret(ctx, &awssm.DescribeSecretInput{
				SecretId: aws.String(input.Name),
			})
			if err != nil {
				return nil, fmt.Errorf("unable to describe secret after create race: %w", err)
			}
			break
		}
		if cerr != nil {
			return nil, fmt.Errorf("unable to create secret: %w", cerr)
		}

		out.ARN = aws.ToString(created.ARN)
		out.Wrote = true

		return out, nil

	case err != nil:
		return nil, fmt.Errorf("unable to describe secret: %w", err)
	}

	arn := aws.ToString(described.ARN)
	// A pending deletion must be undone before any update, otherwise Secrets Manager
	// rejects KMS, tag, and value changes with InvalidRequestException.
	if described.DeletedDate != nil {
		if _, rerr := client.RestoreSecret(ctx, &awssm.RestoreSecretInput{
			SecretId: aws.String(arn),
		}); rerr != nil {
			return nil, fmt.Errorf("unable to restore secret pending deletion: %w", rerr)
		}
		s.l.Info("restored phone home secret that was pending deletion", zap.String("secret_arn", arn))
	}

	if err := s.reconcileKMSKey(ctx, client, arn, aws.ToString(described.KmsKeyId), input.KMSKeyARN); err != nil {
		return nil, err
	}

	if err := s.reconcileTags(ctx, client, arn, input.Tags, described.Tags); err != nil {
		return nil, err
	}

	return s.putExisting(ctx, client, input, arn, out)
}

func (s *service) reconcileKMSKey(ctx context.Context, client api, secretID, current, desired string) error {
	if desired == "" || current == desired {
		return nil
	}

	if _, err := client.UpdateSecret(ctx, &awssm.UpdateSecretInput{
		SecretId: aws.String(secretID),
		KmsKeyId: aws.String(desired),
	}); err != nil {
		return fmt.Errorf("unable to update secret KMS key: %w", err)
	}

	return nil
}

// putExisting writes the value only when it differs from what is stored. Without
// this guard every stack generation across all four provisioning workflows would
// mint a new Secrets Manager version.
func (s *service) putExisting(
	ctx context.Context, client api, input EnsureSecretInput, secretID string, out *EnsureSecretOutput,
) (*EnsureSecretOutput, error) {
	out.ARN = secretID

	current, err := client.GetSecretValue(ctx, &awssm.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	switch {
	case err != nil && !isNotFound(err) && !isNoValue(err):
		return nil, fmt.Errorf("unable to read current secret value: %w", err)
	case err == nil && aws.ToString(current.SecretString) == input.Value:
		return out, nil
	}

	put, err := client.PutSecretValue(ctx, &awssm.PutSecretValueInput{
		SecretId:     aws.String(secretID),
		SecretString: aws.String(input.Value),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to put secret value: %w", err)
	}

	if arn := aws.ToString(put.ARN); arn != "" {
		out.ARN = arn
	}
	out.Wrote = true

	return out, nil
}

func (s *service) PutResourcePolicy(ctx context.Context, secretID, policy string) error {
	client, err := s.newAPI(ctx)
	if err != nil {
		return err
	}

	if _, err := client.PutResourcePolicy(ctx, &awssm.PutResourcePolicyInput{
		SecretId:       aws.String(secretID),
		ResourcePolicy: aws.String(policy),
	}); err != nil {
		return fmt.Errorf("unable to put resource policy: %w", err)
	}

	return nil
}

func (s *service) DeleteSecret(ctx context.Context, secretID string) error {
	client, err := s.newAPI(ctx)
	if err != nil {
		return err
	}

	if _, err := client.DeleteSecret(ctx, &awssm.DeleteSecretInput{
		SecretId:                   aws.String(secretID),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	}); err != nil {
		if isNotFound(err) {
			return nil
		}

		return fmt.Errorf("unable to delete secret: %w", err)
	}

	return nil
}

func awsTags(tags map[string]string) []types.Tag {
	if len(tags) == 0 {
		return nil
	}

	// Sorted so the request is deterministic and two identical reconciles produce
	// identical calls, which matters for the fake in tests.
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]types.Tag, 0, len(keys))
	for _, k := range keys {
		out = append(out, types.Tag{Key: aws.String(k), Value: aws.String(tags[k])})
	}

	return out
}

// reconcileTags adds or corrects the tags this reconciler owns on an existing secret.
//
// Needed because CreateSecret is the only Secrets Manager call that takes tags, so a
// secret provisioned before a tag existed would never acquire it. Skips the call
// entirely when everything already matches — TagResource is cheap but it is one more
// API call on a path that runs on every stack generation.
func (s *service) reconcileTags(
	ctx context.Context, client api, secretID string, want map[string]string, have []types.Tag,
) error {
	if len(want) == 0 {
		return nil
	}

	current := make(map[string]string, len(have))
	for _, t := range have {
		current[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}

	missing := map[string]string{}
	for k, v := range want {
		if current[k] != v {
			missing[k] = v
		}
	}
	if len(missing) == 0 {
		return nil
	}

	if _, err := client.TagResource(ctx, &awssm.TagResourceInput{
		SecretId: aws.String(secretID),
		Tags:     awsTags(missing),
	}); err != nil {
		return fmt.Errorf("unable to tag secret: %w", err)
	}

	s.l.Info("updated phone home secret tags",
		zap.String("secret_id", secretID), zap.Int("tags", len(missing)))

	return nil
}

func isNotFound(err error) bool {
	var notFound *types.ResourceNotFoundException

	return errors.As(err, &notFound)
}

func isAlreadyExists(err error) bool {
	var exists *types.ResourceExistsException

	return errors.As(err, &exists)
}

// isNoValue covers a secret that exists with no current version, which
// GetSecretValue reports as an InvalidRequestException rather than a not-found.
func isNoValue(err error) bool {
	var invalid *types.InvalidRequestException

	return errors.As(err, &invalid)
}

func stringOrNil(s string) *string {
	if s == "" {
		return nil
	}

	return aws.String(s)
}
