export default {
  title: 'Terraform/TerraformWorkspaceCard',
}

import { Button } from '@/components/common/Button'
import { TerraformWorkspaceCard } from './TerraformWorkspaceCard'

export const Default = () => (
  <TerraformWorkspaceCard
    currentRevision={{
      values: {
        outputs: {
          vpc_id: { value: 'vpc-abc123' },
          region: { value: 'us-west-2' },
        },
        root_module: {
          resources: [
            {
              address: 'aws_vpc.main',
              name: 'main',
              type: 'aws_vpc',
              mode: 'managed',
              provider_name: 'registry.terraform.io/hashicorp/aws',
              schema_version: 1,
              values: { cidr_block: '10.0.0.0/16' },
              sensitive_values: {},
            },
          ],
        },
      },
    } as any}
    actions={
      <>
        <Button variant="secondary" size="sm">Use Terraform CLI</Button>
        <Button size="sm">Unlock</Button>
      </>
    }
  />
)

export const Empty = () => (
  <TerraformWorkspaceCard
    currentRevision={null}
    actions={<Button variant="secondary" size="sm">Use Terraform CLI</Button>}
  />
)
