export default {
  title: 'Approvals/PlanDiffs/TerraformValuesDiff',
}

import { TerraformValuesDiff } from './TerraformValuesDiff'

export const Update = () => (
  <TerraformValuesDiff
    values={{
      action: 'update',
      before: {
        instance_type: 't3.micro',
        min_capacity: 1,
        max_capacity: 3,
      },
      after: {
        instance_type: 't3.small',
        min_capacity: 2,
        max_capacity: 5,
      },
    }}
  />
)

export const Create = () => (
  <TerraformValuesDiff
    values={{
      action: 'create',
      before: null,
      after: {
        bucket: 'my-app-assets',
        acl: 'private',
        region: 'us-east-1',
      },
    }}
  />
)

export const Delete = () => (
  <TerraformValuesDiff
    values={{
      action: 'delete',
      before: {
        name: 'legacy-resource',
        value: 'some-old-value',
      },
      after: null,
    }}
  />
)

export const UpdateWithNestedPolicy = () => (
  <TerraformValuesDiff
    values={{
      action: 'update',
      before: {
        name: 'my-iam-role',
        path: '/',
        assume_role_policy: {
          Version: '2012-10-17',
          Statement: [
            {
              Effect: 'Allow',
              Principal: { Service: 'ec2.amazonaws.com' },
              Action: 'sts:AssumeRole',
            },
          ],
        },
        max_session_duration: 3600,
      },
      after: {
        name: 'my-iam-role',
        path: '/',
        assume_role_policy: {
          Version: '2012-10-17',
          Statement: [
            {
              Effect: 'Allow',
              Principal: {
                Service: ['ec2.amazonaws.com', 'ecs-tasks.amazonaws.com'],
              },
              Action: 'sts:AssumeRole',
            },
          ],
        },
        max_session_duration: 7200,
      },
    }}
  />
)

export const CreateWithNestedValues = () => (
  <TerraformValuesDiff
    values={{
      action: 'create',
      before: null,
      after: {
        name: 'my-cluster',
        version: '1.28',
        node_config: {
          machine_type: 'e2-standard-4',
          disk_size_gb: 100,
          oauth_scopes: [
            'https://www.googleapis.com/auth/cloud-platform',
          ],
          metadata: {
            disable_legacy_endpoints: true,
          },
        },
        network_policy: {
          enabled: true,
          provider: 'CALICO',
        },
      },
    }}
  />
)

export const CreateWithNullValues = () => (
  <TerraformValuesDiff
    values={{
      action: 'create',
      before: null,
      after: {
        name: 'my-registry',
        cleanup_policy_dry_run: null,
        description: null,
        docker_config: {
          immutable_tags: false,
        },
        format: 'DOCKER',
        labels: null,
        location: 'us-central1',
      },
    }}
  />
)

export const UpdateWithMixedValues = () => (
  <TerraformValuesDiff
    values={{
      action: 'update',
      before: {
        instance_type: 't3.small',
        ami: 'ami-0123456789abcdef0',
        tags: {
          Name: 'web-server',
          Environment: 'staging',
        },
        root_block_device: {
          volume_size: 20,
          volume_type: 'gp2',
          encrypted: false,
        },
        monitoring: false,
      },
      after: {
        instance_type: 't3.medium',
        ami: 'ami-0123456789abcdef0',
        tags: {
          Name: 'web-server',
          Environment: 'production',
          Team: 'platform',
        },
        root_block_device: {
          volume_size: 50,
          volume_type: 'gp3',
          encrypted: true,
        },
        monitoring: true,
      },
    }}
  />
)

