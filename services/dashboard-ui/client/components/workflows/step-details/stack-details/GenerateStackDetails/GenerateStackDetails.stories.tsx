export default {
  title: 'Workflows/GenerateStackDetails',
}

import { GenerateStackDetails } from './GenerateStackDetails'
import type { TAppConfig, TInstallStackVersionWithCompositeError } from '@/types'

const mockConfig = {
  stack: {
    name: 'my-stack',
    description: 'A test stack',
    runner_nested_template_url: 'https://example.com/runner.yaml',
    vpc_nested_template_url: 'https://example.com/vpc.yaml',
    type: 'aws-cloudformation',
  },
} as TAppConfig

const mockVersionWithError: TInstallStackVersionWithCompositeError = {
  id: 'isv-abc123',
  composite_error: {
    type: 'stack.generation_failed',
    severity: 'error',
    message: 'Stack template generation failed',
    sections: [
      {
        heading: 'Why',
        body: 'The app config references a runner nested template URL that could not be fetched.',
      },
      {
        heading: 'How to fix',
        body: 'Verify that `runner_nested_template_url` in your app config points to a publicly accessible YAML file, then re-sync.',
      },
    ],
  },
}

export const Default = () => (
  <GenerateStackDetails appConfig={mockConfig} isLoading={false} />
)

export const Loading = () => <GenerateStackDetails isLoading={true} />

export const WithCompositeError = () => (
  <GenerateStackDetails
    appConfig={mockConfig}
    isLoading={false}
    stackVersion={mockVersionWithError}
  />
)

export const CompositeErrorOnly = () => (
  <GenerateStackDetails isLoading={false} stackVersion={mockVersionWithError} />
)
