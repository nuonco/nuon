// Package clientcfg builds Kubernetes rest configs.
//
// It lives outside package kube so binaries that only pass ClusterInfo around
// do not link client-go.
package clientcfg

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/impersonate"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/kube"
)

// client-go defaults to 5 QPS / 10 burst, which throttles helm and discovery fan-out.
const (
	restConfigQPS   = 50
	restConfigBurst = 100
)

func withRateLimits(cfg *rest.Config) *rest.Config {
	cfg.QPS = restConfigQPS
	cfg.Burst = restConfigBurst
	return cfg
}

func ConfigForCluster(ctx context.Context, cInfo *kube.ClusterInfo) (*rest.Config, error) {
	if cInfo.KubeConfig != "" {
		config, err := clientcmd.RESTConfigFromKubeConfig([]byte(cInfo.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("unable to parse kube config: %w", err)
		}

		return withRateLimits(config), nil
	}

	if cInfo.AWSAuth == nil && cInfo.AzureAuth == nil && cInfo.GCPAuth == nil {
		return nil, fmt.Errorf("missing auth configuration")
	}

	u, err := url.Parse(cInfo.Endpoint)
	if err != nil {
		return nil, err
	}

	envVars, err := fetchEnv(ctx, cInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch environment: %w", err)
	}

	caData, err := base64.StdEncoding.DecodeString(cInfo.CAData)
	if err != nil {
		return nil, fmt.Errorf("unable to decode CA data: %w", err)
	}

	cfg := &rest.Config{
		Host: cInfo.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			ServerName: u.Hostname(),
			CAData:     []byte(caData),
		},
		ExecProvider: &clientcmdapi.ExecConfig{
			APIVersion:      "client.authentication.k8s.io/v1beta1",
			Command:         "aws-iam-authenticator",
			Env:             envVars,
			Args:            []string{"token", "-i", cInfo.ID},
			InteractiveMode: clientcmdapi.NeverExecInteractiveMode,
		},
	}
	// TODO(jm): this is deprecated and only used in legacy users of this
	if cInfo.TrustedRoleARN != "" {
		cfg.ExecProvider.Args = []string{"token", "-i", cInfo.ID, "-r", cInfo.TrustedRoleARN}
	}

	if cInfo.Inline {
		tok, err := inlineAWSToken(ctx, cInfo)
		if err != nil {
			return nil, err
		}
		cfg.BearerToken = tok
		cfg.ExecProvider = nil
	}

	if cInfo.AzureAuth != nil {
		// TODO(ja): reverse engineer how kubelogin works so we can go back to using a BearerToken
		// // get a credential
		// cred, err := azurecredentials.Fetch(ctx)
		// if err != nil {
		// 	return nil, fmt.Errorf("unable to get azure credential: %w", err)
		// }

		// // use the credentials to get an Entra ID token
		// entraToken, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		// 	Scopes: []string{"https://management.azure.com/.default"}},
		// )
		// if err != nil {
		// 	return nil, fmt.Errorf("unable to get entra ID token: %w", err)
		// }

		// cfg.BearerToken = entraToken.Token
		// cfg.ExecProvider = nil

		// Use kubelogin to authenticate.
		// On Azure VMs the runner has ARM_USE_MSI=true and uses managed identity.
		// Locally (dev mode) we use az login credentials via azurecli mode.
		loginMode := "azurecli"
		if os.Getenv("ARM_USE_MSI") == "true" {
			loginMode = "msi"
		}
		args := []string{
			"get-token",
			"--login",
			loginMode,
			"--server-id",
			"6dae42f8-4368-4678-94ff-3960e28e3630",
			"--tenant-id",
			cInfo.AzureAuth.ServicePrincipal.SubscriptionTenantID,
		}
		// Authenticate as the operation identity, not the runner's system identity.
		if loginMode == "msi" && cInfo.AzureAuth.ManagedIdentityClientID != "" {
			args = append(args, "--client-id", cInfo.AzureAuth.ManagedIdentityClientID)
		}
		cfg.ExecProvider = &clientcmdapi.ExecConfig{
			APIVersion:      "client.authentication.k8s.io/v1beta1",
			Command:         "kubelogin",
			Args:            args,
			InteractiveMode: clientcmdapi.NeverExecInteractiveMode,
		}
	}

	if cInfo.GCPAuth != nil {
		var ts oauth2.TokenSource
		if cInfo.GCPAuth.ImpersonateServiceAccount != "" {
			// K8s auth must run as the operation role SA, not the runner SA — the runner SA
			// holds no container permissions.
			// userinfo.email is required for GKE to map the token to the SA's
			// IAM identity; without it the apiserver sees an unmappable numeric
			// principal and denies everything.
			ts, err = impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
				TargetPrincipal: cInfo.GCPAuth.ImpersonateServiceAccount,
				Scopes: []string{
					"https://www.googleapis.com/auth/cloud-platform",
					"https://www.googleapis.com/auth/userinfo.email",
				},
			})
			if err != nil {
				return nil, fmt.Errorf("unable to get impersonated GCP token source for K8s auth: %w", err)
			}
		} else {
			ts, err = google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
			if err != nil {
				return nil, fmt.Errorf("unable to get GCP token source for K8s auth: %w", err)
			}
		}

		cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			return &oauth2.Transport{Source: ts, Base: rt}
		}

		// Use gke-gcloud-auth-plugin if available (install runners in customer clusters),
		// otherwise rely on oauth2 WrapTransport via Workload Identity (org runners in our cluster).
		// The exec provider is what gets serialized into kubeconfigs for action workflows, so it
		// carries the impersonation target via gcloud's env var.
		if _, err := exec.LookPath("gke-gcloud-auth-plugin"); err == nil {
			execCfg := &clientcmdapi.ExecConfig{
				APIVersion:      "client.authentication.k8s.io/v1beta1",
				Command:         "gke-gcloud-auth-plugin",
				InteractiveMode: clientcmdapi.NeverExecInteractiveMode,
			}
			if cInfo.GCPAuth.ImpersonateServiceAccount != "" {
				execCfg.Env = []clientcmdapi.ExecEnvVar{{
					Name:  "CLOUDSDK_AUTH_IMPERSONATE_SERVICE_ACCOUNT",
					Value: cInfo.GCPAuth.ImpersonateServiceAccount,
				}}
			}
			cfg.ExecProvider = execCfg
		} else {
			cfg.ExecProvider = nil
		}
	}

	return withRateLimits(cfg), nil
}

