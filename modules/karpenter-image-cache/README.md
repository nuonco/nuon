# karpenter-image-cache

Terraform module that creates an EBS snapshot with pre-pulled container images for Karpenter nodes.

## Problem

Large container images (multi-GB) cause slow node scale-out because every new Karpenter node must pull images from the registry. This module pre-bakes images into an EBS snapshot that Karpenter nodes mount on launch, making images available instantly.

## How it works

1. Launches a temporary EC2 builder instance in the install's VPC
2. Attaches a secondary EBS volume
3. Pulls all specified container images into containerd's `k8s.io` namespace on that volume
4. Builder shuts itself down after pulling completes
5. Creates an EBS snapshot of the volume
6. Outputs `snapshot_id` for use in Karpenter EC2NodeClass `blockDeviceMappings`

## Usage in Nuon app config

### 1. Terraform module component (`components/image_cache.toml`)

```toml
name              = "image_cache"
type              = "terraform_module"
terraform_version = "1.11.3"

[public_repo]
repo      = "nuonco/karpenter-image-cache"
directory = "."
branch    = "main"

[vars]
region    = "{{.nuon.install_stack.outputs.region}}"
vpc_id    = "{{.nuon.install.sandbox.outputs.vpc.id}}"
subnet_id = "{{.nuon.install.sandbox.outputs.vpc.runner_subnet_id}}"
images    = '["your-registry.com/big-image:v1", "another/huge-image:latest"]'
```

### 2. EC2NodeClass component (`components/node_class.toml`)

```toml
name      = "node_class"
type      = "kubernetes_manifest"
namespace = "kube-system"
manifest  = "./manifests/ec2nodeclass.yaml"
```

With `manifests/ec2nodeclass.yaml`:
```yaml
apiVersion: karpenter.k8s.aws/v1
kind: EC2NodeClass
metadata:
  name: cached
spec:
  amiSelectorTerms:
    - alias: al2023@latest
  role: "{{.nuon.install.sandbox.outputs.karpenter.instance_profile.name}}"
  subnetSelectorTerms:
    - tags:
        karpenter.sh/discovery: "{{.nuon.install.sandbox.outputs.cluster.name}}"
  securityGroupSelectorTerms:
    - tags:
        karpenter.sh/discovery: "{{.nuon.install.sandbox.outputs.cluster.name}}"
  blockDeviceMappings:
    - deviceName: /dev/xvda
      ebs:
        volumeSize: 20Gi
        volumeType: gp3
        deleteOnTermination: true
    - deviceName: /dev/xvdf
      ebs:
        snapshotID: "{{.nuon.components.image_cache.outputs.snapshot_id}}"
        volumeSize: "{{.nuon.components.image_cache.outputs.volume_size_gb}}Gi"
        volumeType: gp3
        deleteOnTermination: true
  userData: |
    #!/bin/bash
    # Mount the pre-cached image volume into containerd's data directory
    # This runs BEFORE the EKS bootstrap script, so containerd hasn't started yet.
    DEVICE="/dev/xvdf"
    for i in $(seq 1 60); do [ -b "$DEVICE" ] && break; sleep 1; done
    if [ -b "$DEVICE" ]; then
      # Copy cached containerd data before containerd starts
      mkdir -p /tmp/image-cache
      mount "$DEVICE" /tmp/image-cache
      if [ -d /tmp/image-cache/containerd ]; then
        mkdir -p /var/lib/containerd
        cp -a /tmp/image-cache/containerd/* /var/lib/containerd/
      fi
      umount /tmp/image-cache
    fi
```

### 3. NodePool component (`components/node_pool.toml`)

```toml
name         = "node_pool"
type         = "kubernetes_manifest"
namespace    = "kube-system"
dependencies = ["node_class"]
manifest     = "./manifests/nodepool.yaml"
```

## Inputs

| Name | Description | Type | Default |
|------|-------------|------|---------|
| images | Container images to pre-cache | list(string) | required |
| vpc_id | VPC ID for the builder | string | required |
| subnet_id | Subnet with internet access | string | required |
| region | AWS region | string | required |
| volume_size_gb | Cache volume size in GB | number | 100 |
| volume_type | EBS volume type | string | gp3 |
| instance_type | Builder instance type | string | m5.xlarge |
| tags | Additional resource tags | map(string) | {} |

## Outputs

| Name | Description |
|------|-------------|
| snapshot_id | EBS snapshot ID for EC2NodeClass blockDeviceMappings |
| snapshot_arn | EBS snapshot ARN |
| volume_size_gb | Volume size (pass through to blockDeviceMappings) |
