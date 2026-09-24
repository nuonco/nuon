import { useMemo } from 'react'
import { useParams, useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { WorkflowTimelineComponent } from '@/components/workflows/WorkflowTimeline'
import { WorkflowFilters } from '@/components/workflows/filters/WorkflowFilters'
import { getBranchWorkflowRuns } from '@/lib'
import {
  datePresetQueryParameter,
  readWorkflowFilters,
} from '@/utils/workflow-filters'
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
  const filters = readWorkflowFilters(searchParams, 'app')
  const since = searchParams.get('since')
  const createdAtGte = useMemo(() => datePresetQueryParameter(since), [since])
  const basePath = `/${orgId}/apps/${appId}/branches/${branchId}`

  const { data: runsResult, isLoading } = useQuery({
    queryKey: [
      'branch-runs',
      orgId,
      appId,
      branchId,
      offset,
      filters.api.search,
      filters.api.status,
      filters.api.type,
      filters.api.preview,
      createdAtGte,
    ],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId,
        branchId,
        limit: LIMIT,
        offset,
        planonly: true,
        q: filters.api.search,
        status: filters.api.status,
        type: filters.api.type,
        preview: filters.api.preview,
        created_at_gte: createdAtGte,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
    placeholderData: keepPreviousData,
  })

  const runs = runsResult?.data ?? []

  return (
    <BranchTabPage
      tab="Runs"
      tabPath="runs"
      heading="Runs"
      subheading="Every run from this branch, newest first."
    >
      <WorkflowFilters owner="app" />
      <WorkflowTimelineComponent
        workflows={runs}
        pagination={{
          hasNext: runsResult?.pagination?.hasNext ?? false,
          offset,
          limit: LIMIT,
        }}
        orgId={orgId}
        getWorkflowHref={(run) => `${basePath}/runs/${run.id}`}
        isFiltered={filters.filtered}
        isLoading={isLoading}
      />
    </BranchTabPage>
  )
}

export const BranchRunsTab = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchRunsContent /> : <BranchDetail />
}
