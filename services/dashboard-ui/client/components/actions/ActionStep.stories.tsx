export default {
  title: 'Actions/ActionStep',
}

import { ActionStep } from './ActionStep'

export const WithCommand = () => (
  <ActionStep
    index={0}
    step={{
      name: 'Run tests',
      command: 'npm run test -- --coverage',
      env_vars: {
        NODE_ENV: 'test',
        CI: 'true',
      },
    } as any}
  />
)

export const WithInlineContents = () => (
  <ActionStep
    index={1}
    step={{
      name: 'Build Docker image',
      inline_contents: `#!/bin/bash
set -e
docker build -t myapp:latest .
docker push myapp:latest`,
      env_vars: {},
    } as any}
  />
)

export const WithGitHub = () => (
  <ActionStep
    index={2}
    step={{
      name: 'Deploy from GitHub',
      connected_github_vcs_config: {
        repo: 'my-org/my-repo',
        branch: 'main',
        root: '/',
      },
      env_vars: {},
    } as any}
  />
)

export const WithNoContent = () => (
  <ActionStep
    index={0}
    step={{
      name: 'Empty step',
      env_vars: {},
    } as any}
  />
)
