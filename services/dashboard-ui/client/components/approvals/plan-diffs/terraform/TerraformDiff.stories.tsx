export default {
  title: 'Approvals/PlanDiffs/TerraformDiff',
}

import { TerraformDiff } from './TerraformDiff'

export const NoPlan = () => <TerraformDiff plan={undefined} />

export const WithPlan = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'aws_s3_bucket.app_assets',
            type: 'aws_s3_bucket',
            name: 'app_assets',
            change: {
              actions: ['create'],
              before: null,
              after: { bucket: 'my-app-assets', acl: 'private' },
            },
          },
          {
            address: 'aws_instance.web',
            type: 'aws_instance',
            name: 'web',
            change: {
              actions: ['update'],
              before: { instance_type: 't3.micro' },
              after: { instance_type: 't3.small' },
            },
          },
        ],
        output_changes: {},
      } as any
    }
  />
)

export const IAMRoleWithNestedPolicy = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'aws_iam_role.ecs_task_execution',
            type: 'aws_iam_role',
            name: 'ecs_task_execution',
            module_address: 'module.ecs',
            change: {
              actions: ['update'],
              before: {
                name: 'ecs-task-execution',
                path: '/',
                assume_role_policy: JSON.stringify({
                  Version: '2012-10-17',
                  Statement: [
                    {
                      Effect: 'Allow',
                      Principal: { Service: 'ecs-tasks.amazonaws.com' },
                      Action: 'sts:AssumeRole',
                    },
                  ],
                }),
                max_session_duration: 3600,
                tags: { team: 'platform', managed_by: 'terraform' },
              },
              after: {
                name: 'ecs-task-execution',
                path: '/',
                assume_role_policy: JSON.stringify({
                  Version: '2012-10-17',
                  Statement: [
                    {
                      Effect: 'Allow',
                      Principal: {
                        Service: [
                          'ecs-tasks.amazonaws.com',
                          'lambda.amazonaws.com',
                        ],
                      },
                      Action: 'sts:AssumeRole',
                    },
                    {
                      Effect: 'Allow',
                      Principal: {
                        AWS: 'arn:aws:iam::123456789012:role/deploy-role',
                      },
                      Action: 'sts:AssumeRole',
                      Condition: {
                        StringEquals: {
                          'sts:ExternalId': 'nuon-deploy',
                        },
                      },
                    },
                  ],
                }),
                max_session_duration: 7200,
                tags: {
                  team: 'platform',
                  managed_by: 'terraform',
                  updated: '2026-04-10',
                },
              },
            },
          },
          {
            address: 'aws_iam_role_policy.ecs_secrets',
            type: 'aws_iam_role_policy',
            name: 'ecs_secrets',
            module_address: 'module.ecs',
            change: {
              actions: ['create'],
              before: null,
              after: {
                name: 'ecs-secrets-access',
                role: 'ecs-task-execution',
                policy: JSON.stringify({
                  Version: '2012-10-17',
                  Statement: [
                    {
                      Effect: 'Allow',
                      Action: [
                        'secretsmanager:GetSecretValue',
                        'ssm:GetParameters',
                      ],
                      Resource: '*',
                    },
                  ],
                }),
              },
            },
          },
        ],
        output_changes: {
          role_arn: {
            actions: ['update'],
            before: 'arn:aws:iam::123456789012:role/ecs-task-execution',
            after: 'arn:aws:iam::123456789012:role/ecs-task-execution',
          },
        },
      } as any
    }
  />
)

export const EKSClusterCreate = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'aws_eks_cluster.main',
            type: 'aws_eks_cluster',
            name: 'main',
            module_address: 'module.eks',
            change: {
              actions: ['create'],
              before: null,
              after: {
                name: 'production-cluster',
                role_arn: 'Known after apply',
                version: '1.29',
                vpc_config: {
                  subnet_ids: [
                    'subnet-0a1b2c3d4e5f60001',
                    'subnet-0a1b2c3d4e5f60002',
                    'subnet-0a1b2c3d4e5f60003',
                  ],
                  security_group_ids: ['sg-0a1b2c3d4e5f6000'],
                  endpoint_private_access: true,
                  endpoint_public_access: false,
                },
                encryption_config: {
                  provider: {
                    key_arn:
                      'arn:aws:kms:us-west-2:123456789012:key/abc-def-123',
                  },
                  resources: ['secrets'],
                },
                kubernetes_network_config: {
                  service_ipv4_cidr: '172.20.0.0/16',
                  ip_family: 'ipv4',
                },
                tags: {
                  environment: 'production',
                  team: 'platform',
                  cost_center: 'eng-infra',
                },
              },
              after_unknown: {
                arn: true,
                certificate_authority: true,
                cluster_id: true,
                endpoint: true,
                id: true,
                platform_version: true,
                status: true,
              },
            },
          },
          {
            address: 'aws_eks_node_group.default',
            type: 'aws_eks_node_group',
            name: 'default',
            module_address: 'module.eks',
            change: {
              actions: ['create'],
              before: null,
              after: {
                cluster_name: 'production-cluster',
                node_group_name: 'default',
                node_role_arn: 'Known after apply',
                instance_types: ['t3.xlarge'],
                scaling_config: {
                  desired_size: 3,
                  max_size: 10,
                  min_size: 2,
                },
                update_config: { max_unavailable: 1 },
                labels: { role: 'general', managed_by: 'terraform' },
                tags: { environment: 'production' },
              },
              after_unknown: {
                arn: true,
                id: true,
                status: true,
                resources: true,
              },
            },
          },
          {
            address: 'aws_eks_addon.vpc_cni',
            type: 'aws_eks_addon',
            name: 'vpc_cni',
            module_address: 'module.eks',
            change: {
              actions: ['create'],
              before: null,
              after: {
                addon_name: 'vpc-cni',
                cluster_name: 'production-cluster',
                addon_version: 'v1.16.0-eksbuild.1',
                resolve_conflicts_on_create: 'OVERWRITE',
                configuration_values: JSON.stringify({
                  env: {
                    ENABLE_PREFIX_DELEGATION: 'true',
                    WARM_PREFIX_TARGET: '1',
                  },
                }),
              },
            },
          },
        ],
        output_changes: {
          cluster_endpoint: {
            actions: ['create'],
            before: null,
            after: 'Known after apply',
            after_unknown: true,
          },
          cluster_ca_certificate: {
            actions: ['create'],
            before: null,
            after: 'Known after apply',
            after_unknown: true,
          },
        },
      } as any
    }
  />
)

