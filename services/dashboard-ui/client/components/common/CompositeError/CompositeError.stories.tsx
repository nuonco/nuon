import { CompositeError } from './CompositeError'
import type { TCompositeError } from '@/types'

export default {
  title: 'Common/CompositeError',
}

const awsPermissionError: TCompositeError = {
  type: 'aws_permission_error',
  severity: 'error',
  message: 'Deploy failed because the install role is missing required AWS IAM permissions.',
  sections: [
    {
      heading: 'Missing permissions',
      body: '- `ec2:CreateSecurityGroup`\n- `ec2:AuthorizeSecurityGroupIngress`',
    },
    {
      heading: 'How to fix',
      body: 'Add the missing actions to the install role policy, then retry the deploy.',
    },
  ],
}

export const Default = () => <CompositeError error={awsPermissionError} />

export const NoSections = () => (
  <CompositeError
    error={{
      type: 'aws_permission_error',
      severity: 'error',
      message: 'The install role is not authorized to perform this Terraform operation.',
    }}
  />
)

export const Warning = () => (
  <CompositeError
    error={{
      type: 'aws_permission_error',
      severity: 'warning',
      message: 'Some optional permissions are missing and may limit functionality.',
      sections: [{ heading: 'Details', body: 'Missing `s3:GetBucketTagging`.' }],
    }}
  />
)
