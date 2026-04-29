export default {
  title: 'Sandbox/SandboxOutputs',
}

import { SandboxOutputs } from './SandboxOutputs'

const fullAwsOutputs = {
  account: {
    id: '123456789012',
    region: 'us-east-1',
  },
  cluster: {
    arn: 'arn:aws:eks:us-west-2:123456789012:cluster/nuon-cluster',
    certificate_authority_data: 'LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREF2TVMwd0t3WURWUVFERXlRME4yVTEK...',
    cluster_security_group_id: 'sg-0abc123def456',
    endpoint: 'https://A1B2C3D4E5F6.gr7.us-west-2.eks.amazonaws.com',
    name: 'nuon-cluster',
    node_security_group_id: 'sg-0xyz789uvw456',
    oidc_issuer_url: 'https://oidc.eks.us-west-2.amazonaws.com/id/A1B2C3D4E5F6',
    oidc_provider: '123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/A1B2C3D4E5F6',
    oidc_provider_arn: 'arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/A1B2C3D4E5F6',
    platform_version: 'eks.9',
    status: 'ACTIVE',
  },
  ecr: {
    registry_id: '123456789012',
    registry_url: '123456789012.dkr.ecr.us-west-2.amazonaws.com',
    repository_arn: 'arn:aws:ecr:us-west-2:123456789012:repository/nuon-app',
    repository_name: 'nuon-app',
    repository_url: '123456789012.dkr.ecr.us-west-2.amazonaws.com/nuon-app',
  },
  karpenter: {
    discovery_key: 'karpenter-discovery-key',
    discovery_value: 'karpenter-discovery-value',
    instance_profile: {
      arn: 'arn:aws:iam::123456789012:instance-profile/karpenter-profile',
      id: 'karpenter-instance-profile-id',
      name: 'karpenter-instance-profile',
    },
  },
  namespaces: ['default', 'inlilb8l11rez6f02jwm5gz4oj'],
  nuon_dns: {
    alb_ingress_controller: {
      chart: 'aws-load-balancer-controller',
      enabled: true,
      id: 'alb-ingress-controller',
      revision: '1.4.7',
    },
    cert_manager: {
      chart: 'cert-manager',
      enabled: true,
      id: 'cert-manager',
      revision: '1.11.0',
    },
    enabled: true,
    external_dns: {
      chart: 'external-dns',
      enabled: true,
      id: 'external-dns',
      revision: '1.12.1',
    },
    ingress_nginx: {
      chart: 'ingress-nginx',
      enabled: true,
      id: 'ingress-nginx',
      revision: '4.7.1',
    },
    internal_domain: {
      name: 'internal.example.com',
      nameservers: [
        'ns-5678.awsdns-90.org',
        'ns-123.awsdns-12.com',
        'ns-456.awsdns-34.net',
        'ns-789.awsdns-56.co.uk',
      ],
      zone_id: 'Z8G9H0I1J2K3L4',
    },
    public_domain: {
      name: 'example.com',
      nameservers: [
        'ns-1234.awsdns-12.org',
        'ns-567.awsdns-34.com',
        'ns-890.awsdns-56.net',
        'ns-1234.awsdns-78.co.uk',
      ],
      zone_id: 'Z1A2B3C4D5E6F7',
    },
  },
  region: 'us-east-1',
  vpc: {
    arn: 'arn:aws:ec2:us-west-2:123456789012:vpc/vpc-0abc123def456',
    azs: ['us-west-2a', 'us-west-2b', 'us-west-2c'],
    cidr: '10.0.0.0/16',
    default_security_group_id: 'sg-0qrs345tuv678',
    id: 'vpc-0abc123def456',
    private_subnet_cidr_blocks: ['10.0.1.0/24', '10.0.2.0/24', '10.0.3.0/24'],
    private_subnet_ids: [
      'subnet-0abc123def456',
      'subnet-0ghi789jkl012',
      'subnet-0mno345pqr678',
    ],
    public_subnet_cidr_blocks: ['10.0.4.0/24', '10.0.5.0/24', '10.0.6.0/24'],
    public_subnet_ids: [
      'subnet-0stu901vwx234',
      'subnet-0yza567bcd890',
      'subnet-0efg123hij456',
    ],
    runner_subnet_cidr: '10.0.7.0/24',
    runner_subnet_id: 'subnet-0klm789nop012',
  },
}