export const SecurityGroupUpdate = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'aws_security_group.web',
            type: 'aws_security_group',
            name: 'web',
            change: {
              actions: ['update'],
              before: {
                name: 'web-sg',
                description: 'Web-facing security group',
                vpc_id: 'vpc-abc123',
                ingress: [
                  {
                    from_port: 80,
                    to_port: 80,
                    protocol: 'tcp',
                    cidr_blocks: ['0.0.0.0/0'],
                    description: 'HTTP from anywhere',
                  },
                  {
                    from_port: 443,
                    to_port: 443,
                    protocol: 'tcp',
                    cidr_blocks: ['0.0.0.0/0'],
                    description: 'HTTPS from anywhere',
                  },
                ],
                egress: [
                  {
                    from_port: 0,
                    to_port: 0,
                    protocol: '-1',
                    cidr_blocks: ['0.0.0.0/0'],
                    description: 'Allow all outbound',
                  },
                ],
                tags: { Name: 'web-sg', environment: 'production' },
              },
              after: {
                name: 'web-sg',
                description: 'Web-facing security group',
                vpc_id: 'vpc-abc123',
                ingress: [
                  {
                    from_port: 80,
                    to_port: 80,
                    protocol: 'tcp',
                    cidr_blocks: ['10.0.0.0/8'],
                    description: 'HTTP from VPC',
                  },
                  {
                    from_port: 443,
                    to_port: 443,
                    protocol: 'tcp',
                    cidr_blocks: ['10.0.0.0/8'],
                    description: 'HTTPS from VPC',
                  },
                  {
                    from_port: 8080,
                    to_port: 8080,
                    protocol: 'tcp',
                    cidr_blocks: ['10.0.0.0/8'],
                    description: 'App port from VPC',
                  },
                ],
                egress: [
                  {
                    from_port: 0,
                    to_port: 0,
                    protocol: '-1',
                    cidr_blocks: ['0.0.0.0/0'],
                    description: 'Allow all outbound',
                  },
                ],
                tags: { Name: 'web-sg', environment: 'production' },
              },
            },
          },
          {
            address: 'aws_security_group.internal',
            type: 'aws_security_group',
            name: 'internal',
            change: {
              actions: ['create'],
              before: null,
              after: {
                name: 'internal-sg',
                description: 'Internal services security group',
                vpc_id: 'vpc-abc123',
                ingress: [
                  {
                    from_port: 5432,
                    to_port: 5432,
                    protocol: 'tcp',
                    security_groups: ['sg-web123'],
                    description: 'Postgres from web tier',
                  },
                  {
                    from_port: 6379,
                    to_port: 6379,
                    protocol: 'tcp',
                    security_groups: ['sg-web123'],
                    description: 'Redis from web tier',
                  },
                ],
                egress: [
                  {
                    from_port: 0,
                    to_port: 0,
                    protocol: '-1',
                    cidr_blocks: ['0.0.0.0/0'],
                    description: 'Allow all outbound',
                  },
                ],
                tags: { Name: 'internal-sg', environment: 'production' },
              },
            },
          },
        ],
        output_changes: {},
      } as any
    }
  />
)

