export default {
  title: 'InstallComponents/LatestDeployCard',
}

import type { TDeploy } from '@/types'
import { LatestDeployCard } from './LatestDeployCard'

const deploy = {
  id: 'dep01hzk8t3fqp2r9x4m7wcn5vb',
  created_at: '2026-08-24T10:12:00Z',
  updated_at: '2026-08-24T10:19:42Z',
  install_deploy_type: 'apply',
  status_v2: {
    status: 'success',
    status_human_description: 'Deploy applied.',
  },
} as unknown as TDeploy

export const Default = () => (
  <LatestDeployCard deploy={deploy} href="/org-001/installs/inst-001" />
)

export const Empty = () => <LatestDeployCard />

export const Loading = () => <LatestDeployCard isLoading deploy={deploy} />
