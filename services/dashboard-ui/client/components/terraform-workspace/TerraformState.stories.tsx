export default {
  title: 'Terraform/TerraformState',
}

import { TerraformState } from './TerraformState'
import type { TTerraformState } from '@/types'

const mockTerraformState: TTerraformState = {
  format_version: '1.0',
  terraform_version: '1.5.7',
  values: {
    outputs: {
      cluster_endpoint: { value: 'https://k8s.example.com', sensitive: false },
      vpc_id: { value: 'vpc-0abc123def456', sensitive: false },
      region: { value: 'us-west-2', sensitive: false },
    },
    root_module: {
      resources: [
        {
          address: 'aws_eks_cluster.main',
          name: 'main',
          type: 'aws_eks_cluster',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 0,
          mode: 'managed',
          values: {
            name: 'nuon-cluster',
            version: '1.28',
            status: 'ACTIVE',
          },
          sensitive_values: {},
        },
        {
          address: 'aws_vpc.main',
          name: 'main',
          type: 'aws_vpc',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 1,
          mode: 'managed',
          values: {
            cidr_block: '10.0.0.0/16',
            enable_dns_hostnames: true,
          },
          sensitive_values: {},
        },
      ],
      child_modules: [],
    },
  },
} as unknown as TTerraformState

export const Default = () => (
  <div className="max-w-4xl p-4">
    <TerraformState terraformState={mockTerraformState} />
  </div>
)

export const Empty = () => (
  <div className="max-w-4xl p-4">
    <TerraformState terraformState={{ format_version: '1.0', terraform_version: '1.5.7', values: { outputs: {}, root_module: { resources: [], child_modules: [] } } } as unknown as TTerraformState} />
  </div>
)

const networkingState: TTerraformState = {
  format_version: '1.0',
  terraform_version: '1.9.5',
  values: {
    outputs: {
      vpc_id: { value: 'vpc-0a3f7c9e2b1d4e8f6', sensitive: false, type: 'string' },
      public_subnet_ids: {
        value: ['subnet-0a1b2c3d4e5f60001', 'subnet-0a1b2c3d4e5f60002', 'subnet-0a1b2c3d4e5f60003'],
        sensitive: false,
        type: ['list', 'string'],
      },
      private_subnet_ids: {
        value: ['subnet-0f1e2d3c4b5a60001', 'subnet-0f1e2d3c4b5a60002', 'subnet-0f1e2d3c4b5a60003'],
        sensitive: false,
        type: ['list', 'string'],
      },
      nat_gateway_ip: { value: '54.203.112.47', sensitive: false, type: 'string' },
      availability_zones: {
        value: ['us-west-2a', 'us-west-2b', 'us-west-2c'],
        sensitive: false,
        type: ['list', 'string'],
      },
    },
    root_module: {
      resources: [
        {
          address: 'aws_vpc.main',
          name: 'main',
          type: 'aws_vpc',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 1,
          mode: 'managed',
          values: {
            cidr_block: '10.0.0.0/16',
            enable_dns_hostnames: true,
            enable_dns_support: true,
            tags: { Name: 'production-vpc', Environment: 'production', ManagedBy: 'terraform' },
          },
          sensitive_values: {},
        },
        {
          address: 'aws_internet_gateway.main',
          name: 'main',
          type: 'aws_internet_gateway',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 0,
          mode: 'managed',
          values: { vpc_id: 'vpc-0a3f7c9e2b1d4e8f6', tags: { Name: 'production-igw' } },
          sensitive_values: {},
          depends_on: ['aws_vpc.main'],
        },
        {
          address: 'aws_nat_gateway.main',
          name: 'main',
          type: 'aws_nat_gateway',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 0,
          mode: 'managed',
          values: {
            allocation_id: 'eipalloc-0a1b2c3d4e5f6a7b8',
            subnet_id: 'subnet-0a1b2c3d4e5f60001',
            connectivity_type: 'public',
          },
          sensitive_values: {},
          depends_on: ['aws_internet_gateway.main', 'aws_eip.nat'],
        },
        {
          address: 'aws_eip.nat',
          name: 'nat',
          type: 'aws_eip',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 0,
          mode: 'managed',
          values: { domain: 'vpc', public_ip: '54.203.112.47' },
          sensitive_values: {},
        },
      ],
      child_modules: [
        {
          address: 'module.public_subnets',
          resources: [
            {
              address: 'module.public_subnets.aws_subnet.this[0]',
              name: 'this',
              type: 'aws_subnet',
              provider_name: 'registry.terraform.io/hashicorp/aws',
              schema_version: 1,
              index: 0,
              mode: 'managed',
              values: {
                cidr_block: '10.0.1.0/24',
                availability_zone: 'us-west-2a',
                map_public_ip_on_launch: true,
              },
              sensitive_values: {},
              depends_on: ['aws_vpc.main'],
            },
            {
              address: 'module.public_subnets.aws_subnet.this[1]',
              name: 'this',
              type: 'aws_subnet',
              provider_name: 'registry.terraform.io/hashicorp/aws',
              schema_version: 1,
              index: 1,
              mode: 'managed',
              values: {
                cidr_block: '10.0.2.0/24',
                availability_zone: 'us-west-2b',
                map_public_ip_on_launch: true,
              },
              sensitive_values: {},
              depends_on: ['aws_vpc.main'],
            },
            {
              address: 'module.public_subnets.aws_subnet.this[2]',
              name: 'this',
              type: 'aws_subnet',
              provider_name: 'registry.terraform.io/hashicorp/aws',
              schema_version: 1,
              index: 2,
              mode: 'managed',
              values: {
                cidr_block: '10.0.3.0/24',
                availability_zone: 'us-west-2c',
                map_public_ip_on_launch: true,
              },
              sensitive_values: {},
              depends_on: ['aws_vpc.main'],
            },
          ],
        },
      ],
    },
  },
} as unknown as TTerraformState