export const RDSReplace = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'aws_db_instance.main',
            type: 'aws_db_instance',
            name: 'main',
            module_address: 'module.database',
            change: {
              actions: ['delete', 'create'],
              before: {
                identifier: 'myapp-db',
                engine: 'postgres',
                engine_version: '14.10',
                instance_class: 'db.r6g.large',
                allocated_storage: 100,
                max_allocated_storage: 500,
                storage_type: 'gp3',
                storage_encrypted: true,
                multi_az: true,
                db_subnet_group_name: 'myapp-db-subnet',
                vpc_security_group_ids: ['sg-db001'],
                backup_retention_period: 7,
                backup_window: '03:00-04:00',
                maintenance_window: 'Mon:04:00-Mon:05:00',
                parameter_group_name: 'myapp-pg14',
                performance_insights_enabled: true,
                tags: {
                  environment: 'production',
                  team: 'platform',
                  backup: 'daily',
                },
              },
              after: {
                identifier: 'myapp-db-v2',
                engine: 'postgres',
                engine_version: '16.2',
                instance_class: 'db.r7g.large',
                allocated_storage: 200,
                max_allocated_storage: 1000,
                storage_type: 'gp3',
                storage_encrypted: true,
                multi_az: true,
                db_subnet_group_name: 'myapp-db-subnet',
                vpc_security_group_ids: ['sg-db001', 'sg-db002'],
                backup_retention_period: 14,
                backup_window: '03:00-04:00',
                maintenance_window: 'Mon:04:00-Mon:05:00',
                parameter_group_name: 'myapp-pg16',
                performance_insights_enabled: true,
                deletion_protection: true,
                tags: {
                  environment: 'production',
                  team: 'platform',
                  backup: 'daily',
                  migrated_from: 'myapp-db',
                },
              },
            },
          },
          {
            address: 'aws_db_parameter_group.pg16',
            type: 'aws_db_parameter_group',
            name: 'pg16',
            module_address: 'module.database',
            change: {
              actions: ['create'],
              before: null,
              after: {
                name: 'myapp-pg16',
                family: 'postgres16',
                parameter: [
                  {
                    name: 'shared_preload_libraries',
                    value: 'pg_stat_statements,auto_explain',
                  },
                  { name: 'log_min_duration_statement', value: '1000' },
                  { name: 'max_connections', value: '200' },
                ],
                tags: { environment: 'production' },
              },
            },
          },
          {
            address: 'aws_db_parameter_group.pg14',
            type: 'aws_db_parameter_group',
            name: 'pg14',
            module_address: 'module.database',
            change: {
              actions: ['delete'],
              before: {
                name: 'myapp-pg14',
                family: 'postgres14',
                parameter: [
                  {
                    name: 'shared_preload_libraries',
                    value: 'pg_stat_statements',
                  },
                  { name: 'log_min_duration_statement', value: '2000' },
                ],
                tags: { environment: 'production' },
              },
              after: null,
            },
          },
        ],
        output_changes: {
          db_endpoint: {
            actions: ['update'],
            before: 'myapp-db.abc123.us-west-2.rds.amazonaws.com:5432',
            after: 'Known after apply',
            after_unknown: true,
          },
          db_identifier: {
            actions: ['update'],
            before: 'myapp-db',
            after: 'myapp-db-v2',
          },
        },
      } as any
    }
  />
)

export const NoOpAndReadResources = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'data.aws_caller_identity.current',
            type: 'aws_caller_identity',
            name: 'current',
            change: {
              actions: ['read'],
              before: null,
              after: {
                account_id: 'Known after apply',
                arn: 'Known after apply',
                user_id: 'Known after apply',
              },
              after_unknown: {
                account_id: true,
                arn: true,
                user_id: true,
              },
            },
          },
          {
            address: 'data.aws_region.current',
            type: 'aws_region',
            name: 'current',
            change: {
              actions: ['read'],
              before: null,
              after: {
                name: 'Known after apply',
                endpoint: 'Known after apply',
                description: 'Known after apply',
              },
              after_unknown: {
                name: true,
                endpoint: true,
                description: true,
              },
            },
          },
          {
            address: 'data.aws_vpc.selected',
            type: 'aws_vpc',
            name: 'selected',
            change: {
              actions: ['read'],
              before: null,
              after: {
                id: 'Known after apply',
                cidr_block: 'Known after apply',
                enable_dns_support: true,
                enable_dns_hostnames: true,
                tags: { environment: 'production' },
              },
              after_unknown: {
                id: true,
                cidr_block: true,
              },
            },
          },
          {
            address: 'aws_s3_bucket.logs',
            type: 'aws_s3_bucket',
            name: 'logs',
            change: {
              actions: ['no-op'],
              before: {
                bucket: 'myapp-logs-production',
                acl: 'private',
                versioning: { enabled: true },
                server_side_encryption_configuration: {
                  rule: {
                    apply_server_side_encryption_by_default: {
                      sse_algorithm: 'aws:kms',
                    },
                  },
                },
                tags: { environment: 'production', team: 'platform' },
              },
              after: {
                bucket: 'myapp-logs-production',
                acl: 'private',
                versioning: { enabled: true },
                server_side_encryption_configuration: {
                  rule: {
                    apply_server_side_encryption_by_default: {
                      sse_algorithm: 'aws:kms',
                    },
                  },
                },
                tags: { environment: 'production', team: 'platform' },
              },
            },
          },
          {
            address: 'aws_iam_role.lambda_exec',
            type: 'aws_iam_role',
            name: 'lambda_exec',
            change: {
              actions: ['no-op'],
              before: {
                name: 'lambda-exec-role',
                path: '/',
                assume_role_policy: JSON.stringify({
                  Version: '2012-10-17',
                  Statement: [
                    {
                      Effect: 'Allow',
                      Principal: { Service: 'lambda.amazonaws.com' },
                      Action: 'sts:AssumeRole',
                    },
                  ],
                }),
                tags: { managed_by: 'terraform' },
              },
              after: {
                name: 'lambda-exec-role',
                path: '/',
                assume_role_policy: JSON.stringify({
                  Version: '2012-10-17',
                  Statement: [
                    {
                      Effect: 'Allow',
                      Principal: { Service: 'lambda.amazonaws.com' },
                      Action: 'sts:AssumeRole',
                    },
                  ],
                }),
                tags: { managed_by: 'terraform' },
              },
            },
          },
        ],
        output_changes: {
          account_id: {
            actions: ['read'],
            before: null,
            after: 'Known after apply',
            after_unknown: true,
          },
        },
      } as any
    }
  />
)

