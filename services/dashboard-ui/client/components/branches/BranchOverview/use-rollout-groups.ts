import { useMemo } from 'react'
import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { resolveInstallGroupMembership } from '@/components/branches/install-group-membership'
import { getRunTitle } from '@/components/branches/shared/run-title'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import {
  getAppInstalls,
  getBranchInstallGroupRuns,
  getBranchWorkflowRun,
  getBranchWorkflowRuns,
} from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import type {
  TAppBranchInstallGroup,
  TInstall,
  TInstallGroupRun,
} from '@/types'
import type { TOverviewRollout } from './BranchOverview'
import { fetchCommitReady } from './overview-loading'
import { buildRolloutStages } from './rollout-stages'
import type { TTrackGroup, TTrackInstall } from './RolloutTrack'
import { commitUrl, resolveRunSource } from './run-source'

const TERMINAL = new Set(['success', 'failed', 'error', 'cancelled'])

const groupRules = (group?: TAppBranchInstallGroup) => {
  if (!group) return undefined
  const labels = Object.entries(group.label_selector?.match_labels ?? {})
    .map(([key, value]) => `${key}=${value}`)
    .join(', ')
  return [
    labels || (group.default ? 'Every other install' : undefined),
    group.max_parallel ? `up to ${group.max_parallel} at a time` : undefined,
    group.auto_approve_on_policies_passing
      ? 'auto-approves when policies pass'
      : 'manual approval',
  ]
    .filter(Boolean)
    .join(' · ')
}

const installSnapshot = (
  install: TInstall | undefined,
  orgId?: string
): Pick<
  TTrackInstall,
  'resources' | 'deployment' | 'health' | 'overviewHref'
> => {
  if (!install?.id) return {}
  const resourcesActive =
    !!install.sandbox_health_status && install.sandbox_status === 'active'
  return {
    resources: (
      resourcesActive ? install.sandbox_health_status : install.sandbox_status
    )
      ? {
          status: (resourcesActive
            ? install.sandbox_health_status
            : install.sandbox_status) as string,
          detail: resourcesActive
            ? install.sandbox_health_message
            : install.sandbox_status_description,
        }
      : undefined,
    deployment: install.composite_component_status
      ? {
          status: install.composite_component_status,
          detail: install.composite_component_status_description,
        }
      : undefined,
    health: install.composite_health_status
      ? {
          status: install.composite_health_status,
          detail: install.composite_health_status_description,
        }
      : undefined,
    overviewHref: orgId ? `/${orgId}/installs/${install.id}` : undefined,
  }
}

const fromGroupRuns = (
  groupRuns: TInstallGroupRun[],
  groups: TAppBranchInstallGroup[],
  installsById: Record<string, TInstall>,
  orgId?: string
): TTrackGroup[] =>
  groupRuns.map((groupRun) => {
    const group =
      groups.find((item) => item.id === groupRun.install_group_id) ??
      groupRun.install_group
    return {
      id: groupRun.install_group_id ?? groupRun.id ?? '',
      name: groupRun.install_group_name || group?.name || 'Install group',
      status: groupRun.status?.status || 'pending',
      plannedCount: groupRun.total_installs,
      rules: groupRules(group),
      installs: (groupRun.installs ?? []).map((install) => {
        const id = install.install_id ?? ''
        return {
          id,
          name: installsById[id]?.name ?? id,
          status: install.status || 'pending',
          detail: install.runbooks?.length
            ? `${install.runbooks.length} post-deploy runbooks`
            : undefined,
          ...installSnapshot(installsById[id], orgId),
          workflowHref:
            id && orgId && install.workflow_id
              ? `/${orgId}/installs/${id}/workflows/${install.workflow_id}`
              : undefined,
        }
      }),
    }
  })

