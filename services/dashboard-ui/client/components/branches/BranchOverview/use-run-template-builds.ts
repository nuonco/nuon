import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { TTemplateBuildChange } from '@/components/branches/ConfigChanges'
import { useApp } from '@/hooks/use-app'
import { getBranchRunBuilds, getSandboxBuild } from '@/lib'
import type { TInstallWorkflowStep } from '@/types'
import { vcsRepo } from '@/utils/vcs-urls'
import { changedBuildRows, type TBuildMeta } from './changed-builds'

const isBuildStep = (name?: string) =>
  !!name && /build/i.test(name) && !/(?:fetch|sync) app config/i.test(name)

const NO_META_BUILDS: TBuildMeta[] = []

const TERMINAL_RUN_STATUSES = new Set([
  'success',
  'failed',
  'error',
  'cancelled',
])

export const isTerminalRunStatus = (status?: string) =>
  TERMINAL_RUN_STATUSES.has(status ?? '')

interface IRunTemplateBuilds {
  orgId?: string
  appId?: string
  branchId?: string
  branchRunId?: string
  workflowSteps: TInstallWorkflowStep[]
  isRunTerminal?: boolean
}

export const useRunTemplateBuilds = ({
  orgId,
  appId,
  branchId,
  branchRunId,
  workflowSteps,
  isRunTerminal,
}: IRunTemplateBuilds) => {
  const { app } = useApp()
  const {
    data: builds,
    isLoading: buildsLoading,
    isPending: buildsQueryPending,
    isPlaceholderData: buildsPlaceholder,
    isError: buildsError,
  } = useQuery({
    queryKey: ['branch-run-builds', orgId, appId, branchId, branchRunId],
    queryFn: () =>
      getBranchRunBuilds({
        orgId: orgId!,
        appId: appId!,
        branchId,
        runId: branchRunId!,
      }),
    enabled: !!orgId && !!appId && !!branchRunId,
    placeholderData: keepPreviousData,
    // Undefined means this caller does not follow a live run.
    refetchInterval: isRunTerminal === false ? 5000 : false,
  })
  const buildStep = workflowSteps.find((step) => isBuildStep(step.name))
  const buildMetadata = (buildStep?.status?.metadata ?? {}) as {
    builds?: TBuildMeta[]
    sandbox_build_id?: string
  }
  const metaBuilds = buildMetadata.builds ?? NO_META_BUILDS
  const sandboxBuildId = buildMetadata.sandbox_build_id
  const runBuilds = useMemo(
    () =>
      (builds ?? []).map((build) => ({
        id: build.id,
        component_id: build.component_id,
        component_name: build.component_name,
        status: build.status_v2?.status || build.status,
      })),
    [builds]
  )
  const changedBuilds = useMemo(
    () =>
      changedBuildRows({
        metaBuilds,
        runBuilds,
        orgId,
        appId,
        sandboxBuildId,
      }),
    [metaBuilds, runBuilds, orgId, appId, sandboxBuildId]
  )
  const { data: sandboxBuild } = useQuery({
    queryKey: ['sandbox-build', orgId, appId, buildMetadata.sandbox_build_id],
    queryFn: () =>
      getSandboxBuild({
        orgId: orgId!,
        appId: appId!,
        buildId: buildMetadata.sandbox_build_id!,
      }),
    enabled: !!orgId && !!appId && !!buildMetadata.sandbox_build_id,
  })
  const templateBuilds = useMemo<TTemplateBuildChange[]>(
    () =>
      changedBuilds.map((row) => {
        const commit = sandboxBuild?.vcs_connection_commit
        return {
          ...row,
          kind: row.kind ?? 'component',
          commit:
            row.kind === 'sandbox' && commit?.sha
              ? {
                  sha: commit.sha,
                  message: commit.message,
                  author: commit.author_name,
                  repo: vcsRepo(app?.sandbox_config),
                }
              : undefined,
        }
      }),
    [changedBuilds, sandboxBuild, app?.sandbox_config]
  )

  const buildsEnabled = !!orgId && !!appId && !!branchRunId
  return {
    builds,
    buildsLoading,
    buildsPending:
      buildsEnabled &&
      !buildsError &&
      (buildsQueryPending || buildsPlaceholder),
    changedBuilds,
    metaBuilds,
    runBuilds,
    sandboxBuildId,
    templateBuilds,
  }
}
