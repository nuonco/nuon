export default {
  title: 'Approvals/PlanDiffs/TerraformValueCodeBlock',
}

import { TerraformValueCodeBlock } from './TerraformValueCodeBlock'

export const JsonValue = () => (
    <TerraformValueCodeBlock
      value={JSON.stringify({ bucket: 'my-app-assets', acl: 'private', region: 'us-east-1' }, null, 2)}
    />
  )

export const PlainString = () => <TerraformValueCodeBlock value="t3.small" />

export const YamlLike = () => (
    <TerraformValueCodeBlock
      value={`name: my-resource\ntype: aws_instance\nregion: us-east-1\ntags:\n  env: production`}
    />
  )
