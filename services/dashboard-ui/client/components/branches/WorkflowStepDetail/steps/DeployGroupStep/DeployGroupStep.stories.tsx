export default {
  title: 'Branches/WorkflowStepDetail/DeployGroupStep',
}

import { DeployGroupStep } from './DeployGroupStep'

const step = (status: string): any => ({
  id: 'step-deploy',
  name: 'deploy install group: production',
  status: { status, status_human_description: 'Applying terraform' },
})

export const InProgress = () => (
  <DeployGroupStep
    step={step('in-progress')}
    metadata={{
      current_activity: 'Applying terraform to us-east-1',
      installs: [
        { install_id: 'i1', install_name: 'acme-prod', status: 'success', region: 'us-east-1', version: 'v1.4.2', duration: '3m 12s' },
        { install_id: 'i2', install_name: 'globex-prod', status: 'in-progress', region: 'us-west-2', progress: 55, activity: 'Waiting for pods' },
        { install_id: 'i3', install_name: 'initech-prod', status: 'pending', region: 'eu-west-1' },
      ],
    }}
  />
)

export const AllDeployed = () => (
  <DeployGroupStep
    step={step('success')}
    metadata={{
      installs: [
        { install_id: 'i1', install_name: 'acme-prod', status: 'deployed', region: 'us-east-1', version: 'v1.4.2', duration: '3m 12s' },
        { install_id: 'i2', install_name: 'globex-prod', status: 'deployed', region: 'us-west-2', version: 'v1.4.2', duration: '2m 48s' },
      ],
    }}
  />
)

export const Empty = () => <DeployGroupStep step={step('in-progress')} metadata={{}} />
