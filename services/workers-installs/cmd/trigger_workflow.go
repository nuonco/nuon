package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/common/temporalzap"
	"github.com/powertoolsdev/mono/pkg/config"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	shared "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/spf13/cobra"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

var triggerWorkflowCmd = &cobra.Command{
	Use:   "trigger-workflow",
	Short: "Trigger a workflow",
	Run:   runTriggerWorkflow,
}
var triggerWorkflowCmdName string

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(triggerWorkflowCmd)
	triggerWorkflowCmd.Flags().StringVarP(&triggerWorkflowCmdName, "workflow", "w", "org_signup", "The full workflow path such as org_signup or install_provision")
}

// workflowTriggerFn represents the type of workflow to be triggered
type workflowTriggerFn func(context.Context, tclient.Client, []string) (string, error)

// triggerOrgSignupWorkflow triggers an org to be provisioned and expects a single argument (an identifier)
func triggerOrgSignupWorkflow(ctx context.Context, tc tclient.Client, args []string) (string, error) {
	//if len(args) < 1 {
	//return "", fmt.Errorf("expected id arg")
	//}

	//req := signup.SignupRequest{
	//OrgID:  args[0],
	//Region: "us-west-2",
	//}

	//opts := tclient.StartWorkflowOptions{
	//TaskQueue: "org",
	//}

	//run, err := tc.ExecuteWorkflow(ctx, opts, "Signup", req)
	//if err != nil {
	//return "", fmt.Errorf("unable to submit workflow: %w", err)
	//}

	//return run.GetID(), nil
	return "", nil
}

// triggerOrgTeardownWorkflow triggers an org to be provisioned and expects a single argument (an identifier)
func triggerOrgTeardownWorkflow(ctx context.Context, tc tclient.Client, args []string) (string, error) {
	//if len(args) < 1 {
	//return "", fmt.Errorf("expected id arg")
	//}

	//req := teardown.TeardownRequest{
	//OrgID:  args[0],
	//Region: "us-west-2",
	//}

	//opts := tclient.StartWorkflowOptions{
	//TaskQueue: "org",
	//}

	//run, err := tc.ExecuteWorkflow(ctx, opts, "Teardown", req)
	//if err != nil {
	//return "", fmt.Errorf("unable to submit workflow: %w", err)
	//}

	//return run.GetID(), nil
	return "", nil
}

// NOTE(jm): this is not designed to be used in any type of production workflow, and should probably not live in this
// cli long term
func getPreviousInstallProvisionRequest(ctx context.Context, installID string) (*installsv1.ProvisionRequest, error) {
	resp := &installsv1.ProvisionRequest{}

	cfg, err := aws_config.LoadDefaultConfig(ctx)
	if err != nil {
		return resp, err
	}
	client := s3.NewFromConfig(cfg)

	bucketName := "nuon-org-installations-stage"
	req := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}

	subStr := fmt.Sprintf("install=%s", installID)
	var key string
	for {
		s3Resp, err2 := client.ListObjectsV2(ctx, req)
		if err2 != nil {
			return resp, err2
		}

		for _, obj := range s3Resp.Contents {
			if strings.Contains(*obj.Key, subStr) && strings.HasSuffix(*obj.Key, "request.json") {
				key = *obj.Key
				break
			}
		}

		if key != "" || s3Resp.ContinuationToken == nil {
			break
		}

		req.ContinuationToken = s3Resp.ContinuationToken
	}
	if key == "" {
		return resp, fmt.Errorf("unable to find previous request for install")
	}

	fmt.Println("downloading request from: " + key)
	// grab the object
	objReq := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &key,
	}
	objResp, err := client.GetObject(ctx, objReq)
	if err != nil {
		return resp, err
	}
	byts, err := io.ReadAll(objResp.Body)
	if err != nil {
		return resp, err
	}

	if err := json.Unmarshal(byts, &resp); err != nil {
		return resp, fmt.Errorf("unable to decode to request file: %w", err)
	}

	fp := fmt.Sprintf("/tmp/install-request-%s.json", installID)
	fmt.Printf("writing request to %s\n", fp)

	//nolint:all
	if err := os.WriteFile(fp, byts, 0644); err != nil {
		return resp, nil
	}

	return resp, nil
}

