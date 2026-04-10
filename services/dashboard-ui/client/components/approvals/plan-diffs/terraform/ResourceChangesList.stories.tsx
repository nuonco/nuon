export default {
  title: 'Approvals/PlanDiffs/ResourceChangesList',
}

import { ResourceChangesList } from './ResourceChangesList'

const mockChanges = [
  {
    address: 'aws_s3_bucket.app_assets',
    resource: 'aws_s3_bucket',
    name: 'app_assets',
    action: 'create',
    module: null,
    before: null,
    after: { bucket: 'my-app-assets', acl: 'private' },
  },
  {
    address: 'aws_instance.web_server',
    resource: 'aws_instance',
    name: 'web_server',
    action: 'update',
    module: 'module.networking',
    before: { instance_type: 't3.micro' },
    after: { instance_type: 't3.small' },
  },
  {
    address: 'aws_db_instance.legacy',
    resource: 'aws_db_instance',
    name: 'legacy',
    action: 'delete',
    module: null,
    before: { allocated_storage: 20, engine: 'postgres' },
    after: null,
  },
  {
    address: 'aws_iam_role.service_role',
    resource: 'aws_iam_role',
    name: 'service_role',
    action: 'read',
    module: null,
    before: null,
    after: null,
  },
] as any[]

export const Default = () => <ResourceChangesList changes={mockChanges} />

export const Empty = () => <ResourceChangesList changes={[]} />

const complexMockChanges = [
  {
    address: 'aws_iam_role.service_role',
    resource: 'aws_iam_role',
    name: 'service_role',
    action: 'update',
    module: null,
    before: {
      name: 'my-service-role',
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
      tags: { team: 'platform' },
    },
    after: {
      name: 'my-service-role',
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
      tags: { team: 'platform', managed_by: 'terraform' },
    },
  },
  {
    address: 'aws_eks_cluster.main',
    resource: 'aws_eks_cluster',
    name: 'main',
    action: 'create',
    module: 'module.eks',
    before: null,
    after: {
      name: 'production-cluster',
      version: '1.28',
      vpc_config: {
        subnet_ids: ['subnet-1', 'subnet-2', 'subnet-3'],
        security_group_ids: ['sg-main'],
        endpoint_private_access: true,
        endpoint_public_access: false,
      },
      encryption_config: {
        provider: { key_arn: 'arn:aws:kms:us-west-2:123456789012:key/abc' },
        resources: ['secrets'],
      },
      tags: { environment: 'production', team: 'platform' },
    },
  },
  {
    address: 'aws_security_group.web',
    resource: 'aws_security_group',
    name: 'web',
    action: 'update',
    module: null,
    before: {
      name: 'web-sg',
      description: 'Web security group',
      ingress: [
        { from_port: 80, to_port: 80, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
        { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
      ],
      egress: [
        { from_port: 0, to_port: 0, protocol: '-1', cidr_blocks: ['0.0.0.0/0'] },
      ],
    },
    after: {
      name: 'web-sg',
      description: 'Web security group',
      ingress: [
        { from_port: 80, to_port: 80, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
        { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
        { from_port: 8080, to_port: 8080, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
      ],
      egress: [
        { from_port: 0, to_port: 0, protocol: '-1', cidr_blocks: ['0.0.0.0/0'] },
      ],
    },
  },
  {
    address: 'aws_instance.bastion',
    resource: 'aws_instance',
    name: 'bastion',
    action: 'update',
    module: null,
    before: {
      instance_type: 't3.micro',
      ami: 'ami-old123',
      monitoring: false,
    },
    after: {
      instance_type: 't3.small',
      ami: 'ami-new456',
      monitoring: true,
    },
  },
] as any[]

export const WithComplexValues = () => (
  <ResourceChangesList changes={complexMockChanges} />
)
