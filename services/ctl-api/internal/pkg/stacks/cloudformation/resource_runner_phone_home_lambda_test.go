package cloudformation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func phoneHomeTestInput(installID string) *stacks.TemplateInput {
	return &stacks.TemplateInput{
		Install: &app.Install{ID: installID},
		AppCfg:  &app.AppConfig{},
		CloudFormationStackVersion: &app.InstallStackVersion{
			PhoneHomeURL: "https://example.com/phone-home",
		},
	}
}

const (
	testPhoneHomeSecretARN = "arn:aws:secretsmanager:us-west-2:123456789012:secret:nuon/phone-home/inst1-aB3xYz"
	testPhoneHomeCMKARN    = "arn:aws:kms:us-west-2:123456789012:key/abcd-1234"
)

func phoneHomeAuthTestInput(installID string) *stacks.TemplateInput {
	inp := phoneHomeTestInput(installID)
	inp.CloudFormationStackVersion.PhoneHomeID = "phv7g2k9x4m1qz8w3n6b5t0jrc"
	inp.PhoneHomeSecretARN = testPhoneHomeSecretARN
	inp.PhoneHomeSecretRegion = "us-west-2"

	return inp
}

// The Lambda fetches its token at invocation time, so the template only has to carry
// the secret's location and which map entry to read.
func TestGetRunnerPhoneHomeLambda_SecretEnvVars(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")

	fn := tpl.getRunnerPhoneHomeLambda(inp, tagBuilder{installID: inp.Install.ID})

	require.NotNil(t, fn.Environment)
	assert.Equal(t, testPhoneHomeSecretARN, fn.Environment.Variables["NUON_PHONE_HOME_SECRET_ARN"])
	// The region of the secret (Nuon's management region), not the install's.
	assert.Equal(t, "us-west-2", fn.Environment.Variables["NUON_PHONE_HOME_SECRET_REGION"])
	assert.Equal(t, "phv7g2k9x4m1qz8w3n6b5t0jrc", fn.Environment.Variables["NUON_PHONE_HOME_ID"])
}

// Backwards compatibility: an install without phone-home auth must render exactly as
// before, so the script's "send the header only when the env vars are present" branch
// keeps one script version serving both flagged and unflagged orgs.
func TestGetRunnerPhoneHomeLambda_NoEnvWithoutSecret(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	fn := tpl.getRunnerPhoneHomeLambda(inp, tagBuilder{installID: inp.Install.ID})

	assert.Nil(t, fn.Environment, "no phone-home secret means no environment block")
}

// The identity half of the cross-account read. Without kms:Decrypt the read fails no
// matter how permissive the secret's resource policy is, because the AWS-managed
// secretsmanager key cannot be shared cross-account.
func TestGetRunnerPhoneHomeLambdaRole_SecretPolicy(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{AWSPhoneHomeCMKARN: testPhoneHomeCMKARN}}
	inp := phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")

	role := tpl.getRunnerPhoneHomeLambdaRole(inp, tagBuilder{installID: inp.Install.ID})

	var policy *map[string]any
	for i := range role.Policies {
		if role.Policies[i].PolicyName == "PhoneHomeSecretPolicy" {
			doc := role.Policies[i].PolicyDocument.(map[string]any)
			policy = &doc
		}
	}
	require.NotNil(t, policy, "expected a PhoneHomeSecretPolicy inline policy")

	statements := (*policy)["Statement"].([]map[string]any)
	require.Len(t, statements, 2, "expected a GetSecretValue and a kms:Decrypt statement")

	// Scoped to this install's secret, not "*".
	assert.Equal(t, testPhoneHomeSecretARN, statements[0]["Resource"])
	assert.Equal(t, []string{"secretsmanager:GetSecretValue"}, statements[0]["Action"])
	assert.Equal(t, testPhoneHomeCMKARN, statements[1]["Resource"])
	assert.Equal(t, []string{"kms:Decrypt", "kms:DescribeKey"}, statements[1]["Action"])
}

// The CMK is out-of-repo infra that has not landed. Until it does, the role should
// still get its GetSecretValue grant rather than a statement naming an empty ARN.
func TestGetRunnerPhoneHomeLambdaRole_OmitsKMSWhenUnset(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")

	role := tpl.getRunnerPhoneHomeLambdaRole(inp, tagBuilder{installID: inp.Install.ID})

	for i := range role.Policies {
		if role.Policies[i].PolicyName != "PhoneHomeSecretPolicy" {
			continue
		}
		doc := role.Policies[i].PolicyDocument.(map[string]any)
		statements := doc["Statement"].([]map[string]any)
		assert.Len(t, statements, 1, "no CMK configured means no kms statement")

		return
	}
	t.Fatal("expected a PhoneHomeSecretPolicy inline policy")
}