export const ReplaceResources = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'aws_instance.web',
            type: 'aws_instance',
            name: 'web',
            module_address: 'module.compute',
            change: {
              actions: ['delete', 'create'],
              before: {
                ami: 'ami-0a1b2c3d4e5f60001',
                instance_type: 't3.medium',
                subnet_id: 'subnet-old001',
                vpc_security_group_ids: ['sg-old001'],
                root_block_device: {
                  volume_size: 20,
                  volume_type: 'gp2',
                  encrypted: false,
                },
                tags: { Name: 'web-server', environment: 'production' },
              },
              after: {
                ami: 'ami-0f9b8c7d6e5a40002',
                instance_type: 't3.medium',
                subnet_id: 'subnet-new001',
                vpc_security_group_ids: ['sg-new001'],
                root_block_device: {
                  volume_size: 50,
                  volume_type: 'gp3',
                  encrypted: true,
                },
                tags: { Name: 'web-server', environment: 'production' },
              },
            },
          },
          {
            address: 'aws_eip.web',
            type: 'aws_eip',
            name: 'web',
            module_address: 'module.compute',
            change: {
              actions: ['create', 'delete'],
              before: {
                instance: 'i-old12345',
                public_ip: '54.200.100.50',
                domain: 'vpc',
                tags: { Name: 'web-eip' },
              },
              after: {
                instance: 'Known after apply',
                public_ip: 'Known after apply',
                domain: 'vpc',
                tags: { Name: 'web-eip' },
              },
              after_unknown: {
                instance: true,
                public_ip: true,
              },
            },
          },
          {
            address: 'aws_lb_target_group.web',
            type: 'aws_lb_target_group',
            name: 'web',
            module_address: 'module.compute',
            change: {
              actions: ['update'],
              before: {
                name: 'web-tg',
                port: 80,
                protocol: 'HTTP',
                vpc_id: 'vpc-abc123',
                health_check: {
                  path: '/health',
                  interval: 30,
                  timeout: 5,
                  healthy_threshold: 2,
                  unhealthy_threshold: 3,
                },
              },
              after: {
                name: 'web-tg',
                port: 8080,
                protocol: 'HTTP',
                vpc_id: 'vpc-abc123',
                health_check: {
                  path: '/healthz',
                  interval: 15,
                  timeout: 3,
                  healthy_threshold: 2,
                  unhealthy_threshold: 2,
                },
              },
            },
          },
        ],
        output_changes: {
          web_public_ip: {
            actions: ['update'],
            before: '54.200.100.50',
            after: 'Known after apply',
            after_unknown: true,
          },
        },
      } as any
    }
  />
)

export const MixedWithNoOp = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'aws_vpc.main',
            type: 'aws_vpc',
            name: 'main',
            change: {
              actions: ['no-op'],
              before: {
                cidr_block: '10.0.0.0/16',
                enable_dns_support: true,
                enable_dns_hostnames: true,
                tags: { Name: 'main-vpc', environment: 'production' },
              },
              after: {
                cidr_block: '10.0.0.0/16',
                enable_dns_support: true,
                enable_dns_hostnames: true,
                tags: { Name: 'main-vpc', environment: 'production' },
              },
            },
          },
          {
            address: 'aws_subnet.public',
            type: 'aws_subnet',
            name: 'public',
            change: {
              actions: ['update'],
              before: {
                cidr_block: '10.0.1.0/24',
                availability_zone: 'us-west-2a',
                map_public_ip_on_launch: false,
                tags: { Name: 'public-subnet-a' },
              },
              after: {
                cidr_block: '10.0.1.0/24',
                availability_zone: 'us-west-2a',
                map_public_ip_on_launch: true,
                tags: { Name: 'public-subnet-a', tier: 'public' },
              },
            },
          },
          {
            address: 'aws_nat_gateway.main',
            type: 'aws_nat_gateway',
            name: 'main',
            change: {
              actions: ['no-op'],
              before: {
                allocation_id: 'eipalloc-abc123',
                subnet_id: 'subnet-pub001',
                tags: { Name: 'main-nat' },
              },
              after: {
                allocation_id: 'eipalloc-abc123',
                subnet_id: 'subnet-pub001',
                tags: { Name: 'main-nat' },
              },
            },
          },
          {
            address: 'aws_route_table.private',
            type: 'aws_route_table',
            name: 'private',
            change: {
              actions: ['create'],
              before: null,
              after: {
                vpc_id: 'vpc-abc123',
                route: [
                  {
                    cidr_block: '0.0.0.0/0',
                    nat_gateway_id: 'Known after apply',
                  },
                ],
                tags: { Name: 'private-rt', environment: 'production' },
              },
              after_unknown: {
                id: true,
              },
            },
          },
          {
            address: 'aws_security_group_rule.old_ssh',
            type: 'aws_security_group_rule',
            name: 'old_ssh',
            change: {
              actions: ['delete'],
              before: {
                type: 'ingress',
                from_port: 22,
                to_port: 22,
                protocol: 'tcp',
                cidr_blocks: ['0.0.0.0/0'],
                security_group_id: 'sg-web001',
                description: 'SSH from anywhere',
              },
              after: null,
            },
          },
        ],
        output_changes: {},
      } as any
    }
  />
)