func shortIDsToUUIDs(ids ...string) []string {
	output := make([]string, len(ids))

	for idx, id := range ids {
		id, err := shortid.ToUUID(id)
		if err != nil {
			// NOTE(jdt): just fail out completely. This isn't the "right way" but it works for this use case
			zap.L().Fatal("failed to convert to short IDs: %s", zap.Error(err))
		}

		output[idx] = id.String()
	}

	return output
}

// deprovisionInstallProvisionWorkflow deprovisions an installation
func triggerInstallDeprovisionWorkflow(ctx context.Context, tc tclient.Client, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("expected id arg")
	}

	prReq, err := getPreviousInstallProvisionRequest(ctx, args[0])
	if err != nil {
		return "", err
	}

	ids := shortIDsToUUIDs(prReq.OrgId, prReq.AppId, prReq.InstallId)

	req := &installsv1.DeprovisionRequest{
		OrgId:           ids[0],
		AppId:           ids[1],
		InstallId:       ids[2],
		SandboxSettings: prReq.SandboxSettings,
		AccountSettings: prReq.AccountSettings,
	}

	opts := tclient.StartWorkflowOptions{
		TaskQueue: "install",
	}
	run, err := tc.ExecuteWorkflow(ctx, opts, "Deprovision", req)
	if err != nil {
		return "", fmt.Errorf("unable to submit workflow: %w", err)
	}

	return run.GetID(), nil
}

// retriggerInstallProvisionWorkflow re triggers an installation workflow
func retriggerInstallProvisionWorkflowWithVersion(ctx context.Context, tc tclient.Client, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("expected id arg and version")
	}

	req, err := getPreviousInstallProvisionRequest(ctx, args[0])
	if err != nil {
		return "", err
	}

	ids := shortIDsToUUIDs(req.OrgId, req.AppId, req.InstallId)
	req.OrgId = ids[0]
	req.AppId = ids[1]
	req.InstallId = ids[2]

	req.SandboxSettings.Version = args[1]

	opts := tclient.StartWorkflowOptions{
		TaskQueue: "install",
	}
	run, err := tc.ExecuteWorkflow(ctx, opts, "Provision", req)
	if err != nil {
		return "", fmt.Errorf("unable to submit workflow: %w", err)
	}

	return run.GetID(), nil
}

func triggerDeploymentStartWorkflow(ctx context.Context, tc tclient.Client, args []string) (string, error) {
	//if len(args) < 1 {
	//return "", fmt.Errorf("expected an installID so we could look up the org/appID")
	//}

	//instReq, err := getPreviousInstallProvisionRequest(ctx, args[0])
	//if err != nil {
	//return "", err
	//}

	//orgID, err := shortid.ToUUID(instReq.OrgID)
	//if err != nil {
	//return "", fmt.Errorf("unable to convert orgID to uuid: %s", err)
	//}

	//appID, err := shortid.ToUUID(instReq.AppID)
	//if err != nil {
	//return "", fmt.Errorf("unable to convert appID to uuid: %s", err)
	//}
	//installID, err := shortid.ToUUID(instReq.InstallID)
	//if err != nil {
	//return "", fmt.Errorf("unable to convert appID to uuid: %s", err)
	//}

	//req := dstart.StartRequest{
	//OrgID:	orgID.String(),
	//AppID:	appID.String(),
	//DeploymentID: uuid.NewString(),
	//InstallIDs: []string{
	//installID.String(),
	//},
	//Component: waypoint.Component{
	//Name:		     "mario",
	//ContainerImageURL: "pengbai/docker-supermario",
	//Type:		     "PUBLIC_IMAGE",
	//},
	//}

	//opts := tclient.StartWorkflowOptions{
	//TaskQueue: "deployment",
	//}

	//run, err := tc.ExecuteWorkflow(ctx, opts, "Start", req)
	//if err != nil {
	//return "", fmt.Errorf("unable to submit workflow: %w", err)
	//}

	//return run.GetID(), nil
	return "", nil
}

