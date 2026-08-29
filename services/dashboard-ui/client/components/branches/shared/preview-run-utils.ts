import type {
  TAppBranchPreviewConfig,
  TAppBranchRun,
  TAppBranchRunPreview,
  TAppBranchRunPreviewMode,
  TInstall,
  TInstallWorkflow,
} from '@/types'
import {
  defaultPreviewDefaults,
  previewDefaultsFromConfig,
  type IPreviewDefaults,
} from '@/components/branches/shared/PreviewDefaultsEditor'

export const isPreviewBranchRun = (branchRun?: TAppBranchRun): boolean => {
  if (!branchRun) return false
  return (
    !!branchRun.preview ||
    !!branchRun.plan_only ||
    branchRun.run_type === 'git-preview-run'
  )
}

export const getBranchRunFromWorkflow = (
  workflow?: TInstallWorkflow
): TAppBranchRun | undefined => workflow?.app_branch_runs?.[0]

export const isPreviewWorkflow = (workflow?: TInstallWorkflow): boolean =>
  isPreviewBranchRun(getBranchRunFromWorkflow(workflow))

export const previewModeLabel = (preview?: TAppBranchRunPreview): string | undefined =>
  preview?.mode

export const previewSourceLabel = (branchRun?: TAppBranchRun): string | undefined => {
  const preview = branchRun?.preview
  if (preview?.source === 'pr' && branchRun?.pr_number != null) {
    return `PR #${branchRun.pr_number}`
  }
  if (preview?.git_ref) return preview.git_ref
  if (preview?.source === 'branch' && branchRun?.base_branch) return branchRun.base_branch
  return undefined
}

const modeDisplayLabel = (mode: TAppBranchRunPreviewMode): string => {
  switch (mode) {
    case 'apply':
      return 'apply'
    case 'build-only':
      return 'build-only'
    default:
      return 'plan-only'
  }
}

export const resolveInstallName = (
  installId: string,
  config?: TAppBranchPreviewConfig,
  installs?: TInstall[]
): string => {
  if (config?.install_name) return config.install_name
  const match = installs?.find((i) => i.id === installId)
  if (match?.name) return match.name
  return installId
}

export const formatPreviewDefaultsSummary = (
  config?: TAppBranchPreviewConfig,
  installs?: TInstall[],
  options?: { includeGithub?: boolean }
): string => {
  const defaults = previewDefaultsFromConfig(config, installs)
  const parts: string[] = [modeDisplayLabel(defaults.mode)]

  if (defaults.mode !== 'build-only') {
    if (defaults.installTargetMode === 'labels' && Object.keys(defaults.labelSelector).length > 0) {
      const labels = Object.entries(defaults.labelSelector)
        .map(([k, v]) => `${k}=${v}`)
        .join(', ')
      parts.push(`labels (${labels})`)
    } else if (defaults.installId) {
      parts.push(resolveInstallName(defaults.installId, config, installs))
    } else {
      parts.push('no install')
    }
  }

  if (options?.includeGithub !== false) {
    if (defaults.setStatuses) parts.push('statuses on')
    if (defaults.comment) parts.push('PR comments on')
  }

  return parts.join(' · ')
}

export const isPreviewOverride = (
  mode: TAppBranchRunPreviewMode,
  installId: string,
  branchDefaults: IPreviewDefaults
): boolean =>
  mode !== branchDefaults.mode ||
  (mode !== 'build-only' && installId !== branchDefaults.installId)

export const effectivePreviewDefaults = (
  config?: TAppBranchPreviewConfig,
  installs?: TInstall[]
): IPreviewDefaults => previewDefaultsFromConfig(config, installs) ?? defaultPreviewDefaults()

export const sortPreviewInstallCandidates = (
  installs: TInstall[],
  branchId: string
): TInstall[] => {
  const rank = (install: TInstall) => {
    if (install.app_branch_id === branchId) return 0
    if (!install.app_branch_id) return 1
    return 2
  }

  return [...installs].sort((a, b) => {
    const rankDiff = rank(a) - rank(b)
    if (rankDiff !== 0) return rankDiff

    if (rank(a) === 2) {
      const branchNameDiff = (a.app_branch?.name ?? '').localeCompare(b.app_branch?.name ?? '')
      if (branchNameDiff !== 0) return branchNameDiff
    }

    return (a.name ?? a.id).localeCompare(b.name ?? b.id)
  })
}
