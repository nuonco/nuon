export default {
  title: 'Workflows/StepDetails/AwaitAWSDetails',
}

import { AwaitAWSDetails, AwaitAWSDetailsSkeleton } from './AwaitAWSDetails'

const mockStack = {
  versions: [
    {
      template_url: 'https://s3.amazonaws.com/bucket/template.json',
      quick_link_url: 'https://console.aws.amazon.com/cloudformation/home?stackName=nuon-stack&region=us-east-1',
      region: 'us-east-1',
    },
  ],
} as any

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <AwaitAWSDetails
      stack={mockStack}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitAWSDetailsSkeleton />
  </div>
)
