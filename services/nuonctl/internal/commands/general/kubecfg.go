package general

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/briandowns/spinner"
	"github.com/powertoolsdev/mono/pkg/kube"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"gopkg.in/ini.v1"
)

const (
	maxRetries   = 1
	spinnerDelay = 150 * time.Millisecond
)

var allowedRegions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2"}

type state func() state

type kubecfgCli struct {
	ctx     context.Context
	spinner *spinner.Spinner

	profiles []string
	profile  string
	region   string
	awsCfg   *aws.Config

	clusters           []string
	cluster            string
	listClustersAPI    eks.ListClustersAPIClient
	describeClusterAPI eks.DescribeClusterAPIClient
	clusterDetails     *ekstypes.Cluster
	restConfig         *rest.Config
	kcfg               *api.Config
}

func (c *commands) GetKubecfg() error {
	cli := &kubecfgCli{
		ctx:     context.Background(),
		spinner: spinner.New(spinner.CharSets[1], spinnerDelay),
	}
	cli.spinner.FinalMSG = "\n"

	state := cli.getProfiles
	for {
		state = state()
		if state == nil {
			break
		}
	}

	fmt.Println("finished...")
	return nil
}

func (k *kubecfgCli) getProfiles() state {
	cfg, err := ini.Load(config.DefaultSharedConfigFilename())
	if err != nil {
		return k.handleError("get profiles", err)
	}

	for _, s := range cfg.Sections() {
		name := s.Name()
		if strings.Contains(name, ".NuonPowerUser") {
			k.profiles = append(k.profiles, strings.TrimPrefix(name, "profile "))
		}
	}

	return k.promptProfiles
}

func (k *kubecfgCli) promptProfiles() state {
	prompt := &survey.Select{
		Message: "choose the profile for the account with the cluster",
		Options: k.profiles,
	}
	if err := survey.AskOne(prompt, &k.profile, survey.WithValidator(survey.Required)); err != nil {
		return k.handleError("choosing profile", err)
	}
	return k.promptRegion
}

func (k *kubecfgCli) promptRegion() state {
	prompt := &survey.Select{
		Message: "choose the region for the cluster",
		Options: allowedRegions,
	}
	if err := survey.AskOne(prompt, &k.region, survey.WithValidator(survey.Required)); err != nil {
		return k.handleError("choosing region", err)
	}
	return k.loadConfig
}

func (k *kubecfgCli) loadConfig() state {
	cfg, err := config.LoadDefaultConfig(k.ctx, config.WithSharedConfigProfile(k.profile), config.WithRegion(k.region))
	if err != nil {
		return k.handleError("loading aws config", err)
	}
	k.awsCfg = &cfg
	return k.setKubeClient
}

func (k *kubecfgCli) setKubeClient() state {
	c := eks.NewFromConfig(*k.awsCfg)
	k.listClustersAPI = c
	k.describeClusterAPI = c
	return k.listClusters
}

func (k *kubecfgCli) listClusters() state {
	step := "listing clusters"
	k.spinner.Suffix = fmt.Sprintf(" fetching clusters in %s", k.profile)
	k.spinner.Start()

	out, err := k.listClustersAPI.ListClusters(k.ctx, &eks.ListClustersInput{})
	if err != nil {
		return k.handleError(step, err)
	}
	k.spinner.Stop()
	if len(out.Clusters) == 0 {
		return k.handleError(step, fmt.Errorf("account and region do not have any clusters"))
	}

	k.clusters = append(k.clusters, out.Clusters...)

	return k.promptCluster
}

func (k *kubecfgCli) promptCluster() state {
	prompt := &survey.Select{
		Message: "choose the cluster",
		Options: k.clusters,
	}
	if err := survey.AskOne(prompt, &k.cluster, survey.WithValidator(survey.Required)); err != nil {
		return k.handleError("choosing cluster", err)
	}
	return k.fetchClusterDetails
}

func (k *kubecfgCli) fetchClusterDetails() state {
	k.spinner.Suffix = fmt.Sprintf(" fetching cluster %s details", k.cluster)
	k.spinner.Start()

	out, err := k.describeClusterAPI.DescribeCluster(k.ctx, &eks.DescribeClusterInput{Name: &k.cluster})
	if err != nil {
		return k.handleError("", err)
	}
	k.spinner.Stop()
	k.clusterDetails = out.Cluster

	return k.buildClusterInfo
}

func (k *kubecfgCli) buildClusterInfo() state {
	ci := &kube.ClusterInfo{
		ID:             *k.clusterDetails.Name,
		Endpoint:       *k.clusterDetails.Endpoint,
		CAData:         *k.clusterDetails.CertificateAuthority.Data,
		TrustedRoleARN: "arn:aws:iam::618886478608:role/install-k8s-admin-stage",
	}

	rc, err := kube.ConfigForCluster(ci)
	if err != nil {
		return k.handleError("building cluster info", err)
	}
	rc.ExecProvider.Env = []api.ExecEnvVar{{Name: "AWS_PROFILE", Value: "external.NuonPowerUser"}}
	k.restConfig = rc

	return k.buildKubeconfig
}

func (k *kubecfgCli) buildKubeconfig() state {
	rc := k.restConfig
	po := clientcmd.NewDefaultPathOptions()

	kcfg, err := po.GetStartingConfig()
	if err != nil {
		return k.handleError("building kubeconfig", err)
	}

	kcfg.Contexts[k.cluster] = &api.Context{Cluster: k.cluster, AuthInfo: k.cluster, Namespace: "default"}
	kcfg.AuthInfos[k.cluster] = &api.AuthInfo{Exec: rc.ExecProvider}
	kcfg.Clusters[k.cluster] = &api.Cluster{
		Server:                   rc.Host,
		TLSServerName:            rc.TLSClientConfig.ServerName,
		CertificateAuthorityData: rc.TLSClientConfig.CAData,
	}
	kcfg.CurrentContext = k.cluster

	k.kcfg = kcfg

	return k.modifyKubeconfig
}
func (k *kubecfgCli) modifyKubeconfig() state {
	po := clientcmd.NewDefaultPathOptions()
	if err := clientcmd.ModifyConfig(po, *k.kcfg, true); err != nil {
		return k.handleError("modifying kubeconfig", err)
	}

	return nil
}

func (k *kubecfgCli) handleError(step string, err error) state {
	fmt.Printf("issue running step: step: %s: error: %v\n", step, err)
	return nil
}