// retriggerInstallProvisionWorkflow re triggers an installation workflow
func retriggerInstallProvisionWorkflow(ctx context.Context, tc tclient.Client, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("expected id arg")
	}

	req, err := getPreviousInstallProvisionRequest(ctx, args[0])
	if err != nil {
		return "", err
	}
	ids := shortIDsToUUIDs(req.OrgId, req.AppId, req.InstallId)
	req.OrgId = ids[0]
	req.AppId = ids[1]
	req.InstallId = ids[2]

	opts := tclient.StartWorkflowOptions{
		TaskQueue: "install",
	}

	run, err := tc.ExecuteWorkflow(ctx, opts, "Provision", req)
	if err != nil {
		return "", fmt.Errorf("unable to submit workflow: %w", err)
	}

	return run.GetID(), nil
}

func triggerInstanceProvisionWorkflow(ctx context.Context, tc tclient.Client, args []string) (string, error) {
	//if len(args) < 1 {
	//return "", fmt.Errorf("expected id arg")
	//}

	//req, err := getPreviousInstallProvisionRequest(ctx, args[0])
	//if err != nil {
	//return "", err
	//}

	//fmt.Printf("%+v\n", req)

	//deploymentID := uuid.New()
	//deploymentShortID, err := shortid.ParseUUID(deploymentID)
	//if err != nil {
	//return "", err
	//}

	//// NOTE(jm): the provision request accepts short ids
	//instancereq := instanprov.ProvisionRequest{}
	//instancereq.OrgID = req.OrgID
	//instancereq.AppID = req.AppID
	//instancereq.InstallID = req.InstallID
	//instancereq.DeploymentID = deploymentShortID
	//instancereq.Component = waypoint.Component{
	//Name:		     "mario",
	//ContainerImageURL: "https://hub.docker.com/r/pengbai/docker-supermario",
	//Type:		     "PUBLIC_IMAGE",
	//}

	//opts := tclient.StartWorkflowOptions{
	//TaskQueue: "instance",
	//}

	//run, err := tc.ExecuteWorkflow(ctx, opts, "Provision", instancereq)
	//if err != nil {
	//return "", fmt.Errorf("unable to submit workflow: %w", err)
	//}

	//return run.GetID(), nil
	return "", nil
}

// handlerForWorkflow returns the function that should be used to handle the workflow
func handlerForWorkflow(name string) (workflowTriggerFn, error) {
	fn, ok := map[string]workflowTriggerFn{
		"org-signup":          triggerOrgSignupWorkflow,
		"org-teardown":        triggerOrgTeardownWorkflow,
		"install-reprovision": retriggerInstallProvisionWorkflow,
		"install-deprovision": triggerInstallDeprovisionWorkflow,
		"install-upgrade":     retriggerInstallProvisionWorkflowWithVersion,
		"instance-provision":  triggerInstanceProvisionWorkflow,
		"deployment-start":    triggerDeploymentStartWorkflow,
	}[name]
	if !ok {
		return nil, fmt.Errorf("%s is not supported", name)
	}

	return fn, nil
}

func runTriggerWorkflow(cmd *cobra.Command, args []string) {
	var cfg shared.Config
	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err))
	}

	var (
		l   *zap.Logger
		err error
	)
	switch cfg.Env {
	case config.Development:
		l, err = zap.NewDevelopment()
	default:
		l, err = zap.NewProduction()
	}
	if err != nil {
		fmt.Printf("failed to instantiate logger: %s", err)
	}
	zap.ReplaceGlobals(l)

	handlerFn, err := handlerForWorkflow(triggerWorkflowCmdName)
	if err != nil {
		l.Fatal("can't find handler fn: %v\n", zap.Error(err))
	}

	c, err := tclient.Dial(tclient.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.TemporalNamespace,
		Logger:    temporalzap.NewLogger(l),
	})
	if err != nil {
		l.Fatal("failed to instantiate temporal client", zap.Error(err))
	}
	defer c.Close()

	ctx := context.Background()
	ctx, ctxCancel := context.WithCancel(ctx)
	defer ctxCancel()

	workflowID, err := handlerFn(ctx, c, args)
	if err != nil {
		l.Fatal("unable to emit job", zap.Error(err))
	}

	fmt.Printf("submitted workflow: %s", workflowID)
}
