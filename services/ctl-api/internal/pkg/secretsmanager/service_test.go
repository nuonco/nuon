package secretsmanager

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssm "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

const testSecretARN = "arn:aws:secretsmanager:us-west-2:123456789012:secret:nuon/phone-home/inst1-aB3xYz"

// fakeAPI models a single secret's server-side state.
type fakeAPI struct {
	exists          bool
	value           string
	pendingDeletion bool

	creates  int
	puts     int
	gets     int
	restores int
	policies int
	deletes  int

	createErr error
}

func (f *fakeAPI) DescribeSecret(_ context.Context, in *awssm.DescribeSecretInput, _ ...func(*awssm.Options)) (*awssm.DescribeSecretOutput, error) {
	if !f.exists {
		return nil, &types.ResourceNotFoundException{}
	}

	out := &awssm.DescribeSecretOutput{ARN: aws.String(testSecretARN)}
	if f.pendingDeletion {
		out.DeletedDate = aws.Time(time.Now())
	}

	return out, nil
}

func (f *fakeAPI) GetSecretValue(_ context.Context, in *awssm.GetSecretValueInput, _ ...func(*awssm.Options)) (*awssm.GetSecretValueOutput, error) {
	f.gets++
	if !f.exists {
		return nil, &types.ResourceNotFoundException{}
	}

	return &awssm.GetSecretValueOutput{SecretString: aws.String(f.value)}, nil
}

func (f *fakeAPI) CreateSecret(_ context.Context, in *awssm.CreateSecretInput, _ ...func(*awssm.Options)) (*awssm.CreateSecretOutput, error) {
	f.creates++
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.exists = true
	f.value = aws.ToString(in.SecretString)

	return &awssm.CreateSecretOutput{ARN: aws.String(testSecretARN)}, nil
}

func (f *fakeAPI) PutSecretValue(_ context.Context, in *awssm.PutSecretValueInput, _ ...func(*awssm.Options)) (*awssm.PutSecretValueOutput, error) {
	f.puts++
	f.exists = true
	f.value = aws.ToString(in.SecretString)

	return &awssm.PutSecretValueOutput{ARN: aws.String(testSecretARN)}, nil
}

func (f *fakeAPI) PutResourcePolicy(_ context.Context, in *awssm.PutResourcePolicyInput, _ ...func(*awssm.Options)) (*awssm.PutResourcePolicyOutput, error) {
	f.policies++

	return &awssm.PutResourcePolicyOutput{}, nil
}

func (f *fakeAPI) DeleteSecret(_ context.Context, in *awssm.DeleteSecretInput, _ ...func(*awssm.Options)) (*awssm.DeleteSecretOutput, error) {
	f.deletes++
	if !f.exists {
		return nil, &types.ResourceNotFoundException{}
	}
	f.exists = false

	return &awssm.DeleteSecretOutput{}, nil
}

func (f *fakeAPI) RestoreSecret(_ context.Context, in *awssm.RestoreSecretInput, _ ...func(*awssm.Options)) (*awssm.RestoreSecretOutput, error) {
	f.restores++
	f.pendingDeletion = false

	return &awssm.RestoreSecretOutput{}, nil
}

func testService(fake *fakeAPI) *service {
	return &service{
		cfg:    &internal.Config{ManagementRegion: "us-west-2"},
		l:      zap.NewNop(),
		newAPI: func(context.Context) (api, error) { return fake, nil },
	}
}

func TestEnsureSecretCreatesWhenMissing(t *testing.T) {
	fake := &fakeAPI{}
	svc := testService(fake)

	out, err := svc.EnsureSecret(context.Background(), EnsureSecretInput{Name: "nuon/phone-home/inst1", Value: `{"a":"t1"}`})
	if err != nil {
		t.Fatalf("EnsureSecret: %v", err)
	}

	if fake.creates != 1 || fake.puts != 0 {
		t.Errorf("expected one create and no put, got creates=%d puts=%d", fake.creates, fake.puts)
	}
	// The ARN is not derivable from the name, so it has to come back from AWS.
	if out.ARN != testSecretARN {
		t.Errorf("ARN = %q, want %q", out.ARN, testSecretARN)
	}
	if !out.Wrote {
		t.Error("Wrote should be true on create")
	}
}

// The regression test for the idempotency guard. Without it, every stack generation
// across all four provisioning workflows mints a new Secrets Manager version — a
// cost and noise problem, and it makes version history useless for auditing.
func TestEnsureSecretSkipsUnchangedValue(t *testing.T) {
	fake := &fakeAPI{exists: true, value: `{"a":"t1"}`}
	svc := testService(fake)

	out, err := svc.EnsureSecret(context.Background(), EnsureSecretInput{Name: "nuon/phone-home/inst1", Value: `{"a":"t1"}`})
	if err != nil {
		t.Fatalf("EnsureSecret: %v", err)
	}

	if fake.puts != 0 {
		t.Errorf("an unchanged map must not write a new version, got %d puts", fake.puts)
	}
	if out.Wrote {
		t.Error("Wrote should be false when the value is unchanged")
	}
	if out.ARN != testSecretARN {
		t.Errorf("ARN = %q, want %q", out.ARN, testSecretARN)
	}
}

