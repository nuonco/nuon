import { useParams, useSearchParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { BackToTop } from '@/components/common/BackToTop'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { RunnerDetailsCard } from '@/components/runners/RunnerDetailsCard'
import { RunnerHealthCard } from '@/components/runners/RunnerHealthCard'
import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity'
import { ManagementDropdown } from '@/components/runners/management/ManagementDropdown'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { RunnerProvider } from '@/providers/runner-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type {
  TRunner,
  TRunnerSettings,
  TRunnerHeartbeat,
  TRunnerHealthcheck,
  TRunnerJob,
  TRunnerGroup,
} from '@/types'

const RunnerDetailsError = () => (
  <Card className="flex-auto">
    <EmptyState
      emptyMessage="Runner details will display here once available."
      emptyTitle="No runner details"
      variant="table"
    />
  </Card>
)

const RunnerHealthError = () => (
  <Card className="flex-auto">
    <EmptyState
      emptyMessage="Runner health checks will display here once available."
      emptyTitle="No health check data"
      variant="diagram"
    />
  </Card>
)

const RunnerActivityError = () => (
  <div className="w-full">
    <Text>Error fetching recent runner activity</Text>
  </div>
)

const RunnerDetails = ({
  orgId,
  runnerId,
  settings,
}: {
  orgId: string
  runnerId: string
  settings: TRunnerSettings
}) => {
  const {
    data: runnerHeartbeat,
    error,
    isLoading,
  } = usePolling<TRunnerHeartbeat>({
    path: `/api/ctl-api/v1/runners/${runnerId}/heart-beats/latest`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  if (error) {
    return <RunnerDetailsError />
  }

  if (isLoading && !runnerHeartbeat) {
    return <div>Loading runner details...</div>
  }

  return (
    <RunnerDetailsCard
      className="md:flex-initial"
      initHeartbeat={runnerHeartbeat}
      runnerGroup={settings as TRunnerGroup}
      shouldPoll
    />
  )
}

const RunnerHealth = ({
  orgId,
  runnerId,
}: {
  orgId: string
  runnerId: string
}) => {
  const {
    data: healthchecks,
    error,
    isLoading,
  } = usePolling<TRunnerHealthcheck[]>({
    path: `/api/ctl-api/v1/runners/${runnerId}/recent-health-checks`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  if (error) {
    return <RunnerHealthError />
  }

  if (isLoading && !healthchecks) {
    return <div>Loading health checks...</div>
  }

  return (
    <RunnerHealthCard
      className="flex-auto"
      initHealthchecks={healthchecks}
      shouldPoll
    />
  )
}

const RunnerActivity = ({
  orgId,
  runnerId,
  offset,
}: {
  orgId: string
  runnerId: string
  offset: string
}) => {
  const {
    data: jobs,
    error,
    isLoading,
    headers,
  } = usePolling<TRunnerJob[]>({
    path: `/api/ctl-api/v1/runners/${runnerId}/jobs?offset=${offset}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error) {
    return <RunnerActivityError />
  }

  if (isLoading && !jobs) {
    return <div>Loading runner activity...</div>
  }

  return (
    <>
      <Text variant="base" weight="strong">
        Recent activity
      </Text>
      <RunnerRecentActivity
        initJobs={jobs}
        pagination={pagination}
        shouldPoll
      />
    </>
  )
}

export default function InstallRunner() {
  const { installId, orgId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const offset = searchParams.get('offset') || '0'

  const {
    data: runner,
    error: runnerError,
    isLoading: runnerLoading,
  } = usePolling<TRunner>({
    path: `/api/ctl-api/v1/runners/${install?.runner_id}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const {
    data: settings,
    error: settingsError,
    isLoading: settingsLoading,
  } = useQuery<TRunnerSettings>({
    path: `/api/ctl-api/v1/runners/${install?.runner_id}/settings`,
    enabled: !!install?.runner_id,
  })

  const {
    data: heartbeat,
    error: heartbeatError,
    isLoading: heartbeatLoading,
  } = useQuery<TRunnerHeartbeat>({
    path: `/api/ctl-api/v1/runners/${install?.runner_id}/heart-beats/latest`,
    enabled: !!install?.runner_id,
  })

  if (!installId || !orgId || !install?.runner_id) {
    return null
  }

  if (runnerError) {
    return <div>Runner not found</div>
  }

  if (runnerLoading && !runner) {
    return <div>Loading runner...</div>
  }

  if (!runner) {
    return <div>Loading runner...</div>
  }

  return (
    <RunnerProvider initRunner={runner} shouldPoll>
      <SurfacesProvider>
        <PageSection className="@container" isScrollable>
          <Breadcrumbs
            breadcrumbs={[
              {
                path: `/${orgId}`,
                text: org?.name,
              },
              {
                path: `/${orgId}/installs`,
                text: 'Installs',
              },
              {
                path: `/${orgId}/installs/${installId}`,
                text: install?.name,
              },
              {
                path: `/${orgId}/installs/${installId}/runner`,
                text: 'Runner',
              },
            ]}
          />
          <div className="flex gap-4 justify-between">
            <hgroup>
              <Text variant="base" weight="strong">
                Install runner
              </Text>
            </hgroup>
            {settings && (
              <ManagementDropdown
                settings={settings}
                isInstallRunner
                isManagedRunner={Boolean(heartbeat?.mng)}
              />
            )}
          </div>

          <div className="flex flex-col @min-4xl:flex-row gap-6">
            {settings && (
              <RunnerDetails
                orgId={orgId}
                runnerId={install.runner_id}
                settings={settings}
              />
            )}

            <RunnerHealth orgId={orgId} runnerId={install.runner_id} />
          </div>

          <div className="flex flex-col gap-6">
            <RunnerActivity
              orgId={orgId}
              offset={offset}
              runnerId={install.runner_id}
            />
          </div>

          <BackToTop />
        </PageSection>
      </SurfacesProvider>
    </RunnerProvider>
  )
}