export const FullAws = () => <SandboxOutputs outputs={fullAwsOutputs} />

export const ClusterOnly = () => (
  <SandboxOutputs
    outputs={{
      account: { id: '123456789012', region: 'us-west-2' },
      cluster: {
        arn: 'arn:aws:eks:us-west-2:123456789012:cluster/staging',
        endpoint: 'https://ABC123.gr7.us-west-2.eks.amazonaws.com',
        name: 'staging',
        status: 'ACTIVE',
        platform_version: 'eks.9',
      },
    }}
  />
)

export const NetworkHeavy = () => (
  <SandboxOutputs
    outputs={{
      vpc: {
        id: 'vpc-0abc123def456',
        cidr: '10.0.0.0/16',
        azs: ['us-west-2a', 'us-west-2b', 'us-west-2c'],
        private_subnet_ids: [
          'subnet-0abc123def456',
          'subnet-0ghi789jkl012',
          'subnet-0mno345pqr678',
        ],
        public_subnet_ids: [
          'subnet-0stu901vwx234',
          'subnet-0yza567bcd890',
          'subnet-0efg123hij456',
        ],
        private_subnet_cidr_blocks: ['10.0.1.0/24', '10.0.2.0/24', '10.0.3.0/24'],
        public_subnet_cidr_blocks: ['10.0.4.0/24', '10.0.5.0/24', '10.0.6.0/24'],
        default_security_group_id: 'sg-0qrs345tuv678',
      },
      nuon_dns: {
        enabled: true,
        internal_domain: {
          name: 'internal.example.com',
          zone_id: 'Z8G9H0I1J2K3L4',
          nameservers: ['ns-5678.awsdns-90.org', 'ns-123.awsdns-12.com'],
        },
        public_domain: {
          name: 'example.com',
          zone_id: 'Z1A2B3C4D5E6F7',
          nameservers: ['ns-1234.awsdns-12.org', 'ns-567.awsdns-34.com'],
        },
      },
      region: 'us-west-2',
    }}
  />
)

export const MinimalSandbox = () => (
  <SandboxOutputs
    outputs={{
      account: { id: '987654321098', region: 'eu-west-1' },
      region: 'eu-west-1',
      namespaces: ['default'],
    }}
  />
)

export const DnsAndIngress = () => (
  <SandboxOutputs
    outputs={{
      nuon_dns: {
        enabled: true,
        alb_ingress_controller: {
          chart: 'aws-load-balancer-controller',
          enabled: true,
          id: 'alb-ingress-controller',
          revision: '1.4.7',
        },
        cert_manager: {
          chart: 'cert-manager',
          enabled: true,
          id: 'cert-manager',
          revision: '1.11.0',
        },
        external_dns: {
          chart: 'external-dns',
          enabled: true,
          id: 'external-dns',
          revision: '1.12.1',
        },
        ingress_nginx: {
          chart: 'ingress-nginx',
          enabled: true,
          id: 'ingress-nginx',
          revision: '4.7.1',
        },
        internal_domain: {
          name: 'internal.myapp.io',
          zone_id: 'ZABC123',
          nameservers: ['ns-1.awsdns.org', 'ns-2.awsdns.com'],
        },
        public_domain: {
          name: 'myapp.io',
          zone_id: 'ZDEF456',
          nameservers: ['ns-3.awsdns.net', 'ns-4.awsdns.co.uk'],
        },
      },
    }}
  />
)
