export default {
  title: 'Workflows/StepDetails/DeployApply',
}

import { DeployApply, DeployApplySkeleton, DeployLogsSkeleton } from './DeployApply'
import type { TInstallDeploy } from '@/types'

const mockDeploy: TInstallDeploy = {
  id: 'deploy-abc123',
  status_v2: {
    status: 'active',
    status_human_description: 'Deploying components',
  },
  log_stream: null,
} as TInstallDeploy

export const Active = () => <DeployApply initDeploy={mockDeploy} />

export const Error = () => (
  <DeployApply
    initDeploy={{
      ...mockDeploy,
      status_v2: { status: 'error', status_human_description: 'Deploy failed' },
    } as TInstallDeploy}
  />
)

export const Null = () => <DeployApply initDeploy={null as any} />

export const ApplySkeleton = () => <DeployApplySkeleton />

export const LogsSkeleton = () => <DeployLogsSkeleton />