// The secret ARN must reach the Lambda as an environment variable and nowhere else.
// phonehome.py does `props = data.pop("ResourceProperties")` and POSTs every prop, so
// an ARN added there would be echoed into the phone-home body and persisted to the
// install's stack outputs.
func TestGetRunnerPhoneHomeProps_DoesNotEchoSecretARN(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{AWSPhoneHomeCMKARN: testPhoneHomeCMKARN}}
	inp := phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")

	props := tpl.getRunnerPhoneHomeProps(inp, nil)

	require.NotNil(t, props)
	for key, value := range props.Properties {
		str, ok := value.(string)
		if !ok {
			continue
		}
		assert.NotContains(t, str, testPhoneHomeSecretARN, "property %q echoes the secret ARN", key)
		assert.NotContains(t, str, "secretsmanager", "property %q echoes a secrets manager reference", key)
	}
}

// Leaving these to CloudFormation's defaults (3s, 128MB) is what stranded the first
// enrolled install: the token fetch imports boto3 and builds a client, which at 128MB —
// roughly a twelfth of a vCPU — does not finish in 3 seconds. The function was killed
// before it reached Secrets Manager, so nothing was logged, no phone home was sent, and
// the stack sat on the custom resource until it rolled back.
//
// Both values are asserted against the script's own retry arithmetic rather than as bare
// literals, so shortening the timeout below what the ladder needs fails here instead of
// in a customer account.
func TestGetRunnerPhoneHomeLambda_TimeoutAndMemory(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}

	for _, tc := range []struct {
		name string
		inp  *stacks.TemplateInput
	}{
		// Applies with or without auth: the retry ladder predates the token fetch, and
		// was equally unreachable under the 3s default.
		{"with phone home auth", phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")},
		{"without phone home auth", phoneHomeTestInput("instabcdefghijklmnopqrstuv")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fn := tpl.getRunnerPhoneHomeLambda(tc.inp, tagBuilder{installID: tc.inp.Install.ID})

			require.NotNil(t, fn.Timeout, "unset means CloudFormation's 3s default, which is too short for the token fetch")
			require.NotNil(t, fn.MemorySize, "unset means CloudFormation's 128MB default, which is too little CPU to import boto3 in time")

			// phonehome.py: MAX_RETRIES=5, BASE_DELAY=1.75, delay = BASE_DELAY * 2**attempt
			// sleeps after attempts 1..4 => 1.75 * (1+2+4+8) = 26.25s
			const scriptRetrySleepSeconds = 26.25
			assert.Greater(t, float64(*fn.Timeout), scriptRetrySleepSeconds,
				"timeout must clear the script's %.2fs of retry sleeps, or the ladder can never finish",
				scriptRetrySleepSeconds)

			assert.GreaterOrEqual(t, *fn.MemorySize, 512,
				"below ~512MB the boto3 import is slow enough to threaten the timeout again")
		})
	}
}

// An oversized script otherwise fails at CreateStack inside the customer's account,
// where the cause is far harder to attribute.
func TestValidatePhoneHomeScriptSize(t *testing.T) {
	assert.NoError(t, validatePhoneHomeScriptSize(strings.Repeat("x", lambdaInlineCodeLimit)))

	err := validatePhoneHomeScriptSize(strings.Repeat("x", lambdaInlineCodeLimit+1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inline lambda source")
}

// The role name is load-bearing: the secret's resource policy in the management account
// pins its grant to this exact ARN via an aws:PrincipalArn condition, and a condition
// that matches nothing denies silently — so a mismatch surfaces only as an
// AccessDeniedException at phone-home time, never as an error at policy-write time.
// Assert the literal.
//
// (The policy names the account root rather than this role, because Secrets Manager
// validates principals and the role does not exist until this stack creates it. See
// secretsmanager.PhoneHomeResourcePolicy.)
func TestGetRunnerPhoneHomeLambdaRole_DeterministicName(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	role := tpl.getRunnerPhoneHomeLambdaRole(inp, tagBuilder{installID: inp.Install.ID})

	require.NotNil(t, role.RoleName)
	assert.Equal(t, "instabcdefghijklmnopqrstuv-phone-home", *role.RoleName)
	assert.Equal(t, stacks.PhoneHomeRoleName(inp.Install.ID), *role.RoleName)
}

// IAM caps role names at 64 characters. Install IDs are 26, so the suffix leaves
// plenty of room — this guards against the suffix growing past the limit.
func TestPhoneHomeRoleName_WithinIAMLimit(t *testing.T) {
	name := stacks.PhoneHomeRoleName(strings.Repeat("i", 26))

	assert.LessOrEqual(t, len(name), 64,
		"IAM role names cannot exceed 64 characters")
	assert.Equal(t, strings.Repeat("i", 26)+"-phone-home", name)
}

// The role name must not leak into the phone-home payload. phonehome.py POSTs every
// ResourceProperty it receives, and anything echoed there lands in the install's
// stack outputs.
func TestGetRunnerPhoneHomeProps_DoesNotEchoRoleName(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	props := tpl.getRunnerPhoneHomeProps(inp, nil)

	require.NotNil(t, props)
	for key, value := range props.Properties {
		if str, ok := value.(string); ok {
			assert.NotContains(t, str, "-phone-home",
				"property %q echoes the phone-home role name", key)
		}
	}
}
