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
import { buildRolloutStages } from './rollout-stages'
import type { TTrackGroup } from './RolloutTrack'
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

const fromGroupRuns = (
  groupRuns: TInstallGroupRun[],
  groups: TAppBranchInstallGroup[],
  installsById: Record<string, TInstall>,
  installHref: (id: string) => string
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
          href: id
            ? installHref(id) +
              (install.workflow_id ? `/workflows/${install.workflow_id}` : '')
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
    const installHref = (id: string) => `/${orgId}/installs/${id}`
    if (groupRuns?.length) {
      return fromGroupRuns(groupRuns, groups, installsById, installHref)
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
            href: installHref(install.id),
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
      }
    : undefined

  return {
    app,
    branch,
    branchId,
    basePath,
    repoSlug,
    branchRunId,
    rollout,
    groups: trackGroups,
    hasPlan: groups.length > 0,
    isLoading:
      isLoadingLatest || (!!latestId && isLoadingRollout && !rolloutRun),
  }
}
