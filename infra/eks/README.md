## infra-eks-nuon

Deploy all infrastructure needed to run a nuon EKS cluster

## Nodepools

See [`karpenter`](./karpenter)

### Setup

We use Terraform Cloud as remote state store as well as remote execution agent.
The workspaces are named `infra-eks-<env>-<pool>`.

1. `terraform workspace new <env>-<pool>`
1. associate aws environment var set to workspace
1. associate twingate env var set to workspace
1. copy existing vars file to vars/<env>-<pool>.yaml
1. update newly copied vars file with the correct pool name and other settings
1. `terraform plan`
1. set up workspace in TF cloud to automatically plan on PRs and merges
1. create PR, verify PR plan, merge, approve plan in TF cloud

#### Caveats

1. The opentelemetry and grafana IAM roles may not fully exist before the pods
   start so may need to be restarted to pick up the roles appropriately.
1. It may take a second apply to ensure that everything is fully created.
1. It will take several minutes (~15?) for Twingate to be fully functional.
1. Additionally, it may require re-connecting in order to pick up the new
   resources / endpoints.

### Teardown

Unfortunately, a few things are created outside of Terraform that prevent being
able to cleanly `destroy`.

1. The `cert-manager` helm resource has a lifecycle policy that prevents
   deletion. This will need to be commented out beforehand.
1. The `karpenter` managed nodes don't currently get removed and will need to be
   terminated outside of terraform before proceeding.
1. There's a security group that seems to be created by `karpenter` that blocks
   the VPC from being destroyed by Terraform.

## Components

### External DNS

The External DNS module creates a service that monitors for annotations on
`Service`s allowing the team to quickly and easily create DNS entries pointing
to the a `Service`'s ClusterIP.

### Vantage Agent

Docs: https://docs.vantage.sh/kubernetes_agent

We add the agent as a deployment with S3 for state so we don't ever have to
debug a statefulset.

### cert-manager

### Grafana

### Amazon Managed Prometheus

### ALB Load Balancer Controller

## Observability

We are using opentelemetry agents running on each node as the primary mechanism
for scraping and receiving observability data (traces / metrics / logs). Each
agent forwards the data to a centralized collector. Both agent and collector are
defined in [`otel.tf`](./otel.tf).

### Metrics

Metrics are forwarded from the centralized collector to an Amazon managed
Prometheus ([`amp.tf`](./amp.tf)). Grafana is currently running in cluster as
the visualization tool ([`grafana.tf`](./grafana.tf)). It's automatically
configured with the clusters managed prometheus instance as a datasource.
Grafana is exposed over Twingate at
`http://grafana.${pool}.${region}.${env}.${root_domain}` - (e.g.
http://grafana.internal.jtarasovic.nuon.co).

### TBD

- **Logs**: there are many options but we should use the otel agents to scrap
  the logs from each node and the collector should forward to chosen tool
- **Traces**: again, lots of options but the same pattern should hold - agents
  receive traces, forward to collector which forwards to tool
- **Grafana dashboards**: We should be able to use the dashboard sidecar that's
  already running to be able to load dashboards from a configmap.

## ~Lifecycle~ Deprecated Concept

⚠️ We are no longer using this concept as originally conceived. ⚠️

New environments are spun up by env and a new term "pool". A pool is an
identifier for this grouping of resources, so that we can have many clusters in
the same environment. The pool must match the workspace name you create.

As an example, for a cluster in the sandbox environment in the testing pool, in
vars/ there would be a `sandbox-test.yaml` that loads environment specific
terraform vars.