export const AWSNetworkingStack = () => (
  <div className="max-w-4xl p-4">
    <TerraformState terraformState={networkingState} />
  </div>
)

const eksClusterState: TTerraformState = {
  format_version: '1.0',
  terraform_version: '1.9.5',
  values: {
    outputs: {
      cluster_endpoint: { value: 'https://ABCDEF1234567890.gr7.us-west-2.eks.amazonaws.com', sensitive: false, type: 'string' },
      cluster_certificate_authority: { value: 'LS0tLS1CRUdJTi...', sensitive: true, type: 'string' },
      cluster_name: { value: 'prod-eks-cluster', sensitive: false, type: 'string' },
      cluster_version: { value: '1.30', sensitive: false, type: 'string' },
      oidc_provider_arn: { value: 'arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/ABCDEF1234567890', sensitive: false, type: 'string' },
      node_group_status: {
        value: { general: 'ACTIVE', gpu: 'ACTIVE' },
        sensitive: false,
        type: ['object', { general: 'string', gpu: 'string' }],
      },
    },
    root_module: {
      resources: [
        {
          address: 'aws_eks_cluster.main',
          name: 'main',
          type: 'aws_eks_cluster',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 1,
          mode: 'managed',
          values: {
            name: 'prod-eks-cluster',
            role_arn: 'arn:aws:iam::123456789012:role/eks-cluster-role',
            version: '1.30',
            enabled_cluster_log_types: ['api', 'audit', 'authenticator', 'controllerManager', 'scheduler'],
            endpoint_private_access: true,
            endpoint_public_access: false,
            status: 'ACTIVE',
          },
          sensitive_values: {},
          depends_on: ['aws_iam_role.cluster', 'aws_vpc.main'],
        },
        {
          address: 'aws_eks_node_group.general',
          name: 'general',
          type: 'aws_eks_node_group',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 0,
          mode: 'managed',
          values: {
            cluster_name: 'prod-eks-cluster',
            node_group_name: 'general-20240315',
            instance_types: ['m6i.xlarge'],
            scaling_config: { desired_size: 3, max_size: 10, min_size: 2 },
            ami_type: 'AL2_x86_64',
            capacity_type: 'ON_DEMAND',
            disk_size: 100,
            status: 'ACTIVE',
          },
          sensitive_values: {},
          depends_on: ['aws_eks_cluster.main', 'aws_iam_role.node_group'],
        },
        {
          address: 'aws_eks_node_group.gpu',
          name: 'gpu',
          type: 'aws_eks_node_group',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 0,
          mode: 'managed',
          values: {
            cluster_name: 'prod-eks-cluster',
            node_group_name: 'gpu-20240315',
            instance_types: ['g5.2xlarge'],
            scaling_config: { desired_size: 1, max_size: 4, min_size: 0 },
            ami_type: 'AL2_x86_64_GPU',
            capacity_type: 'ON_DEMAND',
            disk_size: 200,
            status: 'ACTIVE',
            labels: { 'nvidia.com/gpu.present': 'true', workload: 'ml-inference' },
            taints: [{ key: 'nvidia.com/gpu', value: 'true', effect: 'NO_SCHEDULE' }],
          },
          sensitive_values: {},
          depends_on: ['aws_eks_cluster.main', 'aws_iam_role.node_group'],
        },
        {
          address: 'data.aws_eks_cluster_auth.main',
          name: 'main',
          type: 'aws_eks_cluster_auth',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 0,
          mode: 'data',
          values: { name: 'prod-eks-cluster' },
          sensitive_values: { token: true },
        },
      ],
      child_modules: [],
    },
  },
} as unknown as TTerraformState

