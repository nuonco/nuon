// Package installer orchestrates AWS-native install stack provisioning.
//
// It wires the AWS SDK clients to the stack package's resource provisioners
// and emits structured logs through the logstream package, which targets
// either stdout (dev) or the Nuon ctl-api OTLP log-stream ingest (production).
package installer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/nuonco/nuon/sdks/nuon-installer-go/logstream"
	"github.com/nuonco/nuon/sdks/nuon-installer-go/stack"
	"github.com/nuonco/nuon/sdks/nuon-installer-go/stackrun"
)

// Installer provisions and tears down an install stack in a customer AWS account.
type Installer struct {
	opts Options
	log  *slog.Logger
	prov *logstream.Provider

	awsCfg aws.Config
	ec2c   *ec2.Client
	eksc   *eks.Client
	iamc   *iam.Client
	s3c    *s3.Client
	stsc   *sts.Client
}

// New builds an Installer. Caller must call Close to flush logs.
func New(ctx context.Context, opts Options) (*Installer, error) {
	if opts.InstallID == "" {
		return nil, fmt.Errorf("InstallID required")
	}
	if opts.AWSRegion == "" {
		return nil, fmt.Errorf("AWSRegion required")
	}

	var prov *logstream.Provider
	if opts.LogStream == nil {
		prov = logstream.NewStdout("stack")
	} else {
		var err error
		prov, err = logstream.New(ctx, logstream.Config{
			RunnerAPIURL: opts.LogStream.RunnerAPIURL,
			LogStreamID:  opts.LogStream.ID,
			WriteToken:   opts.LogStream.WriteToken,
			ServiceName:  "stack",
			Attrs: map[string]string{
				"install_id": opts.InstallID,
				"aws_region": opts.AWSRegion,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(opts.AWSRegion))
	if err != nil {
		_ = prov.Shutdown(ctx)
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &Installer{
		opts:   opts,
		log:    prov.Logger().With("install_id", opts.InstallID, "aws_region", opts.AWSRegion),
		prov:   prov,
		awsCfg: awsCfg,
		ec2c:   ec2.NewFromConfig(awsCfg),
		eksc:   eks.NewFromConfig(awsCfg),
		iamc:   iam.NewFromConfig(awsCfg),
		s3c:    s3.NewFromConfig(awsCfg),
		stsc:   sts.NewFromConfig(awsCfg),
	}, nil
}

func (i *Installer) Close(ctx context.Context) error {
	if i.prov != nil {
		return i.prov.Shutdown(ctx)
	}
	return nil
}

// Provision runs the full provisioning sequence. State is persisted to disk
// after each successful step so a partial failure can be cleaned up by Deprovision.
func (i *Installer) Provision(ctx context.Context) (*stack.State, error) {
	st, err := stack.LoadState(i.opts.InstallID, i.opts.AWSRegion)
	if err != nil {
		return nil, err
	}

	// Report to ctl-api as a stack run. Required when configured: the response
	// also carries the OTLP log-stream credentials we need for visibility, so
	// silently continuing on failure would hide the entire run from the user.
	var runClient *stackrun.Client
	var runID string
	if i.opts.StackRun != nil {
		runClient = stackrun.New(stackrun.Config{
			CtlAPIURL:   i.opts.StackRun.CtlAPIURL,
			InstallID:   i.opts.InstallID,
			PhoneHomeID: i.opts.StackRun.PhoneHomeID,
		})
		resp, err := runClient.CreateRun(ctx)
		if err != nil {
			return st, fmt.Errorf("create stack run: %w", err)
		}
		runID = resp.ID
		i.log.Info("created stack run", "run_id", runID)
		// Swap the stdout-only provider for an OTLP one fed by the credentials
		// ctl-api just minted for this run.
		if resp.LogStream != nil && resp.LogStream.ID != "" {
			otelProv, err := logstream.New(ctx, logstream.Config{
				RunnerAPIURL: resp.LogStream.RunnerAPIURL,
				LogStreamID:  resp.LogStream.ID,
				WriteToken:   resp.LogStream.WriteToken,
				ServiceName:  "stack",
				Attrs: map[string]string{
					"install_id": i.opts.InstallID,
					"aws_region": i.opts.AWSRegion,
				},
			})
			if err != nil {
				i.log.Warn("init otlp log stream failed (continuing with prior logger)", "err", err.Error())
			} else {
				// Flush the prior (stdout/dev) provider, then route through OTLP.
				if i.prov != nil {
					_ = i.prov.Shutdown(ctx)
				}
				i.prov = otelProv
				i.log = otelProv.Logger().With(
					"install_id", i.opts.InstallID,
					"aws_region", i.opts.AWSRegion,
				)
			}
		}
	}

	steps := []struct {
		name string
		run  func(context.Context, *slog.Logger, *stack.State) error
	}{
		{"validate-aws-access", i.stepValidateAWS},
		{"create-s3-bucket", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.CreateS3Bucket(ctx, l, i.s3c, s)
		}},
		{"create-vpc", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.CreateVPC(ctx, l, i.ec2c, s)
		}},
		{"create-iam-roles", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.CreateIAMRoles(ctx, l, i.iamc, s)
		}},
		{"create-eks-cluster", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.CreateEKSCluster(ctx, l, i.eksc, i.iamc, s)
		}},
		{"create-node-group", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.CreateNodeGroup(ctx, l, i.eksc, i.iamc, s)
		}},
	}

	for _, s := range steps {
		log := i.log.With("step", s.name)
		log.Info("step starting")
		if err := s.run(ctx, log, st); err != nil {
			log.Error("step failed", "err", err.Error())
			i.reportRun(ctx, runClient, runID, "failed", err.Error(), st)
			return st, fmt.Errorf("step %s: %w", s.name, err)
		}
		log.Info("step completed")
	}

	i.log.Info("provision complete",
		"cluster_endpoint", st.ClusterEndpoint,
		"oidc_issuer", st.OIDCIssuer,
	)
	i.reportRun(ctx, runClient, runID, "succeeded", "", st)
	return st, nil
}

