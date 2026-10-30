export default {
  title: 'Features / Pull requests / Preview run progress',
}

import type { TOverviewStage } from '@/components/branches/BranchOverview/overview-loading'
import { PreviewRunProgress } from './PreviewRunProgress'

const stages = (active: number): TOverviewStage[] =>
  [
    ['starting', 'Starting workflow'],
    ['fetch-commit', 'Fetch commit'],
    ['app-config', 'Compile and diff app config'],
    ['build-components', 'Build components'],
    ['plan', 'Plan'],
  ].map(([id, label], index) => ({
    id,
    label,
    status:
      index < active ? 'success' : index === active ? 'in-progress' : 'pending',
  }))

export const FetchingCommit = () => (
  <PreviewRunProgress
    stages={stages(1)}
    description="fetching commit"
    sha="a1b2c3d4e5f6"
  />
)

export const CompilingConfig = () => (
  <PreviewRunProgress
    stages={stages(2)}
    description="fetched commit a1b2c3d4"
    baseRef="main"
  />
)

export const Building = () => (
  <PreviewRunProgress
    stages={stages(3)}
    description="building 2 components"
    builds={[
      {
        component_id: 'cmp-api',
        component_name: 'api',
        component_type: 'docker_build',
        status: 'success',
        change_reason: 'source_changed',
      },
      {
        component_id: 'cmp-worker',
        component_name: 'worker',
        component_type: 'helm_chart',
        status: 'building',
        change_reason: 'config_changed',
      },
      {
        component_id: 'sandbox',
        component_type: 'sandbox',
        component_name: 'Sandbox',
        status: 'skipped',
        change_reason: 'no_changes',
      },
    ]}
  />
)

export const Planning = () => (
  <PreviewRunProgress
    stages={stages(4)}
    description="planning against staging"
    install="staging"
  />
)
