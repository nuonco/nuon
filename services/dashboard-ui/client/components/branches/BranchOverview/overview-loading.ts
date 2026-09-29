import {
  getStepDisplayStatus,
  isActiveStepStatus,
  stepStatusCategory,
} from '@/components/branches/shared/step-status'
import type { TCompositeError, TInstallWorkflowStep } from '@/types'

export type TOverviewStageStatus = 'pending' | 'in-progress' | 'success' | 'error'

export interface TOverviewStage {
  id: string
  label: string
  status: TOverviewStageStatus
}

const FETCH_COMMIT = /fetch commit/i
const APP_CONFIG = /(?:fetch|sync) app config/i

const isFetchCommit = (name?: string) => !!name && FETCH_COMMIT.test(name)

const isAppConfig = (name?: string) => !!name && APP_CONFIG.test(name)

const isBuildComponents = (name?: string) =>
  !!name && /build/i.test(name) && !isAppConfig(name)

const findStep = (
  steps: TInstallWorkflowStep[],
  match: (name?: string) => boolean
) => steps.find((step) => match(step.name))

export const stepStageStatus = (
  step?: TInstallWorkflowStep
): TOverviewStageStatus => {
  if (!step) return 'pending'
  const status = getStepDisplayStatus(step)
  if (status === 'queued' || isActiveStepStatus(status)) return 'in-progress'
  const category = stepStatusCategory(status)
  if (category === 'error') return 'error'
  if (category === 'success') return 'success'
  if (category === 'active' || category === 'awaiting') return 'in-progress'
  return 'pending'
}

export const buildOverviewLoadingStages = ({
  steps,
  sha,
}: {
  steps: TInstallWorkflowStep[]
  sha?: string
}): TOverviewStage[] => {
  const fetchCommit = findStep(steps, isFetchCommit)
  const appConfig = findStep(steps, isAppConfig)
  const build = findStep(steps, isBuildComponents)
  const fetchStatus = stepStageStatus(fetchCommit)
  const showCommit: TOverviewStageStatus =
    fetchStatus === 'success' ? (sha ? 'success' : 'in-progress') : 'pending'

  return [
    {
      id: 'waiting',
      label: 'Waiting for workflow',
      status: steps.length > 0 ? 'success' : 'in-progress',
    },
    {
      id: 'fetch-commit',
      label: 'Fetch commit',
      status: fetchStatus,
    },
    {
      id: 'show-commit',
      label: 'Show commit',
      status: showCommit,
    },
    {
      id: 'app-config',
      label: 'Build app config and diff',
      status: stepStageStatus(appConfig),
    },
    {
      id: 'build-components',
      label: 'Build components',
      status: stepStageStatus(build),
    },
  ]
}

export const overviewCompositeError = (
  steps: TInstallWorkflowStep[]
): TCompositeError | undefined => {
  const ordered = [
    findStep(steps, isFetchCommit),
    findStep(steps, isAppConfig),
    findStep(steps, isBuildComponents),
  ]
  return ordered.find(
    (step) => step && stepStageStatus(step) === 'error' && step.status?.composite_error
  )?.status?.composite_error
}

export const fetchCommitReady = (
  steps: TInstallWorkflowStep[],
  sha?: string
) => stepStageStatus(findStep(steps, isFetchCommit)) === 'success' && !!sha
