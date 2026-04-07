export default {
  title: 'Approvals/PlanDiffs/TerraformValuesDiff',
}

import { TerraformValuesDiff } from './TerraformValuesDiff'

export const Update = () => (
    <TerraformValuesDiff
      values={{
        action: 'update',
        before: {
          instance_type: 't3.micro',
          min_capacity: 1,
          max_capacity: 3,
        },
        after: {
          instance_type: 't3.small',
          min_capacity: 2,
          max_capacity: 5,
        },
      }}
    />
  )

export const Create = () => (
    <TerraformValuesDiff
      values={{
        action: 'create',
        before: null,
        after: {
          bucket: 'my-app-assets',
          acl: 'private',
          region: 'us-east-1',
        },
      }}
    />
  )

export const Delete = () => (
    <TerraformValuesDiff
      values={{
        action: 'delete',
        before: {
          name: 'legacy-resource',
          value: 'some-old-value',
        },
        after: null,
      }}
    />
  )
