export default {
  title: 'Actions/InstallActionRunOutputs',
}

import { InstallActionRunOutputs } from './InstallActionRunOutputs'

const mockRunWithOutputs = {
  steps: [
    { id: 'step-run-1', step_id: 'step-1', status: 'finished', execution_duration: 45000000000 },
    { id: 'step-run-2', step_id: 'step-2', status: 'finished', execution_duration: 30000000000 },
  ],
  config: {
    steps: [
      { id: 'step-1', name: 'terraform-apply', idx: 0 },
      { id: 'step-2', name: 'verify', idx: 1 },
    ],
  },
  outputs: {
    steps: {
      'terraform-apply': {
        cluster_endpoint: 'https://eks.us-west-2.amazonaws.com/cluster-1',
        cluster_name: 'prod-cluster',
        vpc_id: 'vpc-abc123',
        status: 'active',
      },
      verify: null,
    },
  },
} as any

export const Default = () => (
  <InstallActionRunOutputs installActionRun={mockRunWithOutputs} />
)

const mockRunNoOutputs = {
  steps: [
    { id: 'step-run-1', step_id: 'step-1', status: 'finished', execution_duration: 10000000000 },
  ],
  config: {
    steps: [{ id: 'step-1', name: 'run-script', idx: 0 }],
  },
  outputs: { steps: {} },
} as any

export const NoOutputs = () => (
  <InstallActionRunOutputs installActionRun={mockRunNoOutputs} />
)

