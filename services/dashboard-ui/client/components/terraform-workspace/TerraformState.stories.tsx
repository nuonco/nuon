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