func TestEnsureSecretWritesChangedValue(t *testing.T) {
	fake := &fakeAPI{exists: true, value: `{"a":"t1"}`}
	svc := testService(fake)

	out, err := svc.EnsureSecret(context.Background(), EnsureSecretInput{Name: "nuon/phone-home/inst1", Value: `{"a":"t1","b":"t2"}`})
	if err != nil {
		t.Fatalf("EnsureSecret: %v", err)
	}

	if fake.puts != 1 || fake.creates != 0 {
		t.Errorf("expected one put and no create, got puts=%d creates=%d", fake.puts, fake.creates)
	}
	if !out.Wrote {
		t.Error("Wrote should be true when the value changed")
	}
	if fake.value != `{"a":"t1","b":"t2"}` {
		t.Errorf("stored value = %q", fake.value)
	}
}

// AWS's default 7-30 day recovery window otherwise makes re-provisioning the same
// install ID fail with InvalidRequestException forever.
func TestEnsureSecretRestoresPendingDeletion(t *testing.T) {
	fake := &fakeAPI{exists: true, value: "{}", pendingDeletion: true}
	svc := testService(fake)

	if _, err := svc.EnsureSecret(context.Background(), EnsureSecretInput{Name: "nuon/phone-home/inst1", Value: `{"a":"t1"}`}); err != nil {
		t.Fatalf("EnsureSecret: %v", err)
	}

	if fake.restores != 1 {
		t.Errorf("expected the secret to be restored, got %d restores", fake.restores)
	}
	if fake.puts != 1 {
		t.Errorf("expected the value to be written after restore, got %d puts", fake.puts)
	}
}

// Two installs in the same target account can provision concurrently, so losing the
// describe/create race must fall through to the update path rather than erroring.
func TestEnsureSecretHandlesCreateRace(t *testing.T) {
	fake := &fakeAPI{createErr: &types.ResourceExistsException{}}
	svc := testService(fake)

	out, err := svc.EnsureSecret(context.Background(), EnsureSecretInput{Name: "nuon/phone-home/inst1", Value: `{"a":"t1"}`})
	if err != nil {
		t.Fatalf("EnsureSecret should recover from a create race: %v", err)
	}

	if fake.creates != 1 || fake.puts != 1 {
		t.Errorf("expected a failed create then a put, got creates=%d puts=%d", fake.creates, fake.puts)
	}
	if !out.Wrote {
		t.Error("Wrote should be true after the fallback put")
	}
}

func TestDeleteSecretIsIdempotent(t *testing.T) {
	fake := &fakeAPI{}
	svc := testService(fake)

	if err := svc.DeleteSecret(context.Background(), testSecretARN); err != nil {
		t.Fatalf("deleting an absent secret should not error: %v", err)
	}
}

func TestUnsupportedCloudSurfacesSentinel(t *testing.T) {
	// Azure-hosted: no federation path to AWS, so ManagementSecretsCreds returns nil.
	svc := NewService(&internal.Config{CloudProvider: "azure", ManagementRegion: "us-west-2"}, zap.NewNop())

	_, err := svc.EnsureSecret(context.Background(), EnsureSecretInput{Name: "n", Value: "{}"})
	if err == nil {
		t.Fatal("expected an error when the control plane cannot reach secrets manager")
	}
	if !strings.Contains(err.Error(), ErrUnsupportedCloud.Error()) {
		t.Errorf("expected ErrUnsupportedCloud, got %v", err)
	}
}

func TestPhoneHomeSecretName(t *testing.T) {
	if got, want := PhoneHomeSecretName("inst123"), "nuon/phone-home/inst123"; got != want {
		t.Errorf("PhoneHomeSecretName = %q, want %q", got, want)
	}
}

// The role named here does not exist when the policy is first applied — it is
// created by the CloudFormation stack the customer has yet to run — so a typo fails
// silently as an AccessDeniedException at phone-home time. Hence pinning the shape.
func TestPhoneHomeResourcePolicy(t *testing.T) {
	raw, err := PhoneHomeResourcePolicy("123456789012", "inst123-phone-home")
	if err != nil {
		t.Fatalf("PhoneHomeResourcePolicy: %v", err)
	}

	var policy struct {
		Statement []struct {
			Effect    string            `json:"Effect"`
			Principal map[string]string `json:"Principal"`
			Action    []string          `json:"Action"`
		} `json:"Statement"`
	}
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		t.Fatalf("unmarshal policy: %v", err)
	}

	if len(policy.Statement) != 1 {
		t.Fatalf("expected exactly one statement, got %d", len(policy.Statement))
	}
	stmt := policy.Statement[0]

	if stmt.Effect != "Allow" {
		t.Errorf("Effect = %q", stmt.Effect)
	}
	if want := "arn:aws:iam::123456789012:role/inst123-phone-home"; stmt.Principal["AWS"] != want {
		t.Errorf("Principal = %q, want %q", stmt.Principal["AWS"], want)
	}
	// Read-only, and only the one action: this grant reaches into a customer account.
	if len(stmt.Action) != 1 || stmt.Action[0] != "secretsmanager:GetSecretValue" {
		t.Errorf("Action = %v, want only secretsmanager:GetSecretValue", stmt.Action)
	}
}
