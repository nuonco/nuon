export default {
  title: 'InstallComponents/LatestDeployCard',
}

import type { TComponentBuild, TDeploy } from '@/types'
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

const build = {
  id: 'bld01hzk8t3fqp2r9x4m7wcn5vb',
  created_at: '2026-08-24T09:58:00Z',
  status_v2: {
    status: 'success',
    status_human_description: 'Build finished.',
  },
  vcs_connection_commit: {
    sha: '4f1c9b2d8a7e5c3f1b0d9a8c7e6f5d4c3b2a1908',
    message: 'Bump acme-api chart to 2.4.0',
    author_name: 'Ada Lovelace',
  },
} as unknown as TComponentBuild

export const Default = () => (
  <LatestDeployCard deploy={deploy} href="/org-001/installs/inst-001" />
)

export const WithBuild = () => (
  <LatestDeployCard
    deploy={deploy}
    build={build}
    buildHref="/org-001/apps/app-001/components/cmp-001/builds/bld-001"
    href="/org-001/installs/inst-001"
  />
)

export const Sync = () => (
  <LatestDeployCard
    deploy={deploy}
    href="/org-001/installs/inst-001"
    variant="sync"
  />
)

export const Empty = () => <LatestDeployCard />

export const EmptySync = () => <LatestDeployCard variant="sync" />

export const Loading = () => <LatestDeployCard isLoading deploy={deploy} />
