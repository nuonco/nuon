import { useMemo } from 'react'
import { useParams, useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useBranch } from '@/hooks/use-branch'
import { BranchProvider } from '@/providers/branch-provider'

import { BranchDetailActions } from '@/components/branches/BranchDetailActions'
import { InstallGroupsSection } from '@/components/branches/install-groups/InstallGroupsSection'
import { WorkflowTimelineComponent } from '@/components/workflows/WorkflowTimeline'
import { getBranchWorkflowRuns, getAppInstalls } from '@/lib'
import type { TInstall } from '@/types'

const LIMIT = 20

const BranchDetailContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const [searchParams] = useSearchParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string
  const offset = Number(searchParams.get('offset') ?? 0)

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
          <BackLink className="mb-4" />
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
        <Text variant="base" weight="strong">
          Workflow runs
        </Text>

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
          <WorkflowTimelineComponent
            workflows={runs}
            pagination={{
              hasNext: runsResult?.pagination?.hasNext ?? false,
              offset,
              limit: LIMIT,
            }}
            orgId={orgId}
            getWorkflowHref={(run) =>
              `/${orgId}/apps/${appId}/branches/${branchId}/runs/${run.id}`
            }
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
