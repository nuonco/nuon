export default {
  title: 'Workflows/GenerateStackDetails',
}

import { GenerateStackDetails, GenerateStackDetailsSkeleton } from './GenerateStackDetails'
import type { TAppConfig } from '@/types'

const mockConfig = {
  stack: {
    name: 'my-stack',
    description: 'A test stack',
    runner_nested_template_url: 'https://example.com/runner.yaml',
    vpc_nested_template_url: 'https://example.com/vpc.yaml',
    type: 'aws-cloudformation',
  },
} as TAppConfig

export const Default = () => (
  <GenerateStackDetails appConfig={mockConfig} isLoading={false} />
)

export const Loading = () => (
  <GenerateStackDetails isLoading={true} />
)

export const Skeleton = () => <GenerateStackDetailsSkeleton />
