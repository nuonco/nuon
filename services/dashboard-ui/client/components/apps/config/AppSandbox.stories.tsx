export default {
  title: 'Apps/Config/AppSandbox',
}

import { AppSandbox } from './AppSandbox'

export const Default = () => (
  <AppSandbox
    appConfig={{
      sandbox: {
        terraform_version: '1.5.7',
        public_git_vcs_config: {
          repo: 'https://github.com/my-org/sandbox-config',
          branch: 'main',
          root: '/terraform',
        },
        variables: {
          instance_type: 't3.small',
          min_nodes: '2',
        },
        operation_roles: {},
      },
    } as any}
  />
)

export const WithGitHub = () => (
  <AppSandbox
    appConfig={{
      sandbox: {
        terraform_version: '1.6.0',
        connected_github_vcs_config: {
          repo: 'my-org/sandbox-terraform',
          branch: 'main',
          root: '/',
        },
        variables: {},
        operation_roles: {},
      },
    } as any}
  />
)
