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
import { installGroupApprovalLabel } from '@/components/branches/install-group-approval'
import { installGroupMatch } from './InstallGroupMatch'
import { fetchCommitReady } from './overview-loading'
import { buildRolloutStages } from './rollout-stages'
import type { TTrackGroup, TTrackInstall } from './RolloutTrack'
import { commitUrl, resolveRunSource } from './run-source'

const TERMINAL = new Set(['success', 'failed', 'error', 'cancelled'])

const installRegion = (install?: TInstall) =>
  install?.aws_account?.region ||
  install?.gcp_account?.region ||
  install?.azure_account?.location

const installSnapshot = (
  install: TInstall | undefined,
  orgId?: string
): Pick<
  TTrackInstall,
  'resources' | 'deployment' | 'health' | 'overviewHref' | 'labels' | 'region'
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
    labels: install.labels,
    region: installRegion(install),
  }
}

const fromGroupRun = (
  groupRun: TInstallGroupRun,
  groups: TAppBranchInstallGroup[],
  installsById: Record<string, TInstall>,
  orgId?: string
): TTrackGroup => {
  const group =
    groups.find((item) => item.id === groupRun.install_group_id) ??
    groupRun.install_group
  return {
    id: groupRun.install_group_id ?? groupRun.id ?? '',
    name: groupRun.install_group_name || group?.name || 'Install group',
    status: groupRun.status?.status || 'pending',
    plannedCount: groupRun.total_installs,
    match: installGroupMatch(group),
    approval: installGroupApprovalLabel(group),
    maxParallel: group ? (group.max_parallel ?? 1) : undefined,
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
        workflowId: install.workflow_id,
        workflowHref:
          id && orgId && install.workflow_id
            ? `/${orgId}/installs/${id}/workflows/${install.workflow_id}`
            : undefined,
      }
    }),
  }
}

const sameName = (a?: string, b?: string) =>
  !!a && !!b && a.toLowerCase() === b.toLowerCase()

export const mergeGroupRuns = (
  planned: TTrackGroup[],
  groupRuns: TInstallGroupRun[],
  groups: TAppBranchInstallGroup[],
  installsById: Record<string, TInstall>,
  orgId?: string
): TTrackGroup[] => {
  const used = new Set<TInstallGroupRun>()
  const merged = planned.map((group) => {
    const groupRun = groupRuns.find(
      (run) =>
        !used.has(run) &&
        (run.install_group_id === group.id ||
          sameName(run.install_group_name, group.name))
    )
    if (!groupRun) return group
    used.add(groupRun)
    return fromGroupRun(groupRun, groups, installsById, orgId)
  })
  const extra = groupRuns
    .filter((run) => !used.has(run))
    .map((run) => fromGroupRun(run, groups, installsById, orgId))
  return [...merged, ...extra]
}

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
    const planned = buildRolloutStages({
      groups,
      steps: rolloutRun?.steps ?? [],
      installsByGroup: membership.installsByGroup,
    })
      .filter((stage) => stage.kind === 'group')
      .map((stage): TTrackGroup => {
        const index = groups.findIndex((item) => item.id === stage.groupId)
        const group = groups[index]
        const members = membership.installsByGroup[index] ?? []
        const installs: {
          id: string
          name: string
          status: string
          region?: string
        }[] =
          stage.installs ??
          members.map((install) => ({
            id: install.id ?? '',
            name: install.name ?? install.id ?? 'Install',
            status: 'pending',
          }))
        return {
          id: stage.groupId ?? stage.id,
          name: stage.name,
          status: stage.status,
          match: installGroupMatch(group),
          approval: installGroupApprovalLabel(group),
          maxParallel: group ? (group.max_parallel ?? 1) : undefined,
          installs: installs.map((install) => ({
            id: install.id,
            name: install.name,
            status: install.status,
            ...installSnapshot(installsById[install.id], orgId),
          })),
        }
      })
    return groupRuns?.length
      ? mergeGroupRuns(planned, groupRuns, groups, installsById, orgId)
      : planned
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
    branchRun,
    showLoadingTrack:
      isLoadingLatest ||
      (!!latestId && isLoadingRollout && !rolloutRun) ||
      !!rollout,
    groups: trackGroups,
    hasPlan: groups.length > 0,
    isLoading:
      isLoadingLatest || (!!latestId && isLoadingRollout && !rolloutRun),
  }
}