export const TruncatedLongStrings = () => (
  <TerraformValuesDiff
    values={{
      action: 'update',
      before: {
        allowed_origins:
          'https://app.example.com,https://staging.example.com,https://dev.example.com',
        iam_role_arn:
          'arn:aws:iam::123456789012:role/short-role',
        policy:
          '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":"arn:aws:s3:::my-bucket/*"}]}',
        description: 'A short desc',
      },
      after: {
        allowed_origins:
          'https://app.example.com,https://staging.example.com,https://dev.example.com,https://preview.example.com,https://canary.example.com,https://internal.example.com',
        iam_role_arn:
          'arn:aws:iam::123456789012:role/very-long-cross-account-role-name-for-production-eks-cluster-access-with-additional-suffix-v2',
        policy:
          '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject","s3:PutObject","s3:DeleteObject","s3:ListBucket"],"Resource":"arn:aws:s3:::my-bucket/*"},{"Effect":"Allow","Action":["kms:Decrypt","kms:GenerateDataKey"],"Resource":"arn:aws:kms:us-west-2:123456789012:key/mrk-1234"}]}',
        description: 'A short desc',
      },
    }}
  />
)

export const TruncatedYamlConfig = () => (
  <TerraformValuesDiff
    values={{
      action: 'update',
      before: {
        name: 'prometheus-config',
        config_yaml:
          'global:\n  scrape_interval: 30s\n  evaluation_interval: 30s\nscrape_configs:\n  - job_name: kubernetes-pods\n    kubernetes_sd_configs:\n      - role: pod',
        nginx_conf:
          'upstream backend { server backend-1.prod.svc.cluster.local:8080; server backend-2.prod.svc.cluster.local:8080; }',
        short_value: 'hello',
      },
      after: {
        name: 'prometheus-config',
        config_yaml:
          'global:\n  scrape_interval: 15s\n  evaluation_interval: 15s\n  external_labels:\n    cluster: production\nscrape_configs:\n  - job_name: kubernetes-pods\n    kubernetes_sd_configs:\n      - role: pod\n  - job_name: istio-mesh\n    kubernetes_sd_configs:\n      - role: pod',
        nginx_conf:
          'upstream backend { server backend-1.prod.svc.cluster.local:8080; server backend-2.prod.svc.cluster.local:8080; server backend-3.prod.svc.cluster.local:8080; } server { listen 80; listen 443 ssl; location / { proxy_pass http://backend; } }',
        short_value: 'world',
      },
    }}
  />
)

export const KubectlManifestYamlBody = () => (
  <TerraformValuesDiff
    values={{
      action: 'update',
      before: {
        api_version: 'apps/v1',
        apply_only: false,
        force_conflicts: false,
        force_new: false,
        kind: 'Deployment',
        name: 'restate-cloud-ingress',
        namespace: 'restate-cloud-ingress',
        server_side_apply: false,
        wait_for_rollout: true,
        yaml_body:
          'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: restate-cloud-ingress\n  namespace: restate-cloud-ingress\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: restate-cloud-ingress\n  template:\n    spec:\n      containers:\n      - name: ingress\n        image: registry.example.com/restate-cloud-ingress:v1.2.0\n        ports:\n        - containerPort: 8080\n        resources:\n          requests:\n            cpu: 100m\n            memory: 128Mi\n          limits:\n            cpu: 500m\n            memory: 256Mi',
        yaml_body_parsed:
          'apiVersion: apps/v1 kind: Deployment metadata: name: restate-cloud-ingress namespace: restate-cloud-ingress spec: replicas: 2',
      },
      after: {
        api_version: 'apps/v1',
        apply_only: false,
        force_conflicts: false,
        force_new: false,
        kind: 'Deployment',
        name: 'restate-cloud-ingress',
        namespace: 'restate-cloud-ingress',
        server_side_apply: false,
        wait_for_rollout: true,
        yaml_body:
          'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: restate-cloud-ingress\n  namespace: restate-cloud-ingress\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: restate-cloud-ingress\n  template:\n    spec:\n      containers:\n      - name: ingress\n        image: registry.example.com/restate-cloud-ingress:v1.3.0\n        ports:\n        - containerPort: 8080\n        - containerPort: 9090\n          name: metrics\n        resources:\n          requests:\n            cpu: 200m\n            memory: 256Mi\n          limits:\n            cpu: "1"\n            memory: 512Mi\n        env:\n        - name: OTEL_ENABLED\n          value: "true"',
        yaml_body_parsed:
          'apiVersion: apps/v1 kind: Deployment metadata: name: restate-cloud-ingress namespace: restate-cloud-ingress spec: replicas: 3',
      },
    }}
  />
)