export const useRolloutGroups = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const orgId = org?.id
  const appId = app?.id
  const branchId = params.branchId as string
  const basePath = `/${orgId}/apps/${appId}/branches/${branchId}`

  const currentConfig = useMemo(() => latestBranchConfig(branch), [branch])
  const groups = useMemo(
    () => currentConfig?.install_groups ?? [],
    [currentConfig]
  )
  const vcs =
    currentConfig?.connected_github_vcs_config ??
    currentConfig?.public_git_vcs_config
  const repoSlug = vcs?.repo

  const { data: installsResult } = useQuery({
    queryKey: ['app-installs', orgId, appId, branchId],
    queryFn: () =>
      getAppInstalls({
        orgId: orgId!,
        appId: appId!,
        app_branch_id: branchId,
        limit: 100,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    placeholderData: keepPreviousData,
  })

  const { data: latestResult, isLoading: isLoadingLatest } = useQuery({
    queryKey: ['branch-latest-run', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId: orgId!,
        appId: appId!,
        branchId,
        limit: 1,
        offset: 0,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
    placeholderData: keepPreviousData,
  })

  const latestId = latestResult?.data?.[0]?.id
  const { data: rolloutRun, isLoading: isLoadingRollout } = useQuery({
    queryKey: ['branch-run', orgId, appId, branchId, latestId],
    queryFn: () =>
      getBranchWorkflowRun({
        orgId: orgId!,
        appId: appId!,
        branchId,
        runId: latestId!,
      }),
    enabled: !!orgId && !!appId && !!branchId && !!latestId,
    refetchInterval: 5000,
    placeholderData: keepPreviousData,
  })

  const branchRun = rolloutRun?.app_branch_runs?.at(0)
  const branchRunId = branchRun?.id
  const isTerminal = TERMINAL.has(rolloutRun?.status?.status ?? '')

  const { data: groupRuns } = useQuery({
    queryKey: [
      'branch-install-group-runs',
      orgId,
      appId,
      branchId,
      branchRunId,
    ],
    queryFn: () =>
      getBranchInstallGroupRuns({
        orgId: orgId!,
        appId: appId!,
        branchId,
        runId: branchRunId!,
      }),
    enabled: !!orgId && !!appId && !!branchRunId,
    refetchInterval: isTerminal ? false : 5000,
    placeholderData: keepPreviousData,
  })

  const installs = useMemo(
    () => installsResult?.data ?? [],
    [installsResult?.data]
  )
  const installsById = useMemo(
    () => Object.fromEntries(installs.map((install) => [install.id, install])),
    [installs]
  )
  const membership = useMemo(
    () => resolveInstallGroupMembership(installs, groups),
    [installs, groups]
  )

  const trackGroups = useMemo<TTrackGroup[]>(() => {
    if (groupRuns?.length) {
      return fromGroupRuns(groupRuns, groups, installsById, orgId)
    }
    return buildRolloutStages({
      groups,
      steps: rolloutRun?.steps ?? [],
      installsByGroup: membership.installsByGroup,
    })
      .filter((stage) => stage.kind === 'group')
      .map((stage) => {
        const group = groups.find((item) => item.id === stage.groupId)
        return {
          id: stage.groupId ?? stage.id,
          name: stage.name,
          status: stage.status,
          rules: groupRules(group),
          installs: (stage.installs ?? []).map((install) => ({
            id: install.id,
            name: install.name,
            status: install.status,
            detail: install.region,
            ...installSnapshot(installsById[install.id], orgId),
          })),
        }
      })
  }, [groupRuns, groups, installsById, rolloutRun?.steps, membership, orgId])

  const sha = branchRun?.vcs_connection_commit?.sha ?? branchRun?.head_sha
  const rollout: TOverviewRollout | undefined = rolloutRun?.id
    ? {
        id: rolloutRun.id,
        href: `${basePath}/runs/${rolloutRun.id}`,
        source: resolveRunSource(branchRun, repoSlug),
        title: getRunTitle(rolloutRun),
        sha,
        shaUrl: commitUrl(repoSlug, sha),
        author: branchRun?.vcs_connection_commit?.author_name,
        status: rolloutRun.status?.status || 'unknown',
        activity: rolloutRun.status?.status_human_description,
        commit: fetchCommitReady(rolloutRun.steps ?? [], sha)
          ? {
              message: branchRun?.vcs_connection_commit?.message,
              author: branchRun?.vcs_connection_commit?.author_name,
              avatarUrl: branchRun?.vcs_connection_commit?.author_avatar_url,
              sha,
              shaUrl: commitUrl(repoSlug, sha),
              createdAt: branchRun?.vcs_connection_commit?.created_at,
            }
          : undefined,
      }
    : undefined

  const workflowSteps = rolloutRun?.steps ?? []

  return {
    app,
    branch,
    orgId,
    appId,
    branchId,
    basePath,
    repoSlug,
    branchRunId,
    rollout,
    workflowSteps,
    showLoadingTrack:
      isLoadingLatest ||
      (!!latestId && isLoadingRollout && !rolloutRun) ||
      (!!rollout && !isTerminal),
    groups: trackGroups,
    hasPlan: groups.length > 0,
    isLoading:
      isLoadingLatest || (!!latestId && isLoadingRollout && !rolloutRun),
  }
}
