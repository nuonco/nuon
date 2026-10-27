import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigDiff, getBranchWorkflowRuns } from '@/lib'
import type { TAppBranchRun } from '@/types'
import {
  extractSections,
  computeSummary,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { TInstallWorkflowStep } from '@/types'
import type { PlanInstallDiff } from './PlanGroupStep'

// An approval's plan and config diff stay the same for that approval id.
const PLAN_DIFF_CACHE = {
  staleTime: Number.POSITIVE_INFINITY,
  refetchOnWindowFocus: false,
} as const

const commitOf = (run: TAppBranchRun) => {
  const commit = run.vcs_connection_commit
  const sha = commit?.sha || run.head_sha || run.metadata?.head_sha
  if (!sha) return undefined
  return {
    sha,
    message: commit?.message,
    author: commit?.author_name,
    createdAt: commit?.created_at,
  }
}

export const usePlanGroupInstalls = (
  step: TInstallWorkflowStep,
  metadata: Record<string, any>
) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const { app } = useApp()
  const appId = app?.id
  const { branchId } = useParams()
  const approvalId = step.approval?.id

  const { data: plan, isLoading: planLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['approval-plan', orgId, step.id, approvalId],
    queryFn: async () => {
      const res = await fetch(
        `/api/orgs/${orgId}/workflows/${step.install_workflow_id}/steps/${step.id}/approvals/${approvalId}/contents`
      )
      if (!res.ok)
        throw new Error(`Failed to fetch approval contents: ${res.status}`)
      return res.json()
    },
    enabled: !!orgId && !!step.id && !!step.install_workflow_id && !!approvalId,
    ...PLAN_DIFF_CACHE,
  })

  const rawInstalls = (plan?.installs || metadata.installs || []) as any[]
  const groupName: string | undefined =
    plan?.install_group ||
    metadata.install_group_name ||
    step.name?.replace(/^plan install group:\s*/i, '')

  const diffQueries = useQueries({
    queries: rawInstalls.map((inst) => ({
      queryKey: [
        'app-config-diff',
        orgId,
        appId,
        inst.new_app_config_id,
        inst.old_app_config_id,
      ],
      queryFn: () =>
        getAppConfigDiff({
          orgId,
          appId: appId!,
          configId: inst.new_app_config_id,
          oldConfigId: inst.old_app_config_id || undefined,
        }),
      enabled: !!orgId && !!appId && !!inst.new_app_config_id,
      ...PLAN_DIFF_CACHE,
    })),
  })

  const configIds = [
    ...new Set(
      rawInstalls
        .flatMap((inst) => [inst.old_app_config_id, inst.new_app_config_id])
        .filter(Boolean) as string[]
    ),
  ]
  const configQueries = useQueries({
    queries: configIds.map((configId) => ({
      queryKey: ['app-config', orgId, appId, configId],
      queryFn: () =>
        getAppConfig({ orgId, appId: appId!, appConfigId: configId }),
      enabled: !!orgId && !!appId,
      ...PLAN_DIFF_CACHE,
    })),
  })
  const configsById = new Map(
    configIds.map((id, i) => [id, configQueries[i]?.data] as const)
  )

  const { data: runsPage } = useQuery({
    queryKey: ['branch-config-commits', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId: appId!,
        branchId: branchId!,
        limit: 100,
        offset: 0,
      }),
    enabled: !!orgId && !!appId && !!branchId && configIds.length > 0,
    staleTime: 30_000,
  })
  const commitByConfigId = new Map<
    string,
    NonNullable<ReturnType<typeof commitOf>>
  >()
  for (const workflow of runsPage?.data ?? []) {
    for (const run of workflow.app_branch_runs ?? []) {
      if (!run.app_config_id || commitByConfigId.has(run.app_config_id))
        continue
      const commit = commitOf(run)
      if (commit) commitByConfigId.set(run.app_config_id, commit)
    }
  }

  const installs: PlanInstallDiff[] = rawInstalls.map((inst, i) => {
    const query = diffQueries[i]
    const next = configsById.get(inst.new_app_config_id)
    const previous = inst.old_app_config_id
      ? configsById.get(inst.old_app_config_id)
      : undefined
    const nextVersion = next?.version
    const previousVersion = previous?.version
    const commit = commitByConfigId.get(inst.new_app_config_id)
    const previousCommit = inst.old_app_config_id
      ? commitByConfigId.get(inst.old_app_config_id)
      : undefined
    const sections = query?.data?.diff ? extractSections(query.data.diff) : []
    const summary =
      sections.length > 0
        ? computeSummary(sections)
        : query?.data?.summary
          ? {
              added: query.data.summary.added,
              removed: query.data.summary.removed,
              changed: query.data.summary.changed,
            }
          : null

    return {
      installId: inst.install_id,
      installName: inst.install_name || inst.install_id,
      installLabels: inst.install_labels,
      sections,
      summary,
      isLoading: !!query?.isLoading,
      isError: !!query?.isError,
      alreadyApplied:
        !!inst.old_app_config_id &&
        inst.old_app_config_id === inst.new_app_config_id,
      versionLabel:
        nextVersion == null
          ? undefined
          : previousVersion == null || previousVersion === nextVersion
            ? `v${nextVersion}`
            : `v${previousVersion} → v${nextVersion}`,
      sha: commit?.sha,
      previousSha:
        previousCommit?.sha && previousCommit.sha !== commit?.sha
          ? previousCommit.sha
          : undefined,
      message: commit?.message,
      author: commit?.author,
      createdAt: commit?.createdAt,
    }
  })

  return {
    orgId,
    approvalId,
    groupName,
    installs,
    isLoading: planLoading && rawInstalls.length === 0,
  }
}
