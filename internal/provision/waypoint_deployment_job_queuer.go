package provision

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/template"

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

	BucketName   string `json:"bucket_name"   validate:"required"`
	BucketPrefix string `json:"bucket_prefix" validate:"required"`

	AppID         string `json:"app_id" validate:"required"`
	InstallID     string `json:"install_id" validate:"required"`
	DeploymentID  string `json:"deployment_id" validate:"required"`
	ComponentName string `json:"component_name" validate:"required"`
}

func (w QueueWaypointDeploymentJobRequest) validate() error {
	validate := validator.New()
	return validate.Struct(w)
}

type QueueWaypointDeploymentJobResponse struct {
	JobID string `json:"job_id" validate:"required"`
}

var waypointDeployTmpl string = `
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

type deployTmplArgs struct {
	Project          string
	InputImage       string
	InputVersion     string
	OutputRepository string
	OutputVersion    string
}

func getWaypointHcl(req QueueWaypointDeploymentJobRequest) ([]byte, error) {
	tmpl, err := template.New("build-config").Parse(waypointDeployTmpl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse template: %w", err)
	}

	buf := new(bytes.Buffer)

	args := deployTmplArgs{
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

	waypointHcl, err := getWaypointHcl(req)
	if err != nil {
		return resp, err
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	artifact, err := getWaypointBuild(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to get waypoint build: %w", err)
	}

	jobID, err := a.queueWaypointDeploymentJob(ctx, client, req, waypointHcl, artifact)
	if err != nil {
		return resp, fmt.Errorf("failed to queue waypoint deployment application: %w", err)
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
		*gen.PushedArtifact,
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
	key := fmt.Sprintf("%s/deploy.hcl", req.BucketPrefix)
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
func getWaypointBuild(ctx context.Context, client gen.WaypointClient, req QueueWaypointDeploymentJobRequest) (*gen.PushedArtifact, error) {
	// get latest artifact if none set
	push, err := client.GetLatestPushedArtifact(ctx, &gen.GetLatestPushedArtifactRequest{
		Application: &gen.Ref_Application{
			Application: req.ComponentName,
			Project:     req.OrgID,
		},
		Workspace: &gen.Ref_Workspace{
			Workspace: req.AppID,
		},
	})
	if err != nil {
		return nil, err
	}

	return push, nil
}

func (w *waypointDeploymentJobQueuerImpl) queueWaypointDeploymentJob(
	ctx context.Context,
	client waypointClientJobQueuer,
	req QueueWaypointDeploymentJobRequest,
	waypointHcl []byte,
	artifact *gen.PushedArtifact,
) (string, error) {
	wpReq := &gen.QueueJobRequest{
		Job: &gen.Job{
			SingletonId: fmt.Sprintf("%s-%s", req.InstallID, req.DeploymentID),
			Operation: &gen.Job_Deploy{
				Deploy: &gen.Job_DeployOp{
					Artifact: artifact,
				},
			},
			Workspace: &gen.Ref_Workspace{
				Workspace: req.InstallID,
			},
			TargetRunner: &gen.Ref_Runner{
				Target: &gen.Ref_Runner_Id{
					Id: &gen.Ref_RunnerId{
						Id: req.InstallID,
					},
				},
			},
			OndemandRunner: &gen.Ref_OnDemandRunnerConfig{
				Name: req.InstallID,
			},
			Application: &gen.Ref_Application{
				Project:     req.InstallID,
				Application: req.ComponentName,
			},

			Labels: map[string]string{
				"deployment_id": req.DeploymentID,
				"install_id":    req.InstallID,
			},
			DataSource: &gen.Job_DataSource{
				Source: &gen.Job_DataSource_Git{
					Git: &gen.Job_Git{
						Url: "https://github.com/jonmorehouse/empty",
					},
				},
			},
			WaypointHcl: &gen.Hcl{
				Contents: waypointHcl,
			},
			Variables: []*gen.Variable{},
		},
		ExpiresIn: defaultJobTimeout,
	}
	jResp, err := client.QueueJob(ctx, wpReq)
	if err != nil {
		return "", fmt.Errorf("unable to queue deployment job: %w", err)
	}

	return jResp.JobId, nil
}