export const EKSCluster = () => (
  <div className="max-w-4xl p-4">
    <TerraformState terraformState={eksClusterState} />
  </div>
)

const k8sManifestState: TTerraformState = {
  format_version: '1.0',
  terraform_version: '1.9.5',
  values: {
    outputs: {
      service_endpoint: { value: 'http://api-gateway.default.svc.cluster.local:8080', sensitive: false, type: 'string' },
      namespace: { value: 'default', sensitive: false, type: 'string' },
    },
    root_module: {
      resources: [
        {
          address: 'kubernetes_deployment.api',
          name: 'api',
          type: 'kubernetes_deployment',
          provider_name: 'registry.terraform.io/hashicorp/kubernetes',
          schema_version: 0,
          mode: 'managed',
          values: {
            metadata: { name: 'api-gateway', namespace: 'default', labels: { app: 'api-gateway', version: 'v2.4.1' } },
            replicas: '3',
            strategy: { type: 'RollingUpdate', rolling_update: { max_surge: '1', max_unavailable: '0' } },
          },
          sensitive_values: {},
        },
        {
          address: 'kubernetes_service.api',
          name: 'api',
          type: 'kubernetes_service',
          provider_name: 'registry.terraform.io/hashicorp/kubernetes',
          schema_version: 0,
          mode: 'managed',
          values: {
            metadata: { name: 'api-gateway', namespace: 'default' },
            spec: { type: 'ClusterIP', port: { port: 8080, target_port: 8080, protocol: 'TCP' } },
          },
          sensitive_values: {},
          depends_on: ['kubernetes_deployment.api'],
        },
        {
          address: 'kubernetes_config_map.api',
          name: 'api',
          type: 'kubernetes_config_map',
          provider_name: 'registry.terraform.io/hashicorp/kubernetes',
          schema_version: 0,
          mode: 'managed',
          values: {
            metadata: { name: 'api-gateway-config', namespace: 'default' },
            data: { LOG_LEVEL: 'info', MAX_CONNECTIONS: '100', RATE_LIMIT_RPS: '500' },
          },
          sensitive_values: {},
        },
        {
          address: 'kubernetes_secret.api_credentials',
          name: 'api_credentials',
          type: 'kubernetes_secret',
          provider_name: 'registry.terraform.io/hashicorp/kubernetes',
          schema_version: 0,
          mode: 'managed',
          values: {
            metadata: { name: 'api-gateway-credentials', namespace: 'default' },
            type: 'Opaque',
          },
          sensitive_values: { data: { DATABASE_URL: true, API_KEY: true, JWT_SECRET: true } },
        },
        {
          address: 'kubernetes_horizontal_pod_autoscaler_v2.api',
          name: 'api',
          type: 'kubernetes_horizontal_pod_autoscaler_v2',
          provider_name: 'registry.terraform.io/hashicorp/kubernetes',
          schema_version: 0,
          mode: 'managed',
          values: {
            metadata: { name: 'api-gateway', namespace: 'default' },
            min_replicas: 3,
            max_replicas: 20,
            target_cpu_utilization_percentage: 70,
          },
          sensitive_values: {},
          depends_on: ['kubernetes_deployment.api'],
        },
      ],
      child_modules: [],
    },
  },
} as unknown as TTerraformState

