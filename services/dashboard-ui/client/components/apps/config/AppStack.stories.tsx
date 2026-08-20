export default {
  title: 'Apps/Config/AppStack',
}

import { AppStack } from './AppStack'

export const Default = () => (
  <AppStack
    appConfig={
      {
        stack: {
          type: 'aws-cloudformation',
          name: 'production-stack',
          description: 'Infrastructure stack for the payments app',
        },
      } as any
    }
  />
)

export const WithTemplateUrls = () => (
  <AppStack
    appConfig={
      {
        stack: {
          type: 'aws-cloudformation',
          name: 'production-stack',
          description: 'Infrastructure stack for the payments app',
          runner_nested_template_url:
            'https://s3.amazonaws.com/my-bucket/runner.yaml',
          vpc_nested_template_url:
            'https://s3.amazonaws.com/my-bucket/vpc.yaml',
        },
      } as any
    }
  />
)

export const AzureSubscriptionScope = () => (
  <AppStack
    appConfig={
      {
        stack: {
          type: 'azure-bicep',
          name: 'payments-{{.nuon.install.id}}',
          description: 'Subscription-scoped install stack',
          runner_nested_template_url:
            'https://raw.githubusercontent.com/acme/install-stacks/main/azure/runner.json',
          vpc_nested_template_url:
            'https://raw.githubusercontent.com/acme/install-stacks/main/azure/vnet.json',
          deployment_scope: 'subscription',
        },
      } as any
    }
  />
)

// deployment_scope is unset rather than 'resource_group', which is how the API
// sends every config that has not opted into subscription scope.
export const AzureDefaultScope = () => (
  <AppStack
    appConfig={
      {
        stack: {
          type: 'azure-bicep',
          name: 'payments-{{.nuon.install.id}}',
          description: 'Resource-group-scoped install stack',
        },
      } as any
    }
  />
)

export const WithCustomStacks = () => (
  <AppStack
    appConfig={
      {
        stack: {
          type: 'aws-cloudformation',
          name: 'production-stack',
          description: 'Infrastructure stack for the payments app',
          custom_nested_stacks: [
            {
              index: 0,
              name: 'monitoring',
              template_url:
                'https://s3.amazonaws.com/my-bucket/monitoring.yaml',
              contents_hash: 'abc123',
            },
            {
              index: 1,
              name: 'logging',
              template_url: 'https://s3.amazonaws.com/my-bucket/logging.yaml',
              contents_hash: 'def456',
            },
          ],
        },
      } as any
    }
  />
)
