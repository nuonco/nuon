import { useMemo } from 'react'
import { useParams, useSearchParams } from 'react-router'
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toast } from '@/components/surfaces/Toast'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useBranch } from '@/hooks/use-branch'
import { useToast } from '@/hooks/use-toast'
import { BranchProvider } from '@/providers/branch-provider'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { getWorkflowBadge } from '@/utils/workflow-utils'

import { BranchDetailActions } from '@/components/branches/BranchDetailActions'
import { InstallGroupsSection } from '@/components/branches/install-groups/InstallGroupsSection'
import { getBranchWorkflowRuns, getAppInstalls, cancelWorkflow } from '@/lib'
import type { TAPIError, TInstall } from '@/types'

const CANCELLABLE = new Set(['pending', 'queued', 'in-progress'])
const LIMIT = 20

const BranchDetailContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const params = useParams()
  const [searchParams] = useSearchParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string
  const offset = Number(searchParams.get('offset') ?? 0)

  const cancelMutation = useMutation({
    mutationFn: (workflowId: string) => cancelWorkflow({ orgId, workflowId }),
    onSuccess: () => {
      addToast(
        <Toast heading="Workflow cancelled" theme="success">
          <Text>The workflow run has been cancelled.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['branch-runs'] })
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Cancel failed" theme="error">
          <Text>{err?.error || 'Unable to cancel workflow.'}</Text>
        </Toast>
      )
    },
  })

  const currentConfig = useMemo(() => {
    if (!branch.configs?.length) return undefined
    return [...branch.configs].sort(
      (a, b) => (b.config_number || 0) - (a.config_number || 0)
    )[0]
  }, [branch.configs])

  const { data: appInstallsResult } = useQuery({
    queryKey: ['app-installs', orgId, appId],
    queryFn: () => getAppInstalls({ appId, orgId, limit: 100 }),
    enabled: !!orgId && !!appId && !!currentConfig,
  })

  const installsById = useMemo(
    () =>
      (appInstallsResult?.data ?? []).reduce<Record<string, TInstall>>(
        (acc, install) => {
          acc[install.id] = install
          return acc
        },
        {}
      ),
    [appInstallsResult]
  )

  const { data: runsResult, isLoading: isLoadingRuns } = useQuery({
    queryKey: ['branch-runs', orgId, appId, branchId, offset],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId,
        branchId,
        limit: LIMIT,
        offset,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
    placeholderData: keepPreviousData,
  })

  const runs = runsResult?.data ?? []

  return (
    <PageSection>
      <PageTitle title={`${branch.name} | ${app.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/apps`, text: 'Apps' },
          { path: `/${org.id}/apps/${app.id}`, text: app.name },
          { path: `/${org.id}/apps/${app.id}/branches`, text: 'Branches' },
          { path: `/${org.id}/apps/${app.id}/branches/${branchId}`, text: branch.name },
        ]}
      />
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <HeadingGroup className="gap-1.5">
          <Text variant="h3" weight="stronger" level={1}>
            {branch.name}
          </Text>
          <Text variant="subtext" theme="info">
            Last updated{' '}
            <Time
              variant="subtext"
              time={branch.updated_at}
              format="relative"
            />
          </Text>
        </HeadingGroup>
        <BranchDetailActions
          branch={branch}
          currentConfig={currentConfig}
          appId={appId}
          orgId={orgId}
        />
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <Text variant="base" weight="strong">
            Install groups
          </Text>
          {currentConfig && (
            <Badge theme="info" size="sm">
              v{currentConfig.config_number}
            </Badge>
          )}
        </div>

        {!currentConfig ? (
          <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg p-6">
            <EmptyState
              variant="diagram"
              emptyTitle="No install groups yet"
              emptyMessage={`Use "Deployment plan" above to group installs for staged deployment.`}
            />
          </div>
        ) : (
          <InstallGroupsSection
            config={currentConfig}
            installsById={installsById}
            orgId={orgId}
          />
        )}
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <Text variant="base" weight="strong">
            Workflow runs
          </Text>
          {runs.length > 0 && (
            <Link href={`/${orgId}/apps/${appId}/branches/${branchId}/runs`}>
              View all <Icon variant="CaretRightIcon" />
            </Link>
          )}
        </div>

        {isLoadingRuns ? (
          <TimelineSkeleton eventCount={3} />
        ) : runs.length === 0 ? (
          <Card>
            <EmptyState
              variant="history"
              emptyTitle="No workflow runs yet"
              emptyMessage={`Use "Trigger run" above to start a deployment.`}
            />
          </Card>
        ) : (
          <Timeline
            events={runs}
            pagination={{ hasNext: runsResult?.pagination?.hasNext ?? false, offset, limit: LIMIT }}
            renderEvent={(run) => {
              const commitSha = run.app_branch_runs?.[0]?.commit_sha
              return (
                <TimelineEvent
                  key={run.id}
                  status={run.status?.status}
                  createdAt={run.created_at}
                  createdBy={run.created_by?.email}
                  badge={getWorkflowBadge(run)}
                  title={
                    <Link href={`/${orgId}/apps/${appId}/branches/${branchId}/runs/${run.id}`}>
                      {run.name || toSentenceCase(snakeToWords(run.type)) || `Run #${run.id?.substring(0, 8)}`}
                    </Link>
                  }
                  caption={<ID>{run.id}</ID>}
                  underline={commitSha ? (
                    <span className="flex items-center gap-1.5">
                      <Icon variant="GitCommitIcon" size={12} />
                      {commitSha.substring(0, 7)}
                    </span>
                  ) : undefined}
                  actions={CANCELLABLE.has(run.status?.status) ? (
                    <Button
                      variant="danger"
                      size="xs"
                      onClick={() => cancelMutation.mutate(run.id)}
                      disabled={cancelMutation.isPending}
                    >
                      <Icon variant="XCircleIcon" size={14} />
                      Cancel
                    </Button>
                  ) : undefined}
                />
              )
            }}
          />
        )}
      </div>
    </PageSection>
  )
}

export const BranchDetail = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId} shouldPoll>
      <BranchDetailContent />
    </BranchProvider>
  )
}
