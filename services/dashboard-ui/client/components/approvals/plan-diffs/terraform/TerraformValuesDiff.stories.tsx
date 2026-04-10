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
