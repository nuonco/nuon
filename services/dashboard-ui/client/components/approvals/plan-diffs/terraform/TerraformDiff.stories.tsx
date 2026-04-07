export default {
  title: 'Approvals/PlanDiffs/TerraformDiff',
}

import { TerraformDiff } from './TerraformDiff'

export const NoPlan = () => <TerraformDiff plan={undefined} />

export const WithPlan = () => (
    <TerraformDiff
      plan={{
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
      } as any}
    />
  )
