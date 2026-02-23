import { useParams, useSearchParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { BackToTop } from '@/components/common/BackToTop'
import { RunnerDetailsCard } from '@/components/runners/RunnerDetailsCard'
import { RunnerHealthCard, RunnerHealthEmptyCard } from '@/components/runners/RunnerHealthCard'
import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import { RunnerProvider } from '@/providers/runner-provider'
import type { TRunner, TRunnerMngHeartbeat, TRunnerHealthCheck, TRunnerJob } from '@/types'

export default function OrgRunner() {
  const { orgId } = useParams()
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const offset = searchParams.get('offset') || '0'

  const runnerGroup = org?.runner_group
  const runner = runnerGroup?.runners?.at(0)
  const runnerId = runner?.id

  // Fetch runner heartbeat
  const {
    data: runnerHeartbeat,
    isLoading: heartbeatLoading,
  } = usePolling<TRunnerMngHeartbeat>({
    path: `/api/ctl-api/v1/runners/${runnerId}/heart-beats/latest`,
    pollInterval: 5000,
    shouldPoll: !!runnerId,
  })

  // Fetch runner health checks
  const {
    data: healthchecks,
    isLoading: healthLoading,
  } = usePolling<TRunnerHealthCheck[]>({
    path: `/api/ctl-api/v1/runners/${runnerId}/recent-health-checks`,
    pollInterval: 60000,
    shouldPoll: !!runnerId,
  })

  // Fetch runner jobs
  const {
    data: jobs,
    isLoading: jobsLoading,
    headers: jobsHeaders,
  } = usePolling<TRunnerJob[]>({
    path: `/api/ctl-api/v1/runners/${runnerId}/jobs?groups=actions,build,deploy,operations,sandbox,sync&limit=10&offset=${offset}`,
    pollInterval: 20000,
    shouldPoll: !!runnerId,
  })

  const pagination = {
    hasNext: jobsHeaders?.['x-nuon-page-next'] === 'true',
    offset: Number(jobsHeaders?.['x-nuon-page-offset'] ?? '0'),
  }

  if (!org?.features?.['org-runner']) {
    return (
      <PageSection isScrollable>
        <Text theme="neutral">Build runner is not available for this organization.</Text>
      </PageSection>
    )
  }

  if (!runnerGroup || !runner) {
    return (
      <PageSection isScrollable>
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${orgId}`, text: org?.name || '' },
            { path: `/${orgId}/runner`, text: 'Builds' },
          ]}
        />
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Builds
          </Text>
          <Text theme="neutral">
            View your organization's build runner performance and activities.
          </Text>
        </HeadingGroup>
        <EmptyState
          variant="table"
          title="No runner available"
          emptyMessage="No build runner has been configured for this organization."
        />
        <BackToTop />
      </PageSection>
    )
  }

  return (
    <RunnerProvider initRunner={runner} shouldPoll>
      <PageSection isScrollable>
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${orgId}`, text: org?.name || '' },
            { path: `/${orgId}/runner`, text: 'Builds' },
          ]}
        />
        
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Builds
          </Text>
          <Text theme="neutral">
            View your organization's build runner performance and activities.
          </Text>
        </HeadingGroup>

        {/* Runner Details and Health Cards */}
        <div className="flex flex-col md:flex-row gap-6 mt-6">
          {heartbeatLoading ? (
            <Card className="flex-initial">
              <Loading variant="stack" loadingText="Loading runner details..." />
            </Card>
          ) : runnerHeartbeat ? (
            <RunnerDetailsCard
              className="flex-initial"
              initHeartbeat={runnerHeartbeat}
              runnerGroup={runnerGroup}
              shouldPoll
            />
          ) : (
            <Card className="flex-initial">
              <EmptyState
                emptyMessage="Runner details will display here once available."
                emptyTitle="No runner details"
                variant="table"
              />
            </Card>
          )}

          {healthLoading ? (
            <Card className="flex-auto">
              <Loading variant="stack" loadingText="Loading health status..." />
            </Card>
          ) : healthchecks && healthchecks.length > 0 ? (
            <RunnerHealthCard
              className="flex-auto"
              initHealthchecks={healthchecks}
              shouldPoll
            />
          ) : (
            <RunnerHealthEmptyCard />
          )}
        </div>

        {/* Recent Activity */}
        <div className="flex flex-col gap-4 mt-6">
          <Text variant="base" weight="strong">
            Recent activity
          </Text>
          {jobsLoading ? (
            <Loading variant="stack" loadingText="Loading recent jobs..." />
          ) : jobs && jobs.length > 0 ? (
            <RunnerRecentActivity
              initJobs={jobs}
              pagination={pagination}
              shouldPoll
            />
          ) : (
            <EmptyState
              variant="table"
              title="No recent activity"
              emptyMessage="Runner job activity will display here once jobs are executed."
            />
          )}
        </div>

        <BackToTop />
      </PageSection>
    </RunnerProvider>
  )
}
