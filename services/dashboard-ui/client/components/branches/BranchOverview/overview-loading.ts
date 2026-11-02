import {
  getStepDisplayStatus,
  isActiveStepStatus,
  stepStatusCategory,
} from '@/components/branches/shared/step-status'
import type { TCompositeError, TInstallWorkflowStep } from '@/types'
import { deployStepForGroup } from './rollout-stages'
import type { TTrackGroup } from './RolloutTrack'

export type TOverviewStageStatus =
  | 'pending'
  | 'in-progress'
  | 'success'
  | 'error'

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

export const buildComponentsStep = (steps: TInstallWorkflowStep[]) =>
  findStep(steps, isBuildComponents)

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

export const groupStageStatus = (status?: string): TOverviewStageStatus => {
  const category = stepStatusCategory(status)
  if (category === 'error') return 'error'
  if (category === 'success') return 'success'
  if (category === 'active' || category === 'awaiting') return 'in-progress'
  return 'pending'
}

const isSkipped = (status?: string) => !!status?.endsWith('skipped')

export const rolloutStageStatus = (
  groupStatuses: string[]
): TOverviewStageStatus => {
  const statuses = groupStatuses.map((status) =>
    isSkipped(status) ? 'success' : groupStageStatus(status)
  )
  if (statuses.length === 0) return 'pending'
  if (statuses.includes('error')) return 'error'
  if (statuses.every((status) => status === 'success')) return 'success'
  if (statuses.every((status) => status === 'pending')) return 'pending'
  return 'in-progress'
}

export type TPreviewProgress = 'build-only' | 'plan-only' | 'apply'

const isPlanPreview = (name?: string) => !!name && /plan preview/i.test(name)

const isApplyPreview = (name?: string) => !!name && /apply preview/i.test(name)

export const buildOverviewLoadingStages = ({
  steps,
  sha,
  groupStatuses = [],
  previewMode,
}: {
  steps: TInstallWorkflowStep[]
  sha?: string
  groupStatuses?: string[]
  previewMode?: TPreviewProgress
}): TOverviewStage[] => {
  const fetchCommit = findStep(steps, isFetchCommit)
  const appConfig = findStep(steps, isAppConfig)
  const build = findStep(steps, isBuildComponents)
  const fetchStatus = stepStageStatus(fetchCommit)
  const fetchCommitStatus: TOverviewStageStatus =
    fetchStatus === 'success' ? (sha ? 'success' : 'in-progress') : fetchStatus

  const stages: TOverviewStage[] = [
    {
      id: 'starting',
      label: 'Starting workflow',
      status: steps.length > 0 ? 'success' : 'in-progress',
    },
    {
      id: 'fetch-commit',
      label: 'Fetch commit',
      status: fetchCommitStatus,
    },
    {
      id: 'app-config',
      label: 'Compile and diff app config',
      status: stepStageStatus(appConfig),
    },
    {
      id: 'build-components',
      label: 'Build components',
      status: stepStageStatus(build),
    },
  ]
  if (previewMode === 'build-only') return stages
  if (previewMode === 'plan-only') {
    return [
      ...stages,
      {
        id: 'plan',
        label: 'Plan',
        status: stepStageStatus(findStep(steps, isPlanPreview)),
      },
    ]
  }
  if (previewMode === 'apply') {
    return [
      ...stages,
      {
        id: 'apply',
        label: 'Apply',
        status: stepStageStatus(findStep(steps, isApplyPreview)),
      },
    ]
  }
  return [
    ...stages,
    {
      id: 'rollout',
      label: 'Rollout',
      status: rolloutStageStatus(groupStatuses),
    },
  ]
}

const isDeployGroup = (name?: string) =>
  !!name && /^deploy install group:/i.test(name)

export const overviewCompositeError = (
  steps: TInstallWorkflowStep[],
  runError?: TCompositeError
): TCompositeError | undefined => {
  if (runError?.message || runError?.type) return runError
  const ordered = [
    findStep(steps, isFetchCommit),
    findStep(steps, isAppConfig),
    findStep(steps, isBuildComponents),
    ...steps.filter((step) => isDeployGroup(step.name)),
  ]
  return ordered.find(
    (step) =>
      step && stepStageStatus(step) === 'error' && step.status?.composite_error
  )?.status?.composite_error
}

const isRolloutError = (error?: TCompositeError) =>
  !!error?.type?.startsWith('install_group.')

export const preRolloutCompositeError = (
  steps: TInstallWorkflowStep[],
  runError?: TCompositeError
): TCompositeError | undefined => {
  const failed = [
    findStep(steps, isFetchCommit),
    findStep(steps, isAppConfig),
    findStep(steps, isBuildComponents),
  ].find((step) => step && stepStageStatus(step) === 'error')
  if (failed) {
    return runError?.message || runError?.type
      ? runError
      : failed.status?.composite_error
  }
  const rolloutFailed = steps.some(
    (step) => isDeployGroup(step.name) && stepStageStatus(step) === 'error'
  )
  if (rolloutFailed || isRolloutError(runError)) return undefined
  return runError?.message || runError?.type ? runError : undefined
}

const failedInstallId = (error?: TCompositeError) =>
  (error?.data as unknown as { install_id?: string } | undefined)?.install_id

const cancelledInstallFinish = (message?: string) =>
  !!message && /finished as cancelled\b/i.test(message)

export const groupCompositeError = (
  steps: TInstallWorkflowStep[],
  group: Pick<TTrackGroup, 'name' | 'status' | 'installs'>,
  runError?: TCompositeError
): TCompositeError | undefined => {
  const deploy = deployStepForGroup(steps, group.name)
  if (deploy && stepStageStatus(deploy) === 'error') {
    const stepError = deploy.status?.composite_error
    if (stepError?.message || stepError?.type) {
      return cancelledInstallFinish(stepError.message) ? undefined : stepError
    }
    const description = deploy.status?.status_human_description
    if (description && !cancelledInstallFinish(description)) {
      return { message: description, severity: 'error' }
    }
  }
  if (
    groupStageStatus(group.status) !== 'error' ||
    !isRolloutError(runError) ||
    cancelledInstallFinish(runError?.message)
  ) {
    return undefined
  }
  const installId = failedInstallId(runError)
  return installId && group.installs.some((install) => install.id === installId)
    ? runError
    : undefined
}

export const installFailureHref = (
  error: TCompositeError | undefined,
  orgId?: string
) => {
  if (
    !error ||
    error.type !== 'install_group.install_update_failed' ||
    !orgId
  ) {
    return undefined
  }
  const data = error.data as unknown as
    | { install_id?: string; workflow_id?: string }
    | undefined
  if (!data?.install_id || !data.workflow_id) return undefined
  return `/${orgId}/installs/${data.install_id}/workflows/${data.workflow_id}`
}

export const fetchCommitReady = (steps: TInstallWorkflowStep[], sha?: string) =>
  stepStageStatus(findStep(steps, isFetchCommit)) === 'success' && !!sha
