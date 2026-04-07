export default {
  title: 'Apps/Config/AppStack',
}

import { AppStack } from './AppStack'

export const Default = () => (
  <AppStack
    appConfig={{
      stack: {
        type: 'eks',
        name: 'production-stack',
      },
    } as any}
  />
)

export const WithTemplateUrls = () => (
  <AppStack
    appConfig={{
      stack: {
        type: 'eks',
        name: 'production-stack',
        runner_nested_template_url: 'https://s3.amazonaws.com/my-bucket/runner.yaml',
        vpc_nested_template_url: 'https://s3.amazonaws.com/my-bucket/vpc.yaml',
      },
    } as any}
  />
)

export const WithCustomStacks = () => (
  <AppStack
    appConfig={{
      stack: {
        type: 'eks',
        name: 'production-stack',
        custom_nested_stacks: [
          {
            name: 'monitoring',
            template_url: 'https://s3.amazonaws.com/my-bucket/monitoring.yaml',
            contents_hash: 'abc123',
          },
          {
            name: 'logging',
            template_url: 'https://s3.amazonaws.com/my-bucket/logging.yaml',
            contents_hash: 'def456',
          },
        ],
      },
    } as any}
  />
)
