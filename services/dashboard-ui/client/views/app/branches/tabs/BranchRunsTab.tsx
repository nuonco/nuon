import { useParams, useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { WorkflowTimelineComponent } from '@/components/workflows/WorkflowTimeline'
import { getBranchWorkflowRuns } from '@/lib'
import { BranchDetail } from '../BranchDetail'
import { BranchTabPage } from './BranchTabPage'

const LIMIT = 20

const BranchRunsContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const [searchParams] = useSearchParams()
  const branchId = params.branchId as string
  const orgId = org.id!
  const appId = app.id!
  const offset = Number(searchParams.get('offset') ?? 0)
  const basePath = `/${orgId}/apps/${appId}/branches/${branchId}`

  const { data: runsResult, isLoading } = useQuery({
    queryKey: ['branch-runs', orgId, appId, branchId, offset],
    queryFn: () =>
      getBranchWorkflowRuns({ orgId, appId, branchId, limit: LIMIT, offset }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
    placeholderData: keepPreviousData,
  })

  const runs = runsResult?.data ?? []

  return (
    <BranchTabPage
      tab="Runs"
      heading="Workflow runs"
      subheading="Every deployment of this branch, newest first."
    >
      {isLoading ? (
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
          pagination={{
            hasNext: runsResult?.pagination?.hasNext ?? false,
            offset,
            limit: LIMIT,
          }}
          orgId={orgId}
          getWorkflowHref={(run) => `${basePath}/runs/${run.id}`}
        />
      )}
    </BranchTabPage>
  )
}

export const BranchRunsTab = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchRunsContent /> : <BranchDetail />
}