// reportRun is best-effort; failures only log.
func (i *Installer) reportRun(ctx context.Context, c *stackrun.Client, runID, status, statusDesc string, st *stack.State) {
	if c == nil || runID == "" {
		return
	}
	data := map[string]any{}
	if st != nil {
		if st.ClusterEndpoint != "" {
			data["cluster_endpoint"] = st.ClusterEndpoint
		}
		if st.OIDCIssuer != "" {
			data["oidc_issuer"] = st.OIDCIssuer
		}
		if st.ClusterName != "" {
			data["cluster_name"] = st.ClusterName
		}
	}
	if err := c.UpdateRun(ctx, runID, stackrun.UpdateRunRequest{
		Status:            status,
		StatusDescription: statusDesc,
		Data:              data,
	}); err != nil {
		i.log.Warn("update stack run failed", "err", err.Error(), "status", status)
	}
}

// Deprovision tears down everything in the state file in reverse order.
func (i *Installer) Deprovision(ctx context.Context) error {
	st, err := stack.LoadState(i.opts.InstallID, i.opts.AWSRegion)
	if err != nil {
		return err
	}
	steps := []struct {
		name string
		run  func(context.Context, *slog.Logger, *stack.State) error
	}{
		{"delete-eks", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.DeleteEKS(ctx, l, i.eksc, s)
		}},
		{"delete-iam-roles", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.DeleteIAMRoles(ctx, l, i.iamc, s)
		}},
		{"delete-vpc", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.DeleteVPC(ctx, l, i.ec2c, s)
		}},
		{"delete-s3-bucket", func(ctx context.Context, l *slog.Logger, s *stack.State) error {
			return stack.DeleteS3Bucket(ctx, l, i.s3c, s)
		}},
	}
	for _, s := range steps {
		log := i.log.With("step", s.name)
		log.Info("step starting")
		if err := s.run(ctx, log, st); err != nil {
			log.Error("step failed", "err", err.Error())
			return fmt.Errorf("step %s: %w", s.name, err)
		}
		log.Info("step completed")
	}
	if err := st.Delete(); err != nil {
		return fmt.Errorf("delete state file: %w", err)
	}
	i.log.Info("deprovision complete")
	return nil
}

// Status returns the current persisted state.
func (i *Installer) Status() (*stack.State, error) {
	return stack.LoadState(i.opts.InstallID, i.opts.AWSRegion)
}

func (i *Installer) stepValidateAWS(ctx context.Context, log *slog.Logger, _ *stack.State) error {
	out, err := i.stsc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("sts get-caller-identity: %w", err)
	}
	log.Info("aws caller identity",
		"account", aws.ToString(out.Account),
		"arn", aws.ToString(out.Arn),
	)
	return nil
}