export const KubernetesWorkload = () => (
  <div className="max-w-4xl p-4">
    <TerraformState terraformState={k8sManifestState} />
  </div>
)

const rdsState: TTerraformState = {
  format_version: '1.0',
  terraform_version: '1.9.5',
  values: {
    outputs: {
      db_endpoint: { value: 'prod-db.c9abcdef1234.us-west-2.rds.amazonaws.com:5432', sensitive: false, type: 'string' },
      db_name: { value: 'app_production', sensitive: false, type: 'string' },
      db_connection_string: { value: null, sensitive: true, type: 'string' },
      read_replica_endpoints: {
        value: [
          'prod-db-r1.c9abcdef1234.us-west-2.rds.amazonaws.com:5432',
          'prod-db-r2.c9abcdef1234.us-west-2.rds.amazonaws.com:5432',
        ],
        sensitive: false,
        type: ['list', 'string'],
      },
    },
    root_module: {
      resources: [
        {
          address: 'aws_db_instance.primary',
          name: 'primary',
          type: 'aws_db_instance',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 2,
          mode: 'managed',
          values: {
            identifier: 'prod-db',
            engine: 'postgres',
            engine_version: '16.3',
            instance_class: 'db.r6g.xlarge',
            allocated_storage: 500,
            max_allocated_storage: 2000,
            storage_type: 'gp3',
            storage_encrypted: true,
            multi_az: true,
            publicly_accessible: false,
            backup_retention_period: 30,
            performance_insights_enabled: true,
            deletion_protection: true,
            db_name: 'app_production',
            status: 'available',
          },
          sensitive_values: { password: true, master_user_secret: true },
          depends_on: ['aws_db_subnet_group.main', 'aws_security_group.rds'],
        },
        {
          address: 'aws_db_instance.replica[0]',
          name: 'replica',
          type: 'aws_db_instance',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 2,
          index: 0,
          mode: 'managed',
          values: {
            identifier: 'prod-db-r1',
            engine: 'postgres',
            engine_version: '16.3',
            instance_class: 'db.r6g.large',
            replicate_source_db: 'prod-db',
            storage_encrypted: true,
            publicly_accessible: false,
            performance_insights_enabled: true,
            status: 'available',
          },
          sensitive_values: { password: true },
          depends_on: ['aws_db_instance.primary'],
        },
        {
          address: 'aws_db_instance.replica[1]',
          name: 'replica',
          type: 'aws_db_instance',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 2,
          index: 1,
          mode: 'managed',
          values: {
            identifier: 'prod-db-r2',
            engine: 'postgres',
            engine_version: '16.3',
            instance_class: 'db.r6g.large',
            replicate_source_db: 'prod-db',
            storage_encrypted: true,
            publicly_accessible: false,
            performance_insights_enabled: true,
            status: 'available',
          },
          sensitive_values: { password: true },
          depends_on: ['aws_db_instance.primary'],
        },
        {
          address: 'aws_db_subnet_group.main',
          name: 'main',
          type: 'aws_db_subnet_group',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 0,
          mode: 'managed',
          values: {
            name: 'prod-db-subnet-group',
            subnet_ids: ['subnet-0f1e2d3c4b5a60001', 'subnet-0f1e2d3c4b5a60002', 'subnet-0f1e2d3c4b5a60003'],
          },
          sensitive_values: {},
        },
        {
          address: 'aws_security_group.rds',
          name: 'rds',
          type: 'aws_security_group',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 1,
          mode: 'managed',
          values: {
            name: 'prod-rds-sg',
            vpc_id: 'vpc-0a3f7c9e2b1d4e8f6',
            ingress: [{ from_port: 5432, to_port: 5432, protocol: 'tcp', cidr_blocks: ['10.0.0.0/16'] }],
            egress: [{ from_port: 0, to_port: 0, protocol: '-1', cidr_blocks: ['0.0.0.0/0'] }],
          },
          sensitive_values: {},
        },
      ],
      child_modules: [],
    },
  },
} as unknown as TTerraformState

