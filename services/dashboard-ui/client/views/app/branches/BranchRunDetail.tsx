import { useCallback, useState } from 'react'
import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { CompositeError } from '@/components/common/CompositeError'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { PageSection } from '@/components/layout/PageSection'
import { ProviderError } from '@/components/layout/ProviderError'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { BranchRunApproval } from '@/components/branches/BranchRunApproval'
import { BranchRunChanges } from '@/components/branches/BranchRunChanges'
import { BranchRunComparisonRuns } from '@/components/branches/BranchRunComparisonRuns'
import { BranchRunSummary } from '@/components/branches/BranchRunSummary'
import { RuntimeChanges } from '@/components/branches/RuntimeChanges'
import { WorkflowRunPanelButton } from '@/components/branches/WorkflowRunPanel'
import {
  isPreviewBranchRun,
  previewModeLabel,
  previewSourceLabel,
} from '@/components/branches/shared/preview-run-utils'
import {
  githubCommitUrl,
  resolvePrLink,
} from '@/components/branches/shared/pr-link'
import { getRunTitle } from '@/components/branches/shared/run-title'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { BranchProvider } from '@/providers/branch-provider'
import {
  ConfigDiffFocusContext,
  type TConfigDiffFocus,
} from '@/components/approvals/plan-diffs/config-diff-focus'
import type { TAPIError } from '@/types'
import { getBranchRunComparison, getBranchWorkflowRun } from '@/lib'

const BranchRunDetailContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string
  const runId = params.runId as string
  const [configFocus, setConfigFocus] = useState<TConfigDiffFocus | null>(null)
  const requestConfigFocus = useCallback(
    (sectionKey: string, entityName?: string) => {
      setConfigFocus((prev) => ({
        sectionKey,
        entityName,
        nonce: (prev?.nonce ?? 0) + 1,
      }))
    },
    []
  )

  const {
    data: run,
    isLoading,
    error,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-run', orgId, appId, branchId, runId],
    queryFn: () => getBranchWorkflowRun({ orgId, appId, branchId, runId }),
    enabled: !!orgId && !!appId && !!branchId && !!runId,
    refetchInterval: 5000,
  })

  const branchRun = run?.app_branch_runs?.at(0)
  const repoSlug =
    branchRun?.app_branch_config?.connected_github_vcs_config?.repo ||
    branchRun?.app_branch_config?.public_git_vcs_config?.repo

  const { data: comparison } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-run-comparison', orgId, appId, branchId, branchRun?.id],
    queryFn: () =>
      getBranchRunComparison({
        orgId: orgId!,
        appId: appId!,
        branchId,
        runId: branchRun!.id!,
        includeDiff: ['config'],
      }),
    enabled: !!orgId && !!appId && !!branchId && !!branchRun?.id,
    retry: 1,
  })

  if (error && !run) {
    return <ProviderError error={error as TAPIError} />
  }

  if (isLoading || !run) {
    return (
      <PageSection>
        <Text variant="body" theme="neutral">
          Loading workflow run...
        </Text>
      </PageSection>
    )
  }

  const status = run.status?.status || 'unknown'
  const statusDescription = run.status?.status_human_description || ''
  const runTitle = getRunTitle(run)
  const isPreview = isPreviewBranchRun(branchRun)
  const previewMode =
    previewModeLabel(branchRun?.preview) ??
    (branchRun?.plan_only ? 'Plan only' : undefined)
  const previewSource = previewSourceLabel(branchRun)
  const previewInstall = branchRun?.preview?.install_name

  const prLink = resolvePrLink({
    repoSlug,
    prNumber: branchRun?.pr_number,
    commitMessage: branchRun?.vcs_connection_commit?.message,
  })
  const commitUrl = githubCommitUrl(
    repoSlug,
    branchRun?.vcs_connection_commit?.sha
  )
  const currentGithubHref = prLink?.url ?? commitUrl

  const showRunComparison =
    !!branchRun?.vcs_connection_commit ||
    !!branchRun?.pr_number ||
    !!comparison?.head_run ||
    !!comparison?.base_run

  return (
    <ConfigDiffFocusContext.Provider
      value={{ requestFocus: requestConfigFocus }}
    >
      <>
        <PageTitle segments={[runTitle, app?.name]} />
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${org?.id}`, text: org?.name },
            { path: `/${org?.id}/apps`, text: 'Apps' },
            { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
            { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
            {
              path: `/${org?.id}/apps/${app?.id}/branches/${branchId}`,
              text: branch?.name,
            },
            {
              path: `/${org?.id}/apps/${appId}/branches/${branchId}/runs/${runId}`,
              text: 'Run',
            },
          ]}
        />

        <DetailPage
          className="max-w-full"
          header={
            <DetailHeader
              title={
                prLink ? (
                  <Link href={prLink.url} isExternal variant="inline">
                    {runTitle}
                  </Link>
                ) : (
                  runTitle
                )
              }
              status={
                <>
                  {branch?.name ? (
                    <Badge size="sm" variant="code" className="shrink-0">
                      {branch.name}
                    </Badge>
                  ) : null}
                  {isPreview ? (
                    <Badge size="sm" variant="code" className="shrink-0">
                      preview
                    </Badge>
                  ) : null}
                  {previewMode ? (
                    <Badge size="sm" variant="code" className="shrink-0">
                      {previewMode}
                    </Badge>
                  ) : null}
                  {previewSource ? (
                    <Badge size="sm" variant="code" className="shrink-0">
                      {previewSource}
                    </Badge>
                  ) : null}
                  {previewInstall ? (
                    <Badge size="sm" variant="code" className="shrink-0">
                      install: {previewInstall}
                    </Badge>
                  ) : null}
                  {branchRun?.event_type === 'manual' ? (
                    <Badge size="sm" variant="code" className="shrink-0">
                      manual
                    </Badge>
                  ) : null}
                </>
              }
              id={run.id}
              actions={
                <>
                  <AdminDashboardLink
                    path={`/workflows/${run.id}`}
                    label="admin"
                  />
                  <WorkflowRunPanelButton runId={run.id!} />
                  <CancelWorkflowButton workflow={run} />
                </>
              }
              metadata={
                <>
                  <LabeledStatus
                    label="Status"
                    statusProps={{ status }}
                    tooltipProps={{
                      tipContent: statusDescription,
                      position: 'bottom',
                    }}
                  />
                  <LabeledValue label="Created">
                    <Time
                      time={run.created_at}
                      format="relative"
                      variant="subtext"
                    />
                  </LabeledValue>
                  {run.started_at ? (
                    <LabeledValue label="Started">
                      <Time
                        time={run.started_at}
                        format="relative"
                        variant="subtext"
                      />
                    </LabeledValue>
                  ) : null}
                  {run.finished_at ? (
                    <LabeledValue label="Finished">
                      <Time
                        time={run.finished_at}
                        format="relative"
                        variant="subtext"
                      />
                    </LabeledValue>
                  ) : null}
                </>
              }
            />
          }
          banners={
            <>
              {branchRun?.composite_error ? (
                <CompositeError error={branchRun.composite_error} />
              ) : null}
              <BranchRunApproval run={run} />
            </>
          }
        >
          <div className="flex flex-col gap-4">
            {showRunComparison && branchRun?.id ? (
              <BranchRunComparisonRuns
                orgId={orgId}
                appId={appId}
                branchId={branchId}
                baseRun={comparison?.base_run}
                headRun={comparison?.head_run}
                repoSlug={repoSlug}
                currentGithubHref={currentGithubHref}
              />
            ) : null}

            <BranchRunSummary
              branchRun={branchRun}
              appId={appId}
              branchId={branchId}
              branchRunId={branchRun?.id}
              runStatus={status}
            />

            {branchRun?.id && (
              <RuntimeChanges
                branchId={branchId}
                appBranchRunId={branchRun.id}
              />
            )}

            {branchRun?.id && (
              <BranchRunChanges
                branchId={branchId}
                appBranchRunId={branchRun.id}
                focus={configFocus}
                repoSlug={repoSlug}
                showRunComparison={false}
              />
            )}
          </div>
        </DetailPage>
      </>
    </ConfigDiffFocusContext.Provider>
  )
}

export const BranchRunDetail = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId}>
      <BranchRunDetailContent />
    </BranchProvider>
  )
}
