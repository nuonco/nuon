# Testing in go

We have outlined the following styles for our `go` unit testing. Our general approach to unit testing is to balance coverage with effort and readability. We prefer to have more readable code when given a choice between higher test coverage and readability.

## all packages have at least one test

Each package should have at least a single unit test, even if it's an empty package. We do this so we can more easily track coverage across our entire code base.

This means that in the default case, we add a "noop" test:

```go
import "testing"

func TestNoop(t *testing.T) {}
```

And we can see the output, to measure coverage:

```go
~pkg/terraform (jm/styleguide*)$ go test -cover -count=1 ./...
ok      github.com/powertoolsdev/mono/pkg/terraform     0.335s  coverage: [no statements]
ok      github.com/powertoolsdev/mono/pkg/terraform/archive     0.432s  coverage: 0.0% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/archive/oci 0.156s  coverage: 70.5% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/archive/s3  0.313s  coverage: 69.6% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/backend     0.629s  coverage: 0.0% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/backend/s3  0.393s  coverage: 85.7% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/binary      0.474s  coverage: 0.0% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/binary/remote       0.766s  coverage: 72.7% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/run 0.908s  coverage: 15.3% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/variables   0.760s  coverage: 0.0% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/variables/static    0.982s  coverage: 40.0% of statements
ok      github.com/powertoolsdev/mono/pkg/terraform/workspace   0.752s  coverage: 44.4% of statements
```

## default to table driven tests

We default to table tests, and follow the guidance of Dave Cheney's [post](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests).

While table tests can sometimes be harder to read, we try to strike a balance by having all table tests structured almost the same.

```go
  // we prefer to have just a few "test-global" variables at the top
  foo := generics.GetFakeObj[Type]()

	tests := map[string]struct {
		// since clients often need to initialize clients, mocks and more we expose methods to handle that
		backendFn   func(*testing.T) *s3
		// when needed, a dedicated assert function is easier to reason about assertions.
		assertFn    func(*testing.T, []byte)
		errExpected error
	}{
		"creds backend": {
			backendFn: func(t *testing.T) *s3 {
				s, err := New(v,
					WithBucketConfig(bucketCfg),
					WithCredentials(staticCreds))
				assert.NoError(t, err)
				return s
			},
			assertFn: func(t *testing.T, byts []byte) {
				var resp map[string]interface{}
				err := json.Unmarshal(byts, &resp)
				assert.NoError(t, err)

				assert.Equal(t, staticCreds.Static.AccessKeyID, resp["access_key"])
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			backend := test.backendFn(t)

			cfg, err := backend.ConfigFile(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, cfg)
		})
	}

```

We prefer table driven tests, because once the initial test is working, we can more easily add additional test cases. Table driven tests also lead to less need of "helper" methods, which mean that the logic for a test is contained in a single place. This makes output easier to read when a failure happens (the call stack is in the same function, instead of a different place).

## standard workflow tests

We use `temporal's` built in test suite, which allows us to test workflow execution.

```go
func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := generics.GetFakeObj[workers.Config]()
	srv := server.NewWorkflow(cfg)
	run := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(run.ProvisionRunner)

	wf := NewWorkflow(cfg)
	a := NewActivities(nil)

	req := generics.GetFakeObj[*orgsv1.SignupRequest]()
	iamResp := generics.GetFakeObj[*iamv1.ProvisionIAMResponse]()
	kmsResp := generics.GetFakeObj[*kmsv1.ProvisionKMSResponse]()
	serverResp := generics.GetFakeObj[*serverv1.ProvisionServerResponse]()

	// Mock activity implementations
	env.OnActivity(a.SendNotification, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, snr SendNotificationRequest) (SendNotificationResponse, error) {
			return SendNotificationResponse{}, nil
		})

	env.OnWorkflow(kmser.ProvisionKMS, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *kmsv1.ProvisionKMSRequest) (*kmsv1.ProvisionKMSResponse, error) {
			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			assert.Equal(t, iamResp.SecretsRoleArn, r.SecretsIamRoleArn)
			return kmsResp, nil
		})

	env.ExecuteWorkflow(wf.Signup, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// assert response
	var resp orgsv1.SignupResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	assert.True(t, proto.Equal(resp.IamRoles, iamResp))
	assert.True(t, proto.Equal(resp.Server, serverResp))
}

```