export const RBACArrayNoise = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'kubernetes_cluster_role.system_node',
            type: 'kubernetes_cluster_role',
            name: 'system_node',
            module_address: 'module.eks',
            change: {
              actions: ['update'],
              before: {
                metadata: { name: 'system:node' },
                rule: [
                  {
                    api_groups: [''],
                    resources: ['nodes'],
                    verbs: ['get', 'list', 'watch'],
                    resource_names: null,
                  },
                  {
                    api_groups: [''],
                    resources: ['pods'],
                    verbs: ['get', 'list', 'watch'],
                    resource_names: null,
                  },
                  {
                    api_groups: [''],
                    resources: ['services'],
                    verbs: ['get', 'list'],
                    resource_names: null,
                  },
                  {
                    api_groups: ['apps'],
                    resources: ['daemonsets'],
                    verbs: ['get', 'list', 'watch'],
                    resource_names: null,
                  },
                  {
                    api_groups: ['coordination.k8s.io'],
                    resources: ['leases'],
                    verbs: ['get', 'create', 'update'],
                    resource_names: null,
                  },
                ],
              },
              after: {
                metadata: { name: 'system:node' },
                rule: [
                  {
                    api_groups: [''],
                    resources: ['nodes'],
                    verbs: ['get', 'list', 'watch'],
                    resource_names: [],
                  },
                  {
                    api_groups: [''],
                    resources: ['pods'],
                    verbs: ['get', 'list', 'watch'],
                    resource_names: [],
                  },
                  {
                    api_groups: [''],
                    resources: ['services'],
                    verbs: ['get', 'list'],
                    resource_names: [],
                  },
                  {
                    api_groups: ['apps'],
                    resources: ['daemonsets'],
                    verbs: ['get', 'list', 'watch'],
                    resource_names: [],
                  },
                  {
                    api_groups: ['coordination.k8s.io'],
                    resources: ['leases'],
                    verbs: ['get', 'create', 'update', 'patch'],
                    resource_names: [],
                  },
                ],
              },
            },
          },
        ],
        output_changes: {},
      } as any
    }
  />
)

