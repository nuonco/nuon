export default {
  title: 'Deploys/TerraformRenderedVariables',
}

import { TerraformRenderedVariables } from './TerraformRenderedVariables'

export const Default = () => (
  <TerraformRenderedVariables
    values={{
      instance_type: 't3.small',
      region: 'us-east-1',
      min_nodes: '2',
      max_nodes: '10',
      cluster_name: 'prod-cluster',
    }}
  />
)

export const Empty = () => <TerraformRenderedVariables values={{}} />
