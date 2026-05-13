import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useBranch } from '@/hooks/use-branch'
import { BranchProvider } from '@/providers/branch-provider'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { getWorkflowBadge } from '@/utils/workflow-utils'

import { BranchDetailActions } from '@/components/branches/BranchDetailActions'
import { InstallGroupsSection } from '@/components/branches/install-groups/InstallGroupsSection'
import { getBranchWorkflowRuns, getAppInstalls } from '@/lib'
import type { TInstall } from '@/types'

const BranchDetailContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string

  const { data: appInstallsResult } = useQuery({
    queryKey: ['app-installs', orgId, appId],
    queryFn: () => getAppInstalls({ appId, orgId, limit: 100 }),
    enabled: !!orgId && !!appId,
  })

  const installsById = (appInstallsResult?.data ?? []).reduce<Record<string, TInstall>>(
    (acc, install) => {
      acc[install.id] = install
      return acc
    },
    {}
  )

  const { data: runs = [] } = useQuery({
    queryKey: ['branch-runs', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId,
        branchId,
        limit: 5,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
  })

  const currentConfig =
    branch.configs && branch.configs.length > 0
      ? branch.configs.sort(
          (a, b) => (b.config_number || 0) - (a.config_number || 0)
        )[0]
      : undefined

  return (
    <PageSection>
      <PageTitle title={`${branch?.name ?? 'Branch'} | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}`, text: branch?.name },
        ]}
      />
      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="strong">
            {branch.name}
          </Text>
          <ID>{branch.id}</ID>
          <Text variant="subtext" theme="neutral">
            Created <Time time={branch?.created_at} format="relative" />
          </Text>
        </HeadingGroup>

        <div className="flex items-center gap-4">
          <BranchDetailActions
            branch={branch}
            currentConfig={currentConfig}
            appId={appId}
            orgId={orgId}
          />
        </div>
      </div>

      <Card className="mb-6">
        <div>
          <div className="flex items-center justify-between mb-4">
            <Text variant="h3" weight="strong">
              Install groups
            </Text>
            {currentConfig && (
              <Badge theme="info" size="sm">
                v{currentConfig.config_number}
              </Badge>
            )}
          </div>

          {!currentConfig ? (
            <div className="text-center py-8">
              <Text variant="body" theme="neutral">
                No configuration yet. Click "Edit" above to set up install groups.
              </Text>
            </div>
          ) : (
            <InstallGroupsSection
              config={currentConfig}
              installsById={installsById}
              orgId={orgId}
            />
          )}
        </div>
      </Card>

      <div>
        <div className="flex items-center justify-between mb-4">
          <Text variant="h3" weight="strong">
            Workflow runs
          </Text>
          <Link href={`/${orgId}/apps/${appId}/branches/${branchId}/runs`}>
            View all <Icon variant="CaretRightIcon" />
          </Link>
        </div>

        {runs.length === 0 ? (
          <div className="text-center py-8 border border-dashed border-cool-grey-300 dark:border-dark-grey-600 rounded-md">
            <Text variant="body" theme="neutral">
              No workflow runs yet. Click "Trigger Run" above to start a deployment.
            </Text>
          </div>
        ) : (
          <Timeline
            events={runs}
            pagination={{ hasNext: false, offset: 0, limit: 5 }}
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