export const RDSPostgresWithReplicas = () => (
  <div className="max-w-4xl p-4">
    <TerraformState terraformState={rdsState} />
  </div>
)

const multiProviderState: TTerraformState = {
  format_version: '1.0',
  terraform_version: '1.9.5',
  values: {
    outputs: {
      datadog_dashboard_url: { value: 'https://app.datadoghq.com/dashboard/abc-def-ghi', sensitive: false, type: 'string' },
      pagerduty_service_id: { value: 'P1234AB', sensitive: false, type: 'string' },
      dns_records: {
        value: {
          api: 'api.prod.example.com',
          app: 'app.prod.example.com',
          status: 'status.prod.example.com',
        },
        sensitive: false,
        type: ['object', { api: 'string', app: 'string', status: 'string' }],
      },
    },
    root_module: {
      resources: [
        {
          address: 'datadog_monitor.cpu_high',
          name: 'cpu_high',
          type: 'datadog_monitor',
          provider_name: 'registry.terraform.io/DataDog/datadog',
          schema_version: 0,
          mode: 'managed',
          values: {
            name: '[Prod] CPU usage > 85% on EKS nodes',
            type: 'metric alert',
            query: 'avg(last_5m):avg:system.cpu.user{env:production,kube_cluster_name:prod-eks-cluster} by {host} > 85',
            message: 'CPU usage is high on {{host.name}}. @pagerduty-prod-infra',
            priority: 2,
            notify_no_data: false,
            renotify_interval: 300,
          },
          sensitive_values: {},
        },
        {
          address: 'datadog_monitor.pod_restarts',
          name: 'pod_restarts',
          type: 'datadog_monitor',
          provider_name: 'registry.terraform.io/DataDog/datadog',
          schema_version: 0,
          mode: 'managed',
          values: {
            name: '[Prod] Pod restart count > 5 in 10m',
            type: 'query alert',
            query: 'change(sum(last_10m),last_10m):avg:kubernetes.containers.restarts{env:production} by {kube_deployment} > 5',
            message: 'Pods restarting frequently in {{kube_deployment.name}}. @slack-prod-alerts',
            priority: 3,
          },
          sensitive_values: {},
        },
        {
          address: 'cloudflare_record.api',
          name: 'api',
          type: 'cloudflare_record',
          provider_name: 'registry.terraform.io/cloudflare/cloudflare',
          schema_version: 0,
          mode: 'managed',
          values: {
            zone_id: 'abc123def456',
            name: 'api.prod',
            type: 'CNAME',
            content: 'prod-alb-1234567890.us-west-2.elb.amazonaws.com',
            proxied: true,
            ttl: 1,
          },
          sensitive_values: {},
        },
        {
          address: 'cloudflare_record.app',
          name: 'app',
          type: 'cloudflare_record',
          provider_name: 'registry.terraform.io/cloudflare/cloudflare',
          schema_version: 0,
          mode: 'managed',
          values: {
            zone_id: 'abc123def456',
            name: 'app.prod',
            type: 'CNAME',
            content: 'prod-alb-1234567890.us-west-2.elb.amazonaws.com',
            proxied: true,
            ttl: 1,
          },
          sensitive_values: {},
        },
        {
          address: 'pagerduty_service.prod_infra',
          name: 'prod_infra',
          type: 'pagerduty_service',
          provider_name: 'registry.terraform.io/PagerDuty/pagerduty',
          schema_version: 0,
          mode: 'managed',
          values: {
            name: 'Production Infrastructure',
            escalation_policy: 'P5678CD',
            alert_creation: 'create_alerts_and_incidents',
            auto_resolve_timeout: 14400,
            acknowledgement_timeout: 1800,
          },
          sensitive_values: {},
        },
      ],
      child_modules: [],
    },
  },
} as unknown as TTerraformState

export const MultiProviderObservability = () => (
  <div className="max-w-4xl p-4">
    <TerraformState terraformState={multiProviderState} />
  </div>
)
