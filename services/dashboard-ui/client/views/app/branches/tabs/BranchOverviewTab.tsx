import { useMemo } from 'react'
import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { PageSection } from '@/components/layout/PageSection'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { BranchSourceCard } from '@/components/branches/BranchSourceCard'
import { DeploymentPlanGraph } from '@/components/branches/DeploymentPlanGraph'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
import { WorkflowTimelineComponent } from '@/components/workflows/WorkflowTimeline'
import { getAppInstalls, getBranchWorkflowRuns } from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import type { TInstall } from '@/types'
import { BranchDetail } from '../BranchDetail'

const RECENT_RUNS_LIMIT = 5

const BranchOverviewContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch, refresh } = useBranch()
  const params = useParams()
  const branchId = params.branchId as string
  const orgId = org.id!
  const appId = app.id!
  const basePath = `/${orgId}/apps/${appId}/branches/${branchId}`

  const currentConfig = useMemo(() => latestBranchConfig(branch), [branch])
  const hasDeploymentPlan = (currentConfig?.install_groups?.length ?? 0) > 0

  const { data: appInstallsResult } = useQuery({
    queryKey: ['app-installs', orgId, appId],
    queryFn: () => getAppInstalls({ appId, orgId, limit: 100 }),
    enabled: !!orgId && !!appId && hasDeploymentPlan,
    refetchInterval: 10000,
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
    queryKey: ['branch-runs-recent', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId,
        branchId,
        limit: RECENT_RUNS_LIMIT,
        offset: 0,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
  })

  const runs = runsResult?.data ?? []

  return (
    <PageSection className="flex flex-col gap-8">
      <BranchSourceCard config={currentConfig} />

      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Text variant="base" weight="strong">
              Deployment plan
            </Text>
            {hasDeploymentPlan && (
              <Badge theme="info" size="sm">
                v{currentConfig?.config_number}
              </Badge>
            )}
          </div>
          {hasDeploymentPlan && <Link href={`${basePath}/plan`}>View plan</Link>}
        </div>
        {hasDeploymentPlan && currentConfig ? (
          <DeploymentPlanGraph
            config={currentConfig}
            installsById={installsById}
            orgId={orgId}
          />
        ) : (
          <div className="border rounded-lg p-6">
            <EmptyState
              variant="diagram"
              emptyTitle="No deployment plan yet"
              emptyMessage="Create a deployment plan to group installs and roll out branch changes in stages."
              action={
                <EditDeploymentPlanButton
                  branch={branch}
                  currentConfig={currentConfig}
                  variant="secondary"
                  label="Create deployment plan"
                  onSuccess={refresh}
                />
              }
            />
          </div>
        )}
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between gap-3">
          <Text variant="base" weight="strong">
            Recent runs
          </Text>
          {runs.length > 0 && <Link href={`${basePath}/runs`}>View all runs</Link>}
        </div>
        {isLoadingRuns ? (
          <TimelineSkeleton eventCount={3} />
        ) : runs.length === 0 ? (
          <Card>
            <EmptyState
              variant="history"
              emptyTitle="No workflow runs yet"
              emptyMessage="Runs will appear here once you trigger a deployment of this branch."
            />
          </Card>
        ) : (
          <WorkflowTimelineComponent
            workflows={runs}
            pagination={{ hasNext: false, offset: 0, limit: RECENT_RUNS_LIMIT }}
            orgId={orgId}
            getWorkflowHref={(run) => `${basePath}/runs/${run.id}`}
          />
        )}
      </div>
    </PageSection>
  )
}

export const BranchOverviewTab = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchOverviewContent /> : <BranchDetail />
}
