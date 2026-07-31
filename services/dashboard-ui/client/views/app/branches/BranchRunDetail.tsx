import { useCallback, useState } from 'react'
import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { BackLink } from '@/components/common/BackLink'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { AppConfigDiff } from '@/components/branches/AppConfigDiff'
import { BranchRunSummary } from '@/components/branches/BranchRunSummary'
import { RuntimeChanges } from '@/components/branches/RuntimeChanges'
import { WorkflowRunPanelButton } from '@/components/branches/WorkflowRunPanel'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { BranchProvider } from '@/providers/branch-provider'
import {
  ConfigDiffFocusContext,
  type TConfigDiffFocus,
} from '@/components/approvals/plan-diffs/config-diff-focus'
import { getBranchWorkflowRun } from '@/lib'
import { getRunTitle } from '@/components/branches/shared/run-title'

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
  const requestConfigFocus = useCallback((sectionKey: string, entityName?: string) => {
    setConfigFocus((prev) => ({ sectionKey, entityName, nonce: (prev?.nonce ?? 0) + 1 }))
  }, [])

  const { data: run, isLoading } = useQuery({
    queryKey: ['branch-run', orgId, appId, branchId, runId],
    queryFn: () => getBranchWorkflowRun({ orgId, appId, branchId, runId }),
    enabled: !!orgId && !!appId && !!branchId && !!runId,
    refetchInterval: 5000,
  })

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

  const branchRun = run.app_branch_runs?.at(0)
  const runTitle = getRunTitle(run)

  const configStep = run.steps?.find(
    (s) => s.name?.toLowerCase().includes('config') && !s.name?.toLowerCase().includes('diff')
  )
  const appConfigId =
    branchRun?.app_config_id || (configStep?.status?.metadata?.app_config_id as string | undefined)

  return (
    <ConfigDiffFocusContext.Provider value={{ requestFocus: requestConfigFocus }}>
      <PageSection className="max-w-full">
        <PageTitle title={`${runTitle} | ${app?.name}`} />
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${org?.id}`, text: org?.name },
            { path: `/${org?.id}/apps`, text: 'Apps' },
            { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
            { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
            { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}`, text: branch?.name },
            {
              path: `/${org?.id}/apps/${app?.id}/branches/${branchId}/runs/${runId}`,
              text: 'Run',
            },
          ]}
        />

        <BackLink />

        <HeadingGroup className="gap-1.5 min-w-0">
          <div className="flex items-center gap-2.5">
            <Text
              as="h1"
              variant="h2"
              weight="strong"
              className="leading-tight min-w-0 truncate"
              title={runTitle}
            >
              {runTitle}
            </Text>
            {branch?.name && (
              <Badge size="sm" variant="code" className="shrink-0">
                {branch.name}
              </Badge>
            )}
          </div>

          <ID className="text-[12px] font-mono text-cool-grey-400 dark:text-cool-grey-500">
            {runId}
          </ID>

          <div className="flex items-center gap-2 mt-0.5">
            <Status status={status} variant="badge" />
            {statusDescription && (
              <Text variant="subtext" theme="neutral">
                {statusDescription}
              </Text>
            )}
          </div>
        </HeadingGroup>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {(branchRun?.vcs_connection_commit || branchRun?.pr_number) && (() => {
            const vcsConfig = branchRun?.app_branch_config?.connected_github_vcs_config
            const repoSlug = vcsConfig?.repo
            const prUrl = repoSlug && branchRun?.pr_number
              ? `https://github.com/${repoSlug}/pull/${branchRun.pr_number}`
              : undefined
            const commitUrl = repoSlug && branchRun?.vcs_connection_commit?.sha
              ? `https://github.com/${repoSlug}/commit/${branchRun.vcs_connection_commit.sha}`
              : undefined
            const githubUrl = prUrl || commitUrl

            return (
              <Card className="!p-4 !gap-3">
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <Icon variant="GitBranchIcon" size={16} className="text-cool-grey-400" />
                    <Text variant="base" weight="strong">Source</Text>
                  </div>
                  {githubUrl && (
                    <Link href={githubUrl} isExternal className="text-xs">
                      <Icon variant="GithubLogoIcon" size={14} />
                      View in GitHub
                      <Icon variant="ArrowSquareOutIcon" size={12} />
                    </Link>
                  )}
                </div>

                {branchRun?.pr_number && (
                  <div className="flex items-center gap-2 flex-wrap">
                    <Badge size="sm" theme="info">
                      PR #{branchRun.pr_number}
                    </Badge>
                    {branchRun?.base_branch && (
                      <Text variant="subtext" theme="neutral">
                        into {branchRun.base_branch}
                      </Text>
                    )}
                    {branchRun?.event_type && (
                      <Badge size="sm" theme="neutral">
                        {branchRun.event_type.replace(/_/g, ' ')}
                      </Badge>
                    )}
                  </div>
                )}

                {branchRun?.vcs_connection_commit && (
                  <div className="flex items-start gap-3">
                    <Icon variant="GitCommitIcon" size={16} className="mt-0.5 shrink-0 text-cool-grey-400" />
                    <div className="flex flex-col gap-1 min-w-0">
                      <Text variant="body" weight="strong" className="truncate">
                        {branchRun.vcs_connection_commit.message?.split('\n')[0]?.trim()}
                      </Text>
                      <div className="flex items-center gap-2 flex-wrap">
                        {branchRun.vcs_connection_commit.sha && (
                          <Badge size="sm" variant="code">
                            {branchRun.vcs_connection_commit.sha.slice(0, 8)}
                          </Badge>
                        )}
                        {branchRun.vcs_connection_commit.author_name && (
                          <Text variant="subtext" theme="neutral">
                            {branchRun.vcs_connection_commit.author_name}
                          </Text>
                        )}
                      </div>
                    </div>
                  </div>
                )}

                {repoSlug && (
                  <div className="flex items-center gap-1.5">
                    <Icon variant="GithubLogoIcon" size={14} className="text-cool-grey-400" />
                    <Text variant="subtext" theme="neutral">{repoSlug}</Text>
                  </div>
                )}
              </Card>
            )
          })()}

          <Card className="!p-4 !gap-3">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <Icon variant="FlowArrowIcon" size={16} className="text-cool-grey-400" />
                <Text variant="base" weight="strong">Workflow details</Text>
                <Status status={status} variant="badge" />
              </div>
              <div className="flex items-center gap-2">
                <AdminDashboardLink path={`/workflows/${runId}`} label="admin" />
                <CancelWorkflowButton workflow={run} />
              </div>
            </div>

            <div className="flex items-center gap-4 flex-wrap">
              <div className="flex items-center gap-1.5">
                <Text variant="subtext" theme="neutral">Created</Text>
                <Time time={run.created_at} format="relative" variant="subtext" />
              </div>
              {run.started_at && (
                <div className="flex items-center gap-1.5">
                  <Text variant="subtext" theme="neutral">Started</Text>
                  <Time time={run.started_at} format="relative" variant="subtext" />
                </div>
              )}
              {run.finished_at && (
                <div className="flex items-center gap-1.5">
                  <Text variant="subtext" theme="neutral">Finished</Text>
                  <Time time={run.finished_at} format="relative" variant="subtext" />
                </div>
              )}
            </div>

            <div className="flex justify-end">
              <WorkflowRunPanelButton runId={runId} />
            </div>
          </Card>
        </div>

        <BranchRunSummary
          branchRun={branchRun}
          appId={appId}
          branchId={branchId}
          branchRunId={branchRun?.id}
          runStatus={status}
        />

        {appConfigId && <AppConfigDiff appConfigId={appConfigId} focus={configFocus} />}

        {branchRun?.id && <RuntimeChanges branchId={branchId} appBranchRunId={branchRun.id} />}
      </PageSection>
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
