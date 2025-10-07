# GitHub Actions Self-Hosted Runners

This Terraform module deploys GitHub Actions self-hosted runners on the infra-shared-ci EKS cluster using the Actions Runner Controller (ARC).

## Architecture

### Two-Chart Deployment

1. **Controller Chart** (`gha-runner-scale-set-controller`)
   - Manages the lifecycle of runner scale sets
   - Handles GitHub API communication
   - Deployed once per cluster

2. **Scale Set Charts** (`gha-runner-scale-set`)
   - Creates autoscaling runner sets
   - Multiple scale sets can be deployed with different configurations
   - Each scale set can have different runner types, resources, and configurations

### Environment-Based Scale Set Configuration

**Key Architecture**: Scale sets are **only defined in environment files**, not in defaults.yaml.

- `vars/defaults.yaml`: Infrastructure defaults, controller config, reusable templates
- `vars/infra-shared-ci.yaml`: Actual scale sets for this environment
- Other environments would have their own var files with their specific scale sets

## Configuration

### Variables Structure

```yaml
scale_sets:
  runner-name:
    github_config_url: "https://github.com/org/repo"
    max_runners: 10
    min_runners: 1
    container_mode:
      type: "dind"  # or "kubernetes"
    template:
      spec:
        containers:
          - name: runner
            resources:
              limits:
                cpu: 4000m
                memory: 8Gi
```

### Required Variables

- `github_token`: GitHub PAT token with repo admin permissions
- `env`: Environment name (defaults to "infra-shared-ci")

### Node Pool Integration

Runners are deployed on the dedicated `self-hosted-runners` node pool:
- **Instance Types**: `c5a.xlarge` (4 vCPU, 8GB RAM)
- **Node Selector**: `karpenter.sh/nodepool: self-hosted-runners`
- **Tolerations**: `pool.nuon.co=self-hosted-runners:NoSchedule`

## Deployment

1. **Set GitHub Token**:
   ```bash
   export TF_VAR_github_token="ghp_your_token_here"
   ```

2. **Initialize and Apply**:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

3. **Verify Deployment**:
   ```bash
   kubectl get pods -n arc-runners
   kubectl get autoscalingrunnersets -n arc-runners
   ```

## Container Modes

### Docker-in-Docker (dind)
- Full Docker daemon in each runner
- Suitable for Docker builds, multi-container workflows
- Higher resource requirements
- Privileged containers

### Kubernetes Mode
- Jobs run as separate Kubernetes pods
- Better resource isolation
- Suitable for simple builds, testing
- More secure (no privileged containers)

## Scaling Behavior

- **Min Runners**: Idle runners always available
- **Max Runners**: Maximum concurrent runners
- **Auto-scaling**: Based on GitHub job queue
- **Node Scaling**: Karpenter provisions nodes as needed

## Security

- **RBAC**: Proper Kubernetes RBAC for controller and runners
- **Service Accounts**: Dedicated service accounts per scale set
- **Network Policies**: (Optional) Can be added for network isolation
- **Resource Limits**: CPU/memory limits prevent resource exhaustion

## Monitoring

- **Metrics**: Exposed via Prometheus metrics
- **Logs**: Available via `kubectl logs`
- **Events**: Kubernetes events for troubleshooting

## Troubleshooting

### Check Controller Status
```bash
kubectl get pods -n arc-runners -l app.kubernetes.io/name=gha-runner-scale-set-controller
kubectl logs -n arc-runners -l app.kubernetes.io/name=gha-runner-scale-set-controller
```

### Check Scale Set Status
```bash
kubectl get autoscalingrunnersets -n arc-runners
kubectl describe autoscalingrunnersets <scale-set-name> -n arc-runners
```

### Check Runner Pods
```bash
kubectl get pods -n arc-runners -l actions.github.com/scale-set-name=<scale-set-name>
```

### Common Issues

1. **GitHub Token Issues**: Verify token has correct permissions
2. **Node Availability**: Check if Karpenter can provision nodes
3. **Resource Limits**: Ensure sufficient cluster resources
4. **Image Pull**: Verify access to ghcr.io images

## Customization

### Adding New Scale Sets

Add to `vars/defaults.yaml`:
```yaml
scale_sets:
  my-custom-runner:
    github_config_url: "https://github.com/myorg/myrepo"
    max_runners: 5
    min_runners: 0
    container_mode:
      type: "dind"
    template:
      spec:
        containers:
          - name: runner
            image: my-custom-image:latest
```

### Environment Overrides

Use `vars/infra-shared-ci.yaml` for environment-specific settings.

### Custom Runner Images

Build custom images extending `ghcr.io/actions/actions-runner:latest` with additional tools.

## Integration with Existing Infrastructure

This deployment integrates with:
- **EKS Cluster**: Uses existing infra-shared-ci cluster
- **Karpenter**: Uses existing node pool configuration
- **VPC/Networking**: Uses cluster's existing network setup
- **IAM**: Uses cluster's existing IAM roles