func inlineAWSToken(ctx context.Context, cInfo *kube.ClusterInfo) (string, error) {
	if cInfo.AWSAuth == nil {
		return "", fmt.Errorf("inline auth requires an aws auth configuration")
	}

	awsCfg, err := credentials.Fetch(ctx, cInfo.AWSAuth)
	if err != nil {
		return "", errors.Wrap(err, "unable to fetch credentials")
	}
	if awsCfg.Credentials == nil {
		return "", fmt.Errorf("no credentials were set on aws config")
	}
	if awsCfg.Region == "" {
		return "", fmt.Errorf("aws region is required for inline eks auth")
	}

	if cInfo.TrustedRoleARN != "" {
		awsCfg.Credentials = stscreds.NewAssumeRoleProvider(sts.NewFromConfig(awsCfg), cInfo.TrustedRoleARN)
	}
	awsCfg.Credentials = aws.NewCredentialsCache(awsCfg.Credentials)
	if _, err := awsCfg.Credentials.Retrieve(ctx); err != nil {
		return "", errors.Wrap(err, "unable to retrieve credentials")
	}

	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return "", err
	}

	tok, err := gen.GetWithSTS(cInfo.ID, sts.NewFromConfig(awsCfg))
	if err != nil {
		return "", errors.Wrap(err, "unable to get token")
	}

	return tok.Token, nil
}
