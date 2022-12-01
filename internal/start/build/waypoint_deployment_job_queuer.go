package build

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

const (
	defaultJobTimeout string = "1h"
)

type QueueWaypointDeploymentJobRequest struct {
	OrgID                string `json:"org_id" validate:"required"`
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`

	BucketName   string `json:"bucket_name" validate:"required"`
	BucketPrefix string `json:"bucket_prefix" validate:"required"`

	AppID         string `json:"app_id" validate:"required"`
	DeploymentID  string `json:"deployment_id" validate:"required"`
	ComponentName string `json:"component_name" validate:"required"`
	ComponentType string `json:"component_type" validate:"required"`
}

func (w QueueWaypointDeploymentJobRequest) validate() error {
	validate := validator.New()
	return validate.Struct(w)
}

type QueueWaypointDeploymentJobResponse struct {
	JobID string `json:"job_id" validate:"required"`
}

var waypointBuildTmpl string = `
project = "{{.Project}}"

app "mario" {
  build {
    use "docker-pull" {
      image = "{{.InputImage}}"
      tag   = "{{.InputVersion}}"
    }

    registry {
      use "aws-ecr" {
	repository = "{{.OutputRepository}}"
	tag	 = "{{.OutputVersion}}"
	region = "us-west-2"
      }
    }
  }

  deploy {
    use "kubernetes" {}
  }
}
`

type buildTmplArgs struct {
	Project          string
	InputImage       string
	InputVersion     string
	OutputRepository string
	OutputVersion    string
}

func getWaypointHcl(req QueueWaypointDeploymentJobRequest) ([]byte, error) {
	tmpl, err := template.New("build-config").Parse(waypointBuildTmpl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse template: %w", err)
	}

	buf := new(bytes.Buffer)

	args := buildTmplArgs{
		Project:          req.AppID,
		InputImage:       "kennethreitz/httpbin",
		InputVersion:     "latest",
		OutputRepository: fmt.Sprintf("%s/%s", req.OrgID, req.AppID),
		OutputVersion:    req.DeploymentID,
	}

	if err := tmpl.Execute(buf, args); err != nil {
		return nil, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func (a *Activities) QueueWaypointDeploymentJob(
	ctx context.Context,
	req QueueWaypointDeploymentJobRequest,
) (QueueWaypointDeploymentJobResponse, error) {
	var resp QueueWaypointDeploymentJobResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return resp, err
	}

	s3Client := s3.NewFromConfig(cfg)

	waypointHcl, err := a.getWaypointHcl(ctx, s3Client, req)
	if err != nil {
		return resp, err
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	jobID, err := a.queueWaypointDeploymentJob(ctx, client, req, waypointHcl)
	if err != nil {
		return resp, fmt.Errorf("failed to create waypoint application: %w", err)
	}
	resp.JobID = jobID

	return resp, nil
}

type waypointDeploymentJobQueuer interface {
	queueWaypointDeploymentJob(
		context.Context,
		waypointClientJobQueuer,
		QueueWaypointDeploymentJobRequest,
		[]byte,
	) (string, error)
	getWaypointHcl(context.Context, s3ClientObjectGetter, QueueWaypointDeploymentJobRequest) ([]byte, error)
}

var _ waypointDeploymentJobQueuer = (*waypointDeploymentJobQueuerImpl)(nil)

type waypointDeploymentJobQueuerImpl struct{}

type waypointClientJobQueuer interface {
	QueueJob(ctx context.Context, in *gen.QueueJobRequest, opts ...grpc.CallOption) (*gen.QueueJobResponse, error)
}

type s3ClientObjectGetter interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func (w *waypointDeploymentJobQueuerImpl) getWaypointHcl(
	ctx context.Context,
	client s3ClientObjectGetter,
	req QueueWaypointDeploymentJobRequest,
) ([]byte, error) {
	key := fmt.Sprintf("%s/build.hcl", req.BucketPrefix)
	objReq := &s3.GetObjectInput{
		Bucket: &req.BucketName,
		Key:    &key,
	}
	objResp, err := client.GetObject(ctx, objReq)
	if err != nil {
		return nil, err
	}

	byts, err := io.ReadAll(objResp.Body)
	if err != nil {
		return nil, err
	}
	return byts, nil
}

func (w *waypointDeploymentJobQueuerImpl) queueWaypointDeploymentJob(
	ctx context.Context,
	client waypointClientJobQueuer,
	req QueueWaypointDeploymentJobRequest,
	waypointHcl []byte,
) (string, error) {
	waypointCfg, err := getWaypointHcl(req)
	if err != nil {
		return "", fmt.Errorf("unable to get waypoint config: %w", err)
	}

	wpReq := &gen.QueueJobRequest{
		Job: &gen.Job{
			Operation: &gen.Job_Build{
				Build: &gen.Job_BuildOp{DisablePush: false},
			},
			SingletonId: fmt.Sprintf("%s-%s", req.DeploymentID, req.ComponentName),
			Workspace: &gen.Ref_Workspace{
				Workspace: req.AppID,
			},
			Application: &gen.Ref_Application{
				Project:     req.OrgID,
				Application: req.ComponentName,
			},
			Labels: map[string]string{
				"temporal-workers": "true",
				"deployment-id":    req.DeploymentID,
				"app-id":           req.AppID,
				"component-name":   req.ComponentName,
				"component-type":   req.ComponentType,
			},
			DataSource: &gen.Job_DataSource{
				Source: &gen.Job_DataSource_Git{
					Git: &gen.Job_Git{
						Url: "https://github.com/jonmorehouse/empty",
					},
				},
			},
			WaypointHcl: &gen.Hcl{
				Contents: waypointCfg,
			},
			TargetRunner: &gen.Ref_Runner{
				Target: &gen.Ref_Runner_Any{
					Any: &gen.Ref_RunnerAny{},
				},
			},
			OndemandRunner: &gen.Ref_OnDemandRunnerConfig{
				Name: req.OrgID,
			},

			Variables: []*gen.Variable{},
		},
		ExpiresIn: defaultJobTimeout,
	}
	jresp, err := client.QueueJob(ctx, wpReq)
	if err != nil {
		return "", fmt.Errorf("unable to queue deployment job: %w", err)
	}

	return jresp.JobId, nil
}