const mockRunManySteps = {
  steps: [
    { id: 'sr-1', step_id: 's-1', status: 'finished', execution_duration: 12000000000 },
    { id: 'sr-2', step_id: 's-2', status: 'finished', execution_duration: 87000000000 },
    { id: 'sr-3', step_id: 's-3', status: 'finished', execution_duration: 34000000000 },
    { id: 'sr-4', step_id: 's-4', status: 'error', execution_duration: 5000000000 },
    { id: 'sr-5', step_id: 's-5', status: 'finished', execution_duration: 120000000000 },
    { id: 'sr-6', step_id: 's-6', status: 'finished', execution_duration: 22000000000 },
  ],
  config: {
    steps: [
      { id: 's-1', name: 'provision-vpc', idx: 0 },
      { id: 's-2', name: 'deploy-eks', idx: 1 },
      { id: 's-3', name: 'configure-dns', idx: 2 },
      { id: 's-4', name: 'run-migrations', idx: 3 },
      { id: 's-5', name: 'deploy-services', idx: 4 },
      { id: 's-6', name: 'health-check', idx: 5 },
    ],
  },
  outputs: {
    steps: {
      'provision-vpc': {
        vpc_id: 'vpc-0a1b2c3d4e5f6a7b8',
        vpc_cidr: '10.0.0.0/16',
        private_subnet_ids: ['subnet-aaa111', 'subnet-bbb222', 'subnet-ccc333'],
        public_subnet_ids: ['subnet-ddd444', 'subnet-eee555', 'subnet-fff666'],
        nat_gateway_id: 'nat-0f1e2d3c4b5a6f7e8',
        internet_gateway_id: 'igw-9a8b7c6d5e4f3a2b1',
        route_table_ids: {
          private: 'rtb-priv001',
          public: 'rtb-pub002',
        },
        availability_zones: ['us-west-2a', 'us-west-2b', 'us-west-2c'],
        tags: {
          Environment: 'production',
          ManagedBy: 'nuon',
          'nuon.install_id': 'instxyz789',
        },
      },
      'deploy-eks': {
        cluster_name: 'prod-us-west-2',
        cluster_arn: 'arn:aws:eks:us-west-2:123456789012:cluster/prod-us-west-2',
        cluster_endpoint: 'https://ABCDEF1234567890.gr7.us-west-2.eks.amazonaws.com',
        cluster_version: '1.29',
        cluster_status: 'ACTIVE',
        oidc_provider_arn: 'arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/ABCDEF',
        node_groups: {
          'workers-general': {
            instance_types: ['m6i.xlarge', 'm6i.2xlarge'],
            min_size: 2,
            max_size: 10,
            desired_size: 3,
            ami_type: 'AL2_x86_64',
            status: 'ACTIVE',
            node_group_arn: 'arn:aws:eks:us-west-2:123456789012:nodegroup/prod/workers-general/abc123',
          },
          'workers-gpu': {
            instance_types: ['g5.xlarge'],
            min_size: 0,
            max_size: 4,
            desired_size: 1,
            ami_type: 'AL2_x86_64_GPU',
            status: 'ACTIVE',
            node_group_arn: 'arn:aws:eks:us-west-2:123456789012:nodegroup/prod/workers-gpu/def456',
          },
        },
        cluster_security_group_id: 'sg-0a1b2c3d4e5f6a7b8',
        service_role_arn: 'arn:aws:iam::123456789012:role/eks-service-role-prod',
        platform_version: 'eks.8',
        certificate_authority: 'LS0tLS1CRUdJTi...',
        addons: {
          'vpc-cni': { version: 'v1.16.2-eksbuild.1', status: 'ACTIVE' },
          coredns: { version: 'v1.11.1-eksbuild.6', status: 'ACTIVE' },
          'kube-proxy': { version: 'v1.29.1-eksbuild.2', status: 'ACTIVE' },
          'ebs-csi-driver': { version: 'v1.28.0-eksbuild.1', status: 'ACTIVE' },
        },
      },
      'configure-dns': {
        hosted_zone_id: 'Z0123456789ABCDEFGHIJ',
        domain_name: 'prod.acme-corp.io',
        nameservers: [
          'ns-512.awsdns-00.net',
          'ns-1024.awsdns-00.org',
          'ns-1536.awsdns-00.co.uk',
          'ns-0.awsdns-00.com',
        ],
        records_created: [
          { name: 'api.prod.acme-corp.io', type: 'A', alias: true },
          { name: 'app.prod.acme-corp.io', type: 'A', alias: true },
          { name: 'grpc.prod.acme-corp.io', type: 'A', alias: true },
          { name: '*.prod.acme-corp.io', type: 'CNAME' },
        ],
        acm_certificate_arn: 'arn:aws:acm:us-west-2:123456789012:certificate/abc-def-ghi',
        certificate_status: 'ISSUED',
        validation_method: 'DNS',
      },
      'run-migrations': {
        error: 'Migration 20240115_add_indexes failed: lock timeout exceeded after 30s on table "events"',
        failed_migration: '20240115_add_indexes',
        migrations_applied: 12,
        migrations_pending: 3,
        database_host: 'prod-db.cluster-abc123.us-west-2.rds.amazonaws.com',
        database_name: 'acme_production',
        rollback_applied: true,
      },
      'deploy-services': {
        deployments: {
          'api-server': {
            image: '123456789012.dkr.ecr.us-west-2.amazonaws.com/api-server:v2.14.3',
            replicas: { desired: 3, ready: 3, available: 3 },
            namespace: 'production',
            service_url: 'http://api-server.production.svc.cluster.local:8080',
            external_url: 'https://api.prod.acme-corp.io',
            status: 'Running',
          },
          'web-frontend': {
            image: '123456789012.dkr.ecr.us-west-2.amazonaws.com/web-frontend:v3.8.1',
            replicas: { desired: 2, ready: 2, available: 2 },
            namespace: 'production',
            service_url: 'http://web-frontend.production.svc.cluster.local:3000',
            external_url: 'https://app.prod.acme-corp.io',
            status: 'Running',
          },
          'worker-processor': {
            image: '123456789012.dkr.ecr.us-west-2.amazonaws.com/worker:v2.14.3',
            replicas: { desired: 5, ready: 5, available: 5 },
            namespace: 'production',
            status: 'Running',
          },
          'grpc-gateway': {
            image: '123456789012.dkr.ecr.us-west-2.amazonaws.com/grpc-gateway:v1.5.0',
            replicas: { desired: 2, ready: 2, available: 2 },
            namespace: 'production',
            service_url: 'http://grpc-gateway.production.svc.cluster.local:9090',
            external_url: 'https://grpc.prod.acme-corp.io',
            status: 'Running',
          },
        },
        ingress: {
          alb_arn: 'arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/prod-ingress/abc123',
          alb_dns_name: 'prod-ingress-123456789.us-west-2.elb.amazonaws.com',
          target_groups: 4,
          listener_rules: 6,
        },
        config_maps_updated: ['api-config', 'feature-flags', 'service-urls'],
        secrets_synced: ['db-credentials', 'api-keys', 'tls-certs'],
      },
      'health-check': {
        checks: {
          'api-server': { endpoint: 'https://api.prod.acme-corp.io/healthz', status: 200, latency_ms: 23 },
          'web-frontend': { endpoint: 'https://app.prod.acme-corp.io/health', status: 200, latency_ms: 45 },
          'grpc-gateway': { endpoint: 'grpc://grpc.prod.acme-corp.io:443/grpc.health.v1.Health/Check', status: 200, latency_ms: 12 },
          database: { endpoint: 'postgres://prod-db.cluster-abc123.us-west-2.rds.amazonaws.com:5432', status: 200, latency_ms: 8 },
          redis: { endpoint: 'redis://prod-cache.abc123.use2.cache.amazonaws.com:6379', status: 200, latency_ms: 3 },
        },
        all_healthy: true,
        total_checks: 5,
        passed: 5,
        failed: 0,
      },
    },
  },
} as any

export const ManySteps = () => (
  <InstallActionRunOutputs installActionRun={mockRunManySteps} />
)
