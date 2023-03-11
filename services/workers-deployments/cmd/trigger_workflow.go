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
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/common/config"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/common/temporalzap"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	shared "github.com/powertoolsdev/mono/services/workers-deployments/internal"
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

func triggerInstallDeploymentWorkflow(ctx context.Context, tc tclient.Client, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("expected an installID so we could look up the org/appID")
	}

	instReq, err := getPreviousInstallProvisionRequest(ctx, args[0])
	if err != nil {
		return "", err
	}

	deploymentID := uuid.NewString()
	ids := shortIDsToUUIDs(instReq.OrgId, instReq.AppId, instReq.InstallId)
	req := &deploymentsv1.StartRequest{
		OrgId:        ids[0],
		AppId:        ids[1],
		DeploymentId: deploymentID,
		InstallIds:   []string{ids[2]},
		Component: &componentv1.Component{
			Id: uuid.NewString(),
		},
		PlanOnly: false,
	}
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "deployment",
	}

	run, err := tc.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return "", fmt.Errorf("unable to submit workflow: %w", err)
	}

	deploymentShortID, err := shortid.ParseString(deploymentID)
	if err != nil {
		return "", fmt.Errorf("unable to parse short id: %w", err)
	}

	fmt.Println(deploymentShortID)

	return run.GetID(), nil
}

// handlerForWorkflow returns the function that should be used to handle the workflow
func handlerForWorkflow(name string) (workflowTriggerFn, error) {
	fn, ok := map[string]workflowTriggerFn{
		"install-deployment": triggerInstallDeploymentWorkflow,
		// TODO(jm): add ability to redeploy
		//"redeployment":	triggerRedeploymentWorkflow,
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
