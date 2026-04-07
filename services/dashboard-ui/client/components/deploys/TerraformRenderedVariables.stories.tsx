import type { Meta, StoryObj } from '@ladle/react'
import { TerraformRenderedVariables } from './TerraformRenderedVariables'

export default {
  title: 'Deploys/TerraformRenderedVariables',
} satisfies Meta

export const Default: StoryObj = {
  render: () => (
    <TerraformRenderedVariables
      values={{
        instance_type: 't3.small',
        region: 'us-east-1',
        min_nodes: '2',
        max_nodes: '10',
        cluster_name: 'prod-cluster',
      }}
    />
  ),
}

export const Empty: StoryObj = {
  render: () => <TerraformRenderedVariables values={{}} />,
}