export const AzureNoOpWithCosmeticDrift = () => (
  <TerraformDiff
    plan={
      {
        resource_drift: [
          {
            address: 'azurerm_container_registry.acr',
            type: 'azurerm_container_registry',
            name: 'acr',
            change: {
              actions: ['update'],
              before: {
                admin_enabled: false,
                location: 'centralindia',
                name: 'myregistry',
                sku: 'Premium',
                tags: null,
              },
              after: {
                admin_enabled: false,
                location: 'centralindia',
                name: 'myregistry',
                sku: 'Premium',
                tags: {},
              },
            },
          },
          {
            address: 'azurerm_dns_zone.public',
            type: 'azurerm_dns_zone',
            name: 'public',
            change: {
              actions: ['update'],
              before: {
                name: 'byoc-azure.nuon.co',
                resource_group_name: 'my-rg',
                tags: null,
              },
              after: {
                name: 'byoc-azure.nuon.co',
                resource_group_name: 'my-rg',
                tags: {},
              },
            },
          },
          {
            address: 'azurerm_private_dns_zone.internal',
            type: 'azurerm_private_dns_zone',
            name: 'internal',
            change: {
              actions: ['update'],
              before: {
                name: 'internal.byoc-azure.nuon.co',
                resource_group_name: 'my-rg',
                tags: null,
              },
              after: {
                name: 'internal.byoc-azure.nuon.co',
                resource_group_name: 'my-rg',
                tags: {},
              },
            },
          },
          {
            address: 'module.aks.azurerm_kubernetes_cluster.main',
            module_address: 'module.aks',
            type: 'azurerm_kubernetes_cluster',
            name: 'main',
            change: {
              actions: ['update'],
              before: {
                name: 'my-aks',
                kubernetes_version: '1.33',
                location: 'centralindia',
                azure_active_directory_role_based_access_control: [
                  { admin_group_object_ids: null, azure_rbac_enabled: true },
                ],
                default_node_pool: [
                  {
                    name: 'agents',
                    node_count: 1,
                    vm_size: 'Standard_D2s_v3',
                    tags: null,
                    zones: null,
                  },
                ],
                identity: [
                  {
                    identity_ids: null,
                    principal_id: '2c80e9b4-xxxx',
                    type: 'SystemAssigned',
                  },
                ],
                tags: null,
              },
              after: {
                name: 'my-aks',
                kubernetes_version: '1.33',
                location: 'centralindia',
                azure_active_directory_role_based_access_control: [
                  { admin_group_object_ids: [], azure_rbac_enabled: true },
                ],
                default_node_pool: [
                  {
                    name: 'agents',
                    node_count: 1,
                    vm_size: 'Standard_D2s_v3',
                    tags: {},
                    zones: [],
                  },
                ],
                identity: [
                  {
                    identity_ids: [],
                    principal_id: '2c80e9b4-xxxx',
                    type: 'SystemAssigned',
                  },
                ],
                tags: {},
              },
            },
          },
        ],
        resource_changes: [
          {
            address: 'azapi_resource.ssh_public_key',
            type: 'azapi_resource',
            name: 'ssh_public_key',
            change: {
              actions: ['no-op'],
              before: { name: 'sshstirringcat' },
              after: { name: 'sshstirringcat' },
            },
          },
          {
            address: 'azurerm_container_registry.acr',
            type: 'azurerm_container_registry',
            name: 'acr',
            change: {
              actions: ['no-op'],
              before: { name: 'myregistry', sku: 'Premium' },
              after: { name: 'myregistry', sku: 'Premium' },
            },
          },
          {
            address: 'azurerm_dns_zone.public',
            type: 'azurerm_dns_zone',
            name: 'public',
            change: {
              actions: ['no-op'],
              before: { name: 'byoc-azure.nuon.co' },
              after: { name: 'byoc-azure.nuon.co' },
            },
          },
          {
            address: 'azurerm_private_dns_zone.internal',
            type: 'azurerm_private_dns_zone',
            name: 'internal',
            change: {
              actions: ['no-op'],
              before: { name: 'internal.byoc-azure.nuon.co' },
              after: { name: 'internal.byoc-azure.nuon.co' },
            },
          },
          {
            address: 'module.aks.azurerm_kubernetes_cluster.main',
            module_address: 'module.aks',
            type: 'azurerm_kubernetes_cluster',
            name: 'main',
            change: {
              actions: ['no-op'],
              before: { name: 'my-aks' },
              after: { name: 'my-aks' },
            },
          },
          {
            address: 'module.aks.azurerm_role_assignment.acr["myregistry"]',
            module_address: 'module.aks',
            type: 'azurerm_role_assignment',
            name: 'acr',
            change: {
              actions: ['no-op'],
              before: { role_definition_name: 'AcrPull' },
              after: { role_definition_name: 'AcrPull' },
            },
          },
          {
            address: 'module.aks.null_resource.kubernetes_cluster_name_keeper',
            module_address: 'module.aks',
            type: 'null_resource',
            name: 'kubernetes_cluster_name_keeper',
            change: {
              actions: ['no-op'],
              before: { id: '12345' },
              after: { id: '12345' },
            },
          },
          {
            address: 'random_pet.ssh_key_name',
            type: 'random_pet',
            name: 'ssh_key_name',
            change: {
              actions: ['no-op'],
              before: { id: 'sshstirringcat' },
              after: { id: 'sshstirringcat' },
            },
          },
        ],
        output_changes: {
          account: {
            actions: ['no-op'],
            before: { location: 'centralindia' },
            after: { location: 'centralindia' },
          },
          cluster: {
            actions: ['no-op'],
            before: { name: 'my-aks' },
            after: { name: 'my-aks' },
          },
          acr: {
            actions: ['no-op'],
            before: { name: 'myregistry' },
            after: { name: 'myregistry' },
          },
        },
      } as any
    }
  />
)

