export default {
  title: 'Approvals/PlanDiffs/TreeDiffValue',
}

import { TreeDiffValue } from './TreeDiffValue'

export const NestedIAMPolicy = () => (
  <TreeDiffValue
    before={{
      assume_role_policy: JSON.stringify({
        Version: '2012-10-17',
        Statement: [
          {
            Effect: 'Allow',
            Principal: { Service: 'ec2.amazonaws.com' },
            Action: 'sts:AssumeRole',
          },
        ],
      }),
      name: 'my-service-role',
      path: '/',
    }}
    after={{
      assume_role_policy: JSON.stringify({
        Version: '2012-10-17',
        Statement: [
          {
            Effect: 'Allow',
            Principal: {
              Service: ['ec2.amazonaws.com', 'ecs-tasks.amazonaws.com'],
            },
            Action: 'sts:AssumeRole',
          },
          {
            Effect: 'Allow',
            Principal: { AWS: 'arn:aws:iam::123456789012:root' },
            Action: 'sts:AssumeRole',
            Condition: {
              StringEquals: { 'sts:ExternalId': 'my-external-id' },
            },
          },
        ],
      }),
      name: 'my-service-role',
      path: '/',
    }}
  />
)

export const KubernetesMetadata = () => (
  <TreeDiffValue
    before={{
      annotations: {
        'kubernetes.io/ingress.class': 'nginx',
        'cert-manager.io/cluster-issuer': 'letsencrypt-prod',
      },
      labels: {
        app: 'my-service',
        environment: 'staging',
      },
    }}
    after={{
      annotations: {
        'kubernetes.io/ingress.class': 'alb',
        'cert-manager.io/cluster-issuer': 'letsencrypt-prod',
        'alb.ingress.kubernetes.io/scheme': 'internet-facing',
      },
      labels: {
        app: 'my-service',
        environment: 'production',
        version: 'v2',
      },
    }}
  />
)

export const SecurityGroupRules = () => (
  <TreeDiffValue
    before={[
      { from_port: 80, to_port: 80, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
      { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
    ]}
    after={[
      { from_port: 80, to_port: 80, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
      { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
      { from_port: 8080, to_port: 8080, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
    ]}
  />
)

export const CreateWithComplexValue = () => (
  <TreeDiffValue
    before={null}
    after={{
      cluster_config: {
        node_pools: [
          { name: 'default', min_size: 1, max_size: 5, instance_type: 't3.medium' },
          { name: 'gpu', min_size: 0, max_size: 2, instance_type: 'p3.2xlarge' },
        ],
        networking: {
          vpc_id: 'vpc-abc123',
          subnet_ids: ['subnet-1', 'subnet-2', 'subnet-3'],
          security_group_ids: ['sg-main'],
        },
        logging: { enabled: true, retention_days: 30 },
      },
    }}
  />
)

export const DeleteWithComplexValue = () => (
  <TreeDiffValue
    before={{
      database: {
        engine: 'postgres',
        version: '14.5',
        allocated_storage: 100,
        multi_az: true,
        backup_retention_period: 7,
        tags: { team: 'platform', cost_center: 'eng-123' },
      },
    }}
    after={null}
  />
)

export const MixedScalarAndComplex = () => (
  <TreeDiffValue
    before={{
      name: 'my-instance',
      tags: { env: 'staging', team: 'backend' },
      count: 2,
    }}
    after={{
      name: 'my-instance',
      tags: { env: 'production', team: 'backend', version: 'v3' },
      count: 3,
    }}
  />
)

export const DeepNesting = () => (
  <TreeDiffValue
    before={{
      level1: {
        level2: {
          level3: {
            level4: {
              level5: { level6: { value: 'old' } },
            },
          },
        },
      },
    }}
    after={{
      level1: {
        level2: {
          level3: {
            level4: {
              level5: { level6: { value: 'new', extra: true } },
            },
          },
        },
      },
    }}
  />
)

export const JsonStringValue = () => (
  <TreeDiffValue
    before={JSON.stringify({ key: 'old-value', nested: { a: 1 } })}
    after={JSON.stringify({ key: 'new-value', nested: { a: 1, b: 2 } })}
  />
)

export const KnownAfterApplyNested = () => (
  <TreeDiffValue
    before={null}
    after={{
      id: 'Known after apply',
      arn: 'Known after apply',
      name: 'my-resource',
      tags: { env: 'prod' },
    }}
  />
)

export const UnchangedComplexValue = () => (
  <TreeDiffValue
    before={{
      ingress: [
        { from_port: 443, to_port: 443, protocol: 'tcp' },
      ],
      egress: [
        { from_port: 0, to_port: 0, protocol: '-1' },
      ],
    }}
    after={{
      ingress: [
        { from_port: 443, to_port: 443, protocol: 'tcp' },
      ],
      egress: [
        { from_port: 0, to_port: 0, protocol: '-1' },
      ],
    }}
  />
)
