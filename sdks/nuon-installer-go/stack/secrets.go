package stack

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// secretAlphabet matches install-stacks/aws/secrets.tf random_password
// (`special = false, length = 63`). Why: app templates may regex over the
// secret value; aligning charset across CFN/TF/SDK keeps that contract.
const secretAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const secretLength = 63

// EnsureSecrets creates auto-generated and customer-provided secrets in
// AWS Secrets Manager. Idempotent: existing secrets keep their stored value
// (we only PutSecretValue when we just created the secret, mirroring TF's
// `lifecycle { ignore_changes = [secret_string] }`).
func EnsureSecrets(ctx context.Context, log *slog.Logger, c *secretsmanager.Client, st *State, cfg *Config) error {
	if st.SecretARNs == nil {
		st.SecretARNs = map[string]string{}
	}
	prefix := cfg.Prefix()

	// Sort for deterministic ordering — stable retries, stable logs.
	auto := append([]string(nil), cfg.AutoGenerateSecrets...)
	sort.Strings(auto)
	for _, name := range auto {
		val, err := randomPassword()
		if err != nil {
			return err
		}
		arn, err := ensureSecret(ctx, log, c, prefix+"-"+name, val, st.InstallID)
		if err != nil {
			return err
		}
		st.SecretARNs[name+"_arn"] = arn
	}

	customerKeys := make([]string, 0, len(cfg.Secrets))
	for k := range cfg.Secrets {
		customerKeys = append(customerKeys, k)
	}
	sort.Strings(customerKeys)
	for _, name := range customerKeys {
		v := cfg.Secrets[name]
		arn, err := ensureSecret(ctx, log, c, prefix+"-"+name, v.Value, st.InstallID)
		if err != nil {
			return err
		}
		st.SecretARNs[name+"_arn"] = arn
	}
	return st.Save()
}

// ensureSecret returns the ARN of an AWS Secrets Manager secret with the
// given name, creating it (and seeding its value) if absent. Existing
// secrets keep their current value to match TF's ignore_changes contract.
func ensureSecret(ctx context.Context, log *slog.Logger, c *secretsmanager.Client, name, value, installID string) (string, error) {
	d, err := c.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{SecretId: &name})
	if err == nil {
		return aws.ToString(d.ARN), nil
	}
	if !IsAWSErrCode(err, "ResourceNotFoundException") {
		return "", fmt.Errorf("describe secret %s: %w", name, err)
	}
	cr, err := c.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         &name,
		SecretString: &value,
		Tags: []smtypes.Tag{
			{Key: aws.String(installIDTagKey), Value: aws.String(installID)},
		},
	})
	if err != nil {
		// Lost race: another invocation created it concurrently.
		if IsAWSErrCode(err, "ResourceExistsException") {
			d, derr := c.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{SecretId: &name})
			if derr != nil {
				return "", fmt.Errorf("describe after race %s: %w", name, derr)
			}
			return aws.ToString(d.ARN), nil
		}
		return "", fmt.Errorf("create secret %s: %w", name, err)
	}
	log.Info("created secret", "name", name)
	return aws.ToString(cr.ARN), nil
}

// randomPassword draws `secretLength` characters from `secretAlphabet`
// using crypto/rand. Why crypto/rand: math/rand is deterministic given
// process state and unsuitable for credential material.
func randomPassword() (string, error) {
	out := make([]byte, secretLength)
	max := big.NewInt(int64(len(secretAlphabet)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("rand: %w", err)
		}
		out[i] = secretAlphabet[n.Int64()]
	}
	return string(out), nil
}

// DeleteSecrets schedules deletion of every secret created for this install.
// Uses ForceDeleteWithoutRecovery so a re-provision can recreate by name
// without waiting out the recovery window.
func DeleteSecrets(ctx context.Context, log *slog.Logger, c *secretsmanager.Client, st *State, cfg *Config) error {
	prefix := cfg.Prefix()
	names := []string{}
	for _, n := range cfg.AutoGenerateSecrets {
		names = append(names, prefix+"-"+n)
	}
	for k := range cfg.Secrets {
		names = append(names, prefix+"-"+k)
	}
	for _, name := range names {
		_, err := c.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(name),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
		if err != nil && !IsAWSErrCode(err, "ResourceNotFoundException") {
			log.Warn("delete secret", "name", name, "err", err)
		}
	}
	st.SecretARNs = nil
	return st.Save()
}