export const DriftDetected = () => (
  <TerraformDiff
    plan={
      {
        resource_drift: [
          {
            address: 'aws_autoscaling_group.web',
            type: 'aws_autoscaling_group',
            name: 'web',
            change: {
              actions: ['update'],
              before: {
                name: 'web-asg',
                desired_capacity: 3,
                min_size: 2,
                max_size: 10,
                launch_template: {
                  id: 'lt-abc123',
                  version: '5',
                },
                tags: [
                  { key: 'environment', value: 'production' },
                  { key: 'team', value: 'platform' },
                ],
              },
              after: {
                name: 'web-asg',
                desired_capacity: 5,
                min_size: 2,
                max_size: 10,
                launch_template: {
                  id: 'lt-abc123',
                  version: '5',
                },
                tags: [
                  { key: 'environment', value: 'production' },
                  { key: 'team', value: 'platform' },
                  { key: 'manually_scaled', value: 'true' },
                ],
              },
            },
          },
        ],
        resource_changes: [
          {
            address: 'aws_autoscaling_group.web',
            type: 'aws_autoscaling_group',
            name: 'web',
            change: {
              actions: ['update'],
              before: {
                name: 'web-asg',
                desired_capacity: 5,
                min_size: 2,
                max_size: 10,
                launch_template: {
                  id: 'lt-abc123',
                  version: '5',
                },
              },
              after: {
                name: 'web-asg',
                desired_capacity: 3,
                min_size: 2,
                max_size: 10,
                launch_template: {
                  id: 'lt-abc123',
                  version: '6',
                },
              },
            },
          },
          {
            address: 'aws_launch_template.web',
            type: 'aws_launch_template',
            name: 'web',
            change: {
              actions: ['update'],
              before: {
                image_id: 'ami-old123',
                instance_type: 't3.large',
                user_data: 'base64encodeddata...',
                block_device_mappings: [
                  {
                    device_name: '/dev/xvda',
                    ebs: {
                      volume_size: 50,
                      volume_type: 'gp3',
                      encrypted: true,
                    },
                  },
                ],
              },
              after: {
                image_id: 'ami-new456',
                instance_type: 't3.large',
                user_data: 'newbase64encodeddata...',
                block_device_mappings: [
                  {
                    device_name: '/dev/xvda',
                    ebs: {
                      volume_size: 100,
                      volume_type: 'gp3',
                      encrypted: true,
                    },
                  },
                ],
              },
            },
          },
        ],
        output_changes: {},
      } as any
    }
  />
)
export const KubectlManifestDeployment = () => (
  <TerraformDiff
    plan={
      {
        resource_changes: [
          {
            address: 'kubectl_manifest.api_deployment',
            type: 'kubectl_manifest',
            name: 'api_deployment',
            module_address: 'module.k8s_workloads',
            change: {
              actions: ['update'],
              before: {
                yaml_body: `apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: api-server\n  namespace: production\n  labels:\n    app: api-server\n    version: v2.14.0\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: api-server\n  template:\n    metadata:\n      labels:\n        app: api-server\n    spec:\n      serviceAccountName: api-server\n      containers:\n      - name: api-server\n        image: 123456789012.dkr.ecr.us-west-2.amazonaws.com/api-server:v2.14.0\n        ports:\n        - containerPort: 8080\n        resources:\n          requests:\n            cpu: 500m\n            memory: 512Mi\n          limits:\n            cpu: 1000m\n            memory: 1Gi`,
                yaml_body_parsed: `apiVersion=apps/v1,kind=Deployment,metadata.name=api-server,metadata.namespace=production,metadata.labels.app=api-server,metadata.labels.version=v2.14.0,spec.replicas=3,spec.template.spec.containers.0.image=123456789012.dkr.ecr.us-west-2.amazonaws.com/api-server:v2.14.0,spec.template.spec.containers.0.resources.requests.cpu=500m,spec.template.spec.containers.0.resources.requests.memory=512Mi`,
                live_manifest_incluster: 'sha256:a3f8c2e1d09b4f7e6c5a2b1d8e3f9c4a7b2e5d8f1c6a9b3e7f2d5c8a1b4e7f0c3d6a9b2e5c8f1a4b7e0d3c6f9a2b5e8c1d4f7a0b3e6c9d2f5a8b1e4c7d0f3a6b9e2c5d8f1a4b7e0c3d6f9a2b5e8c1',
                force_new: false,
                server_side_apply: true,
              },
              after: {
                yaml_body: `apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: api-server\n  namespace: production\n  labels:\n    app: api-server\n    version: v2.15.0\nspec:\n  replicas: 5\n  selector:\n    matchLabels:\n      app: api-server\n  template:\n    metadata:\n      labels:\n        app: api-server\n    spec:\n      serviceAccountName: api-server\n      containers:\n      - name: api-server\n        image: 123456789012.dkr.ecr.us-west-2.amazonaws.com/api-server:v2.15.0\n        ports:\n        - containerPort: 8080\n        resources:\n          requests:\n            cpu: 500m\n            memory: 1Gi\n          limits:\n            cpu: 2000m\n            memory: 2Gi`,
                yaml_body_parsed: `apiVersion=apps/v1,kind=Deployment,metadata.name=api-server,metadata.namespace=production,metadata.labels.app=api-server,metadata.labels.version=v2.15.0,spec.replicas=5,spec.template.spec.containers.0.image=123456789012.dkr.ecr.us-west-2.amazonaws.com/api-server:v2.15.0,spec.template.spec.containers.0.resources.requests.cpu=500m,spec.template.spec.containers.0.resources.requests.memory=1Gi`,
                live_manifest_incluster: 'sha256:b4a9d3f2e1c8b5a2e9d6c3f0a7b4e1d8c5f2a9b6e3c0d7f4a1b8e5c2d9f6a3b0e7c4d1f8a5b2e9c6d3f0a7b4e1d8c5f2a9b6e3c0d7f4a1b8e5c2d9f6a3b0e7',
                force_new: false,
                server_side_apply: true,
              },
            },
          },
          {
            address: 'kubectl_manifest.api_namespace',
            type: 'kubectl_manifest',
            name: 'api_namespace',
            module_address: 'module.k8s_workloads',
            change: {
              actions: ['update'],
              before: {
                yaml_body: `apiVersion: v1\nkind: Namespace\nmetadata:\n  name: production\n  labels:\n    team: platform\n    environment: production\n  annotations:\n    iam.amazonaws.com/permitted: "arn:aws:iam::123456789012:role/production-*"`,
                yaml_body_parsed: `apiVersion=v1,kind=Namespace,metadata.name=production,metadata.labels.team=platform,metadata.labels.environment=production,metadata.annotations.iam.amazonaws.com/permitted=arn:aws:iam::123456789012:role/production-*`,
                live_manifest_incluster: 'sha256:c1e4a7b0d3f6c9a2e5b8d1f4c7a0e3b6d9f2c5a8e1b4d7f0c3a6e9b2d5f8c1a4e7b0d3f6c9a2e5b8d1f4c7a0e3b6d9f2c5a8e1b4d7f0c3a6e9b2d5f8c1a4',
              },
              after: {
                yaml_body: `apiVersion: v1\nkind: Namespace\nmetadata:\n  name: production\n  labels:\n    team: platform\n    environment: production\n    cost-center: eng-platform\n  annotations:\n    iam.amazonaws.com/permitted: "arn:aws:iam::123456789012:role/production-*"\n    scheduler.alpha.kubernetes.io/defaultTolerations: '[{"operator":"Equal","effect":"NoSchedule","key":"dedicated","value":"platform"}]'`,
                yaml_body_parsed: `apiVersion=v1,kind=Namespace,metadata.name=production,metadata.labels.team=platform,metadata.labels.environment=production,metadata.labels.cost-center=eng-platform,metadata.annotations.iam.amazonaws.com/permitted=arn:aws:iam::123456789012:role/production-*,metadata.annotations.scheduler.alpha.kubernetes.io/defaultTolerations=[{"operator":"Equal","effect":"NoSchedule","key":"dedicated","value":"platform"}]`,
                live_manifest_incluster: 'sha256:d2f5b8e1a4c7d0f3b6e9c2d5f8a1b4e7c0d3f6a9b2e5c8d1f4a7b0e3c6d9f2a5b8e1c4d7f0a3b6e9c2d5f8a1b4e7c0d3f6a9b2e5c8d1f4a7b0e3c6d9f2a5',
              },
            },
          },
        ],
        output_changes: {},
      } as any
    }
  />
)

