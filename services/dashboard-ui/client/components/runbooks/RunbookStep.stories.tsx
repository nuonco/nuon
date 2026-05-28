export default {
  title: 'Runbooks/RunbookStep',
}

import { RunbookStep } from './RunbookStep'

export const DeployStep = () => (
  <RunbookStep
    index={0}
    step={{
      id: 'step-1',
      name: 'deploy-whoami',
      type: 'deploy',
      component_name: 'whoami',
      deploy_dependencies: true,
    }}
  />
)

export const ActionReferenceStep = () => (
  <RunbookStep
    index={1}
    step={{
      id: 'step-2',
      name: 'healthcheck',
      type: 'action',
      action_workflow_id: 'acw1u7awfrqw3utj6or2on1zab',
    }}
    actionBasePath="/org-1/installs/install-1"
  />
)

export const InlineActionStep = () => (
  <RunbookStep
    index={2}
    step={{
      id: 'step-3',
      name: 'check-status',
      type: 'action',
      command: './scripts/check-status.sh',
      inline_contents: `#!/bin/bash
set -e
echo "Checking service status..."
kubectl get pods -l app=whoami -o wide
kubectl rollout status deployment/whoami --timeout=120s`,
      timeout: 120000000000,
      role: 'deploy-role',
      env_vars: {
        NAMESPACE: 'default',
        CLUSTER: 'production',
      },
    }}
  />
)

export const MinimalStep = () => (
  <RunbookStep
    index={0}
    step={{
      id: 'step-4',
      name: 'simple-action',
      type: 'action',
    }}
  />
)
