export default {
  title: 'Branches/WorkflowStepDetail/PlanGroupStep',
}

import { PlanGroupStep } from './PlanGroupStep'

const noop = () => {}

const installs = [
  {
    install_id: 'i1',
    install_name: 'acme-prod',
    status: 'pending',
    install_labels: { tier: 'prod', region: 'us-east-1' },
    diff: {
      added: [{ component_id: 'c1', component_name: 'worker', component_type: 'docker_build' }],
      changed: [{ component_id: 'c2', component_name: 'api', component_type: 'helm_chart' }],
      removed: [],
    },
  },
  {
    install_id: 'i2',
    install_name: 'globex-staging',
    status: 'pending',
    sandbox_changed: true,
    diff: { added: [], changed: [], removed: [] },
  },
]

export const AwaitingApproval = () => (
  <PlanGroupStep
    installs={installs}
    groupName="production"
    orgId="org-1"
    hasResponse={false}
    showApproveBar
    isResponding={false}
    isInProgress={false}
    onRespond={noop}
  />
)

export const Approved = () => (
  <PlanGroupStep
    installs={installs}
    groupName="production"
    orgId="org-1"
    hasResponse
    responseType="approve"
    showApproveBar={false}
    isResponding={false}
    isInProgress={false}
    onRespond={noop}
  />
)

export const Computing = () => (
  <PlanGroupStep
    installs={[]}
    groupName="production"
    orgId="org-1"
    hasResponse={false}
    showApproveBar={false}
    isResponding={false}
    isInProgress
    onRespond={noop}
  />
)