export const DriftWithChangesAndOutputs = () => (
  <TerraformDiff
    plan={
      {
        resource_drift: [
          {
            address: 'aws_autoscaling_group.web',
            type: 'aws_autoscaling_group',
            name: 'web',
            change: {
              actions: ['update'],
              before: {
                name: 'web-asg',
                desired_capacity: 3,
                min_size: 2,
                max_size: 10,
              },
              after: {
                name: 'web-asg',
                desired_capacity: 6,
                min_size: 2,
                max_size: 10,
              },
            },
          },
          {
            address: 'aws_security_group.web',
            type: 'aws_security_group',
            name: 'web',
            change: {
              actions: ['update'],
              before: {
                name: 'web-sg',
                ingress: [
                  { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
                ],
              },
              after: {
                name: 'web-sg',
                ingress: [
                  { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
                  { from_port: 22, to_port: 22, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
                ],
              },
            },
          },
        ],
        resource_changes: [
          {
            address: 'aws_autoscaling_group.web',
            type: 'aws_autoscaling_group',
            name: 'web',
            change: {
              actions: ['update'],
              before: {
                name: 'web-asg',
                desired_capacity: 6,
                min_size: 2,
                max_size: 10,
                launch_template: { id: 'lt-abc123', version: '3' },
              },
              after: {
                name: 'web-asg',
                desired_capacity: 3,
                min_size: 2,
                max_size: 10,
                launch_template: { id: 'lt-abc123', version: '4' },
              },
            },
          },
          {
            address: 'aws_security_group.web',
            type: 'aws_security_group',
            name: 'web',
            change: {
              actions: ['update'],
              before: {
                name: 'web-sg',
                ingress: [
                  { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
                  { from_port: 22, to_port: 22, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
                ],
              },
              after: {
                name: 'web-sg',
                ingress: [
                  { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['10.0.0.0/8'] },
                ],
              },
            },
          },
          {
            address: 'aws_launch_template.web',
            type: 'aws_launch_template',
            name: 'web',
            change: {
              actions: ['update'],
              before: { image_id: 'ami-old123', instance_type: 't3.large' },
              after: { image_id: 'ami-new456', instance_type: 't3.xlarge' },
            },
          },
          {
            address: 'aws_cloudwatch_metric_alarm.cpu_high',
            type: 'aws_cloudwatch_metric_alarm',
            name: 'cpu_high',
            change: {
              actions: ['create'],
              before: null,
              after: {
                alarm_name: 'web-cpu-high',
                comparison_operator: 'GreaterThanThreshold',
                threshold: 80,
                evaluation_periods: 2,
              },
            },
          },
        ],
        output_changes: {
          asg_desired_capacity: {
            actions: ['update'],
            before: 6,
            after: 3,
          },
          launch_template_version: {
            actions: ['update'],
            before: '3',
            after: '4',
          },
          cpu_alarm_arn: {
            actions: ['create'],
            before: null,
            after: 'Known after apply',
            after_unknown: true,
          },
        },
      } as any
    }
  />
)