We use our standard approach to splitting tests for activities.

## test validation code

Almost every method in our code base has a `Validate` method. We use use both [protoc-gen-validate](https://github.com/bufbuild/protoc-gen-validate) to generate methods and [go-validator](https://github.com/go-playground/validator) in all structs.

By validating structs are valid, we avoid many hard to debug situations where a value on a struct deep in a call stack is set to nil. A good example of this is how we test both workflow activity and child workflow requests:

```go
  env.OnWorkflow(kmser.ProvisionKMS, mock.Anything, mock.Anything).
    Return(func(_ workflow.Context, r *kmsv1.ProvisionKMSRequest) (*kmsv1.ProvisionKMSResponse, error) {
      assert.Nil(t, r.Validate())
      assert.Equal(t, req.OrgId, r.OrgId)
      assert.Equal(t, iamResp.SecretsRoleArn, r.SecretsIamRoleArn)
      return kmsResp, nil
    })

  env.OnWorkflow(iamer.ProvisionIAM, mock.Anything, mock.Anything).
    Return(func(_ workflow.Context, r *iamv1.ProvisionIAMRequest) (*iamv1.ProvisionIAMResponse, error) {
      assert.Nil(t, r.Validate())
      return iamResp, nil
    })

  env.OnWorkflow(srv.ProvisionServer, mock.Anything, mock.Anything).
    Return(func(_ workflow.Context, r *serverv1.ProvisionServerRequest) (*serverv1.ProvisionServerResponse, error) {
      assert.Nil(t, r.Validate())
      assert.Equal(t, req.OrgId, r.OrgId)
      return serverResp, nil
    })
```

## use mockgen

We use `mockgen` in almost all tests.

## avoid test helpers

We try to avoid tests that have helper methods, to keep logic in a test contained in a single place. This means that an engineer can reason about the test code without having to trace a code path throughout different parts of the codebase.

This was derived from Mitchell Hashimoto's talk [Advanced Testing in Go](https://www.youtube.com/watch?v=8hQG7QlcLBk).

## pkg's expose mock types

We expose `mock`s where we define packages internally. While this might be slightly orthogonal to the advice of "interface" + "mock" where you use a package, our belief is that internal code is tightly coupled regardless and optimizing for isolation for internal codepaths is not in our best interest.

Thus, many packages expose an interface and a mock, that can be used:

```go
import (
  "context"
  "io"
)

// Archive package exposes methods for loading a workspace archive
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=archive_mock.go -source=archive.go -package=archive
type Archive interface {
  // Init should be used for fetching things from s3, or setting up credentials
  Init(context.Context) error

  // Unpack is used to unpack an archive, and should call the unpackFn with each source file
  Unpack(context.Context, Callback) error
}
```

## decouple initialization from methods, for testing

We decouple initialization code from logic code to make testing easier. Generally, a well written function will have a small bit of initialization that is easily detectable as in a bad state - and focus most of the logic in a method that accepts a client:

```go
func (s *s3Downloader) GetBlob(ctx context.Context, key string) ([]byte, error) {
  client, err := s.getClient(ctx)
  if err != nil {
    return nil, err
  }

  downloader := manager.NewDownloader(client)
  return s.getBlob(ctx, downloader, key)
}

type s3BlobGetter interface {
  Download(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*manager.Downloader)) (int64, error)
}

func (s *s3Downloader) getBlob(ctx context.Context, client s3BlobGetter, key string) ([]byte, error) {
  buf := aws.NewWriteAtBuffer([]byte{})
  _, err := client.Download(ctx, buf, &s3.GetObjectInput{
    Bucket: generics.ToPtr(s.Bucket),
    Key:  generics.ToPtr(key),
  })
  if err != nil {
    return nil, fmt.Errorf("unable to download bytes: key=%s: %w", key, err)
  }

  return buf.Bytes(), err
}
```

## avoid struct attributes just for testing

We avoid adding struct attributes, just for testing. This usually makes the code harder to read and the tradeoff in the additional test coverage is not worth the effort.

## don't test initialization code

Testing initialization code is really hard, and requires significant additional effort to interface wrap functions. It often leads to less desirable/readable code, and doesn't actually make the code more robust.

Instead, we break functions down into two parts, and apply a few principles:

* functional options always error out immediately
* functional options are tested
* break "logic" code and "initialization" code into two functions, and test the logic with mocks

Note the following function. We consider "fetching" credentials, and creating a client as initialization code. Adding tests to those calls would require additional interfaces, for code that is a stable "hot path" (used almost everywhere), and not doing any real logic.

We can test `getAuthorizationData` with a mock to enssure the request is created correctly, and the response parsed correctly.

```go
func (e *ecrAuthorizer) GetAuthorization(ctx context.Context) (*Authorization, error) {
  cfg, err := credentials.Fetch(ctx, e.Credentials)
  if err != nil {
    return nil, fmt.Errorf("unable to get credentials: %w", err)
  }

  ecrClient := ecr.NewFromConfig(cfg)
  authData, err := e.getAuthorizationData(ctx, ecrClient)
  if err != nil {
    return nil, fmt.Errorf("unable to get ecr authorization token: %w", err)
  }

  return e.parseAuthorizationData(authData)
}

```

## test functional option code

We heavily use functional options throughout our codebase. On our previous point about not testing initialization code, we have found that the tradeoff of testing functional options + not testing intialization code makes for a good tradeoff in effort / coverage.

Note the following test:

```go
func TestNew(t *testing.T) {
  auth := generics.GetFakeObj[*Auth]()
  img := generics.GetFakeObj[*Image]()
  v := validator.New()

  tests := map[string]struct {
    errExpected error
    optsFn      func() []ociOption
    assertFn    func(*testing.T, *oci)
  }{
    "happy path": {
      optsFn: func() []ociOption {
        return []ociOption{
          WithAuth(auth),
          WithImage(img),
        }
      },
      assertFn: func(t *testing.T, s *oci) {
        assert.Equal(t, auth, s.Auth)
        assert.Equal(t, img, s.Image)
      },
    },
    "missing image": {
      optsFn: func() []ociOption {
        return []ociOption{
          WithAuth(auth),
        }
      },
      errExpected: fmt.Errorf("Image"),
    },
  }

  for name, test := range tests {
    name := name
    test := test
    t.Run(name, func(t *testing.T) {
      e, err := New(v, test.optsFn()...)
      if test.errExpected != nil {
        assert.Error(t, err)
        assert.ErrorContains(t, err, test.errExpected.Error())
        return
      }
      assert.NoError(t, err)
      test.assertFn(t, e)
    })
  }
}
```

These tests are almost always similarly structured, and allow us to easily verify the state of a struct in the test lifecyle.

## gate integration tests with env vars, not build flags

We gate all of our integration tests with env vars, instead of build flags. Env vars are easier on average, and allow engineers to more easily set them.

We use the `INTEGRATION` flag by default.

```go
func TestGetSandboxVersionByID(t *testing.T) {
  integration := os.Getenv("INTEGRATION")
  if integration == "" {
    t.Skip("INTEGRATION=true must be set in environment to run.")
  }
```

Inspired by Peter Bourgon's post: https://peter.bourgon.org/blog/2021/04/02/dont-use-build-tags-for-integration-tests.html

## Use generics + faker

We heavily rely on _good_ fake data when writing tests. We do this from day 1 on any new functionality, as it allows us to have faster, more robust unit tests + makes further updates easier.

`pkg/generics` exposes a function called `generics.GetFakeObj` that we use in most test cases:

```go
type MyStruct struct {
  Foo Foo `faker:"foo"`
  Val string `faker:"shortID"`
}
```

We define custom faker methods to allow us to create real, usable types.
